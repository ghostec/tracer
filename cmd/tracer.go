package main

import (
	"errors"
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
		tracer.DefaultRenderer.Start()
		execute()
		defer pprof.StopCPUProfile()
	}
}

func execute() {
	imageWidth := *width
	imageHeight := int(float64(imageWidth) / *aspectRatio)

	// l := tracer.HitterList{
	// 	tracer.Sphere{Center: tracer.Point3{0, 0, -1}, Radius: 0.5, Material: tracer.Lambertian{Albedo: tracer.Color{0.7, 0.3, 0.3}}},
	// 	tracer.Sphere{Center: tracer.Point3{0, -100.5, -1}, Radius: 100, Material: tracer.Lambertian{Albedo: tracer.Color{0.8, 0.8, 0}}},
	// 	tracer.Sphere{Center: tracer.Point3{-1, 0, -1}, Radius: 0.5, Material: tracer.Dielectric{RefractiveIndex: 1.5}},
	// 	tracer.Sphere{Center: tracer.Point3{-1, 0, -1}, Radius: -0.48, Material: tracer.Dielectric{RefractiveIndex: 1.5}},
	// 	tracer.Sphere{Center: tracer.Point3{1, 0, -1}, Radius: 0.5, Material: tracer.Metal{Albedo: tracer.Color{0.8, 0.6, 0.2}, Fuzz: 0.9}},
	// }

	// var l tracer.HitterList
	// {
	// 	R := math.Cos(math.Pi / 4)
	// 	l = tracer.HitterList{
	// 		tracer.Sphere{Center: tracer.Point3{-R, 0, -1}, Radius: R, Material: tracer.Lambertian{Albedo: tracer.Color{0, 0, 1}}},
	// 		tracer.Sphere{Center: tracer.Point3{R, 0, -1}, Radius: R, Material: tracer.Lambertian{Albedo: tracer.Color{1, 0, 0}}},
	// 	}
	// }

	var l tracer.HitterList
	{
		l = tracer.HitterList{
			tracer.Sphere{Center: tracer.Point3{0, -100.5, -1}, Radius: 100, Material: tracer.Lambertian{Albedo: tracer.Color{0.8, 0.8, 0}}},
			tracer.Sphere{Center: tracer.Point3{0, 0, -1}, Radius: 0.5, Material: tracer.Lambertian{Albedo: tracer.Color{0.1, 0.2, 0.5}}},
			tracer.Sphere{Center: tracer.Point3{-1, 0, -1}, Radius: 0.5, Material: tracer.Dielectric{RefractiveIndex: 1.5}},
			tracer.Sphere{Center: tracer.Point3{-1, 0, -1}, Radius: -0.48, Material: tracer.Dielectric{RefractiveIndex: 1.5}},
			tracer.Sphere{Center: tracer.Point3{1, 0, -1}, Radius: 0.5, Material: tracer.Metal{Albedo: tracer.Color{0.8, 0.6, 0.2}}},
		}
	}

	cam := tracer.Camera{
		AspectRatio: 16.0 / 9.0,
		VFoV:        90,
		LookFrom:    tracer.Point3{-2, 2, 1},
		LookAt:      tracer.Point3{0, 0, -1},
		VUp:         tracer.Vec3{0, 1, 0},
	}

	frame := tracer.NewFrame(imageWidth, imageHeight, false)

	bvh, err := tracer.NewBVHNode(l)
	if err != nil {
		panic(err)
	}

	tracer.Render(tracer.RenderSettings{
		Frame:           frame,
		Camera:          cam,
		Hitter:          bvh,
		RayColorFunc:    tracer.RayColor,
		AggColorFunc:    tracer.AvgSamples,
		SamplesPerPixel: 200,
		MaxDepth:        50,
	}, make(chan bool, 1))

	hr := bvh.Hit(cam.GetRay(tracer.CameraCoordinatesFromPixel(200, 400, imageWidth, imageHeight)))
	if !hr.Hit {
		panic(errors.New("no hit"))
	}

	otherBVH, err := tracer.NewBVHNode(tracer.HitterList{hr.BVHNode.Left})

	otherFrame := tracer.NewFrame(imageWidth, imageHeight, true)
	tracer.Render(tracer.RenderSettings{
		Frame:           otherFrame,
		Camera:          cam,
		Hitter:          otherBVH,
		RayColorFunc:    tracer.RayBVHID,
		AggColorFunc:    tracer.EdgeSamples,
		SamplesPerPixel: 10,
	}, make(chan bool, 1))
	edgesFrame := tracer.ToEdgesFrame(otherFrame, tracer.Color{255, 0, 0})

	// for j := -1; j < 5; j++ {
	// 	for i := -1; i < 5; i++ {
	// 		otherFrame.Set(200-j, 400-i, tracer.Color{1.0, 0, 0})
	// 	}
	// }

	edgesFrame.Blend(frame, 1.0, 1.0)
	frame = edgesFrame

	// finalFrame := tracer.NewFrame(imageWidth, imageHeight)

	// frame = finalFrame

	if err := frame.Save("blend.png"); err != nil {
		panic(err)
	}
}
