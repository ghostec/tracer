package tracer

import (
	"sync/atomic"

	"lukechampine.com/frand"
)

func RayColor(r Ray, n Hitter, depth int) Color {
	if depth <= 0 {
		return Color{}
	}

	if hr := n.Hit(r); hr.Hit {
		if sr := hr.Material.Scatter(r, hr); sr.Scatter {
			return Color(sr.Attenuation.Vec3().MulVec3(RayColor(sr.Ray, n, depth-1).Vec3()))
		}
		return Color{}
	}

	unitDirection := r.Direction.Unit()
	t := 0.5 * (unitDirection[1] + 1.0)
	return Color(Vec3{1, 1, 1}.MulFloat(1.0 - t).Add(Vec3{0.5, 0.7, 1.0}.MulFloat(t)))
}

func Render(frame *Frame, cam Camera, l HitterList, samplesPerPixel, nWorkers int, stop <-chan bool) {
	jobs := make(chan Job, nWorkers)
	results := make(chan JobResult, nWorkers)
	done := make(chan bool, 1)

	settings := RenderSettings{
		FrameWidth:      frame.Width(),
		FrameHeight:     frame.Height(),
		MaxDepth:        50,
		Camera:          cam,
		SamplesPerPixel: samplesPerPixel,
		Hitter:          NewBVHNode(l),
	}

	for i := 0; i < nWorkers; i++ {
		go Worker(jobs, results, done)
	}

	go func() {
		for col := 0; col < frame.Width(); col++ {
			for row := 0; row < frame.Height(); row++ {
				select {
				case <-stop:
					return
				default:
					jobs <- Job{Row: row, Column: col, Settings: &settings}
				}
			}
		}
	}()

	go func() {
		recv := uint64(0)
		for {
			select {
			case result := <-results:
				frame.Set(result.Row, result.Column, result.Color)
				atomic.AddUint64(&recv, 1)
				if recv == uint64(frame.Width()*frame.Height()) {
					close(done)
					return
				}
			case <-stop:
				close(done)
				return
			case <-done:
				return
			}
		}
	}()

	<-done
}

func Worker(in chan Job, out chan JobResult, done chan bool) {
	for {
		select {
		case job := <-in:
			c := Color{}

			for s := 0; s < job.Settings.SamplesPerPixel; s++ {
				u := (float64(job.Column) + frand.Float64()) / float64(job.Settings.FrameWidth-1)
				v := (float64(job.Row) + frand.Float64()) / float64(job.Settings.FrameHeight-1)
				r := job.Settings.Camera.GetRay(u, v)
				c = Color(c.Vec3().Add(RayColor(r, job.Settings.Hitter, job.Settings.MaxDepth).Vec3()))
			}

			c = Color(c.Vec3().MulFloat(1.0 / float64(job.Settings.SamplesPerPixel)))

			out <- JobResult{Row: job.Row, Column: job.Column, Color: c}
		case <-done:
			return
		}
	}
}

type Job struct {
	Row, Column int
	Settings    *RenderSettings
}

type RenderSettings struct {
	FrameWidth      int
	FrameHeight     int
	SamplesPerPixel int
	MaxDepth        int
	Camera          Camera
	Hitter          Hitter
}

type JobResult struct {
	Row, Column int
	Color       Color
}

type Ray struct {
	Origin    Point3
	Direction Vec3
}

func (r Ray) At(t float64) Point3 {
	return Point3(Vec3(r.Origin).Add(r.Direction.MulFloat(t)))
}
