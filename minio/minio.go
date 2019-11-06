package minio

import (
	"fmt"
	"net/http"

	"github.com/brainupdaters/drlm-core/cfg"

	"github.com/brainupdaters/drlm-common/pkg/fs"
	"github.com/minio/minio-go/v6"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// Minio is the connection with the Minio server
var Minio *minio.Client

// Init creates the Minio connection
func Init() {
	var err error
	Minio, err = minio.New(
		fmt.Sprintf("%s:%d", cfg.Config.Minio.Host, cfg.Config.Minio.Port),
		cfg.Config.Minio.AccessKey,
		cfg.Config.Minio.SecretKey,
		cfg.Config.Minio.SSL,
	)
	if err != nil {
		log.Fatalf("error creating the connection to minio: %v", err)
	}

	// If the certificate is self signed, add it to the transport certificates pool
	if cfg.Config.Minio.SSL && cfg.Config.Minio.CertPath != "" {
		tr, err := minio.DefaultTransport(true)
		if err != nil {
			log.Fatalf("error creating the minio connection: error creating the default transport layer: %v", err)
		}
		transport := tr.(*http.Transport)

		b, err := afero.ReadFile(fs.FS, cfg.Config.Minio.CertPath)
		if err != nil {
			log.Fatalf("error creating the minio connection: error reading the certificate: %v", err)
		}

		if ok := transport.TLSClientConfig.RootCAs.AppendCertsFromPEM(b); !ok {
			log.Fatalf("error creating the minio connection: error parsing the certificate")
		}

		Minio.SetCustomTransport(transport)
	}

	log.Info("successfully created the connection to minio")
}
