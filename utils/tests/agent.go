// SPDX-License-Identifier: AGPL-3.0-only

package tests

import (
	"context"
	"errors"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

var errNotImplemented = errors.New("mock method not implemented")

// AgentConnectionServerMock is a mock for the Agent Connection gRPC server
type AgentConnectionServerMock struct {
	mock.Mock
}

// Send mocks the AgentConnectionServer Send method
func (a *AgentConnectionServerMock) Send(req *drlm.AgentConnectionFromCore) error {
	args := a.Called(req)
	return args.Error(0)
}

// Recv mocks the AgentConnnectionServer Recv method
func (a *AgentConnectionServerMock) Recv() (*drlm.AgentConnectionFromAgent, error) {
	return nil, errNotImplemented
}

// SetHeader mocks the grpc.ServerStream SetHeader method
func (a *AgentConnectionServerMock) SetHeader(metadata.MD) error {
	return errNotImplemented
}

// SendHeader mocks the grpc.ServerStream SendHeader method
func (a *AgentConnectionServerMock) SendHeader(metadata.MD) error {
	return errNotImplemented
}

// SetTrailer mocks the grpc.ServerStream SetTrailer method
func (a *AgentConnectionServerMock) SetTrailer(metadata.MD) {}

// Context mocks the grpc.ServerStream Context method
func (a *AgentConnectionServerMock) Context() context.Context {
	return context.TODO()
}

// SendMsg mocks the grpc.ServerStream SendMsg method
func (a *AgentConnectionServerMock) SendMsg(m interface{}) error {
	return errNotImplemented
}

// RecvMsg mocks the grpc.ServerStream RecvMsg method
func (a *AgentConnectionServerMock) RecvMsg(m interface{}) error {
	return errNotImplemented
}
