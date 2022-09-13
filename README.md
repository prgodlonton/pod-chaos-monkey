# README

This CLI program is a skeleton implementation of a chaos monkey program. 
It randomly lists all pods from the namespace given that satisfy the selector provided, randomly selects one from the 
list and deletes it from the cluster.
The kubernetes API server should schedule another pod to run in response to the deletion.
This testing exposes any services that may have resiliency or performance issues when pods are evicted from nodes or
unexpectedly crash.

## Structure

```
.
├── Dockerfile
├── Makefile
├── README.md
├── cli
│   ├── cmd
│   │   └── main.go
│   └── internal
│       ├── commands
│       │   ├── disruptor.go
│       │   └── disruptor_test.go
│       └── kubernetes
│           └── pods
│               ├── client.go
│               ├── client_int_test.go
│               └── client_test.go
├── go.mod
├── go.sum
└── manifests
    ├── run.yaml
    └── setup.yaml
```

The source code for the project is all held under the `cli` directory. 
All kubernetes manifests as requested are held under `manifests`.

## Setup

The project Makefile assumes that you have `golangci-lint` installed and accessible on your system path.
It is also assumed that `minikube` is locally installed and has been started with `minikube start`.
Prior to running any build targets run `eval $(minikube docker-env)` to ensure that your local docker commands will use
the docker daemon running within the minikube cluster.


## Build

The docker image is built with `make build`. 
This image should be available with your minikube cluster.
Alternatively, use `make build.local` to build the binary locally. 

## Running

### Locally

The program can be run locally with `./pod-chaos-monkey workloads --local` or with `make start.local`.
The `--local` argument is required for running locally and instructs the CLI to read cluster data from `~/.kube` directory.
This setup is useful for debugging.
The arguments `--selector` and `--interval` allow you to specify a custom pod selector and interval between pod deletions.
Only those pods that satisfy the selector will be candidates for deletion.
The default selector is none and the interval is 10 seconds. 
For example, `./pod-chaos-monkey workloads --local --selector app=nginx,env=dev --interval 2s` will delete a pod with 
labels `app=nginx,env-dev` every 2 seconds from the workloads namespace.

### In Cluster

Run `kubectl apply -f ./manifests/run.yaml` to run the program within the cluster or with `make start`.
It will be deployed to the workloads namespace and will select only those pods created via `./manifests/setup.yaml` as 
candidates for deletion.
The `./manifests/run.yaml` will also create the necessary cluster role and bind this cluster role onto the default 
service account for the workloads namespace.

## Tests

Both unit and integration test have been provided and are run with `make test` and `make test.int` respectively. 
The build tag `integration` discerns between these two types of test. 
The integration tests assume that the manifest at `manifests/setup.yaml` has been applied cluster. 
This can be done via `kubectl apply -f ./manifests/setup.yaml` or `make setup`.

## Improvements

There is no logging in the program at the present due to time constraints; however, this should be added to get the 
program production-ready.