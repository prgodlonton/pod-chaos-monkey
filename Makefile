# Usage:
# make                           # defaults to all
# make all                       # Runs integration, unit tests, builds the Dockerfile and tags it as pod-chaos-monkey:latest
# make build                     # runs the linter and builds the Dockerfile and tags the image as pod-chaos-monkey:latest
# make build.local               # runs the linter and builds the binary locally
# make clean                     # removes the binary
# make start                     # starts the pod chaos monkey running in your Kubernetes cluster
# make start.local               # starts the pod chaos monkey running locally
# make test                      # runs unit tests
# make test.int                  # runs integration tests

.PHONY: all build clean start.local test test.int

all: test.int test build

clean:
	@rm pod-chaos-monkey

build:
	@golangci-lint run ./...
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

