package agent

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/brainupdaters/drlm-core/models"

	"github.com/brainupdaters/drlm-common/pkg/fs"
	"github.com/brainupdaters/drlm-common/pkg/os"
	"github.com/brainupdaters/drlm-common/pkg/os/client"
	"github.com/brainupdaters/drlm-common/pkg/ssh"
	"github.com/spf13/afero"
)

// Add connects to the Agent host, creates the drlm user and copies the keys to that user, which has to be admin
func Add(usr, pwd string, isAdmin bool, a *models.Agent) error {
	// Set default values
	if a.Port == 0 {
		a.Port = 22
	}

	if a.User == "" {
		a.User = "drlm"
	}

	coreCli := &client.Local{}
	coreOS, err := os.DetectOS(coreCli)
	if err != nil {
		return err
	}

	keys, err := coreOS.CmdSSHGetHostKeys(coreCli, a.Host, a.Port)
	if err != nil {
		return err
	}

	a.HostKeys = strings.Join(keys, "|||")

	// Connect to the host
	s, err := ssh.NewSessionWithPassword(a.Host, a.Port, usr, pwd, keys)
	if err != nil {
		return err
	}
	defer s.Close()

	agentCli := &client.SSH{
		Session: s,
	}

	a.OS, err = os.DetectOS(agentCli)
	if err != nil {
		return err
	}

	a.Arch, err = os.DetectArch(agentCli)
	if err != nil {
		return err
	}

	if err := a.OS.CmdUserCreate(agentCli, a.User, "changeme"); err != nil {
		return err
	}

	u, err := user.Current()
	if err != nil {
		return err
	}

	keysPath, err := coreOS.CmdSSHGetKeysPath(coreCli, u.Username)
	if err != nil {
		return err
	}

	pubKey, err := afero.ReadFile(fs.FS, filepath.Join(keysPath, "id_rsa.pub"))
	if err != nil {
		return fmt.Errorf("error reading the Core SSH public key: %v", err)
	}

	if err := a.OS.CmdSSHCopyID(agentCli, a.User, pubKey); err != nil {
		return err
	}

	if err := a.OS.CmdUserDisable(agentCli, a.User); err != nil {
		return err
	}

	if err := a.OS.CmdUserMakeAdmin(agentCli, a.User); err != nil {
		return err
	}

	return nil
}

// Sync updates the agent OS information, and all the plugins specific info such as OS, OS version, program versions...
// func Sync(a *models.Agent) error {
// 	s, err := ssh.NewSessionWithKey(a.Host, a.Port, a.User, cfg.Config.Security.SSHKeysPath, a.HostKeys)
// 	if err != nil {
// 		return err
// 	}
// 	defer s.Close()

// 	c := &os.ClientSSH{Session: s}
// 	a.OS = os.DetectOS(c)

// 	a.OSVersion, err = a.OS.DetectVersion(c)
// 	if err != nil {
// 		return err
// 	}

// 	a.Distro, a.DistroVersion, err = a.OS.DetectDistro(c)
// 	if err != nil {
// 		return err
// 	}

// 	a.Arch, err = os.DetectArch(c)
// 	if err != nil {
// 		return err
// 	}

// 	// TODO: Call plugin updates

// 	return nil
// }
