package grpc

import (
	"context"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
)

// UserLogin mocks the UserLogin gRPC method
func (c *CoreServer) UserLogin(ctx context.Context, req *drlm.UserLoginRequest) (*drlm.UserLoginResponse, error) {
	return &drlm.UserLoginResponse{}, nil
}

// UserAdd mocks the UserAdd gRPC method
func (c *CoreServer) UserAdd(ctx context.Context, req *drlm.UserAddRequest) (*drlm.UserAddResponse, error) {
	return &drlm.UserAddResponse{}, nil
}

// UserDelete mocks the UserDelete gRPC method
func (c *CoreServer) UserDelete(ctx context.Context, req *drlm.UserDeleteRequest) (*drlm.UserDeleteResponse, error) {
	return &drlm.UserDeleteResponse{}, nil
}

// UserList mocks the UserList gRPC method
func (c *CoreServer) UserList(ctx context.Context, req *drlm.UserListRequest) (*drlm.UserListResponse, error) {
	return &drlm.UserListResponse{}, nil
}
