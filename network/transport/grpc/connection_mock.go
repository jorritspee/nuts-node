// Code generated by MockGen. DO NOT EDIT.
// Source: network/transport/grpc/connection.go

// Package grpc is a generated GoMock package.
package grpc

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	transport "github.com/nuts-foundation/nuts-node/network/transport"
	grpc "google.golang.org/grpc"
)

// MockConnection is a mock of Connection interface.
type MockConnection struct {
	ctrl     *gomock.Controller
	recorder *MockConnectionMockRecorder
}

// MockConnectionMockRecorder is the mock recorder for MockConnection.
type MockConnectionMockRecorder struct {
	mock *MockConnection
}

// NewMockConnection creates a new mock instance.
func NewMockConnection(ctrl *gomock.Controller) *MockConnection {
	mock := &MockConnection{ctrl: ctrl}
	mock.recorder = &MockConnectionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConnection) EXPECT() *MockConnectionMockRecorder {
	return m.recorder
}

// IsConnected mocks base method.
func (m *MockConnection) IsConnected() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsConnected")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsConnected indicates an expected call of IsConnected.
func (mr *MockConnectionMockRecorder) IsConnected() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsConnected", reflect.TypeOf((*MockConnection)(nil).IsConnected))
}

// IsProtocolConnected mocks base method.
func (m *MockConnection) IsProtocolConnected(protocol Protocol) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsProtocolConnected", protocol)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsProtocolConnected indicates an expected call of IsProtocolConnected.
func (mr *MockConnectionMockRecorder) IsProtocolConnected(protocol interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsProtocolConnected", reflect.TypeOf((*MockConnection)(nil).IsProtocolConnected), protocol)
}

// Peer mocks base method.
func (m *MockConnection) Peer() transport.Peer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Peer")
	ret0, _ := ret[0].(transport.Peer)
	return ret0
}

// Peer indicates an expected call of Peer.
func (mr *MockConnectionMockRecorder) Peer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Peer", reflect.TypeOf((*MockConnection)(nil).Peer))
}

// Send mocks base method.
func (m *MockConnection) Send(protocol Protocol, envelope interface{}, ignoreSoftLimit bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", protocol, envelope, ignoreSoftLimit)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockConnectionMockRecorder) Send(protocol, envelope, ignoreSoftLimit interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockConnection)(nil).Send), protocol, envelope, ignoreSoftLimit)
}

// disconnect mocks base method.
func (m *MockConnection) disconnect() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "disconnect")
}

// disconnect indicates an expected call of disconnect.
func (mr *MockConnectionMockRecorder) disconnect() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "disconnect", reflect.TypeOf((*MockConnection)(nil).disconnect))
}

// outboundConnector mocks base method.
func (m *MockConnection) outboundConnector() *outboundConnector {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "outboundConnector")
	ret0, _ := ret[0].(*outboundConnector)
	return ret0
}

// outboundConnector indicates an expected call of outboundConnector.
func (mr *MockConnectionMockRecorder) outboundConnector() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "outboundConnector", reflect.TypeOf((*MockConnection)(nil).outboundConnector))
}

// registerStream mocks base method.
func (m *MockConnection) registerStream(protocol Protocol, stream Stream) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "registerStream", protocol, stream)
	ret0, _ := ret[0].(bool)
	return ret0
}

// registerStream indicates an expected call of registerStream.
func (mr *MockConnectionMockRecorder) registerStream(protocol, stream interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "registerStream", reflect.TypeOf((*MockConnection)(nil).registerStream), protocol, stream)
}

// setPeer mocks base method.
func (m *MockConnection) setPeer(peer transport.Peer) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "setPeer", peer)
}

// setPeer indicates an expected call of setPeer.
func (mr *MockConnectionMockRecorder) setPeer(peer interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "setPeer", reflect.TypeOf((*MockConnection)(nil).setPeer), peer)
}

// startConnecting mocks base method.
func (m *MockConnection) startConnecting(config connectorConfig, backoff Backoff, callback func(*grpc.ClientConn) bool) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "startConnecting", config, backoff, callback)
}

// startConnecting indicates an expected call of startConnecting.
func (mr *MockConnectionMockRecorder) startConnecting(config, backoff, callback interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "startConnecting", reflect.TypeOf((*MockConnection)(nil).startConnecting), config, backoff, callback)
}

// stopConnecting mocks base method.
func (m *MockConnection) stopConnecting() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "stopConnecting")
}

// stopConnecting indicates an expected call of stopConnecting.
func (mr *MockConnectionMockRecorder) stopConnecting() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "stopConnecting", reflect.TypeOf((*MockConnection)(nil).stopConnecting))
}

// verifyOrSetPeerID mocks base method.
func (m *MockConnection) verifyOrSetPeerID(id transport.PeerID) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "verifyOrSetPeerID", id)
	ret0, _ := ret[0].(bool)
	return ret0
}

// verifyOrSetPeerID indicates an expected call of verifyOrSetPeerID.
func (mr *MockConnectionMockRecorder) verifyOrSetPeerID(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "verifyOrSetPeerID", reflect.TypeOf((*MockConnection)(nil).verifyOrSetPeerID), id)
}

// waitUntilDisconnected mocks base method.
func (m *MockConnection) waitUntilDisconnected() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "waitUntilDisconnected")
}

// waitUntilDisconnected indicates an expected call of waitUntilDisconnected.
func (mr *MockConnectionMockRecorder) waitUntilDisconnected() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "waitUntilDisconnected", reflect.TypeOf((*MockConnection)(nil).waitUntilDisconnected))
}
