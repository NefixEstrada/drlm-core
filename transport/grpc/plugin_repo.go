// SPDX-License-Identifier: AGPL-3.0-only

package grpc

import (
	"context"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PluginRepoAdd adds a new plugin repository
func (c *CoreServer) PluginRepoAdd(ctx context.Context, req *drlm.PluginRepoAddRequest) (*drlm.PluginRepoAddResponse, error) {
	return &drlm.PluginRepoAddResponse{}, status.Error(codes.Unimplemented, "not implemented yet")
}

// PluginRepoDelete removes a plugin repository
func (c *CoreServer) PluginRepoDelete(ctx context.Context, req *drlm.PluginRepoDeleteRequest) (*drlm.PluginRepoDeleteResponse, error) {
	return &drlm.PluginRepoDeleteResponse{}, status.Error(codes.Unimplemented, "not implemented yet")
}

// PluginRepoList lists the plugin repositories
func (c *CoreServer) PluginRepoList(ctx context.Context, req *drlm.PluginRepoListRequest) (*drlm.PluginRepoListResponse, error) {
	return &drlm.PluginRepoListResponse{}, status.Error(codes.Unimplemented, "not implemented yet")
}
