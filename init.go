package tracer

import "runtime"

var DefaultRenderer *Renderer

func init() {
	DefaultRenderer = NewRenderer(runtime.NumCPU())
	Render = DefaultRenderer.Render
}
