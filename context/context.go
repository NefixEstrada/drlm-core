// SPDX-License-Identifier: AGPL-3.0-only

package context

import (
	"context"
	"sync"
	"time"

	"github.com/brainupdaters/drlm-core/cfg/types"

	"github.com/jinzhu/gorm"
	"github.com/minio/minio-go/v6"
	"github.com/minio/minio/pkg/madmin"
	"github.com/spf13/afero"
)

// Context has the "global" values of the app (such as the config, the DB connection...). It also implements context.Context
type Context struct {
	ctx context.Context
	WG  sync.WaitGroup

	FS            afero.Fs
	Cfg           *types.DRLMCoreConfig
	DB            *gorm.DB
	MinioCli      *minio.Client
	MinioAdminCli *madmin.AdminClient
}

// Deadline implements context.Context.Deadline
func (ctx *Context) Deadline() (deadline time.Time, ok bool) {
	return ctx.ctx.Deadline()
}

// Done implements context.Context.Done
func (ctx *Context) Done() <-chan struct{} {
	return ctx.ctx.Done()
}

// Err implements context.Context.Err
func (ctx *Context) Err() error {
	return ctx.ctx.Err()
}

// Value implements context.Context.Value
func (ctx *Context) Value(key interface{}) interface{} {
	return ctx.ctx.Value(key)
}

// Background creates a new context with context.Background as the inner context
func Background() *Context {
	ctx := &Context{}
	ctx.ctx = context.Background()

	return ctx
}

// TODO creates a new context with context.TODO as the inner context
func TODO() *Context {
	ctx := &Context{}
	ctx.ctx = context.TODO()

	return ctx
}

// WithCancel creates a new context with context.WithCancel as the inner context
func WithCancel() (*Context, context.CancelFunc) {
	ctx := &Context{}
	cancelCtx, cancel := context.WithCancel(context.Background())
	ctx.ctx = cancelCtx

	return ctx, cancel
}
