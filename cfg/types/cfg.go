// SPDX-License-Identifier: AGPL-3.0-only

package types

import (
	"time"

	logger "github.com/brainupdaters/drlm-common/pkg/log"
)

// DRLMCoreConfig is the configuration of the Core of DRLM
type DRLMCoreConfig struct {
	GRPC     DRLMCoreGRPCConfig     `mapstructure:"grpc"`
	Security DRLMCoreSecurityConfig `mapstructure:"security"`
	DB       DRLMCoreDBConfig       `mapstructure:"db"`
	Minio    DRLMCoreMinioConfig    `mapstructure:"minio"`
	Log      logger.Config          `mapstructure:"log"`
}

// DRLMCoreGRPCConfig is the configuration related with the GRPC of DRLM Core
type DRLMCoreGRPCConfig struct {
	Port     int    `mapstructure:"port"`
	TLS      bool   `mapstructure:"tls"`
	CertPath string `mapstructure:"cert_path"`
	KeyPath  string `mapstructure:"key_path"`
}

// DRLMCoreSecurityConfig is the configuration related with the security of DLRM Core
type DRLMCoreSecurityConfig struct {
	BcryptCost     int           `mapstructure:"bcrypt_cost"`
	TokensSecret   string        `mapstructure:"tokens_secret"`
	TokensLifespan time.Duration `mapstructure:"tokens_lifespan"`
	LoginLifespan  time.Duration `mapstructure:"login_lifespan"`
	SSHKeysPath    string        `mapstructure:"ssh_keys_path"`
}

// DRLMCoreDBConfig is the configuration related wtih the DB of the DRLM Core
type DRLMCoreDBConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Usr  string `mapstructure:"username"`
	Pwd  string `mapstructure:"password"`
	DB   string `mapstructure:"database"`
}

// DRLMCoreMinioConfig is the configuration related wtih the Minio of the DRLM Core
type DRLMCoreMinioConfig struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	SSL       bool   `mapstructure:"ssl"`
	CertPath  string `mapstructure:"cert_path"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Location  string `mapstructure:"location"`
}
