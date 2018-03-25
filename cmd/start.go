package cmd

import (
	"github.com/spf13/cobra"
	"github.com/xmudrii/etcdproxy-proof-of-concept/grpcproxy"
	"github.com/xmudrii/etcdproxy-proof-of-concept/server"
)

// apiServerCmd starts the etcd-proxy-api server.
var (
	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start etcd-gPRC Proxy API server",
		Long:  `Set up an etcd-gRPC proxy to a namespace to access your data.`,
		Run: RunApiServer,
	}

	etcdProxyBindAddress string
	etcdNamespace string
)

func init() {
	startCmd.Flags().StringVarP(&etcdProxyBindAddress, "proxy-bind-address", "a", "127.0.0.1:23790", "Start etcd-gRPC proxy on this address.")
	startCmd.Flags().StringVarP(&etcdNamespace, "namespace", "n", "default", "Namespace to proxy to.")
}

func RunApiServer(cmd *cobra.Command, args []string) {
	// this is a bad idea but prototype idea is important.
	go grpcproxy.StartUnsecureGRPCProxy(etcdProxyBindAddress, etcdNamespace)
	go server.RunServer()

	for {}
}
