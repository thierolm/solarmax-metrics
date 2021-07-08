EXECUTABLE=solarmax-metrics
WINDOWS=./bin/$(EXECUTABLE)_windows_amd64.exe
LINUX=./bin/$(EXECUTABLE)_linux_amd64
DARWIN=./bin/$(EXECUTABLE)_darwin_amd64
VERSION=$(shell git describe --tags --always --long --dirty)

.PHONY: all clean

all: build ## Build and run tests

build: windows linux darwin ## Build binaries
	@echo version: $(VERSION)

windows: $(WINDOWS) ## Build for Windows

linux: $(LINUX) ## Build for Linux

darwin: $(DARWIN) ## Build for Darwin (macOS)

$(WINDOWS):
	GOOS=windows GOARCH=amd64 go build -v -o $(WINDOWS) -ldflags="-s -w -X $(EXECUTABLE).version=$(VERSION)"  ./$(EXECUTABLE).go

$(LINUX):
	env GOOS=linux GOARCH=amd64 go build -v -o $(LINUX) -ldflags="-s -w -X $(EXECUTABLE).version=$(VERSION)"  ./$(EXECUTABLE).go

$(DARWIN):
	env GOOS=darwin GOARCH=amd64 go build -v -o $(DARWIN) -ldflags="-s -w -X $(EXECUTABLE).version=$(VERSION)"  ./$(EXECUTABLE).go

clean: ## Remove previous build
	rm -f $(WINDOWS) $(LINUX) $(DARWIN)

help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
