package tracer

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
