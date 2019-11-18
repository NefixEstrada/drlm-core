// SPDX-License-Identifier: AGPL-3.0-only

package secret

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// New creates a secret
func New(id string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(id), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error generating the secret hash: %v", err)
	}

	hasher := md5.New()
	hasher.Write(h)

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
