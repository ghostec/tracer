package tracer

import (
	"errors"
	"math"
	"sort"

	"lukechampine.com/frand"
)

type Hitter interface {
	Hit(Ray) HitRecord
	BoundingBox() AABB
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

func (s Sphere) BoundingBox() AABB {
	return AABB{
		Point3(s.Center.Vec3().Sub(Vec3{s.Radius, s.Radius, s.Radius})),
		Point3(s.Center.Vec3().Add(Vec3{s.Radius, s.Radius, s.Radius})),
	}
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

func (h HitterList) BoundingBox() AABB {
	if len(h) == 0 {
		return AABB{}
	}

	var outputBox AABB
	firstBox := true
	for _, hitter := range h {
		bb := hitter.BoundingBox()
		if bb.Zero() {
			return AABB{}
		}
		outputBox = bb
		if !firstBox {
			outputBox.Surrounding(bb)
		}
		firstBox = false
	}

	return outputBox
}

type BVHNode struct {
	Box         AABB
	Left, Right Hitter
}

func NewBVHNode(l HitterList) BVHNode {
	axis := frand.Intn(3)

	node := BVHNode{}

	switch len(l) {
	case 1:
		node.Left, node.Right = l[0], l[0]
	case 2:
		if l[0].BoundingBox().Compare(l[1].BoundingBox(), axis) {
			node.Left, node.Right = l[0], l[1]
		} else {
			node.Left, node.Right = l[1], l[0]
		}
	default:
		sort.Slice(l, func(i, j int) bool {
			return l[i].BoundingBox().Compare(l[j].BoundingBox(), axis)
		})

		mid := len(l) / 2
		node.Left = NewBVHNode(l[:mid])
		node.Right = NewBVHNode(l[mid:])
	}

	if node.Left.BoundingBox().Zero() || node.Right.BoundingBox().Zero() {
		panic(errors.New("hahaha"))
	}

	node.Box = node.Left.BoundingBox().Surrounding(node.Right.BoundingBox())

	return node
}

func (n BVHNode) BoundingBox() AABB {
	return n.Box
}

func (n BVHNode) Hit(ray Ray) HitRecord {
	if hr := n.Box.Hit(ray); !hr.Hit {
		return HitRecord{}
	}

	hrLeft := n.Left.Hit(ray)
	hrRight := n.Right.Hit(ray)

	if hrLeft.Hit && hrRight.Hit {
		if hrLeft.T < hrRight.T {
			return hrLeft
		}
		return hrRight
	}
	if hrLeft.Hit {
		return hrLeft
	}
	return hrRight
}
