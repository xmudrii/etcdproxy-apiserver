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

package proxy

import (
	"fmt"
	"math"
	"os"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
	"github.com/coreos/etcd/etcdserver/api/v3election/v3electionpb"
	"github.com/coreos/etcd/etcdserver/api/v3lock/v3lockpb"
	pb "github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/proxy/grpcproxy"
	"github.com/coreos/pkg/capnslog"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"time"
	"github.com/coreos/etcd/pkg/transport"
	"net"
)

// Server provides information about etcd and proxy server.
type Server struct {
	// Address to bind proxy server on in format of <ip>:<port>.
	BindAddress string
	// etcd namespace to provide over the proxy.
	Namespace string

	// EtcdAddresses in format of http://<ip>:<port>.
	EtcdAddresses []string
}

// NewGRPCServer creates new gRPC structure.
func NewGRPCServer(bindAddress, namespace string, etcdAddresses []string) *Server {
	return &Server{
		BindAddress:   bindAddress,
		Namespace:     namespace,
		EtcdAddresses: etcdAddresses,
	}
}

// StartNonSecureServer starts non-secure etcd-gRPC proxy.
func (s *Server) StartNonSecureServer() {
	// gRPC logging.
	capnslog.SetGlobalLogLevel(capnslog.DEBUG)
	grpc.EnableTracing = true
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(os.Stderr, os.Stderr, os.Stderr))

	// Start server.
	m := s.mustListenCMux(nil)
	grpcl := m.Match(cmux.HTTP2())

	client := s.mustNewClient()
	errc := make(chan error)
	go func() { errc <- s.newGRPCProxyServer(client).Serve(grpcl) }()
	go func() { errc <- m.Serve() }()

	fmt.Fprintln(os.Stderr, <-errc)
}

func (s *Server) newGRPCProxyServer(client *clientv3.Client) *grpc.Server {
	if len(s.Namespace) > 0 {
		client.KV = namespace.NewKV(client.KV, s.Namespace)
		client.Watcher = namespace.NewWatcher(client.Watcher, s.Namespace)
		client.Lease = namespace.NewLease(client.Lease, s.Namespace)
	}

	kvp, _ := grpcproxy.NewKvProxy(client)
	watchp, _ := grpcproxy.NewWatchProxy(client)
	clusterp, _ := grpcproxy.NewClusterProxy(client, s.BindAddress, "")
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

func (s *Server) mustNewClient() *clientv3.Client {
	cfg := clientv3.Config{
		Endpoints:   s.EtcdAddresses,
		DialTimeout: 5 * time.Second,
	}

	cfg.DialOptions = append(cfg.DialOptions,
		grpc.WithUnaryInterceptor(grpcproxy.AuthUnaryClientInterceptor))
	cfg.DialOptions = append(cfg.DialOptions,
		grpc.WithStreamInterceptor(grpcproxy.AuthStreamClientInterceptor))
	client, err := clientv3.New(cfg)
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

	fmt.Printf("listening for grpc-proxy client requests on %s", s.BindAddress)
	return cmux.New(l)
}