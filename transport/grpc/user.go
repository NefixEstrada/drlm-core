package grpc

import (
	"context"

	"github.com/brainupdaters/drlm-core/models"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserLogin mocks the UserLogin gRPC method
func (c *CoreServer) UserLogin(ctx context.Context, req *drlm.UserLoginRequest) (*drlm.UserLoginResponse, error) {
	return &drlm.UserLoginResponse{}, nil
}

// UserAdd mocks the UserAdd gRPC method
func (c *CoreServer) UserAdd(ctx context.Context, req *drlm.UserAddRequest) (*drlm.UserAddResponse, error) {
	u := models.User{
		Username: req.Usr,
		Password: req.Pwd,
	}

	if err := u.Add(); err != nil {
		if models.IsErrUsrPwdStrength(err) {
			return &drlm.UserAddResponse{}, status.Error(codes.InvalidArgument, err.Error())
		}

		return &drlm.UserAddResponse{}, status.Error(codes.Unknown, err.Error())
	}

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
