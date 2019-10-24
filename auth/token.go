package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/brainupdaters/drlm-core/cfg"
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
func NewToken(usr string) (Token, time.Time, error) {
	expiresAt := time.Now().Add(cfg.Config.Security.TokensLifespan)

	claims := &TokenClaims{
		Usr:         usr,
		FirstIssued: time.Now(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt.Unix(),
		},
	}
	tkn := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signedTkn, err := tkn.SignedString([]byte(cfg.Config.Security.TokensSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("error signing the token: %v", err)
	}

	return Token(signedTkn), expiresAt, nil
}

// Validate checks whether a token is valid or not
func (t *Token) Validate() bool {
	claims := &TokenClaims{}

	tkn, err := jwt.ParseWithClaims(t.String(), claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Config.Security.TokensSecret), nil
	})
	if err != nil || !tkn.Valid {
		return false
	}

	return true
}

// Renew renews the validity of the token
func (t *Token) Renew() (time.Time, error) {
	claims := &TokenClaims{}

	tkn, err := jwt.ParseWithClaims(t.String(), claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Config.Security.TokensSecret), nil
	})
	if err != nil || !tkn.Valid {
		if err != nil {
			// If the token has expired, but the user hasn't been modified since before the token was issued (the password hasn't changed), the token can be renewed
			if strings.HasPrefix(err.Error(), "token is expired by ") {
				if time.Since(claims.FirstIssued) > cfg.Config.Security.LoginLifespan {
					return time.Time{}, fmt.Errorf("error renewing the token: login lifespan exceeded, login again")
				}

				u := models.User{
					Username: claims.Usr,
				}

				if err = u.Load(); err != nil {
					return time.Time{}, fmt.Errorf("error renewing the token: %v", err)
				}

				if u.UpdatedAt.Before(time.Unix(claims.IssuedAt, 0)) {
					signedTkn, expiresAt, err := renew(claims)
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

	signedTkn, expiresAt, err := renew(claims)
	if err != nil {
		return time.Time{}, fmt.Errorf("error renewing the token: %v", err)
	}

	*t = Token(signedTkn)
	return expiresAt, nil
}

func renew(claims *TokenClaims) (string, time.Time, error) {
	expiresAt := time.Now().Add(cfg.Config.Security.TokensLifespan)
	claims.ExpiresAt = expiresAt.Unix()

	tkn := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signedTkn, err := tkn.SignedString([]byte(cfg.Config.Security.TokensSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("error singing the token: %v", err)
	}

	return signedTkn, expiresAt, nil
}

// String returns the token as a string
func (t Token) String() string {
	return string(t)
}
