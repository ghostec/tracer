package tracer

import (
	"sync/atomic"

	"lukechampine.com/frand"
)

type RayColorFunc func(ray Ray, hitter Hitter, depth int, bounces int) Color
type AggColorFunc func([]Color) Color

func RayColor(ray Ray, n Hitter, depth, bounces int) Color {
	if bounces >= depth {
		return Color{}
	}

	hr := n.Hit(ray)
	if !hr.Hit {
		unitDirection := ray.Direction.Unit()
		t := 0.5 * (unitDirection[1] + 1.0)
		return Color(Vec3{1, 1, 1}.MulFloat(1.0 - t).Add(Vec3{0.5, 0.7, 1.0}.MulFloat(t)))
	}

	sr := hr.Material.Scatter(ray, hr)
	if !sr.Scatter {
		return Color{}
	}

	return Color(sr.Attenuation.Vec3().MulVec3(RayColor(sr.Ray, n, depth, bounces+1).Vec3()))
}

func RayBVHID(ray Ray, n Hitter, _, _ int) Color {
	hr := n.Hit(ray)
	if !hr.Hit {
		return Color{}
	}

	return Uint64ToColor(hr.BVHNode.ID)
}

func RayDistance(ray Ray, n Hitter, _, _ int) Color {
	hr := n.Hit(ray)
	if !hr.Hit {
		return Color{}
	}

	distVec := ray.Origin.Vec3().Sub(hr.P.Vec3())
	dist := distVec.Len()

	if dist > 100 {
		return Color{}
	}

	col := (-25.5*dist + 255.0) / 255.0
	return Color{col, col, col}
}

func ColorToUint64(color Color) uint64 {
	// return uint64(math.Ceil(math.Pow(color[0], 3.0)))
	return uint64(color[0])
}

func Uint64ToColor(val uint64) Color {
	// pow := math.Pow(float64(val), 1.0/3.0)
	// return Color{pow, pow, pow}
	return Color{float64(val), 0, 0}
}

func Render(frame *Frame, cam Camera, l HitterList, rayColorFun RayColorFunc, aggColorFun AggColorFunc, samplesPerPixel, nWorkers int, stop <-chan bool) {
	jobs := make(chan Job, nWorkers)
	results := make(chan JobResult, nWorkers)
	done := make(chan bool, 1)

	bvh, err := NewBVHNode(l)
	if err != nil {
		panic(err)
	}

	settings := RenderSettings{
		FrameWidth:      frame.Width(),
		FrameHeight:     frame.Height(),
		MaxDepth:        50,
		Camera:          cam,
		SamplesPerPixel: samplesPerPixel,
		Hitter:          bvh,
	}

	for i := 0; i < nWorkers; i++ {
		go Worker(rayColorFun, aggColorFun, jobs, results, done)
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

func AvgSamples(samples []Color) Color {
	v := Vec3{}
	for _, cc := range samples {
		v = v.Add(cc.Vec3())
	}
	return Color(v.MulFloat(1.0 / float64(len(samples))))
}

func EdgeSamples(samples []Color) Color {
	freq := map[uint64]int{}
	mostFreqKey := uint64(0)
	for _, cc := range samples {
		key := ColorToUint64(cc)
		freq[key] += 1
		if freq[key] > freq[mostFreqKey] {
			mostFreqKey = key
		}
	}

	return Uint64ToColor(mostFreqKey)
}

func Worker(rayColorFun RayColorFunc, aggColorFun AggColorFunc, in chan Job, out chan JobResult, done chan bool) {
	for {
		select {
		case job := <-in:
			samples := make([]Color, 0, job.Settings.SamplesPerPixel)
			for s := 0; s < job.Settings.SamplesPerPixel; s++ {
				u := (float64(job.Column) + frand.Float64()) / float64(job.Settings.FrameWidth-1)
				v := (float64(job.Row) + frand.Float64()) / float64(job.Settings.FrameHeight-1)
				r := job.Settings.Camera.GetRay(u, v)
				c := rayColorFun(r, job.Settings.Hitter, job.Settings.MaxDepth, 0)
				samples = append(samples, c)
			}
			c := aggColorFun(samples)
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
