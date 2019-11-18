// SPDX-License-Identifier: AGPL-3.0-only

package grpc

import (
	"context"

	"github.com/brainupdaters/drlm-core/auth"
	"github.com/brainupdaters/drlm-core/auth/types"
	"github.com/brainupdaters/drlm-core/models"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UserLogin logs in users and returns tokens
func (c *CoreServer) UserLogin(ctx context.Context, req *drlm.UserLoginRequest) (*drlm.UserLoginResponse, error) {
	tkn, expiresAt, err := auth.LoginLocal(req.Usr, req.Pwd)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return &drlm.UserLoginResponse{}, status.Errorf(codes.NotFound, `error logging in: user "%s" not found`, req.Usr)
		}

		if err == bcrypt.ErrMismatchedHashAndPassword {
			return &drlm.UserLoginResponse{}, status.Error(codes.Unauthenticated, "error logging in: incorrect password")
		}

		return &drlm.UserLoginResponse{}, status.Errorf(codes.Unknown, "error logging in: %v", err)
	}

	return &drlm.UserLoginResponse{
		Tkn: tkn.String(),
		TknExpiration: &timestamp.Timestamp{
			Seconds: expiresAt.Unix(),
		},
	}, nil
}

// UserTokenRenew renews the token of the user
func (c *CoreServer) UserTokenRenew(ctx context.Context, req *drlm.UserTokenRenewRequest) (*drlm.UserTokenRenewResponse, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if len(md.Get("tkn")) > 0 {
			tkn := auth.Token(md.Get("tkn")[0])
			expiresAt, err := tkn.Renew()
			if err != nil {
				return &drlm.UserTokenRenewResponse{}, status.Error(codes.Unknown, err.Error())
			}

			return &drlm.UserTokenRenewResponse{
				Tkn:           tkn.String(),
				TknExpiration: &timestamp.Timestamp{Seconds: expiresAt.Unix()},
			}, nil
		}
	}

	return &drlm.UserTokenRenewResponse{}, status.Error(codes.Unauthenticated, "not authenticated")
}

// UserAdd creates new users in the DB
func (c *CoreServer) UserAdd(ctx context.Context, req *drlm.UserAddRequest) (*drlm.UserAddResponse, error) {
	u := models.User{
		Username: req.Usr,
		Password: req.Pwd,
		AuthType: types.Local,
	}

	if err := u.Add(); err != nil {
		if models.IsErrUsrPwdStrength(err) {
			return &drlm.UserAddResponse{}, status.Error(codes.InvalidArgument, err.Error())
		}

		return &drlm.UserAddResponse{}, status.Error(codes.Unknown, err.Error())
	}

	return &drlm.UserAddResponse{}, nil
}

// UserDelete deletes an user from the DB
func (c *CoreServer) UserDelete(ctx context.Context, req *drlm.UserDeleteRequest) (*drlm.UserDeleteResponse, error) {
	u := models.User{
		Username: req.Usr,
	}

	if err := u.Delete(); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return &drlm.UserDeleteResponse{}, status.Errorf(codes.NotFound, `error deleting the user "%s": not found`, req.Usr)
		}

		return &drlm.UserDeleteResponse{}, status.Errorf(codes.Unknown, `error deleting the user "%s": %v`, req.Usr, err)
	}

	return &drlm.UserDeleteResponse{}, nil
}

// UserList lists all the users from the DB
func (c *CoreServer) UserList(ctx context.Context, req *drlm.UserListRequest) (*drlm.UserListResponse, error) {
	users, err := models.UserList()
	if err != nil {
		return &drlm.UserListResponse{}, status.Error(codes.Unknown, err.Error())
	}

	rsp := &drlm.UserListResponse{}
	for _, u := range users {
		usr := &drlm.UserListResponse_User{
			Usr:       u.Username,
			AuthType:  parseAuthType(u.AuthType),
			CreatedAt: &timestamp.Timestamp{Seconds: u.CreatedAt.Unix()},
			UpdatedAt: &timestamp.Timestamp{Seconds: u.UpdatedAt.Unix()},
		}

		rsp.Users = append(rsp.Users, usr)
	}

	return rsp, nil
}
