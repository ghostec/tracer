package tracer

import (
	"errors"
	"math"
)

type AABB struct {
	Min, Max Point3
}

func Min(x, y float64) float64 {
	if x < y {
		return x
	}
	return y
}

func Max(x, y float64) float64 {
	if x > y {
		return x
	}
	return y
}

func (a AABB) Hit(ray Ray) bool {
	tMin, tMax := 0.0, math.Inf(+1)

	for axis := 0; axis < 3; axis++ {
		invRayDA := 1.0 / ray.Direction[axis]
		t0 := Min(
			(a.Min[axis]-ray.Origin[axis])*invRayDA,
			(a.Max[axis]-ray.Origin[axis])*invRayDA,
		)
		t1 := Max(
			(a.Min[axis]-ray.Origin[axis])*invRayDA,
			(a.Max[axis]-ray.Origin[axis])*invRayDA,
		)

		tMin = Max(t0, tMin)
		tMax = Min(t1, tMax)

		if tMax <= tMin {
			return false
		}
	}

	return true
}

func (a AABB) Zero() bool {
	return Vec3(a.Min).Zero() && Vec3(a.Max).Zero()
}

func (a AABB) Surrounding(b AABB) AABB {
	small := Point3{
		Min(a.Min[0], b.Min[0]),
		Min(a.Min[1], b.Min[1]),
		Min(a.Min[2], b.Min[2]),
	}
	big := Point3{
		Max(a.Max[0], b.Max[0]),
		Max(a.Max[1], b.Max[1]),
		Max(a.Max[2], b.Max[2]),
	}
	return AABB{small, big}
}

func (a AABB) Compare(b AABB, axis int) bool {
	if a.Zero() || b.Zero() {
		panic(errors.New("HAHAHA"))
	}

	return a.Min[axis] < b.Min[axis]
}
