package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/ghostec/tracer"
)

var width = flag.Int("width", 300, "image width")
var aspectRatio = flag.Float64("aspect-ratio", 1.0, "image width")
var parallelism = flag.Int("parallelism", runtime.NumCPU(), "number of render routines to run")
var cpuProfile = flag.String("cpu-profile", "", "write cpu profile to file")

func main() {
	flag.Parse()

	gob.Register(tracer.Sphere{})
	gob.Register(tracer.Lambertian{})
	gob.Register(tracer.Metal{})
	gob.Register(tracer.Dielectric{})

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	tracer.DefaultRenderer.Start()

	// generateScenes(10, 0)

	// return

	scenes, err := loadScenes()
	if err != nil {
		panic(err)
	}

	for i, scene := range scenes {
		render(fmt.Sprintf("frames/%d.png", i), scene)
	}
}

type Scene struct {
	Camera     tracer.Camera
	HitterList tracer.HitterList
}

func randSign() float64 {
	if rand.Float64() < 0.5 {
		return -1.0
	}

	return 1.0
}

func randCenter() tracer.Point3 {
	return tracer.Point3{randSign() * rand.Float64() * 5, randSign() * rand.Float64() * 20, randSign() * rand.Float64() * 5}
}

func randMaterial() tracer.Material {
	m := []tracer.Material{
		tracer.Lambertian{Albedo: tracer.Color{0.8, 0.8, 0}},
		tracer.Lambertian{Albedo: tracer.Color{0.1, 0.2, 0.5}},
		tracer.Dielectric{RefractiveIndex: 1.5},
		tracer.Metal{Albedo: tracer.Color{0.8, 0.6, 0.2}},
	}

	return m[rand.Intn(len(m))]
}

func generateScenes(amount, startIdx int) {
	for i := startIdx; i < startIdx+amount; i++ {
		var l tracer.HitterList
		n := rand.Intn(20) + 1
		for j := 0; j < n; j++ {
			l = append(l, tracer.NewSphere(randCenter(), rand.Float64()*10, randMaterial()))
		}

		cam := tracer.Camera{
			AspectRatio: *aspectRatio,
			VFoV:        90,
			LookFrom:    tracer.Point3{-2, 2, 1},
			LookAt:      tracer.Point3{0, 0, -1},
			VUp:         tracer.Vec3{0, 1, 0},
		}

		var buf bytes.Buffer

		enc := gob.NewEncoder(&buf)

		if err := enc.Encode(&Scene{
			Camera:     cam,
			HitterList: l,
		}); err != nil {
			panic(err)
		}

		if err := os.WriteFile(fmt.Sprintf("scenes/%d.gob", i), buf.Bytes(), 0755); err != nil {
			panic(err)
		}
	}
}

func loadScenes() (scenes []Scene, err error) {
	if err := filepath.Walk("scenes", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".gob") {
			return nil
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var scene Scene

		dec := gob.NewDecoder(bytes.NewReader(b))
		if err := dec.Decode(&scene); err != nil {
			return err
		}

		scenes = append(scenes, scene)

		return nil
	}); err != nil {
		return nil, err
	}

	return scenes, nil
}

func render(dst string, scene Scene) {
	imageWidth := *width
	imageHeight := int(float64(imageWidth) / *aspectRatio)

	frame := tracer.NewFrame(imageWidth, imageHeight, false)

	bvh, err := tracer.NewBVHNode(scene.HitterList)
	if err != nil {
		panic(err)
	}

	tracer.Render(tracer.RenderSettings{
		Frame:           frame,
		Camera:          &scene.Camera,
		Hitter:          bvh,
		RayColorFunc:    tracer.RayColor,
		AggColorFunc:    tracer.AvgSamples,
		SamplesPerPixel: 512,
		MaxDepth:        20,
	}, make(chan bool, 1))

	if err := frame.Save(dst); err != nil {
		panic(err)
	}
}
