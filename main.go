package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
)

func main() {
	aspectRatio := 16.0 / 9.0
	imageWidth := 1000
	imageHeight := int(float64(imageWidth) / aspectRatio)
	samplesPerPixel := 100
	maxDepth := 50

	l := HitterList{
		Sphere{Center: Point3{0, 0, -1}, Radius: 0.5},
		Sphere{Center: Point3{0, -100.5, -1}, Radius: 100},
	}

	cam := DefaultCamera()

	f, err := os.OpenFile("image.ppm", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	if _, err := f.WriteString(fmt.Sprintf("P3\n%d %d\n255\n", imageWidth, imageHeight)); err != nil {
		panic(err)
	}
	for j := imageHeight; j >= 0; j-- {
		for i := 0; i < imageWidth; i++ {
			c := Color{0, 0, 0}
			for s := 0; s < samplesPerPixel; s++ {
				u := (float64(i) + rand.Float64()) / float64(imageWidth-1)
				v := (float64(j) + rand.Float64()) / float64(imageHeight-1)
				r := cam.GetRay(u, v)
				c = Color(c.Vec3().Add(RayColor(r, l, maxDepth).Vec3()))
			}
			c = Color(c.Vec3().Mul(1.0 / float64(samplesPerPixel)))
			if _, err := f.WriteString(fmt.Sprintf("%s ", c.String())); err != nil {
				panic(err)
			}
		}
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
}

func RayColor(r Ray, l HitterList, depth int) Color {
	if depth <= 0 {
		return Color{}
	}

	if hr := l.Hit(r); hr.Hit {
		target := hr.P.Vec3().Add(hr.N.Add(RandomInUnitSphere()))
		return Color(RayColor(Ray{Origin: hr.P, Direction: target.Sub(hr.P.Vec3())}, l, depth-1).Vec3().Mul(0.5))
	}

	unitDirection := r.Direction.Unit()
	t := 0.5 * (unitDirection[1] + 1.0)
	return Color(Vec3{1, 1, 1}.Mul(1.0 - t).Add(Vec3{0.5, 0.7, 1.0}.Mul(t)))
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

func (v Vec3) Mul(s float64) Vec3 {
	return Vec3{s * v[0], s * v[1], s * v[2]}
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
	return v.Mul(1 / l)
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
	return fmt.Sprintf("%d %d %d", int(255.99*Clamp(c[0], 0, 0.9999)), int(255.99*Clamp(c[1], 0, 0.9999)), int(255.99*Clamp(c[2], 0, 0.9999)))
}

type Ray struct {
	Origin    Point3
	Direction Vec3
}

func (r Ray) At(t float64) Point3 {
	return Point3(Vec3(r.Origin).Add(r.Direction.Mul(t)))
}

type Hitter interface {
	Hit(Ray) HitRecord
}

type HitRecord struct {
	Hit       bool
	FrontFace bool
	T         float64
	P         Point3
	N         Vec3
}

type Sphere struct {
	Center Point3
	Radius float64
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

	tMin, tMax := float64(0), math.Inf(+1)

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
		Hit: true,
		T:   root,
		P:   r.At(root),
	}

	outwardNormal := hr.P.Vec3().Sub(s.Center.Vec3()).Mul(1 / s.Radius)
	hr.FrontFace = r.Direction.Dot(outwardNormal) < 0
	hr.N = outwardNormal
	if !hr.FrontFace {
		hr.N = hr.N.Neg()
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
	lowerLeftCorner := origin.Vec3().Sub(horizontal.Mul(0.5)).Sub(vertical.Mul(0.5)).Sub(Vec3{0, 0, c.FocalLength})

	return Ray{
		Origin:    origin,
		Direction: lowerLeftCorner.Add(horizontal.Mul(u)).Add(vertical.Mul(v)).Sub(origin.Vec3()),
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
	return rand.Float64()*(max-min) + min
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
