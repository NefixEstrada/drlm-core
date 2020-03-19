// SPDX-License-Identifier: AGPL-3.0-only

package auth

import (
	"testing"
	"time"

	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

func TestRenew(t *testing.T) {
	assert := assert.New(t)

	t.Run("should renew the token correctly", func(t *testing.T) {
		ctx := tests.GenerateCtx()
		tests.GenerateCfg(t, ctx)

		originalExpirationTime := time.Now().Add(1 * time.Minute)

		signedTkn, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &TokenClaims{
			Usr: "nefix",
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: originalExpirationTime.Unix(),
			},
		}).SignedString([]byte(ctx.Cfg.Security.TokensSecret))
		assert.Nil(err)

		tkn := Token(signedTkn)

		claims := &TokenClaims{}
		parsedTkn, err := jwt.ParseWithClaims(tkn.String(), claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(ctx.Cfg.Security.TokensSecret), nil
		})
		assert.Nil(err)
		assert.True(parsedTkn.Valid)

		signedTkn, expiresAt, err := renew(ctx, claims)
		assert.Nil(err)

		tkn = Token(signedTkn)
		assert.True(expiresAt.After(originalExpirationTime))
	})
}
