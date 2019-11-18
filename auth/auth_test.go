// SPDX-License-Identifier: AGPL-3.0-only

package auth_test

import (
	"testing"

	"github.com/brainupdaters/drlm-core/auth"
	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/brainupdaters/drlm-common/pkg/test"
	"github.com/stretchr/testify/suite"
)

type TestAuthSuite struct {
	test.Test
}

func TestAuth(t *testing.T) {
	suite.Run(t, new(TestAuthSuite))
}

func (s *TestAuthSuite) TestInit() {
	s.Run("should not fail", func() {
		tests.GenerateCfg(s.T())
		cfg.Config.Security.TokensSecret = "Q~-Z{P|ahLL/7L~&UJnU~x+!t+7x-n^V~M&b$O.5[sZ+lS4zfG8Mz.:'#B,Iwr]5"

		auth.Init()
	})

	s.Run("should exit if there's no token secret", func() {
		tests.GenerateCfg(s.T())
		s.Exits(auth.Init)
	})

	s.Run("should exit if the secret has less than 32 characters", func() {
		tests.GenerateCfg(s.T())
		cfg.Config.Security.TokensSecret = "lorem ipsum dolor"

		s.Exits(auth.Init)
	})
}
