package etcdapi

import (
	"context"
	"time"

	"github.com/coreos/etcd/clientv3"
)

func WriteEtcd(address, key, value string) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{address},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err = cli.Put(ctx, key, value)
	cancel()
	return err
}
