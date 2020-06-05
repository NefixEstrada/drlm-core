// SPDX-License-Identifier: AGPL-3.0-only

package grpc

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/brainupdaters/drlm-core/agent"
	"github.com/brainupdaters/drlm-core/auth"
	"github.com/brainupdaters/drlm-core/models"
	"github.com/brainupdaters/drlm-core/plugin"

	"github.com/brainupdaters/drlm-common/pkg/os"
	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// AgentAdd adds a new Agent to the DB
func (c *CoreServer) AgentAdd(ctx context.Context, req *drlm.AgentAddRequest) (*drlm.AgentAddResponse, error) {
	a := &models.Agent{Host: req.Host}

	if err := agent.Add(c.ctx, a); err != nil {
		return &drlm.AgentAddResponse{}, status.Errorf(codes.Unknown, "error adding the agent: %v", err)
	}

	return &drlm.AgentAddResponse{}, nil
}

// AgentAccept accepts a request to join DRLM that has made an Agent
func (c *CoreServer) AgentAccept(ctx context.Context, req *drlm.AgentAcceptRequest) (*drlm.AgentAcceptResponse, error) {
	a := &models.Agent{Host: req.Host}

	if err := a.Load(c.ctx); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &drlm.AgentAcceptResponse{}, status.Error(codes.NotFound, "agent not found")
		}

		return &drlm.AgentAcceptResponse{}, status.Error(codes.Unknown, err.Error())
	}

	if err := agent.Accept(c.ctx, a); err != nil {
		if errors.Is(err, agent.ErrAgentAlreadyAccepted) {
			return &drlm.AgentAcceptResponse{}, status.Error(codes.FailedPrecondition, err.Error())
		}

		return &drlm.AgentAcceptResponse{}, status.Errorf(codes.Unknown, "error accepting the agent: %v", err)
	}

	// The check is done inside the agent.Accept method
	conn, _ := agent.Connections.Get(req.Host)
	if err := conn.Send(&drlm.AgentConnectionFromCore{
		MessageType: drlm.AgentConnectionFromCore_MESSAGE_TYPE_JOIN_RESPONSE,
		JoinResponse: &drlm.AgentConnectionFromCore_JoinResponse{
			Status:         drlm.AgentConnectionFromCore_JoinResponse_STATUS_ACCEPT,
			CoreSecret:     a.Secret,
			MinioAccessKey: a.MinioAccess(),
			MinioSecretKey: a.MinioKey,
		},
	}); err != nil {
		return &drlm.AgentAcceptResponse{}, status.Errorf(codes.Unknown, "error sending the credentials to the agent: %v", err)
	}

	return &drlm.AgentAcceptResponse{}, nil
}

// AgentInstall installs the agent binary to the agent machine
func (c *CoreServer) AgentInstall(stream drlm.DRLM_AgentInstallServer) error {
	var host string
	var sshPort int
	var sshUser string
	var sshPwd string
	var b []byte

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				a := &models.Agent{
					Host:    host,
					SSHPort: int(sshPort),
					SSHUser: sshUser,
				}

				if err := a.Load(c.ctx); err != nil {
					if gorm.IsRecordNotFoundError(err) {
						return status.Error(codes.NotFound, "agent not found")
					}

					return status.Errorf(codes.Unknown, "error loading the agent from the DB: %v", err)
				}

				if err := agent.Install(c.ctx, a, sshPwd, b); err != nil {
					return status.Error(codes.Unknown, err.Error())
				}

				return stream.SendAndClose(&drlm.AgentInstallResponse{
					Code: drlm.AgentInstallResponse_OK,
				})
			}

			return status.Errorf(codes.Unknown, "unexpected error while reading chunks from the stream: %v", err)
		}

		host = req.Host
		sshPort = int(req.SshPort)
		sshUser = req.SshUser
		sshPwd = req.SshPassword
		b = append(b, req.Bin...)
	}
}

// AgentDelete removes the agent from the DB and might do a clenup in the agent machine
func (c *CoreServer) AgentDelete(ctx context.Context, req *drlm.AgentDeleteRequest) (*drlm.AgentDeleteResponse, error) {
	return &drlm.AgentDeleteResponse{}, status.Error(codes.Unimplemented, "not implemented yet")
}

// AgentList returns a list of all the agents
func (c *CoreServer) AgentList(ctx context.Context, req *drlm.AgentListRequest) (*drlm.AgentListResponse, error) {
	agents, err := models.AgentList(c.ctx)
	if err != nil {
		return &drlm.AgentListResponse{}, status.Error(codes.Unknown, err.Error())
	}

	rsp := &drlm.AgentListResponse{}
	for _, a := range agents {
		rsp.Agents = append(rsp.Agents, &drlm.AgentListResponse_Agent{
			Host:          a.Host,
			SshPort:       int32(a.SSHPort),
			SshUser:       a.SSHUser,
			Version:       a.Version,
			Arch:          drlm.Arch(a.Arch),
			Os:            drlm.OS(a.OS),
			OsVersion:     a.OSVersion,
			Distro:        a.Distro,
			DistroVersion: a.DistroVersion,
		})
	}

	return rsp, nil
}

// AgentRequestList returns a list of all the agents that have requested to join DRLM
func (c *CoreServer) AgentRequestList(ctx context.Context, req *drlm.AgentRequestListRequest) (*drlm.AgentRequestListResponse, error) {
	agents, err := models.AgentRequestList(c.ctx)
	if err != nil {
		return &drlm.AgentRequestListResponse{}, status.Error(codes.Unknown, err.Error())
	}

	rsp := &drlm.AgentRequestListResponse{}
	for _, a := range agents {
		rsp.Agents = append(rsp.Agents, &drlm.AgentRequestListResponse_Agent{
			Host:          a.Host,
			Version:       a.Version,
			Arch:          drlm.Arch(a.Arch),
			Os:            drlm.OS(a.OS),
			OsVersion:     a.OSVersion,
			Distro:        a.Distro,
			DistroVersion: a.DistroVersion,
		})
	}

	return rsp, nil
}

// AgentGet returns a specific agent by host
func (c *CoreServer) AgentGet(ctx context.Context, req *drlm.AgentGetRequest) (*drlm.AgentGetResponse, error) {
	a := &models.Agent{
		Host: req.Host,
	}

	if err := a.Load(c.ctx); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return &drlm.AgentGetResponse{}, status.Error(codes.NotFound, "agent not found")
		}

		return &drlm.AgentGetResponse{}, status.Errorf(codes.Unknown, "error getting the agent from the DB: %v", err)
	}

	return &drlm.AgentGetResponse{
		Host:          a.Host,
		SshPort:       int32(a.SSHPort),
		SshUser:       a.SSHUser,
		Version:       a.Version,
		Arch:          drlm.Arch(a.Arch),
		Os:            drlm.OS(a.OS),
		OsVersion:     a.OSVersion,
		Distro:        a.Distro,
		DistroVersion: a.DistroVersion,
	}, nil
}

// AgentPluginAdd adds a new plugin to the Agent
func (c *CoreServer) AgentPluginAdd(stream drlm.DRLM_AgentPluginAddServer) error {
	var (
		host,
		repo,
		pName,
		version string
		arch []os.Arch
		pOS  []os.OS
		b    []byte
	)

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				a := &models.Agent{
					Host: host,
				}

				if err = a.Load(c.ctx); err != nil {
					if gorm.IsRecordNotFoundError(err) {
						return status.Error(codes.NotFound, "error adding the plugin: agent not found")
					}

					return status.Errorf(codes.Unknown, "error adding the plugin: %v", err)
				}

				if _, ok := agent.Connections.Get(a.Host); !ok {
					return status.Error(codes.Unavailable, "error adding the plugin: the agent is unavailable")
				}

				p := &models.Plugin{
					AgentHost: a.Host,
					Repo:      repo,
					Name:      pName,
					Version:   version,
					Arch:      arch,
					OS:        pOS,
				}

				if p.Add(c.ctx); err != nil {
					return status.Errorf(codes.Unknown, "error adding the plugin: %v", err)
				}

				if err := plugin.Install(c.ctx, a, p, b); err != nil {
					return status.Errorf(codes.Unknown, "error installing the plugin: %v", err)
				}

				return stream.SendAndClose(&drlm.AgentPluginAddResponse{})
			}
		}

		host = req.Host
		repo = req.Repo
		pName = req.Plugin
		version = req.Version
		arch = []os.Arch{}
		for _, a := range req.Arch {
			arch = append(arch, os.Arch(a))
		}
		pOS = []os.OS{}
		for _, o := range req.Os {
			pOS = append(pOS, os.OS(o))
		}
		b = append(b, req.Bin...)
	}
}

// AgentPluginRemove removes a plugin from the Agent
func (c *CoreServer) AgentPluginRemove(ctx context.Context, req *drlm.AgentPluginRemoveRequest) (*drlm.AgentPluginRemoveResponse, error) {
	return &drlm.AgentPluginRemoveResponse{}, status.Error(codes.Unimplemented, "not implemented yet")
}

// AgentPluginUpdate updates a plugin of the Agent
func (c *CoreServer) AgentPluginUpdate(stream drlm.DRLM_AgentPluginUpdateServer) error {
	return status.Error(codes.Unimplemented, "not implemented yet")
}

// AgentPluginList lists the plugins of the Agent
func (c *CoreServer) AgentPluginList(ctx context.Context, req *drlm.AgentPluginListRequest) (*drlm.AgentPluginListResponse, error) {
	return &drlm.AgentPluginListResponse{}, status.Error(codes.Unimplemented, "not implemented yet")
}

// AgentConnection creates the connection between the Agent and the Core. It's used for both notifying new jobs and for returning the response / updates of them
func (c *CoreServer) AgentConnection(stream drlm.DRLM_AgentConnectionServer) error {
	for {
		req, err := stream.Recv()

		var host string
		if req != nil && req.MessageType != drlm.AgentConnectionFromAgent_MESSAGE_TYPE_JOIN_REQUEST {
			md, ok := metadata.FromIncomingContext(stream.Context())
			if ok && len(md.Get("tkn")) > 0 {
				tkn := auth.Token(md.Get("tkn")[0])
				host, ok = tkn.ValidateAgent(c.ctx)
				if !ok {
					return status.Error(codes.InvalidArgument, "invalid token")
				}
			} else {
				return status.Error(codes.InvalidArgument, "unable to parse the token")
			}

		} else {
			if p, ok := peer.FromContext(stream.Context()); ok {
				host = strings.Split(p.Addr.String(), ":")[0]
			} else {
				return status.Error(codes.InvalidArgument, "unable to parse the agent host")
			}
		}

		if err != nil {
			if err == io.EOF {
				if _, ok := agent.Connections.Get(host); ok {
					agent.Connections.Delete(host)
				}

				return nil
			}
		}

		if req != nil {
			switch req.MessageType {
			case drlm.AgentConnectionFromAgent_MESSAGE_TYPE_JOIN_REQUEST:
				agent.PendingConnections.Add(host, stream)
				agent.AddRequest(c.ctx, &models.Agent{
					Host: host,
					Arch: os.Arch(req.JoinRequest.Arch),
					OS:   os.OS(req.JoinRequest.Os),
				})

			case drlm.AgentConnectionFromAgent_MESSAGE_TYPE_CONN_ESTABLISH:
				log.Infof("agent '%s' has established a connection", host)
				agent.Connections.Add(host, stream)

			case drlm.AgentConnectionFromAgent_MESSAGE_TYPE_JOB_UPDATE:
				j := &models.Job{
					Model: gorm.Model{
						ID: uint(req.JobUpdate.JobId),
					},
				}

				if err := j.Load(c.ctx); err != nil {
					if gorm.IsRecordNotFoundError(err) {
						return status.Error(codes.NotFound, "job not found")
					}

					return status.Errorf(codes.Unknown, "error loading the job: %v", err)
				}

				j.Status = models.JobStatus(req.JobUpdate.Status)
				j.Info += "\n" + req.JobUpdate.Info

				if err := j.Update(c.ctx); err != nil {
					return status.Errorf(codes.Unknown, "error updating the job: %v", err)
				}

			default:
				return status.Error(codes.InvalidArgument, "unknown message type")
			}
		}
	}
}
