// SPDX-License-Identifier: AGPL-3.0-only

package plugin

import (
	"bytes"
	"fmt"
	"log"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
	"github.com/brainupdaters/drlm-core/agent"
	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/models"
	"github.com/minio/minio-go/v6"
)

// Install installs a plugin on a Agent
func Install(ctx *context.Context, a *models.Agent, p *models.Plugin, b []byte) error {
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

	bName := fmt.Sprintf("drlm-agent-%d-bin", a.ID)
	pName := fmt.Sprintf("drlm-plugin-%s-%s-%s", p.Repo, p.Name, p.Version)

	if _, err := ctx.MinioCli.PutObject(bName, pName, bytes.NewReader(b), -1, minio.PutObjectOptions{}); err != nil {
		return fmt.Errorf("error uploading the plugin to minio: %v", err)
	}

	stream, ok := agent.Connections.Get(a.Host)
	if !ok {
		return fmt.Errorf("agent unavailable")
	}

	log.Println("sending msg")

	if err := stream.Send(&drlm.AgentConnectionFromCore{
		MessageType: drlm.AgentConnectionFromCore_MESSAGE_TYPE_INSTALL_BINARY,
		InstallBinary: &drlm.AgentConnectionFromCore_InstallBinary{
			Bucket: bName,
			Name:   pName,
		},
	}); err != nil {
		log.Printf("error!!!!: %v", err)
		return err
	}

	return nil
}
