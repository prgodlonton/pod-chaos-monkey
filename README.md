# README

This CLI is a skeleton implementation of a chaos monkey program. It randomly selects a pod from a namespace filtered by 
the supplied pod selector and deletes it. The kubernetes API server should run another pod in response. This testing
exposes any services that may have resiliency or performance issues when pods are evicted from nodes or unexpectedly
crash.

## Setup

Install minikube locally. Start it with `minikube start` and then run `eval $(minikube docker-env)` to set up your local 
docker commands to use the docker daemon found within the minikube cluster.

## Build

Run `make build` to build the docker image. This is best run after `eval $(minikube docker-env)` to ensure that the 
docker image has been built by the docker daemon resident within the minikube virtual machine. Alternatively, use 
`make build.local` to build the binary locally. 

## Running

### Locally

The chaos monkey can be run locally using `./pod-chaos-monkey workloads --local`. The `--local` argument is required 
for running locally and instructs the program to read cluster data from the `~/.kube` directory. The arguments
`--selector` and `--interval` may  be supplied to provide custom pod selector and the interval between pod deletions
(default being 10 seconds). For example, `./pod-chaos-monkey workloads --local --selector app=nginx,env=dev --interval 2s`
will delete a pod with labels `app=nginx,env-dev` every 2 seconds from the workloads namespace.

### In Cluster

Run `kubectl apply -f ./manifests/run.yaml` to run the chaos monkey within the cluster. It will be deployed to the 
workloads namespace and will select only those pods created via `./manifests/setup.yaml` as candidates for deletion.
The `./manifests/run.yaml` will also create the necessary cluster role and bind this cluster role onto the default 
service account of the workloads namespace.

## Tests

Both unit and integration test have been provided and are run with `make test` and `make test.int` respectively. The 
build tag `integration` discerns between these two types of test. The integration tests assume that the manifest at 
`test/manifests/setup.yaml` has been applied cluster. This can be done via `kubectl apply -f ./test/manifests/setup.yaml`.
