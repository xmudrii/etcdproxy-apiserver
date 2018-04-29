/*
Copyright 2018 etcdproxy-proof-of-concept Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"

	"github.com/xmudrii/etcdproxy-proof-of-concept/pkg/proxy"
	"strings"
)

func main() {
	// Flags.
	bind := flag.String("bindAddress", "127.0.0.1:23790", "Bind etcd-gRPC proxy to address. "+
		"Format ip:port. Default '127.0.0.1:23790'")
	ns := flag.String("namespace", "default", "Proxy namespace. Default: 'default'")
	etcdAddr := flag.String("etcdAddresses", "http://127.0.0.1:2379", "Comma separated list of etcd endpoints. "+
		"Required format: http://ip:port,http://ip:port,... Default: 'http://127.0.0.1:2379'")
	flag.Parse()

	// Create array of etcd endpoints separated by comma.
	addr := strings.Split(*etcdAddr, ",")
	// Add leading slash to the namespace name to be more Kubernetes-like.
	*ns = "/" + *ns

	// Start GRPC server.
	s := proxy.NewGRPCServer(*bind, *ns, addr)
	s.StartNonSecureServer()
}
