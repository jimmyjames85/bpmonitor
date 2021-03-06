// Copyright ©2017 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plotter

import (
	"image"

	"github.com/gonum/plot"
	"github.com/gonum/plot/palette"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"
)

// ColorBar is a plot.Plotter that draws a color bar legend for a ColorMap.
type ColorBar struct {
	ColorMap palette.ColorMap

	// Vertical determines wether the legend will be
	// plotted vertically or horizontally.
	// The default is false (horizontal).
	Vertical bool

	// Colors specifies the number of colors to be
	// shown in the legend. If Colors is not specified,
	// a default will be used.
	Colors int
}

// colors returns the number of colors to be shown
// in the legend, substituting invalid values
// with the default of one color per vg.Point.
func (l *ColorBar) colors(c draw.Canvas) int {
	if l.Colors > 0 {
		return l.Colors
	}
	if l.Vertical {
		return int((c.Max.Y - c.Min.Y).Points())
	}
	return int((c.Max.X - c.Min.X).Points())
}

// check determines whether the ColorBar is
// valid in its current configuration.
func (l *ColorBar) check() {
	if l.ColorMap == nil {
		panic("plotter: nil ColorMap in ColorBar")
	}
	if l.ColorMap.Max() == l.ColorMap.Min() {
		panic("plotter: ColorMap Max==Min")
	}
}

// Plot implements the Plot method of the plot.Plotter interface.
func (l *ColorBar) Plot(c draw.Canvas, p *plot.Plot) {
	l.check()
	colors := l.colors(c)
	var img *image.NRGBA64
	var xmin, xmax, ymin, ymax vg.Length
	if l.Vertical {
		trX, trY := p.Transforms(&c)
		xmin = trX(l.ColorMap.Min())
		ymin = trY(0)
		xmax = trX(l.ColorMap.Max())
		ymax = trY(1)
		img = image.NewNRGBA64(image.Rectangle{
			Min: image.Point{X: 0, Y: 0},
			Max: image.Point{X: 1, Y: colors},
		})
		for i := 0; i < colors; i++ {
			color, err := l.ColorMap.At(float64(i) / float64(colors-1))
			if err != nil {
				panic(err)
			}
			img.Set(0, colors-1-i, color)
		}
	} else {
		trX, trY := p.Transforms(&c)
		ymin = trY(l.ColorMap.Min())
		xmin = trX(0)
		ymax = trY(l.ColorMap.Max())
		xmax = trX(1)
		img = image.NewNRGBA64(image.Rectangle{
			Min: image.Point{X: 0, Y: 0},
			Max: image.Point{X: colors, Y: 1},
		})
		for i := 0; i < colors; i++ {
			color, err := l.ColorMap.At(float64(i) / float64(colors-1))
			if err != nil {
				panic(err)
			}
			img.Set(i, 0, color)
		}
	}
	rect := vg.Rectangle{
		Min: vg.Point{X: xmin, Y: ymin},
		Max: vg.Point{X: xmax, Y: ymax},
	}
	c.DrawImage(rect, img)
}

// DataRange implements the DataRange method
// of the plot.DataRanger interface.
func (l *ColorBar) DataRange() (xmin, xmax, ymin, ymax float64) {
	l.check()
	if l.Vertical {
		return 0, 1, l.ColorMap.Min(), l.ColorMap.Max()
	}
	return l.ColorMap.Min(), l.ColorMap.Max(), 0, 1
}
