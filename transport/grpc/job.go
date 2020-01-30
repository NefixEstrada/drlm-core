// SPDX-License-Identifier: AGPL-3.0-only

package grpc

import (
	"context"
	"time"

	"github.com/brainupdaters/drlm-core/models"
	"github.com/brainupdaters/drlm-core/scheduler"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// JobSchedule schedules a new job
func (c *CoreServer) JobSchedule(ctx context.Context, req *drlm.JobScheduleRequest) (*drlm.JobScheduleResponse, error) {
	var t time.Time
	if req.Time == nil {
		t = time.Now()
	} else {
		t = time.Unix(req.Time.Seconds, int64(req.Time.Nanos))
	}

	if err := scheduler.AddJob(req.AgentHost, req.Name, req.Config, t); err != nil {
		return &drlm.JobScheduleResponse{}, status.Error(codes.Unknown, err.Error())
	}

	return &drlm.JobScheduleResponse{}, nil
}

// JobCancel cancels an scheduled or running Job
func (c *CoreServer) JobCancel(ctx context.Context, req *drlm.JobCancelRequest) (*drlm.JobCancelResponse, error) {
	return &drlm.JobCancelResponse{}, status.Error(codes.Unimplemented, "not implemented yet")
}

// JobList returns a list with the the jobs of an agent. If the agent Host is "", it will return all the jobs
func (c *CoreServer) JobList(ctx context.Context, req *drlm.JobListRequest) (*drlm.JobListResponse, error) {
	if req.AgentHost == "" {
		jobs, err := models.JobList()
		if err != nil {
			return &drlm.JobListResponse{}, status.Error(codes.Unknown, err.Error())
		}

		rsp := &drlm.JobListResponse{}
		for _, j := range jobs {
			rsp.Jobs = append(rsp.Jobs, &drlm.JobListResponse_Job{
				Id:        uint32(j.ID),
				Name:      j.Plugin.String(),
				AgentHost: j.AgentHost,
				Status:    drlm.JobStatus(j.Status),
				// Info: ,
			})
		}

		return rsp, nil
	}

	a := &models.Agent{Host: req.AgentHost}

	if err := a.Load(); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return &drlm.JobListResponse{}, status.Error(codes.NotFound, "agent not found")
		}

		return &drlm.JobListResponse{}, status.Error(codes.Unknown, err.Error())
	}

	if err := a.LoadJobs(); err != nil {
		return &drlm.JobListResponse{}, status.Error(codes.Unknown, err.Error())
	}

	rsp := &drlm.JobListResponse{}
	for _, j := range a.Jobs {
		rsp.Jobs = append(rsp.Jobs, &drlm.JobListResponse_Job{
			Id: uint32(j.ID),
			// TODO: !!!!
			// Name:      j.Name,
			AgentHost: j.AgentHost,
			Status:    drlm.JobStatus(j.Status),
			// Info: ,
		})
	}

	return rsp, nil
}
