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

package uik

import (
	"image"
	"image/color"

	"github.com/dddaisuke/draw2d/draw2d"
	"github.com/dddaisuke/freetype-go/freetype/truetype"
)

var DefaultFontData = draw2d.FontData{
	Name:   "Fira Sans Light",
	Family: draw2d.FontFamilySans,
	Style:  draw2d.FontStyleNormal,
}

func init() {
	font, err := truetype.Parse(defaultFont())
	if err != nil {
		// TODO: log error
		println("error!")
		println(err.Error())
	}

	draw2d.RegisterFont(DefaultFontData, font)
}

func GetFontHeight(fd draw2d.FontData, size float64) (height float64) {
	font := draw2d.GetFont(fd)
	fupe := font.FUnitsPerEm()
	bounds := font.Bounds(fupe)
	height = float64(bounds.YMax-bounds.YMin) * size / float64(fupe)
	return
}

func RenderString(text string, fd draw2d.FontData, size float64, color color.Color) (buffer image.Image) {

	const stretchFactor = 1.2

	height := GetFontHeight(fd, size) * stretchFactor
	widthMax := float64(len(text)) * size

	buf := image.NewRGBA(image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{int(widthMax + 1), int(height + 1)},
	})

	gc := draw2d.NewGraphicContext(buf)
	gc.Translate(0, height/stretchFactor)
	gc.SetFontData(fd)
	gc.SetFontSize(size)
	gc.SetStrokeColor(color)
	width := gc.FillString(text)

	buffer = buf.SubImage(image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{int(width + 1), int(height + 1)},
	})

	return
}
