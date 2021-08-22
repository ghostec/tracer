package tracer

import (
	"bufio"
	"bytes"
	"errors"
	"image/png"
	"os"
	"sync"
)

type Frame struct {
	content [][]Color
	// TODO: readwrite lock
	mu      sync.Mutex
	samples int
}

func NewFrame(width, height int, transparentBackground bool) *Frame {
	content := make([][]Color, height)
	for j := 0; j < height; j++ {
		content[j] = make([]Color, width)
		if transparentBackground {
			for i := 0; i < width; i++ {
				content[j][i] = Transparent
			}
		}
	}
	return &Frame{
		content: content,
	}
}

func (frame *Frame) Set(row, col int, color Color) {
	frame.mu.Lock()
	defer frame.mu.Unlock()

	frame.content[row][col] = color
}

func (frame *Frame) Get(row, col int) Color {
	frame.mu.Lock()
	defer frame.mu.Unlock()

	return frame.content[row][col]
}

func (frame *Frame) Width() int {
	return len(frame.content[0])
}

func (frame *Frame) Height() int {
	return len(frame.content)
}

func (frame *Frame) Avg(other *Frame) error {
	frame.mu.Lock()
	other.mu.Lock()
	defer frame.mu.Unlock()
	defer other.mu.Unlock()

	width, height := frame.Width(), frame.Height()

	if width != other.Width() || height != other.Height() {
		return errors.New("placeholder")
	}

	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			frameColor := Vec3(frame.content[row][col]).MulFloat(float64(frame.samples + 1))
			otherColor := Vec3(other.content[row][col]).MulFloat(float64(other.samples + 1))
			var color Vec3
			switch {
			case Color(frameColor).Transparent():
				color = otherColor
			case Color(otherColor).Transparent():
				color = frameColor
			default:
				color = frameColor.Add(otherColor).MulFloat(float64(1.0) / float64(frame.samples+other.samples+2))
			}
			frame.content[row][col] = Color(color)
		}
	}

	frame.samples += other.samples + 1

	return nil
}

func (frame *Frame) Save(path string) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	if err := png.Encode(f, NewPPM(frame)); err != nil {
		return err
	}
	return f.Close()
}

func (frame *Frame) Bytes() ([]byte, error) {
	var buf bytes.Buffer

	w := bufio.NewWriter(&buf)
	if err := png.Encode(w, NewPPM(frame)); err != nil {
		return nil, err
	}
	if err := w.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (frame *Frame) Blend(other *Frame, frameAlpha, otherAlpha float64) error {
	if frame.Width() != other.Width() || frame.Height() != other.Height() {
		return errors.New("placeholder")
	}

	for row := 0; row < frame.Height(); row++ {
		for col := 0; col < frame.Width(); col++ {
			blend := frame.Get(row, col).Blend(other.Get(row, col), frameAlpha, otherAlpha)
			frame.Set(row, col, blend)
		}
	}

	return nil
}
