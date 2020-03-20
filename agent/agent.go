// SPDX-License-Identifier: AGPL-3.0-only

package agent

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/minio"
	"github.com/brainupdaters/drlm-core/models"

	"github.com/brainupdaters/drlm-common/pkg/os"
	"github.com/brainupdaters/drlm-common/pkg/os/client"
	"github.com/brainupdaters/drlm-common/pkg/ssh"
	"github.com/spf13/afero"
)

// Developer notes: all the agent minio usernames are `drlm-agent-ID`
// all the agent minio binary buckets are `drlm-agent-ID-bin`

// Add connects to the Agent host, creates the drlm user and copies the keys to that user, which has to be admin
func Add(ctx *context.Context, a *models.Agent) error {
	a.Accepted = true

	usr := fmt.Sprintf("drlm-agent-%d", a.ID)

	var err error
	if a.MinioKey, err = minio.CreateUser(ctx, usr); err != nil {
		return fmt.Errorf("error creating the agent minio user: %v", err)
	}

	if _, err := minio.MakeBucketForUser(ctx, usr, usr+"-bin"); err != nil {
		return fmt.Errorf("error creating the agent binary buclet: %v", err)
	}

	// Add the Agent to the DB
	if err := a.Add(ctx); err != nil {
		return err
	}

	return nil
}

// AddRequest adds a new agent request
func AddRequest(ctx *context.Context, a *models.Agent) error {
	a.Accepted = false

	if err := a.Add(ctx); err != nil {
		return err
	}

	return nil
}

// Install installs the agent binary, sets up the daemon and config and starts the service
func Install(ctx *context.Context, a *models.Agent, sshPwd string, f []byte) error {
	// Set default values
	if a.SSHPort == 0 {
		a.SSHPort = 22
	}

	coreCli := &client.Local{}
	coreOS, err := os.DetectOS(coreCli)
	if err != nil {
		return err
	}

	keys, err := coreOS.CmdSSHGetHostKeys(coreCli, a.Host, a.SSHPort)
	if err != nil {
		return err
	}
	a.SSHHostKeys = strings.Join(keys, "|||")

	if err := a.Update(ctx); err != nil {
		return fmt.Errorf("error updating the agent in the DB: %v", err)
	}

	// Connect to the host through user and password
	s, err := ssh.NewSessionWithPassword(a.Host, a.SSHPort, a.SSHUser, sshPwd, keys)
	if err != nil {
		return err
	}

	agentCli := &client.SSH{Session: s}

	a.OS, err = os.DetectOS(agentCli)
	if err != nil {
		return err
	}

	a.Arch, err = os.DetectArch(agentCli)
	if err != nil {
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

	pubKey, err := afero.ReadFile(ctx.FS, filepath.Join(keysPath, "id_rsa.pub"))
	if err != nil {
		return fmt.Errorf("error reading the Core SSH public key: %v", err)
	}

	if err := a.OS.CmdSSHCopyID(agentCli, a.SSHUser, pubKey); err != nil {
		return err
	}

	s.Close()

	// Connect to the host through user and key
	s, err = ssh.NewSessionWithKey(ctx.FS, a.Host, a.SSHPort, a.SSHUser, keysPath, keys)
	if err != nil {
		return fmt.Errorf("error opening the ssh session with the agent: %v", err)
	}
	defer s.Close()

	agentCli.Session = s

	if err := a.OS.CmdPkgInstallBinary(agentCli, a.SSHUser, "drlm-agent", f); err != nil {
		return fmt.Errorf("error installing DRLM Agent: %v", err)
	}

	if err := a.OS.CmdPkgWriteConfig(agentCli, a.SSHUser, "agent.toml", []byte(fmt.Sprintf(`[core]
secret = "%s"

[minio]
access_key = "%s"
secret_key = "%s"`, a.Secret, "drlm-agent-"+strconv.Itoa(int(a.ID)), a.MinioKey))); err != nil {
		return fmt.Errorf("error configuring the DRLM Agent: %v", err)
	}

	return nil
}

// Sync updates the agent OS information, and all the plugins specific info such as OS, OS version, program versions...
func Sync(ctx *context.Context, a *models.Agent) error {
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

	keys := strings.Split(a.SSHHostKeys, "|||")
	s, err := ssh.NewSessionWithKey(ctx.FS, a.Host, a.SSHPort, a.SSHUser, keysPath, keys)
	if err != nil {
		return fmt.Errorf("error opening the ssh session with the agent: %v", err)
	}
	defer s.Close()
	c := &client.SSH{Session: s}

	a.OS, err = os.DetectOS(c)
	if err != nil {
		return err
	}

	a.OSVersion, err = a.OS.DetectVersion(c)
	if err != nil {
		return err
	}

	a.Distro, a.DistroVersion, err = a.OS.DetectDistro(c)
	if err != nil {
		return err
	}

	a.Arch, err = os.DetectArch(c)
	if err != nil {
		return err
	}

	return nil
}
