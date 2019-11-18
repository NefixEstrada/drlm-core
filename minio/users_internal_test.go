// SPDX-License-Identifier: AGPL-3.0-only

package minio

import (
	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/minio/minio/pkg/madmin"
)

func (s *TestMinioSuite) TestNewAdminClient() {
	s.Run("should create the admin client correctly", func() {
		tests.GenerateCfg(s.T())
		cfg.Config.Minio.SSL = true
		s.GenerateCert("minio", "cert")

		cli, err := newAdminClient()

		s.NotEqual(&madmin.AdminClient{}, cli)
		s.Nil(err)
	})

	s.Run("should return an error if there's an error creating the client", func() {
		tests.GenerateCfg(s.T())
		cfg.Config.Minio.Host = ""
		cfg.Config.Minio.Port = 9443

		cli, err := newAdminClient()

		s.Nil(cli)
		s.EqualError(err, "error connecting to the Minio admin API: Endpoint: :9443 does not follow ip address or domain name standards.")
	})

	s.Run("should return an error if there's an error creating the transport", func() {
		tests.GenerateCfg(s.T())
		cfg.Config.Minio.SSL = true

		cli, err := newAdminClient()

		s.Nil(cli)
		s.EqualError(err, "error creating the minio http transport: error reading the certificate: open cert/minio.crt: file does not exist")
	})
}
