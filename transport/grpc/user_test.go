package grpc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/brainupdaters/drlm-core/transport/grpc"
	"github.com/brainupdaters/drlm-core/utils/tests"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
	mocket "github.com/selvatico/go-mocket"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUserAdd(t *testing.T) {
	assert := assert.New(t)

	t.Run("should add the new user", func(t *testing.T) {
		tests.GenerateDB(t)
		mocket.Catcher.NewMock().WithQuery(`INSERT INTO "users" ("created_at","updated_at","deleted_at","username","password") VALUES(?,?,?,?,?)`).WithReply([]map[string]interface{}{})

		ctx := context.Background()
		req := &drlm.UserAddRequest{
			Usr: "nefix",
			Pwd: "f0cKT3rF$",
		}

		c := grpc.CoreServer{}
		rsp, err := c.UserAdd(ctx, req)

		assert.Nil(err)
		assert.Equal(&drlm.UserAddResponse{}, rsp)
	})

	t.Run("should return an error if the password is too weak", func(t *testing.T) {
		ctx := context.Background()
		req := &drlm.UserAddRequest{
			Usr: "nefix",
			Pwd: "",
		}

		c := grpc.CoreServer{}
		rsp, err := c.UserAdd(ctx, req)

		assert.Equal(err, status.Error(codes.InvalidArgument, "the password requires, at least, a length of 8 characters"))
		assert.Equal(&drlm.UserAddResponse{}, rsp)
	})

	t.Run("should return an error if there's an error adding the user to the DB", func(t *testing.T) {
		tests.GenerateDB(t)
		mocket.Catcher.NewMock().WithQuery(`INSERT  INTO "users" ("created_at","updated_at","deleted_at","username","password") VALUES (?,?,?,?,?)`).WithError(errors.New("testing error"))

		ctx := context.Background()
		req := &drlm.UserAddRequest{
			Usr: "nefix",
			Pwd: "f0cKT3rF$",
		}

		c := grpc.CoreServer{}
		rsp, err := c.UserAdd(ctx, req)

		assert.Equal(err, status.Error(codes.Unknown, "error adding the user to the DB: testing error"))
		assert.Equal(&drlm.UserAddResponse{}, rsp)
	})
}
