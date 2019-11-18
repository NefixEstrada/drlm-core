// SPDX-License-Identifier: AGPL-3.0-only

package minio

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"

	"github.com/brainupdaters/drlm-core/cfg"

	"github.com/brainupdaters/drlm-common/pkg/fs"
	sdk "github.com/minio/minio-go/v6"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// Minio is the connection with the Minio server
var Minio *sdk.Client

// conn returns the parameters for the minio connections
func conn() (string, string, string, bool) {
	return fmt.Sprintf("%s:%d", cfg.Config.Minio.Host, cfg.Config.Minio.Port),
		cfg.Config.Minio.AccessKey,
		cfg.Config.Minio.SecretKey,
		cfg.Config.Minio.SSL
}

// transport returns the http transport for the minio connections
func transport(tr *http.Transport) error {
	b, err := afero.ReadFile(fs.FS, cfg.Config.Minio.CertPath)
	if err != nil {
		return fmt.Errorf("error creating the minio http transport: error reading the certificate: %v", err)
	}

	if tr.TLSClientConfig == nil {
		// Taken from minio/minio-go
		// Keep TLS config.
		tlsConfig := &tls.Config{
			// Can't use SSLv3 because of POODLE and BEAST
			// Can't use TLSv1.0 because of POODLE and BEAST using CBC cipher
			// Can't use TLSv1.1 because of RC4 cipher usage
			MinVersion: tls.VersionTLS12,
		}
		tr.TLSClientConfig = tlsConfig
	}

	if tr.TLSClientConfig.RootCAs == nil {
		// Taken from minio/minio-go
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			// In some systems (like Windows) system cert pool is
			// not supported or no certificates are present on the
			// system - so we create a new cert pool.
			rootCAs = x509.NewCertPool()
		}
		tr.TLSClientConfig.RootCAs = rootCAs
	}

	if ok := tr.TLSClientConfig.RootCAs.AppendCertsFromPEM(b); !ok {
		return fmt.Errorf("error creating the minio http transport: error parsing the certificate")
	}

	return nil
}

// Init creates the Minio connection
func Init() {
	var err error
	Minio, err = sdk.New(conn())
	if err != nil {
		log.Fatalf("error creating the connection to minio: %v", err)
	}

	// If the certificate is self signed, add it to the transport certificates pool
	if cfg.Config.Minio.SSL && cfg.Config.Minio.CertPath != "" {
		defaultTransport, err := sdk.DefaultTransport(true)
		if err != nil {
			log.Fatalf("error creating the minio connection: error creating the default transport layer: %v", err)
		}

		tr := defaultTransport.(*http.Transport)
		if err = transport(tr); err != nil {
			log.Fatalf("error creating the minio connection: %v", err)
		}

		Minio.SetCustomTransport(tr)
	}

	log.Info("successfully created the connection to minio")
}
