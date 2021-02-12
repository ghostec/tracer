package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime/pprof"
	"sync"

	"lukechampine.com/frand"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		tracer()
		defer pprof.StopCPUProfile()
	}
}

func tracer() {
	aspectRatio := 16.0 / 9.0
	imageWidth := 800
	imageHeight := int(float64(imageWidth) / aspectRatio)

	l := HitterList{
		Sphere{Center: Point3{0, 0, -1}, Radius: 0.5, Material: Lambertian{Albedo: Color{0.7, 0.3, 0.3}}},
		Sphere{Center: Point3{0, -100.5, -1}, Radius: 100, Material: Lambertian{Albedo: Color{0.8, 0.8, 0}}},
	}

	cam := DefaultCamera()

	frame := NewFrame(imageWidth, imageHeight)

	Render(frame, cam, l, 20)

	if err := frame.ToPPM("image.ppm"); err != nil {
		panic(err)
	}
}

func Render(frame *Frame, cam Camera, l HitterList, nWorkers int) {
	jobs := make(chan Job, nWorkers)
	results := make(chan JobResult, nWorkers)
	done := make(chan bool, 1)

	wg := sync.WaitGroup{}
	wg.Add(frame.Width() * frame.Height())

	settings := RenderSettings{
		FrameWidth:      frame.Width(),
		FrameHeight:     frame.Height(),
		MaxDepth:        50,
		Camera:          cam,
		SamplesPerPixel: 100,
		HitterList:      l,
	}

	for i := 0; i < nWorkers; i++ {
		go Worker(jobs, results, done)
	}

	go func() {
		for col := 0; col < frame.Width(); col++ {
			for row := 0; row < frame.Height(); row++ {
				jobs <- Job{Row: row, Column: col, Settings: &settings}
			}
		}
	}()

	go func() {
		for {
			select {
			case result := <-results:
				frame.Set(result.Row, result.Column, result.Color)
				wg.Done()
			case <-done:
				return
			}
		}
	}()

	wg.Wait()

	close(done)
}

func Worker(in chan Job, out chan JobResult, done chan bool) {
	for {
		select {
		case job := <-in:
			c := Color{0, 0, 0}

			for s := 0; s < job.Settings.SamplesPerPixel; s++ {
				u := (float64(job.Column) + frand.Float64()) / float64(job.Settings.FrameWidth-1)
				v := (float64(job.Row) + frand.Float64()) / float64(job.Settings.FrameHeight-1)
				r := job.Settings.Camera.GetRay(u, v)
				c = Color(c.Vec3().Add(RayColor(r, job.Settings.HitterList, job.Settings.MaxDepth).Vec3()))
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
	HitterList      HitterList
}

type JobResult struct {
	Row, Column int
	Color       Color
}

type Frame struct {
	content [][]Color
	mu      sync.Mutex
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
	frame.mu.Lock()
	defer frame.mu.Unlock()

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

func RayColor(r Ray, l HitterList, depth int) Color {
	if depth <= 0 {
		return Color{}
	}

	if hr := l.Hit(r); hr.Hit {
		if sr := hr.Material.Scatter(r, hr); sr.Scatter {
			return Color(sr.Attenuation.Vec3().MulVec3(RayColor(sr.Ray, l, depth-1).Vec3()))
		}
		return Color{0, 0, 0}
	}

	unitDirection := r.Direction.Unit()
	t := 0.5 * (unitDirection[1] + 1.0)
	return Color(Vec3{1, 1, 1}.MulFloat(1.0 - t).Add(Vec3{0.5, 0.7, 1.0}.MulFloat(t)))
}

type Vec3 [3]float64

func (v Vec3) Add(o Vec3) Vec3 {
	return Vec3{v[0] + o[0], v[1] + o[1], v[2] + o[2]}
}

func (v Vec3) Sub(o Vec3) Vec3 {
	return v.Add(o.Neg())
}

func (v Vec3) Neg() Vec3 {
	return Vec3{-v[0], -v[1], -v[2]}
}

func (v Vec3) Dot(o Vec3) float64 {
	return v[0]*o[0] + v[1]*o[1] + v[2]*o[2]
}

func (v Vec3) Cross(o Vec3) Vec3 {
	return Vec3{
		v[1]*o[2] - v[2]*o[1],
		v[2]*o[0] - v[0]*o[2],
		v[0]*o[1] - v[1]*o[0],
	}
}

func (v Vec3) MulFloat(s float64) Vec3 {
	return Vec3{s * v[0], s * v[1], s * v[2]}
}

func (v Vec3) MulVec3(o Vec3) Vec3 {
	return Vec3{v[0] * o[0], v[1] * o[1], v[2] * o[2]}
}

func (v Vec3) LenSq() float64 {
	return v.Dot(v)
}

func (v Vec3) Len() float64 {
	return math.Sqrt(v.LenSq())
}

func (v Vec3) Unit() Vec3 {
	l := v.Len()
	if l == 0.0 {
		return v
	}
	return v.MulFloat(1 / l)
}

func (v Vec3) NearZero() bool {
	s := 1e-8
	return (math.Abs(v[0]) < s) && (math.Abs(v[1])) < s && (math.Abs(v[2]) < s)
}

type Color Vec3
type Point3 Vec3

func (c Color) Vec3() Vec3 {
	return Vec3(c)
}

func (p Point3) Vec3() Vec3 {
	return Vec3(p)
}

func (c Color) String() string {
	for i, cc := range c {
		cc = math.Sqrt(cc)
		cc = 256 * Clamp(cc, 0, 0.999)
		c[i] = cc
	}
	return fmt.Sprintf("%d %d %d", int(c[0]), int(c[1]), int(c[2]))
}

type Ray struct {
	Origin    Point3
	Direction Vec3
}

func (r Ray) At(t float64) Point3 {
	return Point3(Vec3(r.Origin).Add(r.Direction.MulFloat(t)))
}

type Hitter interface {
	Hit(Ray) HitRecord
}

type HitRecord struct {
	Hit       bool
	FrontFace bool
	T         float64
	P         Point3
	Normal    Vec3
	Material  Material
}

type Sphere struct {
	Center   Point3
	Radius   float64
	Material Material
}

func (s Sphere) Hit(r Ray) HitRecord {
	oc := Vec3(r.Origin).Sub(Vec3(s.Center))
	a := r.Direction.Dot(r.Direction)
	halfB := oc.Dot(r.Direction)
	c := oc.Dot(oc) - s.Radius*s.Radius

	discriminant := halfB*halfB - a*c
	if discriminant < 0 {
		return HitRecord{}
	}

	tMin, tMax := 0.0001, math.Inf(+1)

	sqrtd := math.Sqrt(discriminant)
	// Find the nearest root that lies in the acceptable range.
	root := (-halfB - sqrtd) / a
	if root < tMin || tMax < root {
		root = (-halfB + sqrtd) / a
		if root < tMin || tMax < root {
			return HitRecord{}
		}
	}

	hr := HitRecord{
		Hit:      true,
		T:        root,
		P:        r.At(root),
		Material: s.Material,
	}

	outwardNormal := hr.P.Vec3().Sub(s.Center.Vec3()).MulFloat(1 / s.Radius)
	hr.FrontFace = r.Direction.Dot(outwardNormal) < 0
	hr.Normal = outwardNormal
	if !hr.FrontFace {
		hr.Normal = hr.Normal.Neg()
	}

	return hr
}

type HitterList []Hitter

func (h HitterList) Hit(r Ray) (hr HitRecord) {
	hr.T = math.Inf(+1)
	for _, hh := range h {
		hhr := hh.Hit(r)
		if hhr.Hit && hhr.T < hr.T {
			hr = hhr
		}
	}
	return
}

type Camera struct {
	AspectRatio    float64
	ViewportHeight float64
	FocalLength    float64
}

func (c Camera) GetRay(u, v float64) Ray {
	viewportWidth := c.AspectRatio * c.ViewportHeight
	origin := Point3{0, 0, 0}
	horizontal := Vec3{viewportWidth, 0, 0}
	vertical := Vec3{0, c.ViewportHeight, 0}
	lowerLeftCorner := origin.Vec3().Sub(horizontal.MulFloat(0.5)).Sub(vertical.MulFloat(0.5)).Sub(Vec3{0, 0, c.FocalLength})

	return Ray{
		Origin:    origin,
		Direction: lowerLeftCorner.Add(horizontal.MulFloat(u)).Add(vertical.MulFloat(v)).Sub(origin.Vec3()),
	}
}

func DefaultCamera() Camera {
	return Camera{
		AspectRatio:    16.0 / 9.0,
		ViewportHeight: 2.0,
		FocalLength:    1.0,
	}
}

func Clamp(x, min, max float64) float64 {
	switch {
	case x < min:
		return min
	case x > max:
		return max
	default:
		return x
	}
}

func RandomVec3(min, max float64) Vec3 {
	return Vec3{Random(min, max), Random(min, max), Random(min, max)}
}

func Random(min, max float64) float64 {
	return frand.Float64()*(max-min) + min
}

func RandomInUnitSphere() Vec3 {
	for {
		p := RandomVec3(-1, 1)
		if p.LenSq() >= 1 {
			continue
		}
		return p
	}
}

func RandomUnitVector() Vec3 {
	return RandomInUnitSphere().Unit()
}

func RandomInHemisphere(normal Vec3) Vec3 {
	inUnitSphere := RandomInUnitSphere()
	if inUnitSphere.Dot(normal) > 0.0 {
		return inUnitSphere
	}
	return inUnitSphere.Neg()
}

type Material interface {
	Scatter(Ray, HitRecord) ScatterRecord
}

type ScatterRecord struct {
	Scatter     bool
	Ray         Ray
	Attenuation Color
}

type Lambertian struct {
	Albedo Color
}

func (l Lambertian) Scatter(ray Ray, hr HitRecord) ScatterRecord {
	scatterDirection := hr.Normal.Add(RandomUnitVector())

	// Catch degenerate scatter direction
	if scatterDirection.NearZero() {
		scatterDirection = hr.Normal
	}

	scattered := Ray{Origin: hr.P, Direction: scatterDirection}

	return ScatterRecord{
		Scatter:     true,
		Ray:         scattered,
		Attenuation: l.Albedo,
	}
}
