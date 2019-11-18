// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"github.com/brainupdaters/drlm-core/auth"
	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/cli"
	"github.com/brainupdaters/drlm-core/db"
	"github.com/brainupdaters/drlm-core/db/migrations"
	"github.com/brainupdaters/drlm-core/minio"

	"github.com/brainupdaters/drlm-common/pkg/fs"
	logger "github.com/brainupdaters/drlm-common/pkg/log"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cfgFile string
var verbose bool

var rootCmd = &cobra.Command{
	Use:   "drlm-core",
	Short: "TODO",
	Long:  "TODO",
	Run: func(cmd *cobra.Command, args []string) {
		cli.Main()
	},
}

// Execute is the main function of the CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", `configuration file to use instead of the defaults ("/etc/drlm/core.toml", "~/.config/drlm/core.toml", "~/.drlm/core.toml", "./core.toml")`)
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging output")
}

func initConfig() {
	fs.Init()
	cfg.Init(cfgFile)
	logger.Init(cfg.Config.Log)
	auth.Init()
	db.Init()
	migrations.Migrate() // Migrations are done here to avoid import cycles
	minio.Init()
}
