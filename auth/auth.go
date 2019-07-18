package auth

import (
	"github.com/brainupdaters/drlm-core/cfg"

	log "github.com/sirupsen/logrus"
)

// Init checks whether the tokens secret is set and secure or not
func Init() {
	if cfg.Config.Security.TokensSecret == "" {
		log.Fatal(`you need to set a tokens secret in the configuration. You can generate one with "< /dev/urandom tr -dc 'A-Za-z0-9!"#$%&'\''()*+,-./:;<=>?@[\]^_{|}~' | head -c${1:-64};echo;"`)
	}

	if len(cfg.Config.Security.TokensSecret) < 32 {
		log.Fatal("the tokens secret needs to be at least 32 characters long")
	}
}
