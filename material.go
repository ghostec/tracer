package tracer

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

type Metal struct {
	Albedo Color
	Fuzz   float64
}

func (m Metal) Scatter(ray Ray, hr HitRecord) ScatterRecord {
	scatterDirection := Reflect(ray.Direction.Unit(), hr.Normal)
	scattered := Ray{Origin: hr.P, Direction: scatterDirection.Add(RandomInUnitSphere().MulFloat(m.Fuzz))}

	return ScatterRecord{
		Scatter:     true,
		Ray:         scattered,
		Attenuation: m.Albedo,
	}
}
