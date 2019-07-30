package auth_test

import (
	"testing"

	"github.com/brainupdaters/drlm-core/auth"
	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/utils/tests"

	cmnTests "github.com/brainupdaters/drlm-common/pkg/tests"
)

func TestInit(t *testing.T) {
	t.Run("should not fail or anything", func(t *testing.T) {
		tests.GenerateCfg(t)
		cfg.Config.Security.TokensSecret = "Q~-Z{P|ahLL/7L~&UJnU~x+!t+7x-n^V~M&b$O.5[sZ+lS4zfG8Mz.:'#B,Iwr]5"

		auth.Init()
	})

	t.Run("should exit if there's no tokens secret", func(t *testing.T) {
		tests.GenerateCfg(t)
		cmnTests.AssertExits(t, auth.Init)
	})

	t.Run("should exit if the secret has less than 32 characters", func(t *testing.T) {
		tests.GenerateCfg(t)
		cfg.Config.Security.TokensSecret = "lorem ipsum dolor"

		cmnTests.AssertExits(t, auth.Init)
	})
}
