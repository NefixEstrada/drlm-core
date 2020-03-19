// SPDX-License-Identifier: AGPL-3.0-only

package minio_test

import (
	"testing"

	"github.com/brainupdaters/drlm-core/minio"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/brainupdaters/drlm-common/pkg/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

type TestMinioSuite struct {
	test.Test
}

func TestMinio(t *testing.T) {
	suite.Run(t, new(TestMinioSuite))
}

func (s *TestMinioSuite) TestInit() {
	s.Run("should init correctly", func() {
		ctx := tests.GenerateCtx()
		tests.GenerateCfg(s.T(), ctx)
		s.GenerateCert(ctx.FS, "minio", "cert")

		ctx.Cfg.Minio.SSL = true

		minio.Init(ctx)
	})

	s.Run("should exit if there's an error creating the connection", func() {
		ctx := tests.GenerateCtx()
		tests.GenerateCfg(s.T(), ctx)
		ctx.Cfg.Minio.Host = "\\"

		s.Exits(func() { minio.Init(ctx) })
	})

	s.Run("should exit if there's an error creating the connection", func() {
		ctx := tests.GenerateCtx()
		tests.GenerateCfg(s.T(), ctx)
		ctx.Cfg.Minio.Host = "\\"

		s.Exits(func() { minio.Init(ctx) })
	})

	s.Run("should exit if there's an error reading the certificate", func() {
		ctx := tests.GenerateCtx()
		tests.GenerateCfg(s.T(), ctx)
		ctx.Cfg.Minio.SSL = true

		s.Exits(func() { minio.Init(ctx) })
	})

	s.Run("should exit if there's an error parsing the certificate", func() {
		ctx := tests.GenerateCtx()
		tests.GenerateCfg(s.T(), ctx)
		ctx.Cfg.Minio.SSL = true

		s.Require().Nil(afero.WriteFile(ctx.FS, "cert/minio.crt", []byte("invalid cert"), 0400))

		s.Exits(func() { minio.Init(ctx) })
	})
}
