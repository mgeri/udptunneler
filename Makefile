BINARY=udptunneler
PLATFORMS=linux windows
ARCHITECTURES=amd64
PACKAGE=github.com/mgeri/udptunneler
VERSION?=0.0.0

# setup linker flags option for build that interoperate with variable names in src code
LDFLAGS=-ldflags "-w -s -X '${PACKAGE}/pkg/version.version=${VERSION}' -X '${PACKAGE}/pkg/version.commit=${GIT_COMMIT}' -X '${PACKAGE}/pkg/version.branch=${GIT_BRANCH}'"

default: mod vet fmt build

all: clean mod test vet fmt build_all

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

mod:
	go mod tidy
	go mod vendor

build:
	go build ${LDFLAGS} -o bin/$(BINARY)

build_all:
	$(foreach GOOS, $(PLATFORMS),\
	$(foreach GOARCH, $(ARCHITECTURES), $(shell export GOOS=$(GOOS); export GOARCH=$(GOARCH); go build ${LDFLAGS} -v -o bin/$(BINARY)-$(GOOS)-$(GOARCH))))

run:
	chmod +x bin/udptunneler
	./bin/udptunneler

install:
	go install -v ./...

clean:
	rm -f bin/*

.PHONY: default all test vet fmt mod build build_all run install clean