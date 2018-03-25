package grpcproxy

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"os"
	"time"

	pb "github.com/coreos/etcd/etcdserver/etcdserverpb"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
	"github.com/coreos/etcd/etcdserver/api/v3election/v3electionpb"
	"github.com/coreos/etcd/etcdserver/api/v3lock/v3lockpb"
	"github.com/coreos/etcd/pkg/srv"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/coreos/etcd/proxy/grpcproxy"
	"github.com/coreos/pkg/capnslog"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

var (
	grpcProxyAddress   string
	grpcProxyNamespace string
)

// StartUnsercureGRPCProxy start unsercure etcd-gRPC proxy.
func StartUnsecureGRPCProxy(address string, namespace string) {
	grpcProxyAddress = address
	if namespace[len(namespace)-1:] == "/" {
		grpcProxyNamespace = namespace
	} else {
		grpcProxyNamespace = namespace + "/"
	}

	// gRPC logging.
	capnslog.SetGlobalLogLevel(capnslog.DEBUG)
	grpc.EnableTracing = true
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(os.Stderr, os.Stderr, os.Stderr))

	m := mustListenCMux(nil)
	grpcl := m.Match(cmux.HTTP2())

	client := mustNewClient()
	srvhttp, httpl := mustHTTPListener(m, nil, client)
	errc := make(chan error)
	go func() { errc <- newGRPCProxyServer(client).Serve(grpcl) }()
	go func() { errc <- srvhttp.Serve(httpl) }()
	go func() { errc <- m.Serve() }()

	fmt.Fprintln(os.Stderr, <-errc)
}

func mustNewClient() *clientv3.Client {
	srvs := discoverEndpoints("", "", false)
	eps := srvs.Endpoints
	if len(eps) == 0 {
		eps = []string{"http://127.0.0.1:2379"}
	}

	cfg, err := newClientCfg(eps)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	cfg.DialOptions = append(cfg.DialOptions,
		grpc.WithUnaryInterceptor(grpcproxy.AuthUnaryClientInterceptor))
	cfg.DialOptions = append(cfg.DialOptions,
		grpc.WithStreamInterceptor(grpcproxy.AuthStreamClientInterceptor))
	client, err := clientv3.New(*cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return client
}

func mustHTTPListener(m cmux.CMux, tlsinfo *transport.TLSInfo, c *clientv3.Client) (*http.Server, net.Listener) {
	httpmux := http.NewServeMux()
	httpmux.HandleFunc("/", http.NotFound)
	srvhttp := &http.Server{Handler: httpmux}
	return srvhttp, m.Match(cmux.HTTP1())
}

func newGRPCProxyServer(client *clientv3.Client) *grpc.Server {
	if len(grpcProxyNamespace) > 0 {
		client.KV = namespace.NewKV(client.KV, grpcProxyNamespace)
		client.Watcher = namespace.NewWatcher(client.Watcher, grpcProxyNamespace)
		client.Lease = namespace.NewLease(client.Lease, grpcProxyNamespace)
	}

	kvp, _ := grpcproxy.NewKvProxy(client)
	watchp, _ := grpcproxy.NewWatchProxy(client)
	clusterp, _ := grpcproxy.NewClusterProxy(client, grpcProxyAddress, "")
	leasep, _ := grpcproxy.NewLeaseProxy(client)
	mainp := grpcproxy.NewMaintenanceProxy(client)
	authp := grpcproxy.NewAuthProxy(client)
	electionp := grpcproxy.NewElectionProxy(client)
	lockp := grpcproxy.NewLockProxy(client)

	server := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
		grpc.MaxConcurrentStreams(math.MaxUint32),
	)

	pb.RegisterKVServer(server, kvp)
	pb.RegisterWatchServer(server, watchp)
	pb.RegisterClusterServer(server, clusterp)
	pb.RegisterLeaseServer(server, leasep)
	pb.RegisterMaintenanceServer(server, mainp)
	pb.RegisterAuthServer(server, authp)
	v3electionpb.RegisterElectionServer(server, electionp)
	v3lockpb.RegisterLockServer(server, lockp)

	// set zero values for metrics registered for this grpc server
	grpc_prometheus.Register(server)

	return server
}

func mustListenCMux(tlsinfo *transport.TLSInfo) cmux.CMux {
	l, err := net.Listen("tcp", grpcProxyAddress)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if l, err = transport.NewKeepAliveListener(l, "tcp", nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("listening for grpc-proxy client requests on %s\n", grpcProxyAddress)
	return cmux.New(l)
}

func discoverEndpoints(dns string, ca string, insecure bool) (s srv.SRVClients) {
	if dns == "" {
		return s
	}
	srvs, err := srv.GetClient("etcd-client", dns)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	endpoints := srvs.Endpoints
	fmt.Printf("discovered the cluster %s from %s\n", endpoints, dns)
	if insecure {
		return *srvs
	}
	// confirm TLS connections are good
	tlsInfo := transport.TLSInfo{
		TrustedCAFile: ca,
		ServerName:    dns,
	}
	fmt.Printf("validating discovered endpoints %v\n", endpoints)
	endpoints, err = transport.ValidateSecureEndpoints(tlsInfo, endpoints)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("using discovered endpoints %v\n", endpoints)

	// map endpoints back to SRVClients struct with SRV data
	eps := make(map[string]struct{})
	for _, ep := range endpoints {
		eps[ep] = struct{}{}
	}
	for i := range srvs.Endpoints {
		if _, ok := eps[srvs.Endpoints[i]]; !ok {
			continue
		}
		s.Endpoints = append(s.Endpoints, srvs.Endpoints[i])
		s.SRVs = append(s.SRVs, srvs.SRVs[i])
	}

	return s
}

func newClientCfg(eps []string) (*clientv3.Config, error) {
	// set tls if any one tls option set
	cfg := clientv3.Config{
		Endpoints:   eps,
		DialTimeout: 5 * time.Second,
	}

	return &cfg, nil
}
