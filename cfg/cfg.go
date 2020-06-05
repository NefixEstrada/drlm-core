// SPDX-License-Identifier: AGPL-3.0-only

package cfg

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/brainupdaters/drlm-core/context"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Init prepares the configuration and reads it
func Init(ctx *context.Context, cfgFile string) {
	v := viper.New()
	v.SetFs(ctx.FS)
	SetDefaults(v)

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatalf("error reading the configuration: %v", err)
		}

		log.Warnln("configuration file not found, using default and environment vaules")
	}

	if err := v.Unmarshal(&ctx.Cfg); err != nil {
		log.Fatalf("error decoding the configuration: invalid configuration: %v", err)
	}
}

// SetDefaults sets the default configurations for Viper
func SetDefaults(v *viper.Viper) {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatalf("error getting the home directory: %v", err)
	}

	v.SetConfigName("core")
	v.AddConfigPath(".")
	v.AddConfigPath(filepath.Join(home, ".drlm"))
	v.AddConfigPath(filepath.Join(home, ".config/drlm"))
	v.AddConfigPath("/etc/drlm")

	v.SetDefault("grpc", map[string]interface{}{
		"port":      50051,
		"tls":       true,
		"cert_path": "cert/server.crt",
		"key_path":  "cert/server.key",
	})
	v.SetDefault("security", map[string]interface{}{
		"bcrypt_cost":     14,
		"tokens_lifespan": 5 * time.Minute,
		"login_lifespan":  240 * time.Hour,
		"ssh_keys_path":   "./ssh",
	})
	v.SetDefault("db", map[string]interface{}{
		"host":     "mariadb",
		"port":     3306,
		"username": "drlm3",
		"password": "drlm3db",
		"database": "drlm3",
	})
	v.SetDefault("minio", map[string]interface{}{
		"host":       "minio",
		"port":       9443,
		"ssl":        true,
		"cert_path":  "cert/minio.crt",
		"access_key": "drlm3minio",
		"secret_key": "drlm3minio",
		"location":   "eu-west-3",
	})
	v.SetDefault("log", map[string]interface{}{
		"level": "info",
		"file":  "/var/log/drlm/core.log",
	})

	v.SetEnvPrefix("DRLM_CORE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
}
