# These will be provided to the target
VERSION := 0.0.1
BUILD := `git rev-parse HEAD`

# Use linker flags to provide version/build settings to the target
LDFLAGS=-tags release -ldflags "-s -w -X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"
ARCH ?= `go env GOHOSTARCH`
GOOS ?= `go env GOOS`

all: build

test:
	@cd pkg/builder && V=$(V) go test -timeout 3s -cover

build: build_dir
	go build $(LDFLAGS) -o bin/image-builder

build_dir:
	mkdir -p bin

run: clean build
	bin/image-builder

clean:
	rm -rf bin/*