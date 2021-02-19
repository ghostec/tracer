package tracer

import (
	"math"

	"lukechampine.com/frand"
)

type Vec3 [3]float64

func (v Vec3) Add(o Vec3) Vec3 {
	return Vec3{v[0] + o[0], v[1] + o[1], v[2] + o[2]}
}

func (v Vec3) Sub(o Vec3) Vec3 {
	return Vec3{v[0] - o[0], v[1] - o[1], v[2] - o[2]}
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

func (v Vec3) MulFloat(s float64) Vec3 {
	return Vec3{s * v[0], s * v[1], s * v[2]}
}

func (v Vec3) MulVec3(o Vec3) Vec3 {
	return Vec3{v[0] * o[0], v[1] * o[1], v[2] * o[2]}
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
	return v.MulFloat(1 / l)
}

func (v Vec3) Zero() bool {
	return v[0] == 0.0 && v[1] == 0.0 && v[2] == 0.0
}

func (v Vec3) NearZero() bool {
	s := 1e-8
	return (math.Abs(v[0]) < s) && (math.Abs(v[1])) < s && (math.Abs(v[2]) < s)
}

type Color Vec3
type Point3 Vec3

func (c Color) RGBA() [4]uint8 {
	if c.Transparent() {
		return [4]uint8{0, 0, 0, 0}
	}
	ret := [4]uint8{0, 0, 0, 255}
	for i, cc := range c {
		cc = math.Sqrt(cc)
		cc = 256 * Clamp(cc, 0, 0.999)
		ret[i] = uint8(cc)
	}
	return ret
}

func (c Color) Transparent() bool {
	return c == [3]float64{-1, -1, -1}
}

func (c Color) Blend(o Color, cA, oA float64) Color {
	switch {
	case c.Transparent():
		return o
	case o.Transparent():
		return c
	}

	alpha := cA + oA*(1-cA)
	if alpha == 0.0 {
		return Color{}
	}
	vec := Vec3(c).MulFloat(cA).Add(Vec3(o).MulFloat(oA).MulFloat(1.0 - cA)).MulFloat(1.0 / alpha)
	return Color(vec.MulFloat(alpha))
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
	return frand.Float64()*(max-min) + min
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

func RandomUnitVector() Vec3 {
	return RandomInUnitSphere().Unit()
}

func RandomInHemisphere(normal Vec3) Vec3 {
	inUnitSphere := RandomInUnitSphere()
	if inUnitSphere.Dot(normal) > 0.0 {
		return inUnitSphere
	}
	return inUnitSphere.Neg()
}

func DegreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

func ClosestVertex(box AABB, target Point3) Point3 {
	var closest Point3
	closestDistance := math.Inf(+1)
	for _, p := range []Point3{
		box.Min,
		box.Max,
		{box.Min[0], box.Min[1], box.Max[2]},
		{box.Min[0], box.Max[1], box.Min[2]},
		{box.Max[0], box.Min[1], box.Min[2]},
		{box.Min[0], box.Max[1], box.Max[2]},
		{box.Max[0], box.Min[1], box.Max[2]},
		{box.Max[0], box.Max[1], box.Min[2]},
	} {
		dist := Vec3(target).Sub(Vec3(p)).Len()
		if dist < closestDistance {
			closestDistance = dist
			closest = p
		}
	}
	return closest
}
