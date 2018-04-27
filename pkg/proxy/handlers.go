package proxy

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/srv"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/coreos/etcd/proxy/grpcproxy"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
)

func (s *Server) mustNewClient() *clientv3.Client {
	srvs := discoverEndpoints("", "", false)
	eps := srvs.Endpoints
	if len(eps) == 0 {
		eps = []string{s.EtcdAddress}
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

func (s *Server) mustListenCMux(tlsinfo *transport.TLSInfo) cmux.CMux {
	l, err := net.Listen("tcp", s.BindAddress)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if l, err = transport.NewKeepAliveListener(l, "tcp", nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("listening for grpc-proxy client requests on %s\n", s.BindAddress)
	return cmux.New(l)
}

func mustHTTPListener(m cmux.CMux, tlsinfo *transport.TLSInfo, c *clientv3.Client) (*http.Server, net.Listener) {
	httpmux := http.NewServeMux()
	httpmux.HandleFunc("/", http.NotFound)
	srvhttp := &http.Server{Handler: httpmux}
	return srvhttp, m.Match(cmux.HTTP1())
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
