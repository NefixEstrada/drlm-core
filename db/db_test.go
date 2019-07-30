package db_test

import (
	"testing"

	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/db"
	"github.com/brainupdaters/drlm-core/utils/tests"

	cmnTests "github.com/brainupdaters/drlm-common/pkg/tests"
)

func TestInit(t *testing.T) {
	t.Run("should exit if there's an error connecting to the DB", func(t *testing.T) {
		tests.GenerateCfg(t)
		cfg.Config.DB.Host = "localhost"
		cfg.Config.DB.Port = cmnTests.GetFreePort(t)

		cmnTests.AssertExits(t, db.Init)
	})
}
