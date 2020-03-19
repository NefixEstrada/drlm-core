// SPDX-License-Identifier: AGPL-3.0-only

package scheduler

import (
	"testing"

	"github.com/brainupdaters/drlm-core/models"

	"github.com/stretchr/testify/suite"
)

type TestJobListInternalSuite struct {
	suite.Suite
}

func TestJobListInternal(t *testing.T) {
	suite.Run(t, &TestJobListInternalSuite{})
}

func (s *TestJobListInternalSuite) TestList() {
	job := &models.Job{AgentHost: "laptop"}

	j := &jobList{}
	j.v = []*models.Job{job}

	s.Len(j.List(), 1)
}

func (s *TestJobListInternalSuite) TestAdd() {
	job := &models.Job{AgentHost: "laptop"}

	j := &jobList{}
	j.Add(job)

	s.Len(j.v, 1)
}
