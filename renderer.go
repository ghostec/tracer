package tracer

import (
	"sync"
)

type Renderer struct {
	mu sync.Mutex

	done     chan struct{}
	jobs     chan Job
	nWorkers int
}

func NewRenderer(nWorkers int) *Renderer {
	return &Renderer{
		done:     make(chan struct{}, 1),
		jobs:     make(chan Job, nWorkers),
		nWorkers: nWorkers,
	}
}

func (renderer *Renderer) Start() {
	renderer.mu.Lock()
	defer renderer.mu.Unlock()

	for i := 0; i < renderer.nWorkers; i++ {
		go Worker(renderer.jobs, renderer.done)
	}
}

func (renderer *Renderer) Stop() {
	close(renderer.done)
}

func (renderer *Renderer) Render(settings RenderSettings, stop <-chan bool) {
	wg := sync.WaitGroup{}

	width := settings.Frame.Width()
	height := settings.Frame.Height()

	for row := 0; row < height; row++ {
		row := row
		wg.Add(1)

		renderer.jobs <- func() {
			for col := 0; col < width; col++ {
				col := col
				samples := make([]Color, settings.SamplesPerPixel)
				for s := 0; s < settings.SamplesPerPixel; s++ {
					u, v := JitteredCameraCoordinatesFromPixel(row, col, width, height)
					r := settings.Camera.GetRay(u, v)
					samples[s] = settings.RayColorFunc(r, settings.Hitter, settings.MaxDepth, 0)
				}

				settings.Frame.Set(row, col, settings.AggColorFunc(samples))
			}
			wg.Done()
		}
	}

	// TODO: won't stop on `stop`
	wg.Wait()
}

func Worker(in chan Job, done chan struct{}) {
	for {
		select {
		case job := <-in:
			job()
		case <-done:
			return
		}
	}
}

type Job func()

type RenderSettings struct {
	Frame           *Frame
	Camera          *Camera
	Hitter          Hitter
	SamplesPerPixel int
	MaxDepth        int
	RayColorFunc    RayColorFunc
	AggColorFunc    AggColorFunc
}
