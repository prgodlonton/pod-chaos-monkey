# Usage:
# make                           # defaults to all
# make all                       # compiles the CLI binary and runs integration and unit tests
# make build                     # compiles the CLI binary
# make build.local               # builds the Dockerfile and tags the image as pod-chaos-monkey:latest
# make clean                     # remove the CLI binary and build artifacts
# make server                    # compile the server binary
# make start.local               # starts the pod chaos monkey running locally
# make test                      # runs unit tests for the CLI
# make test.int                  # runs integration tests for the CLI

.PHONY: all build clean start.local test test.int

all: test.int build

clean:
	@rm pod-chaos-monkey

build:
	@docker build . -t pod-chaos-monkey:latest

build.local: $(shell find ./cli/ -type f)
	@golangci-lint run ./...
	@go build -o pod-chaos-monkey ./cli/cmd/main.go

start: build
	@kubectl apply -f ./manifests/run.yaml

start.local: build.local
	./pod-chaos-monkey workloads --local --selector app=nginx,env=dev

test:
	@go test -v --cover ./...

test.int:
	@go test -v --cover --tags=integration ./...

