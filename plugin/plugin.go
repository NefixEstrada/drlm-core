// SPDX-License-Identifier: AGPL-3.0-only

package plugin

import (
	"fmt"
	"os/user"
	"strings"

	"github.com/brainupdaters/drlm-common/pkg/os"
	"github.com/brainupdaters/drlm-common/pkg/os/client"
	"github.com/brainupdaters/drlm-common/pkg/ssh"
	"github.com/brainupdaters/drlm-core/models"
)

// Install installs a plugin on a Agent
func Install(p *models.Plugin, a *models.Agent, f []byte) error {
	if len(p.OS) != 0 {
		found := false
		for _, pOS := range p.OS {
			if pOS == a.OS {
				found = true
			}
		}

		if !found {
			return fmt.Errorf("unsupported os")
		}
	}

	if len(p.Arch) != 0 {
		found := false
		for _, arch := range p.Arch {
			if arch == a.Arch {
				found = true
			}
		}

		if !found {
			return fmt.Errorf("unsupported arch")
		}
	}

	u, err := user.Current()
	if err != nil {
		return fmt.Errorf("error getting the current user: %v", err)
	}

	coreCli := &client.Local{}
	coreOS, err := os.DetectOS(coreCli)
	if err != nil {
		return err
	}

	keysPath, err := coreOS.CmdSSHGetKeysPath(coreCli, u.Username)
	if err != nil {
		return err
	}

	keys := strings.Split(a.HostKeys, "|||")
	s, err := ssh.NewSessionWithKey(a.Host, a.Port, a.User, keysPath, keys)
	if err != nil {
		return fmt.Errorf("error opening the ssh session with the agent: %v", err)
	}
	defer s.Close()
	agentCli := &client.SSH{Session: s}

	if err := a.OS.CmdPkgInstallBinary(agentCli, a.User, fmt.Sprintf("drlm-plugin-%s-%s-%s", p.Repo, p.Name, p.Version), f); err != nil {
		return err
	}

	return nil
}
