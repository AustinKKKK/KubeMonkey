# KubeMonkey

A minimal [Chaos Monkey](https://netflix.github.io/chaosmonkey/) clone for Kubernetes, written in Go with [client-go](https://github.com/kubernetes/client-go). It randomly kills pods in a running cluster to verify that your workloads actually self-heal the way Kubernetes promises they will.

## Why

"The deployment has 5 replicas, so it's resilient" is an assumption until you've actually watched it recover from a failure. KubeMonkey forces that recovery to happen on a schedule, so you get to see self-healing instead of trusting it.

![KubeMonkey Architecture](./kubemonkey-architecture.svg)

## How it works

Each cycle runs three stages:

```
Lister  → fetch all pods across the cluster
Filter  → keep only pods in namespaces explicitly allowed for chaos
Killer  → pick one at random from what's left and delete it
```

### Design decision: whitelist, not blacklist

Early versions excluded known system namespaces (`kube-system`, etc.) one at a time. That approach silently breaks the moment a new system namespace shows up — which happened here with `local-path-storage`, a component `kind` installs that isn't `kube-system` but is just as unsafe to kill.

KubeMonkey instead requires namespaces to be explicitly allowed:

```go
allowedNamespaces := map[string]bool{
    "default": true,
}
```

Anything not on the list is left alone by default. Safer default, less to maintain.

## Usage

```bash
go run main.go                          # kill a random pod every 30s
go run main.go -dry-run                 # log what would be killed, don't actually delete
go run main.go -interval=10s -dry-run   # faster cycle, still dry-run
```

Flags:

| Flag | Default | Description |
|---|---|---|
| `-dry-run` | `false` | Log the target instead of deleting it |
| `-interval` | `30s` | How often a cycle runs (Go duration format: `10s`, `1m`, ...) |

## Local testing setup

```bash
# spin up a local cluster
kind create cluster --name chaos-lab

# create something for KubeMonkey to test against
kubectl create deployment nginx --image=nginx --replicas=5

# watch pods recover in real time, in a second terminal
kubectl get pods -w
```

## Example run

```
KubeMonkey Activate (dry-run=false, interval=30s)
Killed: default / nginx-69b9cdbbdd-r6w4c
Killed: default / nginx-69b9cdbbdd-8cxvf
Killed: default / nginx-69b9cdbbdd-h4xgc
```

## Stack

- Go
- [client-go](https://github.com/kubernetes/client-go) — official Kubernetes Go client
- [kind](https://kind.sigs.k8s.io/) — local cluster for development/testing
# KubeMonkey
