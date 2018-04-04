etcdproxy-proof-of-concept
==========================

This is a prototype to test how [etcd namespaces](https://github.com/coreos/etcd/blob/3239641a0c0e421769224b4e6c1dc06ce4dc3e48/Documentation/op-guide/grpc_proxy.md#namespacing) works along with [etcd-gRPC proxy](https://github.com/coreos/etcd/blob/3239641a0c0e421769224b4e6c1dc06ce4dc3e48/Documentation/op-guide/grpc_proxy.md).

## Requirements

In order to use this prototype you need to have an etcd instance.
The easiest way to run it is to use Docker:
```
docker run -d -v /usr/share/ca-certificates/:/etc/ssl/certs -p 4001:4001 -p 2380:2380 -p 2379:2379 \
 --name etcd quay.io/coreos/etcd:v2.3.8 \
 -name etcd0 \
 -advertise-client-urls http://${HostIP}:2379,http://${HostIP}:4001 \
 -listen-client-urls http://0.0.0.0:2379,http://0.0.0.0:4001 \
 -initial-advertise-peer-urls http://${HostIP}:2380 \
 -listen-peer-urls http://0.0.0.0:2380 \
 -initial-cluster-token etcd-cluster-1 \
 -initial-cluster etcd0=http://${HostIP}:2380 \
 -initial-cluster-state new
```

This command runs etcd on port 2379 in the background.

For other ways to run etcd, check out the [`Getting started` portion of `etcd` README](https://github.com/coreos/etcd#getting-started).

## Installing `etcdproxy-proof-of-concept`

In order to install this prototype you need to have [Go installed and configured](https://golang.org/doc/install).

Then, you can install `etcdproxy-proof-of-concept` by using the following `go get` command:
```
go get github.com/xmudrii/etcdproxy-proof-of-concept
```

## Running

Execute the following command to run `etcd-proxy-api`:
```
etcdproxy-proof-of-concept start
```

By default it runs proxied `etcd` instance on port `23790` and uses namespace called default `default`.
If you want to change those values, you can use `--proxy-bind-address` and `--namespace` flags. For more details check the help command (`etcdproxy-proof-of-concept start --help`).

## Writing to namespace

To write to the namespace, you need to send a `PUT` request with `name` and `value` keys to `/write` endpoint running on the port `:8000`.
```
curl -X PUT -d '{"name": "test", "value": "testval"}' http://127.0.0.1:8000/write
```
