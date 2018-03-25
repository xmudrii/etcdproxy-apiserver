package cmd

import (
	"github.com/spf13/cobra"
	"github.com/xmudrii/etcd-proxy-api/grpcproxy"
	"github.com/xmudrii/etcd-proxy-api/server"
)

// apiServerCmd starts the etcd-proxy-api server.
var (
	apiServerCmd = &cobra.Command{
		Use:   "apiserver",
		Short: "Start etcd-gPRC Proxy API server",
		Long:  `Set up an etcd-gRPC proxy to a namespace to access your data.`,
		Run: RunApiServer,
	}

	etcdProxyBindAddress string
	etcdNamespace string
)

func init() {
	apiServerCmd.Flags().StringVarP(&etcdProxyBindAddress, "proxy-bind-address", "a", "127.0.0.1:23790", "Start etcd-gRPC proxy on this address. Do NOT include http://!")
	apiServerCmd.Flags().StringVarP(&etcdNamespace, "namespace", "n", "default", "Namespace to proxy to.")
}

func RunApiServer(cmd *cobra.Command, args []string) {
	// this is a bad idea but prototype idea is important.
	go grpcproxy.StartUnsecureGRPCProxy(etcdProxyBindAddress, etcdNamespace)
	go server.RunServer()

	for {}
}
