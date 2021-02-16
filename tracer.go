package tracer

type RayColorFunc func(ray Ray, hitter Hitter, depth int, bounces int) Color
type AggColorFunc func([]Color) Color

var Transparent = Color{-1, -1, -1}

var Render func(RenderSettings, <-chan bool)

func RayColor(ray Ray, scene Hitter, depth, bounces int) Color {
	if bounces >= depth {
		return Transparent
	}

	hr := scene.Hit(ray)
	if !hr.Hit {
		unitDirection := ray.Direction.Unit()
		t := 0.5 * (unitDirection[1] + 1.0)
		return Color(Vec3{1, 1, 1}.MulFloat(1.0 - t).Add(Vec3{0.5, 0.7, 1.0}.MulFloat(t)))
	}

	sr := hr.Material.Scatter(ray, hr)
	if !sr.Scatter {
		return Transparent
	}

	return Color(sr.Attenuation.Vec3().MulVec3(RayColor(sr.Ray, scene, depth, bounces+1).Vec3()))
}

func RayBVHID(ray Ray, scene Hitter, _, _ int) Color {
	hr := scene.Hit(ray)
	if !hr.Hit {
		return Transparent
	}

	return Uint64ToColor(hr.BVHNode.ID)
}

func RayDistance(ray Ray, n Hitter, _, _ int) Color {
	hr := n.Hit(ray)
	if !hr.Hit {
		return Color{}
	}

	distVec := ray.Origin.Vec3().Sub(hr.P.Vec3())
	dist := distVec.Len()

	if dist > 100 {
		return Color{}
	}

	col := (-25.5*dist + 255.0) / 255.0
	return Color{col, col, col}
}

func ColorToUint64(color Color) uint64 {
	if color.Transparent() {
		return 0
	}
	return uint64(color[0])
}

func Uint64ToColor(val uint64) Color {
	// pow := math.Pow(float64(val), 1.0/3.0)
	// return Color{pow, pow, pow}
	return Color{float64(val), 0, 0}
}

func AvgSamples(samples []Color) Color {
	v := Vec3{}
	for _, cc := range samples {
		if cc.Transparent() {
			continue
		}
		v = v.Add(cc.Vec3())
	}
	return Color(v.MulFloat(1.0 / float64(len(samples))))
}

func EdgeSamples(samples []Color) Color {
	freq := map[uint64]int{}
	mostFreqKey := uint64(0)
	for _, cc := range samples {
		key := ColorToUint64(cc)
		freq[key] += 1
		if freq[key] > freq[mostFreqKey] {
			mostFreqKey = key
		}
	}

	if mostFreqKey == 0 {
		return Transparent
	}

	return Uint64ToColor(mostFreqKey)
}

type Ray struct {
	Origin    Point3
	Direction Vec3
}

func (r Ray) At(t float64) Point3 {
	return Point3(Vec3(r.Origin).Add(r.Direction.MulFloat(t)))
}

func ToEdgesFrame(frame *Frame, edgeColor Color) *Frame {
	edgesFrame := NewFrame(frame.Width(), frame.Height(), true)

	for row := 0; row < frame.Height(); row++ {
		for col := 0; col < frame.Width(); col++ {
			if !isEdge(frame, row, col) {
				edgesFrame.Set(row, col, Transparent)
				continue
			}
			edgesFrame.Set(row, col, edgeColor)
		}
	}

	return edgesFrame
}

func isEdge(frame *Frame, row, col int) bool {
	if frame.Get(row, col).Transparent() {
		return false
	}

	for i := -1; i < 2; i++ {
		for j := -1; j < 2; j++ {
			if i == 0 && j == 0 {
				continue
			}
			x, y := row+i, col+j
			if x < 0 || y < 0 || x >= frame.Height() || y >= frame.Width() {
				continue
			}
			if frame.Get(x, y).Transparent() {
				return true
			}
			if ColorToUint64(frame.Get(row, col)) != ColorToUint64(frame.Get(x, y)) {
				return true
			}
		}
	}
	return false
}
