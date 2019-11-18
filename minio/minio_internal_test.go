package minio

import (
	"net/http"
	"testing"

	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/brainupdaters/drlm-common/pkg/fs"
	"github.com/brainupdaters/drlm-common/pkg/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

type TestMinioSuite struct {
	test.Test
}

func TestMino(t *testing.T) {
	suite.Run(t, new(TestMinioSuite))
}

func (s *TestMinioSuite) TestConn() {
	tests.GenerateCfg(s.T())
	cfg.Config.Minio.Port = 9443

	endpoint, aKey, sKey, ssl := conn()

	s.Equal("localhost:9443", endpoint)
	s.Equal("drlm3minio", aKey)
	s.Equal("drlm3minio", sKey)
	s.False(ssl)
}

func (s *TestMinioSuite) TestTransport() {
	s.Run("should return a correct transport", func() {
		tests.GenerateCfg(s.T())
		s.GenerateCert("minio", "cert")

		tr := http.DefaultTransport.(*http.Transport)
		tr.TLSClientConfig = nil
		s.Nil(transport(tr))
	})

	s.Run("should return an error if there's an error reading the cert", func() {
		tests.GenerateCfg(s.T())

		tr := http.DefaultTransport.(*http.Transport)
		s.EqualError(transport(tr), "error creating the minio http transport: error reading the certificate: open cert/minio.crt: file does not exist")
	})

	s.Run("should return an error if there's SOMETHING", func() {
		tests.GenerateCfg(s.T())

		s.Require().Nil(afero.WriteFile(fs.FS, "cert/minio.crt", []byte(`invalid cert`), 0644))

		tr := http.DefaultTransport.(*http.Transport)
		s.EqualError(transport(tr), "error creating the minio http transport: error parsing the certificate")
	})
}
