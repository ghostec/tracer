package tracer

import (
	"fmt"
	"os"
	"sync"
)

type Frame struct {
	content [][]Color
	// TODO: readwrite lock
	mu sync.Mutex
}

func NewFrame(width, height int) *Frame {
	content := make([][]Color, height)
	for i := 0; i < height; i++ {
		content[i] = make([]Color, width)
	}
	return &Frame{
		content: content,
		mu:      sync.Mutex{},
	}
}

func (frame *Frame) Set(row, col int, color Color) {
	frame.content[row][col] = color
}

func (frame *Frame) Get(row, col int) Color {
	return frame.content[row][col]
}

func (frame *Frame) Width() int {
	return len(frame.content[0])
}

func (frame *Frame) Height() int {
	return len(frame.content)
}

func (frame *Frame) ToPPM(path string) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(fmt.Sprintf("P3\n%d %d\n255\n", frame.Width(), frame.Height())); err != nil {
		return err
	}
	for row := frame.Height() - 1; row >= 0; row-- {
		for column := 0; column < frame.Width(); column++ {
			if _, err := f.WriteString(fmt.Sprintf("%s ", frame.content[row][column].String())); err != nil {
				return err
			}
		}
	}
	return f.Close()
}
