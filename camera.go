package tracer

import (
	"math"

	"lukechampine.com/frand"
)

type Camera struct {
	AspectRatio float64
	VFoV        float64 // vertical field-of-view in degrees
	LookFrom    Point3
	LookAt      Point3
	VUp         Vec3

	lowerLeftCorner      Vec3
	horizontal, vertical Vec3
	clean                bool
}

func (c *Camera) GetRay(s, t float64) Ray {
	if !c.clean {
		theta := DegreesToRadians(c.VFoV)
		h := math.Tan(theta / 2.0)
		viewportHeight := 2.0 * h
		viewportWidth := c.AspectRatio * viewportHeight

		w := Vec3(c.LookFrom).Sub(Vec3(c.LookAt)).Unit()
		u := c.VUp.Cross(w).Unit()
		v := w.Cross(u)

		origin := c.LookFrom
		c.horizontal = u.MulFloat(viewportWidth)
		c.vertical = v.MulFloat(viewportHeight)
		c.lowerLeftCorner = Vec3(origin).Sub(c.horizontal.MulFloat(0.5)).Sub(c.vertical.MulFloat(0.5)).Sub(w)
		c.clean = true
	}

	return Ray{
		Origin:    c.LookFrom,
		Direction: c.lowerLeftCorner.Add(c.horizontal.MulFloat(s)).Add(c.vertical.MulFloat(t)).Sub(Vec3(c.LookFrom)).Unit(),
	}
}

func CameraCoordinatesFromPixel(row, col, frameWidth, frameHeight int) (float64, float64) {
	u := float64(col) / float64(frameWidth-1)
	v := float64(row) / float64(frameHeight-1)
	return u, v
}

func JitteredCameraCoordinatesFromPixel(row, col, frameWidth, frameHeight int) (float64, float64) {
	u := (float64(col) + frand.Float64()) / float64(frameWidth-1)
	v := (float64(row) + frand.Float64()) / float64(frameHeight-1)
	return u, v
}
