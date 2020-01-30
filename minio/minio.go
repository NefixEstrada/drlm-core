// SPDX-License-Identifier: AGPL-3.0-only

package minio

import (
	"github.com/brainupdaters/drlm-core/cfg"

	cmnMinio "github.com/brainupdaters/drlm-common/pkg/minio"
	sdk "github.com/minio/minio-go/v6"
	"github.com/minio/minio/pkg/madmin"
	log "github.com/sirupsen/logrus"
)

var (
	// Minio is the connection with the Minio server
	Minio *sdk.Client
	cli   *madmin.AdminClient
)

// Init creates the Minio connection
func Init() {
	var err error
	Minio, err = cmnMinio.NewSDK(
		cfg.Config.Minio.Host,
		cfg.Config.Minio.Port,
		cfg.Config.Minio.AccessKey,
		cfg.Config.Minio.SecretKey,
		cfg.Config.Minio.SSL,
		cfg.Config.Minio.CertPath,
	)
	if err != nil {
		log.Fatal(err)
	}

	cli, err = cmnMinio.NewAdminClient(
		cfg.Config.Minio.Host,
		cfg.Config.Minio.Port,
		cfg.Config.Minio.AccessKey,
		cfg.Config.Minio.SecretKey,
		cfg.Config.Minio.SSL,
		cfg.Config.Minio.CertPath,
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("successfully created the connection to minio")
}
