package main

import (
	"flag"
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/ghostec/tracer"
)

var width = flag.Int("width", 1000, "image width")
var aspectRatio = flag.Float64("aspect-ratio", 16.0/9.0, "image width")
var parallelism = flag.Int("parallelism", runtime.NumCPU(), "number of render routines to run")
var cpuProfile = flag.String("cpu-profile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		execute()
		defer pprof.StopCPUProfile()
	}
}

func execute() {
	imageWidth := *width
	imageHeight := int(float64(imageWidth) / *aspectRatio)

	l := tracer.HitterList{
		tracer.Sphere{Center: tracer.Point3{0, 0, -1}, Radius: 0.5, Material: tracer.Lambertian{Albedo: tracer.Color{0.7, 0.3, 0.3}}},
		tracer.Sphere{Center: tracer.Point3{0, -100.5, -1}, Radius: 100, Material: tracer.Lambertian{Albedo: tracer.Color{0.8, 0.8, 0}}},
		tracer.Sphere{Center: tracer.Point3{-1, 0, -1}, Radius: 0.5, Material: tracer.Metal{Albedo: tracer.Color{0.8, 0.8, 0.8}, Fuzz: 0.1}},
		tracer.Sphere{Center: tracer.Point3{1, 0, -1}, Radius: 0.5, Material: tracer.Metal{Albedo: tracer.Color{0.8, 0.6, 0.2}, Fuzz: 0.9}},
	}

	cam := tracer.DefaultCamera()

	frame := tracer.NewFrame(imageWidth, imageHeight)

	tracer.Render(frame, cam, l, 20)

	if err := frame.ToPPM("image.ppm"); err != nil {
		panic(err)
	}
}
