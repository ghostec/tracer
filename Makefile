run:
	time go run cmd/tracer.go -cpu-profile=trace.out

build:
	go build cmd/tracer.go

view:
	eog image.ppm
