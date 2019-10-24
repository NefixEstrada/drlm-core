package grpc

import (
	"context"

	"github.com/brainupdaters/drlm-core/models"
	"github.com/brainupdaters/drlm-core/scheduler"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// JobSchedule schedules a new job
func (c *CoreServer) JobSchedule(ctx context.Context, req *drlm.JobScheduleRequest) (*drlm.JobScheduleResponse, error) {
	if err := scheduler.AddJob(req.Name, req.AgentHost); err != nil {
		return &drlm.JobScheduleResponse{}, status.Error(codes.Unknown, err.Error())
	}

	return &drlm.JobScheduleResponse{}, nil
}

// JobCancel cancels an scheduled or running Job
func (c *CoreServer) JobCancel(ctx context.Context, req *drlm.JobCancelRequest) (*drlm.JobCancelResponse, error) {
	return &drlm.JobCancelResponse{}, status.Error(codes.Unimplemented, "not implemented yet")
}

// JobList returns a list with the the jobs of an agent. If the agent ID is 0, it will return all the jobs
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
				Name:      j.Name,
				AgentHost: j.AgentHost,
				Status:    drlm.JobStatus(j.Status),
				// Info: ,
			})
		}

		return rsp, nil
	}

	return &drlm.JobListResponse{}, status.Error(codes.Unimplemented, "not implemented yet")
}

// JobNotify notifies a change in a job (this is called from the agent) (status change, stdout...)
func (c *CoreServer) JobNotify(ctx context.Context, req *drlm.JobNotifyRequest) (*drlm.JobNotifyResponse, error) {
	return &drlm.JobNotifyResponse{}, status.Error(codes.Unimplemented, "not implemented yet")
}
