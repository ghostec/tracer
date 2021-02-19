package tracer

import (
	"errors"
	"math"
	"sort"
	"sync/atomic"

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
	BVHNode   *BVHNode
}

type Sphere struct {
	Center   Point3
	Radius   float64
	Material Material
	Box      AABB
}

func NewSphere(center Point3, radius float64, material Material) *Sphere {
	box := AABB{
		Point3(Vec3(center).Sub(Vec3{radius, radius, radius})),
		Point3(Vec3(center).Add(Vec3{radius, radius, radius})),
	}
	return &Sphere{
		Center:   center,
		Radius:   radius,
		Material: material,
		Box:      box,
	}
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

	outwardNormal := Vec3(hr.P).Sub(Vec3(s.Center)).MulFloat(1 / s.Radius)
	hr.FrontFace = r.Direction.Dot(outwardNormal) < 0
	hr.Normal = outwardNormal
	if !hr.FrontFace {
		hr.Normal = hr.Normal.Neg()
	}

	return hr
}

func (s Sphere) BoundingBox() AABB {
	return s.Box
}

type HitterList []Hitter

func (h HitterList) Hit(r Ray) (hr HitRecord) {
	hr.T = math.Inf(+1)
	for i := range h {
		hhr := h[i].Hit(r)
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
	for i := range h {
		bb := h[i].BoundingBox()
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
	ID          uint64
	Box         AABB
	Left, Right Hitter
}

// 0 is reserved
var bvhCounter = uint64(0)

func NewBVHNode(l HitterList) (*BVHNode, error) {
	if len(l) == 0 {
		return nil, errors.New("empty list")
	}

	axis := frand.Intn(3)
	node := &BVHNode{ID: atomic.AddUint64(&bvhCounter, 1)}

	switch len(l) {
	case 1:
		node.Left, node.Right = l[0], l[0]
	case 2:
		var err error
		if l[0].BoundingBox().Compare(l[1].BoundingBox(), axis) {
			node.Left, err = NewBVHNode(l[:1])
			if err != nil {
				return node, err
			}
			node.Right, err = NewBVHNode(l[1:])
			if err != nil {
				return node, err
			}
		} else {
			node.Left, err = NewBVHNode(l[1:])
			if err != nil {
				return node, err
			}
			node.Right, err = NewBVHNode(l[:1])
			if err != nil {
				return node, err
			}
		}
	default:
		sort.Slice(l, func(i, j int) bool {
			return l[i].BoundingBox().Compare(l[j].BoundingBox(), axis)
		})

		var err error
		mid := len(l) / 2
		node.Left, err = NewBVHNode(l[:mid])
		if err != nil {
			return node, err
		}
		node.Right, err = NewBVHNode(l[mid:])
		if err != nil {
			return node, err
		}
	}

	if node.Left.BoundingBox().Zero() || node.Right.BoundingBox().Zero() {
		panic(errors.New("hahaha"))
	}

	node.Box = node.Left.BoundingBox().Surrounding(node.Right.BoundingBox())

	return node, nil
}

func (n *BVHNode) BoundingBox() AABB {
	return n.Box
}

func (n *BVHNode) Hit(ray Ray) HitRecord {
	if !n.Box.Hit(ray) {
		return HitRecord{}
	}

	hrLeft := n.Left.Hit(ray)
	hrRight := n.Right.Hit(ray)

	if _, ok := n.Left.(*BVHNode); !ok {
		hrLeft.BVHNode = n
	}
	if _, ok := n.Right.(*BVHNode); !ok {
		hrRight.BVHNode = n
	}

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

type Plane struct {
	Origin Point3
	// Unit
	Normal Vec3
	// Width, Height float64
	Axis [2]Vec3
}

func NewPlane(origin Point3, normal Vec3) Plane {
	normal = normal.Unit()

	var tangent0 Vec3
	switch {
	case normal[0] != 0:
		tangent0 = Vec3{-normal[1] / normal[0], 1, 0}
	case normal[1] != 0:
		tangent0 = Vec3{1, -normal[0] / normal[1], 0}
	case normal[2] != 0:
		tangent0 = Vec3{1, 0, -normal[0] / normal[2]}
	}
	tangent0 = tangent0.Unit()
	tangent1 := normal.Cross(tangent0).Unit()

	return Plane{
		Origin: origin,
		Axis:   [2]Vec3{tangent0, tangent1},
	}
}

func (p Plane) BoundingBox() AABB {
	return AABB{}
}

func (p Plane) Hit(ray Ray) HitRecord {
	return HitRecord{}
}

func (p Plane) Project(point Point3) Point3 {
	return Point3{}
	// v := point.Vec3().Sub(p.Origin.Vec3())
	// dist := v.Dot(p.Normal)
	// return Point3(v.Add(p.Normal.MulFloat(dist)))
}
