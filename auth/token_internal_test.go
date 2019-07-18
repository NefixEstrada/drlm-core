package auth

import (
	"testing"
	"time"

	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

func TestRenew(t *testing.T) {
	assert := assert.New(t)

	t.Run("should renew the token correctly", func(t *testing.T) {
		tests.GenerateCfg(t)

		originalExpirationTime := time.Now().Add(1 * time.Minute)

		signedTkn, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &TokenClaims{
			Usr: "nefix",
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: originalExpirationTime.Unix(),
			},
		}).SignedString([]byte(cfg.Config.Security.TokensSecret))
		assert.Nil(err)

		tkn := Token(signedTkn)

		claims := &TokenClaims{}
		parsedTkn, err := jwt.ParseWithClaims(tkn.String(), claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Config.Security.TokensSecret), nil
		})
		assert.Nil(err)
		assert.True(parsedTkn.Valid)

		signedTkn, err = renew(claims)
		assert.Nil(err)

		tkn = Token(signedTkn)

		claims = &TokenClaims{}
		parsedTkn, err = jwt.ParseWithClaims(tkn.String(), claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Config.Security.TokensSecret), nil
		})
		assert.Nil(err)
		assert.True(parsedTkn.Valid)

		newExpirationTime := time.Unix(claims.ExpiresAt, 0)
		assert.True(newExpirationTime.After(originalExpirationTime))
	})
}
