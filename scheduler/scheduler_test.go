// SPDX-License-Identifier: AGPL-3.0-only

package scheduler_test

import (
	"errors"
	"regexp"
	"testing"

	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/scheduler"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/brainupdaters/drlm-common/pkg/test"
	"github.com/stretchr/testify/suite"
)

type TestSchedulerSuite struct {
	test.Test
	ctx  *context.Context
	mock sqlmock.Sqlmock
}

func (s *TestSchedulerSuite) SetupTest() {
	s.ctx = tests.GenerateCtx()
	s.mock = tests.GenerateDB(s.T(), s.ctx)
}

func TestScheduler(t *testing.T) {
	suite.Run(t, &TestSchedulerSuite{})
}

func (s *TestSchedulerSuite) TestInit() {
	s.Run("should initialize the scheduler correctly", func() {
		ctx, cancel := context.WithCancel()
		mock := tests.GenerateDB(s.T(), ctx)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"  WHERE "jobs"."deleted_at" IS NULL`)).WillReturnRows(sqlmock.NewRows([]string{}))

		ctx.WG.Add(1)

		scheduler.Init(ctx)

		cancel()
	})

	s.Run("should exit if there's an error getting the job list", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"  WHERE "jobs"."deleted_at" IS NULL`)).WillReturnError(errors.New("testing error"))

		s.Exits(func() { scheduler.Init(s.ctx) })
	})
}
