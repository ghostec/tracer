package tracer

import (
	"image"
	"image/color"
)

type PPM struct {
	rgba *image.RGBA
}

func NewPPM(frame *Frame) *PPM {
	bounds := image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: frame.Width(), Y: frame.Height()},
	}

	rgba := image.NewRGBA(bounds)

	for row := 0; row < frame.Height(); row++ {
		for col := 0; col < frame.Width(); col++ {
			c := frame.Get(row, col).RGBA()
			rgba.Set(col, frame.Height()-row-1, color.NRGBA{R: c[0], G: c[1], B: c[2], A: c[3]})
		}
	}

	return &PPM{rgba: rgba}
}

func (ppm *PPM) ColorModel() color.Model {
	return color.RGBAModel
}

func (ppm *PPM) Bounds() image.Rectangle {
	return ppm.rgba.Bounds()
}

func (ppm *PPM) At(x, y int) color.Color {
	return ppm.rgba.At(x, y)
}
