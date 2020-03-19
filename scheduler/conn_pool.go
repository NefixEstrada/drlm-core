// SPDX-License-Identifier: AGPL-3.0-only

package scheduler

import (
	"sync"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
)

var (
	// AgentConnections are all the active agent connections
	AgentConnections = connPool{v: map[string]drlm.DRLM_AgentConnectionServer{}}
	// PendingAgentConnections are all the active connections from agents that havent been accepted yet
	PendingAgentConnections = connPool{v: map[string]drlm.DRLM_AgentConnectionServer{}}
)

type connPool struct {
	v   map[string]drlm.DRLM_AgentConnectionServer
	mux sync.Mutex
}

func (c *connPool) Get(agent string) (stream drlm.DRLM_AgentConnectionServer, ok bool) {
	c.mux.Lock()
	defer c.mux.Unlock()
	stream, ok = c.v[agent]
	return
}

func (c *connPool) Add(agent string, stream drlm.DRLM_AgentConnectionServer) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.v[agent] = stream
}

func (c *connPool) Delete(agent string) {
	c.mux.Lock()
	defer c.mux.Unlock()

	delete(c.v, agent)
}
