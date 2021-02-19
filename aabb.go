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

func (a AABB) Hit(ray Ray) HitRecord {
	tMin, tMax := 0.0, math.Inf(+1)

	for axis := 0; axis < 3; axis++ {
		t0 := Min(
			(a.Min[axis]-ray.Origin[axis])/ray.Direction[axis],
			(a.Max[axis]-ray.Origin[axis])/ray.Direction[axis],
		)
		t1 := Max(
			(a.Min[axis]-ray.Origin[axis])/ray.Direction[axis],
			(a.Max[axis]-ray.Origin[axis])/ray.Direction[axis],
		)

		tMin = Max(t0, tMin)
		tMax = Min(t1, tMax)

		if tMax <= tMin {
			return HitRecord{}
		}
	}
	return HitRecord{Hit: true}
}

func (a AABB) Zero() bool {
	return a.Min.Vec3().Zero() && a.Max.Vec3().Zero()
}

func (a AABB) Surrounding(b AABB) AABB {
	small := Point3{
		Min(a.Min.Vec3()[0], b.Min.Vec3()[0]),
		Min(a.Min.Vec3()[1], b.Min.Vec3()[1]),
		Min(a.Min.Vec3()[2], b.Min.Vec3()[2]),
	}
	big := Point3{
		Max(a.Max.Vec3()[0], b.Max.Vec3()[0]),
		Max(a.Max.Vec3()[1], b.Max.Vec3()[1]),
		Max(a.Max.Vec3()[2], b.Max.Vec3()[2]),
	}
	return AABB{small, big}
}

func (a AABB) Compare(b AABB, axis int) bool {
	if a.Zero() || b.Zero() {
		panic(errors.New("HAHAHA"))
	}

	return a.Min.Vec3()[axis] < b.Min.Vec3()[axis]
}
