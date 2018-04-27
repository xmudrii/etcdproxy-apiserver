package main

import (
	"flag"
	"github.com/xmudrii/etcdproxy-proof-of-concept/pkg/proxy"
)

func main() {

	bind := flag.String("bindAddress", "127.0.0.1:23790", "Bind etcd-gRPC proxy to address. "+
		"Format ip:port. Default '127.0.0.1:23790'")
	ns := flag.String("namespace", "default", "Proxy namespace. Default: 'default'")
	etcdAddr := flag.String("etcdAddress", "http://127.0.0.1:2379", "Etcd address. "+
		"Required format: http://<ip>:port. Default: 'http://127.0.0.1:2379'")
	flag.Parse()

	s := proxy.NewGRPCServer(*bind, *ns, *etcdAddr)
	go s.StartNonSecureServer()

	for {
	}
}
