package cfg

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestSetDefaults(t *testing.T) {
	assert := assert.New(t)

	t.Run("should work as expected", func(t *testing.T) {
		v = viper.New()

		SetDefaults()

		assert.Equal(50051, v.GetInt("grpc.port"))
		assert.Equal(true, v.GetBool("grpc.tls"))
		assert.Equal("cert/server.crt", v.GetString("grpc.cert_path"))
		assert.Equal("cert/server.key", v.GetString("grpc.key_path"))

		assert.Equal(14, v.GetInt("security.bcrypt_cost"))
		assert.Equal("", v.GetString("security.tokens_secret"))
		assert.Equal(5*time.Minute, v.GetDuration("security.tokens_lifespan"))
		assert.Equal(240*time.Hour, v.GetDuration("security.login_lifespan"))

		assert.Equal("mariadb", v.GetString("db.host"))
		assert.Equal(3306, v.GetInt("db.port"))
		assert.Equal("drlm3", v.GetString("db.username"))
		assert.Equal("drlm3db", v.GetString("db.password"))
		assert.Equal("drlm3", v.GetString("db.database"))

		assert.Equal("minio", v.GetString("minio.host"))
		assert.Equal(9000, v.GetInt("minio.port"))
		assert.Equal(true, v.GetBool("minio.ssl"))
		assert.Equal("drlm3minio", v.GetString("minio.access_key"))
		assert.Equal("drlm3minio", v.GetString("minio.secret_key"))

		assert.Equal("info", v.GetString("log.level"))
		assert.Equal("/var/log/drlm/core.log", v.GetString("log.file"))
	})
}
