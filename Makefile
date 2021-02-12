run:
	go run cmd/tracer.go -cpu-profile=trace.out && eog image.ppm
