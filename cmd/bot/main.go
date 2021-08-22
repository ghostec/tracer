package main

import (
	"encoding/json"

	"github.com/ghostec/botnet/bot"
	"github.com/ghostec/tracer"
)

var width = 1000
var aspectRatio = 16.0 / 9.0
var parallelism = 1

func main() {
	tracer.DefaultRenderer.Start()

	b0 := bot.New("ray", "mac")

	b0.Handle("render", func(b []byte) ([]byte, error) {
		var args struct {
			LineA int `json:"a"`
			LineB int `json:"b"`
		}

		if err := json.Unmarshal(b, &args); err != nil {
			return nil, err
		}

		return execute(args.LineA, args.LineB), nil
	})

	b0.Connect("192.168.15.5", 8333)

	ch := make(chan bool)
	<-ch
}

func execute(lineA, lineB int) []byte {
	var l tracer.HitterList
	l = tracer.HitterList{
		tracer.NewSphere(tracer.Point3{0, -100.5, -1}, 100, tracer.Lambertian{Albedo: tracer.Color{0.8, 0.8, 0}}),
		tracer.NewSphere(tracer.Point3{0, 0, -1}, 0.5, tracer.Lambertian{Albedo: tracer.Color{0.1, 0.2, 0.5}}),
		tracer.NewSphere(tracer.Point3{-1, 0, -1}, 0.5, tracer.Dielectric{RefractiveIndex: 1.5}),
		tracer.NewSphere(tracer.Point3{-1, 0, -1}, -0.48, tracer.Dielectric{RefractiveIndex: 1.5}),
		tracer.NewSphere(tracer.Point3{1, 0, -1}, 0.5, tracer.Metal{Albedo: tracer.Color{0.8, 0.6, 0.2}}),
	}

	bvh, err := tracer.NewBVHNode(l)
	if err != nil {
		panic(err)
	}

	cam := &tracer.Camera{
		AspectRatio: aspectRatio,
		VFoV:        90,
		LookFrom:    tracer.Point3{-2, 2, 1},
		LookAt:      tracer.Point3{0, 0, -1},
		VUp:         tracer.Vec3{0, 1, 0},
	}

	frame := tracer.NewFrame(width, lineB-lineA, false)

	println("before render")

	r := tracer.NewRenderer(4)
	r.Start()
	r.Render(tracer.RenderSettings{
		Frame:           frame,
		Camera:          cam,
		Hitter:          bvh,
		RayColorFunc:    tracer.RayColor,
		AggColorFunc:    tracer.AvgSamples,
		SamplesPerPixel: 48,
		MaxDepth:        48,
		Lines:           int(float64(frame.Width()) / aspectRatio),
		LineA:           lineA,
		LineB:           lineB,
	}, make(chan bool, 1))

	b, err := frame.Bytes()
	if err != nil {
		panic(err)
	}

	println("after render")

	return b
}
