// SPDX-License-Identifier: AGPL-3.0-only

package context_test

import (
	"testing"

	"github.com/brainupdaters/drlm-core/context"

	"github.com/stretchr/testify/suite"
)

type TestContextSuite struct {
	suite.Suite
}

func TestContext(t *testing.T) {
	suite.Run(t, &TestContextSuite{})
}

func (s *TestContextSuite) TestContextDeadline() {
	ctx := context.Background()
	_, ok := ctx.Deadline()

	s.False(ok)
}

func (s *TestContextSuite) TestContextDone() {
	ctx, cancel := context.WithCancel()

	go func() {
		cancel()
	}()

	<-ctx.Done()
}

func (s *TestContextSuite) TestContextErr() {
	ctx, cancel := context.WithCancel()

	go func() {
		cancel()
	}()

	<-ctx.Done()

	s.EqualError(ctx.Err(), "context canceled")
}

func (s *TestContextSuite) TestContextValue() {
	ctx := context.Background()

	_, ok := ctx.Value("key").(bool)

	s.False(ok)
}

func (s *TestContextSuite) TestBackground() {
	context.Background()
}

func (s *TestContextSuite) TestTODO() {
	context.TODO()
}

func (s *TestContextSuite) TestWithCancel() {
	context.WithCancel()
}
