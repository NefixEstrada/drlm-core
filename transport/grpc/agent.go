package grpc

import (
	"context"
	"io"

	"github.com/brainupdaters/drlm-core/agent"
	"github.com/brainupdaters/drlm-core/models"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AgentAdd adds a new Agent to the DB
func (c *CoreServer) AgentAdd(ctx context.Context, req *drlm.AgentAddRequest) (*drlm.AgentAddResponse, error) {
	a := &models.Agent{}
	a.Host = req.Host
	a.Port = int(req.Port)

	if err := agent.Add(req.User, req.Password, req.IsAdmin, a); err != nil {
		return &drlm.AgentAddResponse{}, status.Errorf(codes.Unknown, "error adding the agent: %v", err)
	}

	return &drlm.AgentAddResponse{}, nil
}

// AgentInstall installs the agent binary to the agent machine
func (c *CoreServer) AgentInstall(stream drlm.DRLM_AgentInstallServer) error {
	var host string
	var f []byte

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				if err := agent.Install(host, f); err != nil {
					return status.Error(codes.Unknown, err.Error())
				}

				return stream.SendAndClose(&drlm.AgentInstallResponse{
					Code: drlm.AgentInstallResponse_OK,
				})
			}

			return status.Errorf(codes.Unknown, "unexpected error while reading chunks from the stream: %v", err)
		}

		host = req.Host
		f = append(f, req.Bin...)
	}
}

// AgentDelete removes the agent from the DB and might do a clenup in the agent machine
func (c *CoreServer) AgentDelete(ctx context.Context, req *drlm.AgentDeleteRequest) (*drlm.AgentDeleteResponse, error) {
	return &drlm.AgentDeleteResponse{}, status.Error(codes.Unimplemented, "not implemented yet")
}

// AgentList returns a list of all the agents
func (c *CoreServer) AgentList(ctx context.Context, req *drlm.AgentListRequest) (*drlm.AgentListResponse, error) {
	agents, err := models.AgentList()
	if err != nil {
		return &drlm.AgentListResponse{}, status.Error(codes.Unknown, err.Error())
	}

	rsp := &drlm.AgentListResponse{}
	for _, a := range agents {
		rsp.Agents = append(rsp.Agents, &drlm.AgentListResponse_Agent{
			Host:          a.Host,
			Port:          int32(a.Port),
			User:          a.User,
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

	if err := a.Load(); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return &drlm.AgentGetResponse{}, status.Error(codes.NotFound, "agent not found")
		}

		return &drlm.AgentGetResponse{}, status.Errorf(codes.Unknown, "error getting the agent from the DB: %v", err)
	}

	return &drlm.AgentGetResponse{
		Host:          a.Host,
		Port:          int32(a.Port),
		User:          a.User,
		Version:       a.Version,
		Arch:          drlm.Arch(a.Arch),
		Os:            drlm.OS(a.OS),
		OsVersion:     a.OSVersion,
		Distro:        a.Distro,
		DistroVersion: a.DistroVersion,
	}, nil
}
