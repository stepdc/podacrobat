BUILD_NUMBER=$(shell git rev-parse --short HEAD)
BUILD_DATE=$(shell date +%FT%T%z)
BUILD_TAG=$(shell date +%Y%m%d%H%M%S)

IMAGE:=podacrobat:$(BUILD_TAG)
LATEST:=podacrobat:latest

all: build

build:
	mkdir -p make/output
	go build -o make/output/podacrobat -ldflags '-X github.com/stepdc/podacrobat/app.Version=$(BUILD_NUMBER)' github.com/stepdc/podacrobat/cmd

img: build
	cd make && docker build -f Dockerfile -t stepdc/$(IMAGE) .

img-dev: build
	cd make && docker build -f Dockerfile -t stepdc/$(LATEST) .

clean:
	rm -rf make/output
