// SPDX-License-Identifier: AGPL-3.0-only

package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"

	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/minio"
)

// GenerateMinio generates a test minio server
func GenerateMinio(ctx *context.Context, f http.Handler) *httptest.Server {
	ts := httptest.NewServer(f)

	addr := strings.Split(ts.Listener.Addr().String(), ":")
	port, err := strconv.Atoi(addr[len(addr)-1])
	if err != nil {
		panic(fmt.Sprintf("mino test: invalid port: %s", addr[len(addr)-1]))
	}

	ctx.Cfg.Minio.Port = port

	minio.Init(ctx)

	return ts
}
