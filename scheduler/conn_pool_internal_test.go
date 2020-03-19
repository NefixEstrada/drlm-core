// SPDX-License-Identifier: AGPL-3.0-only

package scheduler

import (
	"testing"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
	"github.com/brainupdaters/drlm-core/utils/tests"
	"github.com/stretchr/testify/suite"
)

type TestConnPoolInternalSuite struct {
	suite.Suite
}

func TestConnPoolInternal(t *testing.T) {
	suite.Run(t, &TestConnPoolInternalSuite{})
}

func (s *TestConnPoolInternalSuite) TestGet() {
	s.Run("should return true and the value if the connection is in the pool", func() {
		conn := &tests.AgentConnectionServerMock{}

		c := &connPool{v: map[string]drlm.DRLM_AgentConnectionServer{
			"127.0.0.1": conn,
		}}
		poolConn, ok := c.Get("127.0.0.1")

		s.True(ok)
		s.Equal(conn, poolConn)
	})

	s.Run("should return false if the connection isn't in the pool", func() {
		c := &connPool{v: map[string]drlm.DRLM_AgentConnectionServer{}}
		_, ok := c.Get("127.0.0.1")

		s.False(ok)
	})
}

func (s *TestConnPoolInternalSuite) TestAdd() {
	conn := &tests.AgentConnectionServerMock{}

	c := &connPool{v: map[string]drlm.DRLM_AgentConnectionServer{}}
	c.Add("127.0.0.1", conn)

	s.Equal(conn, c.v["127.0.0.1"])
}

func (s *TestConnPoolInternalSuite) TestDelete() {
	conn := &tests.AgentConnectionServerMock{}

	c := &connPool{v: map[string]drlm.DRLM_AgentConnectionServer{
		"127.0.0.1": conn,
	}}
	c.Delete("127.0.0.1")

	_, ok := c.v["127.0.0.1"]
	s.False(ok)
}
