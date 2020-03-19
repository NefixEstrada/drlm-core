// SPDX-License-Identifier: AGPL-3.0-only

package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/models"

	"github.com/dgrijalva/jwt-go"
)

// Token is a JWT token
type Token string

// TokenClaims are going to be embedded inside the JWT Token
type TokenClaims struct {
	Usr         string
	FirstIssued time.Time
	jwt.StandardClaims
}

// NewToken issues a new token
func NewToken(ctx *context.Context, usr string) (Token, time.Time, error) {
	expiresAt := time.Now().Add(ctx.Cfg.Security.TokensLifespan)

	claims := &TokenClaims{
		Usr:         usr,
		FirstIssued: time.Now(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt.Unix(),
		},
	}
	tkn := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signedTkn, err := tkn.SignedString([]byte(ctx.Cfg.Security.TokensSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("error signing the token: %v", err)
	}

	return Token(signedTkn), expiresAt, nil
}

// Validate checks whether a token is valid or not
func (t *Token) Validate(ctx *context.Context) bool {
	claims := &TokenClaims{}

	tkn, err := jwt.ParseWithClaims(t.String(), claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(ctx.Cfg.Security.TokensSecret), nil
	})
	if err != nil || !tkn.Valid {
		return false
	}

	return true
}

// ValidateAgent checks whether an agent token (secret) is valid or not
func (t *Token) ValidateAgent(ctx *context.Context) (string, bool) {
	agents, err := models.AgentList(ctx)
	if err != nil {
		return "", false
	}

	for _, a := range agents {
		if a.Secret == t.String() {
			return a.Host, true
		}
	}

	return "", false
}

// Renew renews the validity of the token
func (t *Token) Renew(ctx *context.Context) (time.Time, error) {
	claims := &TokenClaims{}

	tkn, err := jwt.ParseWithClaims(t.String(), claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(ctx.Cfg.Security.TokensSecret), nil
	})
	if err != nil || !tkn.Valid {
		if err != nil {
			// If the token has expired, but the user hasn't been modified since before the token was issued (the password hasn't changed), the token can be renewed
			if strings.HasPrefix(err.Error(), "token is expired by ") {
				if time.Since(claims.FirstIssued) > ctx.Cfg.Security.LoginLifespan {
					return time.Time{}, fmt.Errorf("error renewing the token: login lifespan exceeded, login again")
				}

				u := models.User{
					Username: claims.Usr,
				}

				if err = u.Load(ctx); err != nil {
					return time.Time{}, fmt.Errorf("error renewing the token: %v", err)
				}

				if u.UpdatedAt.Before(time.Unix(claims.IssuedAt, 0)) {
					signedTkn, expiresAt, err := renew(ctx, claims)
					if err != nil {
						return time.Time{}, fmt.Errorf("error renewing the token: %v", err)
					}

					*t = Token(signedTkn)
					return expiresAt, nil
				}
			}
		}

		return time.Time{}, errors.New("error renewing the token: the token is invalid or can't be renewed")
	}

	signedTkn, expiresAt, err := renew(ctx, claims)
	if err != nil {
		return time.Time{}, fmt.Errorf("error renewing the token: %v", err)
	}

	*t = Token(signedTkn)
	return expiresAt, nil
}

func renew(ctx *context.Context, claims *TokenClaims) (string, time.Time, error) {
	expiresAt := time.Now().Add(ctx.Cfg.Security.TokensLifespan)
	claims.ExpiresAt = expiresAt.Unix()

	tkn := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signedTkn, err := tkn.SignedString([]byte(ctx.Cfg.Security.TokensSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("error singing the token: %v", err)
	}

	return signedTkn, expiresAt, nil
}

// String returns the token as a string
func (t Token) String() string {
	return string(t)
}
