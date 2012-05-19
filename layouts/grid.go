/*
   Copyright 2012 the go.uik authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package layouts

import (
	"code.google.com/p/draw2d/draw2d"
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.uik"
	"image/color"
	"image/draw"
	"math"
)

func VBox(config GridConfig, blocks ...*uik.Block) (g *Grid) {
	g = NewGrid(config)
	for i, b := range blocks {
		g.Add(BlockData{
			Block: b,
			GridX: 0, GridY: i,
			AnchorX: AnchorMin,
		})
	}
	return
}

func HBox(config GridConfig, blocks ...*uik.Block) (g *Grid) {
	g = NewGrid(config)
	for i, b := range blocks {
		g.Add(BlockData{
			Block: b,
			GridX: i, GridY: 0,
			AnchorY: AnchorMin,
		})
	}
	return
}

type Anchor uint8

const (
	AnchorMin Anchor = 1 << iota
	AnchorMax
)

type BlockData struct {
	Block *uik.Block
	// The coordinates for the top-left of the block's placement
	GridX, GridY int
	// How many extra columns and rows the block occupies
	ExtraX, ExtraY int
	// AnchorX and AnchorY get the bit flags from AnchorMin and AnchorMax. The
	// zero value means they will float in the center.
	AnchorX, AnchorY Anchor
	// The zero-values for MinSize, PreferredSize and MaxSize tell the grid to ignore them
	MinSize, PreferredSize, MaxSize geom.Coord
}

type GridConfig struct {
}

type Grid struct {
	uik.Foundation

	children          map[*uik.Block]bool
	childrenBlockData map[*uik.Block]BlockData
	config            GridConfig

	vflex, hflex   *flex
	velems, helems map[*uik.Block]*elem

	add       chan BlockData
	remove    chan *uik.Block
	setConfig chan GridConfig
	getConfig chan GridConfig
}

func NewGrid(cfg GridConfig) (g *Grid) {
	g = new(Grid)

	g.config = cfg

	g.Initialize()
	if uik.ReportIDs {
		uik.Report(g.ID, "grid")
	}

	go g.handleEvents()

	return
}

func (g *Grid) Initialize() {
	g.Foundation.Initialize()
	g.DrawOp = draw.Over

	g.children = map[*uik.Block]bool{}
	g.childrenBlockData = map[*uik.Block]BlockData{}

	g.add = make(chan BlockData, 1)
	g.remove = make(chan *uik.Block, 1)
	g.setConfig = make(chan GridConfig, 1)
	g.getConfig = make(chan GridConfig, 1)

	g.hflex = &flex{}
	g.vflex = &flex{}

	g.helems = map[*uik.Block]*elem{}
	g.velems = map[*uik.Block]*elem{}

	// g.Paint = func(gc draw2d.GraphicContext) {
	// 	g.draw(gc)
	// }
}

func (g *Grid) Add(bd BlockData) {
	g.add <- bd
}

func (g *Grid) Remove(b *uik.Block) {
	g.remove <- b
}

func (g *Grid) SetConfig(cfg GridConfig) {
	g.setConfig <- cfg
}

func (g *Grid) GetConfig() (cfg GridConfig) {
	cfg = <-g.getConfig
	return
}

func (g *Grid) addBlock(bd BlockData) {
	g.AddBlock(bd.Block)
	g.children[bd.Block] = true
	g.childrenBlockData[bd.Block] = bd

	helem := &elem{
		index: bd.GridX,
		extra: bd.ExtraX,
	}
	g.helems[bd.Block] = helem
	velem := &elem{
		index: bd.GridY,
		extra: bd.ExtraY,
	}
	g.velems[bd.Block] = velem

	g.hflex.add(helem)
	g.vflex.add(velem)
}

func (g *Grid) remBlock(b *uik.Block) {
	if !g.children[b] {
		return
	}
	g.RemoveBlock(b)

	delete(g.ChildrenHints, b)
	delete(g.childrenBlockData, b)

	g.hflex.rem(g.helems[b])
	g.vflex.rem(g.velems[b])

	g.regrid()
}

func (g *Grid) reflex(b *uik.Block) {
	bd := g.childrenBlockData[b]
	sh := g.ChildrenHints[b]

	helem := g.helems[bd.Block]
	helem.minSize = sh.MinSize.X
	helem.prefSize = sh.PreferredSize.X
	helem.maxSize = math.Inf(1) //sh.MaxSize.X
	if bd.MinSize.X != 0 {
		helem.minSize = math.Max(bd.MinSize.X, helem.minSize)
	}
	if bd.PreferredSize.X != 0 {
		helem.prefSize = bd.PreferredSize.X
	}
	if bd.MaxSize.X != 0 {
		helem.maxSize = math.Min(bd.MaxSize.X, helem.maxSize)
	}
	helem.prefSize = math.Min(helem.maxSize, math.Max(helem.minSize, helem.prefSize))
	helem.fix()

	velem := g.velems[bd.Block]
	velem.minSize = sh.MinSize.Y
	velem.prefSize = sh.PreferredSize.Y
	velem.maxSize = math.Inf(1) //sh.MaxSize.Y
	if bd.MinSize.Y != 0 {
		velem.minSize = math.Max(bd.MinSize.Y, velem.minSize)
	}
	if bd.PreferredSize.Y != 0 {
		velem.prefSize = bd.PreferredSize.Y
	}
	if bd.MaxSize.Y != 0 {
		velem.maxSize = math.Min(bd.MaxSize.Y, velem.maxSize)
	}
	velem.prefSize = math.Min(velem.maxSize, math.Max(velem.minSize, velem.prefSize))
	velem.fix()
}

func (g *Grid) makePreferences() {
	var sizeHint uik.SizeHint
	hmin, hpref, hmax := g.hflex.makePrefs()
	vmin, vpref, vmax := g.vflex.makePrefs()
	sizeHint.MinSize = geom.Coord{hmin, vmin}
	sizeHint.PreferredSize = geom.Coord{hpref, vpref}
	sizeHint.MaxSize = geom.Coord{hmax, vmax}
	// uik.Report("prefs", g.Block.ID, sizeHint)
	g.SetSizeHint(sizeHint)
}

func (g *Grid) regrid() {

	_, minXs, maxXs := g.hflex.constrain(g.Size.X)
	_, minYs, maxYs := g.vflex.constrain(g.Size.Y)

	// if g.Block.ID == 2 {
	// 	uik.Report("regrid", g.Block.ID, g.Size, whs, wvs)
	// }
	for child, csh := range g.ChildrenHints {
		bd := g.childrenBlockData[child]
		gridBounds := geom.Rect{
			Min: geom.Coord{
				X: minXs[bd.GridX],
				Y: minYs[bd.GridY],
			},
			Max: geom.Coord{
				X: maxXs[bd.GridX+bd.ExtraX],
				Y: maxYs[bd.GridY+bd.ExtraY],
			},
		}

		gridSizeX, gridSizeY := gridBounds.Size()
		if gridSizeX > csh.MaxSize.X {
			diff := gridSizeX - csh.MaxSize.X
			switch bd.AnchorX {
			case 0:
				gridBounds.Min.X += diff / 2
				gridBounds.Max.X -= diff / 2
			case AnchorMin:
				gridBounds.Max.X -= diff
			case AnchorMax:
				gridBounds.Min.X += diff
			case AnchorMin | AnchorMax:
			}
		}
		if gridSizeY > csh.MaxSize.Y {
			diff := gridSizeY - csh.MaxSize.Y
			switch bd.AnchorY {
			case 0:
				gridBounds.Min.Y += diff / 2
				gridBounds.Max.Y -= diff / 2
			case AnchorMin:
				gridBounds.Max.Y -= diff
			case AnchorMax:
				gridBounds.Min.Y += diff
			case AnchorMin | AnchorMax:
			}
		}

		gridSizeX, gridSizeY = gridBounds.Size()
		if gridSizeX > csh.PreferredSize.X {
			diff := gridSizeX - csh.PreferredSize.X
			switch bd.AnchorX {
			case 0:
				gridBounds.Min.X += diff / 2
				gridBounds.Max.X -= diff / 2
			case AnchorMin:
				gridBounds.Max.X -= diff
			case AnchorMax:
				gridBounds.Min.X += diff
			case AnchorMin | AnchorMax:
			}
		}
		if gridSizeY > csh.PreferredSize.Y {
			diff := gridSizeY - csh.PreferredSize.Y
			switch bd.AnchorY {
			case 0:
				gridBounds.Min.Y += diff / 2
				gridBounds.Max.Y -= diff / 2
			case AnchorMin:
				gridBounds.Max.Y -= diff
			case AnchorMax:
				gridBounds.Min.Y += diff
			case AnchorMin | AnchorMax:
			}
		}

		g.PlaceBlock(child, gridBounds)
	}

	g.Invalidate()
}

func safeRect(path draw2d.GraphicContext, min, max geom.Coord) {
	x1, y1 := min.X, min.Y
	x2, y2 := max.X, max.Y
	x, y := path.LastPoint()
	path.MoveTo(x1, y1)
	path.LineTo(x2, y1)
	path.LineTo(x2, y2)
	path.LineTo(x1, y2)
	path.Close()
	path.MoveTo(x, y)
}

func (g *Grid) draw(gc draw2d.GraphicContext) {
	gc.Clear()
	gc.SetFillColor(color.RGBA{150, 150, 150, 255})
	safeRect(gc, geom.Coord{0, 0}, g.Size)
	gc.FillStroke()

	_, minXs, _ := g.hflex.constrain(g.Size.X)
	for _, x := range minXs[1:] {
		gc.MoveTo(x, 0)
		gc.LineTo(x, g.Size.Y)

	}
	_, minYs, _ := g.vflex.constrain(g.Size.Y)
	for _, y := range minYs[1:] {
		gc.MoveTo(0, y)
		gc.LineTo(g.Size.X, y)

	}
	gc.Stroke()
	// _, _, maxYs := g.vflex.constrain(g.Size.Y)
}

func (g *Grid) handleEvents() {
	for {
		select {
		case e := <-g.UserEvents:
			switch e := e.(type) {
			default:
				g.Foundation.HandleEvent(e)
			}
		case bsh := <-g.BlockSizeHints:
			if !g.children[bsh.Block] {
				// Do I know you?
				break
			}
			g.ChildrenHints[bsh.Block] = bsh.SizeHint

			// if g.Block.ID == 2 {
			// 	uik.Report("gotpr", g.Block.ID, bsh.Block.ID, bsh.SizeHint)
			// }
			g.reflex(bsh.Block)
			g.makePreferences()
			g.regrid()
		case e := <-g.BlockInvalidations:
			g.DoBlockInvalidation(e)
			// go uik.ShowBuffer("grid", g.Buffer)
		case e := <-g.ResizeEvents:
			g.Size = e.Size

			// if g.Block.ID == 2 {
			// 	uik.Report("sized", g.Block.ID, g.Size)
			// }
			g.regrid()
			g.Invalidate()
		case g.config = <-g.setConfig:
			g.makePreferences()
		case g.getConfig <- g.config:
		case bd := <-g.add:
			g.addBlock(bd)
		case b := <-g.remove:
			g.remBlock(b)
		}
	}
}
