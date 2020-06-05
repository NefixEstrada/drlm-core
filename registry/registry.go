package registry

import (
	"fmt"
	"log"
	"time"

	"github.com/brainupdaters/drlm-core/context"

	"go.etcd.io/etcd/v3/embed"
)

func Init(ctx *context.Context) {
	cfg := embed.NewConfig()
	cfg.Dir = "./etcd-data"
	cfg.DNSClusterServiceName = "core.default.svc.cluster.local"

	e, err := embed.StartEtcd(cfg)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			log.Println(cfg.GetDNSClusterNames())

			for _, p := range e.Peers {
				fmt.Println(p.Addr())
			}

			time.Sleep(1 * time.Second)
		}
	}()
}
