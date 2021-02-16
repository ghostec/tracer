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
}

func (c Camera) GetRay(s, t float64) Ray {
	theta := DegreesToRadians(c.VFoV)
	h := math.Tan(theta / 2.0)
	viewportHeight := 2.0 * h
	viewportWidth := c.AspectRatio * viewportHeight

	w := c.LookFrom.Vec3().Sub(c.LookAt.Vec3()).Unit()
	u := c.VUp.Cross(w).Unit()
	v := w.Cross(u)

	origin := c.LookFrom
	horizontal := u.MulFloat(viewportWidth)
	vertical := v.MulFloat(viewportHeight)
	lowerLeftCorner := origin.Vec3().Sub(horizontal.MulFloat(0.5)).Sub(vertical.MulFloat(0.5)).Sub(w)

	return Ray{
		Origin:    origin,
		Direction: lowerLeftCorner.Add(horizontal.MulFloat(s)).Add(vertical.MulFloat(t)).Sub(origin.Vec3()).Unit(),
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
