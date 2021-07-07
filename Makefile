.PHONY: default build run compile

build:
	go build -o bin/solarmax-metrics solarmax-metrics.go

run:
	go run solarmax-metrics.go -h

default:
	build

compile:
	echo "Compiling for every OS and Platform"
	GOOS=windows GOARCH=386 go build -o bin/solarmax-metrics-win.exe solarmax-metrics.go
	GOOS=linux GOARCH=amd64 go build -o bin/solarmax-metrics-linux-amd64 solarmax-metrics.go
	GOOS=linux GOARCH=arm go build -o bin/solarmax-metrics-linux-arm solarmax-metrics.go
	GOOS=linux GOARCH=arm64 go build -o bin/solarmax-metrics-linux-arm64 solarmax-metrics.go
	GOOS=freebsd GOARCH=386 go build -o bin/solarmax-metrics-freebsd-386 solarmax-metrics.go	