// SPDX-License-Identifier: AGPL-3.0-only

package minio

import (
	"github.com/brainupdaters/drlm-core/context"

	cmnMinio "github.com/brainupdaters/drlm-common/pkg/minio"
	log "github.com/sirupsen/logrus"
)

// Init creates the Minio connection
func Init(ctx *context.Context) {
	var err error
	ctx.MinioCli, err = cmnMinio.NewSDK(
		ctx.FS,
		ctx.Cfg.Minio.Host,
		ctx.Cfg.Minio.Port,
		ctx.Cfg.Minio.AccessKey,
		ctx.Cfg.Minio.SecretKey,
		ctx.Cfg.Minio.SSL,
		ctx.Cfg.Minio.CertPath,
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx.MinioAdminCli, err = cmnMinio.NewAdminClient(
		ctx.FS,
		ctx.Cfg.Minio.Host,
		ctx.Cfg.Minio.Port,
		ctx.Cfg.Minio.AccessKey,
		ctx.Cfg.Minio.SecretKey,
		ctx.Cfg.Minio.SSL,
		ctx.Cfg.Minio.CertPath,
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("successfully created the connection to minio")
}
