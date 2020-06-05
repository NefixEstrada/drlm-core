// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"github.com/brainupdaters/drlm-core/auth"
	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/cli"
	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/db"
	"github.com/brainupdaters/drlm-core/db/migrations"
	"github.com/brainupdaters/drlm-core/minio"
	"github.com/brainupdaters/drlm-core/registry"

	logger "github.com/brainupdaters/drlm-common/pkg/log"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "drlm-core",
	Short: "TODO",
	Long:  "TODO",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel()

		ctx.FS = afero.NewOsFs()

		cfg.Init(ctx, cfgFile)
		logger.Init(ctx.Cfg.Log)
		db.Init(ctx)
		migrations.Migrate(ctx)
		auth.Init(ctx)
		minio.Init(ctx)
		registry.Init(ctx)

		cli.Main(ctx, cancel)
	},
}

// Execute is the main function of the CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", `configuration file to use instead of the defaults ("/etc/drlm/core.toml", "~/.config/drlm/core.toml", "~/.drlm/core.toml", "./core.toml")`)
}
