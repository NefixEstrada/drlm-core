// SPDX-License-Identifier: AGPL-3.0-only

package plugin_test

import (
	"testing"

	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/models"
	"github.com/brainupdaters/drlm-core/plugin"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/brainupdaters/drlm-common/pkg/os"
	"github.com/stretchr/testify/suite"
)

type TestPluginSuite struct {
	suite.Suite
	ctx *context.Context
}

func (s *TestPluginSuite) SetupTest() {
	s.ctx = tests.GenerateCtx()
}

func TestPlugin(t *testing.T) {
	suite.Run(t, &TestPluginSuite{})
}

func (s *TestPluginSuite) TestInstall() {
	s.Run("should install the plugin correctly", func() {

	})

	s.Run("should fail if the arch is unsupported", func() {
		p := &models.Plugin{Arch: []os.Arch{os.ArchAmd64}}
		a := &models.Agent{Arch: os.Arch(999)}

		plugin.Install(s.ctx, p, a, []byte("plugin"))
	})

	s.Run("should fail if the OS is unsupported", func() {
		p := &models.Plugin{OS: []os.OS{os.Linux}}
		a := &models.Agent{OS: os.OS(999)}

		plugin.Install(s.ctx, p, a, []byte("plugin"))
	})
}
