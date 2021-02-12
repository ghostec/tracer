package tracer

import (
	"math"
)

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
