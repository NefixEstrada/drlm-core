// SPDX-License-Identifier: AGPL-3.0-only

package db_test

import (
	"testing"

	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/db"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/brainupdaters/drlm-common/pkg/test"
	"github.com/stretchr/testify/suite"
)

type TestDBSuite struct {
	test.Test
}

func TestDB(t *testing.T) {
	suite.Run(t, new(TestDBSuite))
}

func (s *TestDBSuite) TestInit() {
	tests.GenerateCfg(s.T())

	cfg.Config.DB.Host = "localhost"
	cfg.Config.DB.Port = s.FreePort()

	s.Exits(db.Init)
}
