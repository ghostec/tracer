package tracer

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

type Frame struct {
	content [][]Color
	// TODO: readwrite lock
	mu      sync.Mutex
	samples int
}

func NewFrame(width, height int) *Frame {
	content := make([][]Color, height)
	for i := 0; i < height; i++ {
		content[i] = make([]Color, width)
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

func (frame *Frame) PPM() string {
	frame.mu.Lock()
	defer frame.mu.Unlock()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("P3\n%d %d\n255\n", frame.Width(), frame.Height()))
	for row := frame.Height() - 1; row >= 0; row-- {
		for column := 0; column < frame.Width(); column++ {
			sb.WriteString(fmt.Sprintf("%s", frame.content[row][column].String()))
			if column < frame.Width()-1 {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func (frame *Frame) Avg(other *Frame) error {
	frame.mu.Lock()
	other.mu.Lock()
	defer frame.mu.Unlock()
	defer other.mu.Unlock()

	if frame.Width() != other.Width() || frame.Height() != other.Height() {
		return errors.New("placeholder")
	}

	for row := 0; row < frame.Height(); row++ {
		for col := 0; col < frame.Width(); col++ {
			frameColor := frame.content[row][col].Vec3().MulFloat(float64(frame.samples + 1))
			otherColor := other.content[row][col].Vec3().MulFloat(float64(other.samples + 1))
			color := frameColor.Add(otherColor).MulFloat(float64(1.0) / float64(frame.samples+other.samples+2))
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
	ppm := frame.PPM()
	if _, err := f.WriteString(ppm); err != nil {
		return err
	}
	return f.Close()
}
