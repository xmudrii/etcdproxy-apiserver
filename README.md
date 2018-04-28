etcdproxy-proof-of-concept [![Build Status](https://travis-ci.org/xmudrii/etcdproxy-proof-of-concept.svg?branch=master)](https://travis-ci.org/xmudrii/etcdproxy-proof-of-concept) [![GoDoc](https://godoc.org/github.com/xmudrii/etcdproxy-proof-of-concept?status.svg)](https://godoc.org/github.com/xmudrii/etcdproxy-proof-of-concept) [![Go Report Card](https://goreportcard.com/badge/github.com/xmudrii/etcdproxy-proof-of-concept)](https://goreportcard.com/report/github.com/xmudrii/etcdproxy-proof-of-concept)
==========================

The `etcdproxy-proof-of-concept` is a proof of concept made for [my GSoC project](https://summerofcode.withgoogle.com/projects/#6400208972283904) to demonstrate using [`etcd` Namespaces](https://github.com/coreos/etcd/blob/3239641a0c0e421769224b4e6c1dc06ce4dc3e48/Documentation/op-guide/grpc_proxy.md#namespacing) exposed by [`etcd-gRPC` server](https://github.com/coreos/etcd/blob/3239641a0c0e421769224b4e6c1dc06ce4dc3e48/Documentation/op-guide/grpc_proxy.md) with Kubernetes Aggregated API servers.

If you want to learn more about my GSoC project, check out the following resources:
* [GSoC project page](https://summerofcode.withgoogle.com/projects/#6400208972283904)
* [Submitted proposal (PDF)](https://github.com/xmudrii/gsoc-2018-meta-k8s/blob/master/proposal/proposal.pdf)
* [Proposal draft (public Google Document)](https://docs.google.com/document/d/10IpBTo1dnaQ9H4u9Uwek-fL-gP1om4Zte0ZSvPbPLnY/edit)
* [SIG-API-Machinery mailing list post](https://groups.google.com/d/msg/kubernetes-sig-api-machinery/rHEoQ8cgYwk/iglsNeBwCgAJ)

You can follow the project's progress by following:
* [Project's GitHub Tracker repository](https://github.com/xmudrii/gsoc-2018-meta-k8s)
* [Project's Trello](https://trello.com/b/XeaS0l5E)
* [Daily updated Google Document](https://docs.google.com/document/d/1LoqDnhb-1WV4Ja-8iS5n5Tm3NPVG50DndxsVbE17imE/edit?usp=sharing)

## Prerequisites

* [`etcd`](https://github.com/coreos/etcd) version 3.2 or newer
* [`etcdctl`](https://github.com/coreos/etcd/tree/master/etcdctl) is recommended if you want to test is the proxy working as expected
* If you want to test this project with aggregated API server, you need:
	* minimal Kubernetes cluster ([Minikube](https://kubernetes.io/docs/getting-started-guides/minikube/) or [`local-up-cluster.sh`](https://kubernetes-v1-4.github.io/docs/getting-started-guides/locally/))
	* [`sample-apiserver`](https://github.com/kubernetes/sample-apiserver)

### `etcd`

If you're developing for Kubernetes, the easiest way to install `etcd` is to use the provided [`install-etcd.sh`](https://github.com/kubernetes/kubernetes/blob/master/hack/install-etcd.sh) script. After running it, the script will output instructions needed to complete installation.

Other ways include running `etcd` using Docker or running standalone, which are covered in the [Getting Started section of `etcd`'s README](https://github.com/coreos/etcd#getting-started).

### Kubernetes

The [`sample-apiserver`](https://github.com/kubernetes/sample-apiserver) requires `kubeconfig` and a minimal Kubernetes cluster. The easiest way to run an cluster is to use the [`local-up-cluster.sh`](https://kubernetes-v1-4.github.io/docs/getting-started-guides/locally/) script located in the Kubernetes repository. You can also use Minikube.

### `sample-apiserver`

The [`sample-apiserver`](https://github.com/kubernetes/sample-apiserver) is a minimal aggregated API server and a great foundation for building your own API servers. It works with `etcd`, so we can use it to test does `etcdproxy-proof-of-concept` works as expected. To do so, you need to point API server to use the `etcdproxy-proof-of-concept` as its `etcd` server, by using the
`--etcd-servers` flag. Before running the API server, make sure the `etcdproxy` is running as well.

Instructions for running `sample-apiserver` will be added to the [project's `README.md` file](https://github.com/kubernetes/sample-apiserver#sample-apiserver). In meanwhile, you can check
out the [`kubernetes/kubernetes#55476`](https://github.com/kubernetes/kubernetes/pull/55476) PR for instructions.

## Installing `etcdproxy-proof-of-concept`

In order to install and compile the `etcdproxy-proof-of-conecpt` project, you need configured [Go environment](https://golang.org/doc/install).

Installing the project can be done using the Go's toolchain:
```
go get -u github.com/xmudrii/etcdproxy-proof-of-concept
```

You can also use [precompiled releases](https://github.com/xmudrii/etcdproxy-proof-of-concept/releases), but keep in
mind that this is a fast-moving project, so you could be missing the latest features.

## Running `etcdproxy-proof-of-concept`

Once installed, you can run the project by using the `etcdproxy-proof-of-concept` command, which runs `etcd` server on
the port `23790`, by proxying the namespace called `default`. By default it uses `etcd` server running on the port
`2379`.

The `etcdproxy-proof-of-concept` command has the following flags you can use to change default values:
* `--etcdAddress` - address of the `etcd` instance (formatted as `http://127.0.0.1:2379`)
* `--namespace` - namespace to use
* `--bindAddress` - bind `etcdproxy` to the provided address and port (formatted as `<address>:<port>`)


## Testing proxy using `sample-apiserver`

To test this project using `sample-apiserver` you need to point `sample-apiserver` to use `etcdproxy-proof-of-concept`
as its `etcd` server, by using the `--etcd-servers` flag.

The `sample-apiserver` comes with a [`Flunder` resource](https://github.com/xmudrii/etcdproxy-proof-of-concept/blob/master/artifacts/flunders/flunder.yml), which you can create to test does API server writes to `etcd` as
expected.

In the `artifacts/flunders/flunder.yml` file you can find the manifest for a simple `Flunder`. You can API server's API
to apply it, such as:
```command
http --verify no -j --cert-key client.key --cert client.crt https://localhost:8443/apis/wardle.k8s.io/v1alpha1/namespaces/default/flunders < <(python -c 'import sys, yaml, json; json.dump(yaml.load(sys.stdin), sys.stdout, indent=4)' < artifacts/flunders/flunder.yml)
```

The above command uses Python to convert the YAML manifest to the JSON formatted request, so API can parse it. The
response contains information about your request.

You can use `etcdctl` to make sure the `Flunder` is defined in `etcd`:
```
ETCDCTL_API=3 etcdctl get / --prefix --keys-only --endpoints 127.0.0.1:23790 | grep wardle
```
This should return output such as:
```
/registry/wardle.kubernetes.io/wardle.k8s.io/flunders/default/my-first-flunder
```

Now, point `etcdctl` to the main `etcd` server. This time, the key should have `/default` (or namespace's name) as a
prefix.
```
ETCDCTL_API=3 etcdctl get / --prefix --keys-only | grep wardle
```

```
/default/registry/wardle.kubernetes.io/wardle.k8s.io/flunders/default/my-first-flunder
```
