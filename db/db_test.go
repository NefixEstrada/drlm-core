// SPDX-License-Identifier: AGPL-3.0-only

package db_test

import (
	"testing"

	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/db"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/brainupdaters/drlm-common/pkg/test"
	"github.com/stretchr/testify/suite"
)

type TestDBSuite struct {
	test.Test

	ctx *context.Context
}

func (s *TestDBSuite) SetupTest() {
	s.ctx = tests.GenerateCtx()
}

func TestDB(t *testing.T) {
	suite.Run(t, new(TestDBSuite))
}

func (s *TestDBSuite) TestInit() {
	tests.GenerateCfg(s.T(), s.ctx)

	s.ctx.Cfg.DB.Host = "localhost"
	s.ctx.Cfg.DB.Port = s.FreePort()

	s.Exits(func() { db.Init(s.ctx) })
}
