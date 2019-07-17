package cli

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/brainupdaters/drlm-core/db"
	"github.com/brainupdaters/drlm-core/transport/grpc"

	log "github.com/sirupsen/logrus"
)

// Main is the main function of DRLM Core
func Main() {
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, "wg", &wg)

	go grpc.Serve(ctx)

	wg.Add(1)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	select {
	case <-stop:
		log.Info("stopping DRLM Core...")

		cancel()
		wg.Wait()

		db.DB.Close()
	}
}
