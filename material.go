package tracer

import (
	"math"

	"lukechampine.com/frand"
)

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

	return ScatterRecord{
		Scatter:     true,
		Ray:         Ray{Origin: hr.P, Direction: scatterDirection},
		Attenuation: l.Albedo,
	}
}

type Metal struct {
	Albedo Color
	Fuzz   float64
}

func (m Metal) Scatter(ray Ray, hr HitRecord) ScatterRecord {
	scatterDirection := reflect(ray.Direction.Unit(), hr.Normal).Add(RandomInUnitSphere().MulFloat(m.Fuzz))

	return ScatterRecord{
		Scatter:     true,
		Ray:         Ray{Origin: hr.P, Direction: scatterDirection},
		Attenuation: m.Albedo,
	}
}

func reflect(vector, normal Vec3) Vec3 {
	return vector.Sub(normal.MulFloat(2 * vector.Dot(normal)))
}

type Dielectric struct {
	RefractiveIndex float64
}

func (d Dielectric) Scatter(ray Ray, hr HitRecord) ScatterRecord {
	refractionRatio := d.RefractiveIndex
	if hr.FrontFace {
		refractionRatio = 1.0 / refractionRatio
	}

	unitDirection := ray.Direction.Unit()
	cosTheta := math.Min(unitDirection.Neg().Dot(hr.Normal), 1.0)
	sinTheta := math.Sqrt(1.0 - cosTheta*cosTheta)

	cannotRefract := refractionRatio*sinTheta > 1.0

	var scatterDirection Vec3
	switch {
	case cannotRefract || d.reflectance(cosTheta, refractionRatio) > frand.Float64():
		scatterDirection = reflect(unitDirection, hr.Normal)
	default:
		scatterDirection = refract(ray.Direction.Unit(), hr.Normal, refractionRatio)
	}

	return ScatterRecord{
		Scatter:     true,
		Ray:         Ray{Origin: hr.P, Direction: scatterDirection},
		Attenuation: Color{1, 1, 1},
	}
}

func (d Dielectric) reflectance(cosine, refIdx float64) float64 {
	r0 := (1.0 - refIdx) / (1.0 + refIdx)
	r0 = r0 * r0
	return r0 + (1-r0)*math.Pow(1.0-cosine, 5)
}

func refract(vector, normal Vec3, refractionRatio float64) Vec3 {
	cosTheta := math.Min(vector.Neg().Dot(normal), 1.0)
	rOutPerp := vector.Add(normal.MulFloat(cosTheta)).MulFloat(refractionRatio)
	rOutParallel := normal.MulFloat(-math.Sqrt(math.Abs(1.0 - rOutPerp.LenSq())))
	return rOutPerp.Add(rOutParallel)
}
