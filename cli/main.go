// SPDX-License-Identifier: AGPL-3.0-only

package cli

import (
	stdContext "context"
	"fmt"
	"os"
	"os/signal"

	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/scheduler"
	"github.com/brainupdaters/drlm-core/transport/grpc"

	log "github.com/sirupsen/logrus"
)

// Main is the main function of DRLM Core
func Main(ctx *context.Context, cancel stdContext.CancelFunc) {
	scheduler.Init(ctx)
	ctx.WG.Add(1)

	go grpc.Serve(ctx)
	ctx.WG.Add(1)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	select {
	case <-stop:
		fmt.Println("")
		log.Info("stopping DRLM Core...")

		cancel()
		ctx.WG.Wait()

		ctx.DB.Close()
	}
}
