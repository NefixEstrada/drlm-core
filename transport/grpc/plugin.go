// SPDX-License-Identifier: AGPL-3.0-only

package grpc

import (
	"context"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PluginList returns a list of all the available plugins
func (c *CoreServer) PluginList(ctx context.Context, req *drlm.PluginListRequest) (*drlm.PluginListResponse, error) {
	return &drlm.PluginListResponse{}, status.Error(codes.Unimplemented, "not implemented yet")
}
