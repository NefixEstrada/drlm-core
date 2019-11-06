package minio_test

import (
	"testing"

	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/minio"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/brainupdaters/drlm-common/pkg/fs"
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
		tests.GenerateCfg(s.T())
		fs.FS = afero.NewMemMapFs()
		s.GenerateCert("minio", "cert")

		minio.Init()
	})

	s.Run("should exit if there's an error creating the connection", func() {
		tests.GenerateCfg(s.T())
		cfg.Config.Minio.Host = "\\"

		s.Exits(minio.Init)
	})

	s.Run("should exit if there's an error reading the certificate", func() {
		tests.GenerateCfg(s.T())
		fs.FS = afero.NewMemMapFs()

		s.Exits(minio.Init)
	})

	s.Run("should exit if there's an error parsing the certificate", func() {
		tests.GenerateCfg(s.T())
		fs.FS = afero.NewMemMapFs()

		s.Require().Nil(afero.WriteFile(fs.FS, "cert/minio.crt", []byte("invalid cert"), 0400))

		s.Exits(minio.Init)
	})
}
