package cfg

import (
	"path/filepath"
	"strings"

	"github.com/brainupdaters/drlm-common/pkg/fs"
	logger "github.com/brainupdaters/drlm-common/pkg/log"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config has the values of the user configuration
var Config *DRLMCoreConfig

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
	BcryptCost     int    `mapstructure:"bcrypt_cost"`
	TokensLifespan int    `mapstructure:"tokens_lifespan"`
	TokensSecret   string `mapstructure:"tokens_secret"`
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
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
}

// v is the viper instance for the configuration
var v *viper.Viper

// Init prepares the configuration and reads it
func Init(cfgFile string) {
	v = viper.New()
	v.SetFs(fs.FS)
	SetDefaults()

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	}

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("error reading the configuration: %v", err)
	}

	if err := v.Unmarshal(&Config); err != nil {
		log.Fatalf("error decoding the configuration: invalid configuration: %v", err)
	}
}

// SetDefaults sets the default configurations for Viper
func SetDefaults() {
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
		"tokens_lifespan": 5,
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
		"port":       9000,
		"ssl":        true,
		"access_key": "drlm3minio",
		"secret_key": "drlm3minio",
	})
	v.SetDefault("log", map[string]interface{}{
		"level": "info",
		"file":  "/var/log/drlm/core.log",
	})

	v.SetEnvPrefix("DRLMCORE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
}
