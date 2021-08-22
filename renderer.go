package tracer

import (
	"sync"
)

type Renderer struct {
	mu sync.Mutex

	done     chan bool
	jobs     chan Job
	nWorkers int
}

func NewRenderer(nWorkers int) *Renderer {
	return &Renderer{
		done:     make(chan bool, 1),
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

func (renderer *Renderer) Render(settings RenderSettings, stop <-chan bool) {
	// send jobs
	sent := uint(0)
	spawnerDone := make(chan bool, 1)
	results := make(chan JobResult, renderer.nWorkers)

	go func() {
		defer close(spawnerDone)

		for col := 0; col < settings.Frame.Width(); col++ {
			for row := 0; row < settings.Frame.Height(); row++ {
				select {
				case <-stop:
					return
				default:
					renderer.jobs <- Job{Results: results, Row: row, Column: col, Settings: &settings}
					sent += 1
				}
			}
		}
	}()

	// receive results
	recv := uint(0)
	done := make(chan bool, 1)
	go func() {
	ReceiveResults:
		for {
			select {
			case result := <-results:
				settings.Frame.Set(result.Row, result.Column, result.Color)
				recv += 1
				if recv == uint(settings.Frame.Width()*settings.Frame.Height()) {
					break ReceiveResults
				}
			case <-stop:
				break ReceiveResults
			case <-renderer.done:
				break ReceiveResults
			}
		}

		// wait until all remaining jobs are sent in case of early stop
		<-spawnerDone

		for recv != sent {
			select {
			case <-results:
				recv += 1
				continue
			}
		}

		close(done)
		close(results)
	}()

	<-done
}

func Worker(in chan Job, done chan bool) {
	for {
		select {
		case job := <-in:
			samples := make([]Color, 0, job.Settings.SamplesPerPixel)
			for s := 0; s < job.Settings.SamplesPerPixel; s++ {
				u, v := JitteredCameraCoordinatesFromPixel(job.Settings.LineA+job.Row, job.Column, job.Settings.Frame.Width(), job.Settings.Lines)
				r := job.Settings.Camera.GetRay(u, v)
				c := job.Settings.RayColorFunc(r, job.Settings.Hitter, job.Settings.MaxDepth, 0)
				samples = append(samples, c)
			}
			c := job.Settings.AggColorFunc(samples)
			job.Results <- JobResult{Row: job.Row, Column: job.Column, Color: c}
		case <-done:
			return
		}
	}
}

type Job struct {
	Results     chan JobResult
	Row, Column int
	Settings    *RenderSettings
}

type RenderSettings struct {
	Frame           *Frame
	Camera          *Camera
	Hitter          Hitter
	SamplesPerPixel int
	MaxDepth        int
	RayColorFunc    RayColorFunc
	AggColorFunc    AggColorFunc
	Lines           int
	LineA           int
	LineB           int
}

type JobResult struct {
	Row, Column int
	Color       Color
}
