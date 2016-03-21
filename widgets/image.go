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

package widgets

import (
	"image"
	"math"

	"github.com/dddaisuke/draw2d/draw2d"
	"github.com/kirillDanshin/go.uik"
	"github.com/skelterjohn/geom"
)

type ImageConfig struct {
	Image image.Image
}

func (ic ImageConfig) ImageSize() (s geom.Coord) {
	ib := ic.Image.Bounds()
	is := ib.Size()
	s.X = float64(is.X)
	s.Y = float64(is.Y)
	return
}

type Image struct {
	uik.Block

	config    ImageConfig
	setConfig chan ImageConfig
	getConfig chan ImageConfig
}

func NewImage(cfg ImageConfig) (i *Image) {
	i = new(Image)

	i.Initialize()
	i.updateConfig(cfg)

	go i.handleEvents()

	return
}

func (i *Image) Initialize() {
	i.Block.Initialize()

	i.setConfig = make(chan ImageConfig, 1)
	i.getConfig = make(chan ImageConfig, 1)

	i.Paint = func(gc draw2d.GraphicContext) {
		i.draw(gc)
	}
}

func (i *Image) SetConfig(cfg ImageConfig) {
	i.setConfig <- cfg
}

func (i *Image) GetConfig() (cfg ImageConfig) {
	cfg = <-i.getConfig
	return
}

func (i *Image) draw(gc draw2d.GraphicContext) {
	ib := i.config.Image.Bounds()
	s := ib.Size()
	w := float64(s.X)
	h := float64(s.Y)
	sx := i.Size.X / w
	sy := i.Size.Y / h
	// uik.Report(i.Size, sx, sy)
	gc.Scale(sx, sy)
	gc.DrawImage(i.config.Image)
}

func (i *Image) updateConfig(config ImageConfig) {
	i.config = config
	i.SetSizeHint(uik.SizeHint{
		MinSize:       geom.Coord{},
		PreferredSize: i.config.ImageSize(),
		MaxSize:       geom.Coord{math.Inf(1), math.Inf(1)},
	})
	i.Invalidate()
}

func (i *Image) handleEvents() {

	for {
		select {
		case e := <-i.UserEvents:
			switch e := e.(type) {
			default:
				i.HandleEvent(e)
			}
		case e := <-i.ResizeEvents:
			i.DoResizeEvent(e)
		case config := <-i.setConfig:
			if i.config == config {
				break
			}
			i.updateConfig(config)
		case i.getConfig <- i.config:
		}
	}
}
