// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/qdm12/goservices (interfaces: Service,Hooks)

// Package goservices is a generated GoMock package.
package goservices

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockService is a mock of Service interface.
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
}

// MockServiceMockRecorder is the mock recorder for MockService.
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance.
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// Start mocks base method.
func (m *MockService) Start(arg0 context.Context) (<-chan error, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start", arg0)
	ret0, _ := ret[0].(<-chan error)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Start indicates an expected call of Start.
func (mr *MockServiceMockRecorder) Start(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockService)(nil).Start), arg0)
}

// Stop mocks base method.
func (m *MockService) Stop() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop")
	ret0, _ := ret[0].(error)
	return ret0
}

// Stop indicates an expected call of Stop.
func (mr *MockServiceMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockService)(nil).Stop))
}

// String mocks base method.
func (m *MockService) String() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "String")
	ret0, _ := ret[0].(string)
	return ret0
}

// String indicates an expected call of String.
func (mr *MockServiceMockRecorder) String() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "String", reflect.TypeOf((*MockService)(nil).String))
}

// MockHooks is a mock of Hooks interface.
type MockHooks struct {
	ctrl     *gomock.Controller
	recorder *MockHooksMockRecorder
}

// MockHooksMockRecorder is the mock recorder for MockHooks.
type MockHooksMockRecorder struct {
	mock *MockHooks
}

// NewMockHooks creates a new mock instance.
func NewMockHooks(ctrl *gomock.Controller) *MockHooks {
	mock := &MockHooks{ctrl: ctrl}
	mock.recorder = &MockHooksMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHooks) EXPECT() *MockHooksMockRecorder {
	return m.recorder
}

// OnCrash mocks base method.
func (m *MockHooks) OnCrash(arg0 string, arg1 error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "OnCrash", arg0, arg1)
}

// OnCrash indicates an expected call of OnCrash.
func (mr *MockHooksMockRecorder) OnCrash(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnCrash", reflect.TypeOf((*MockHooks)(nil).OnCrash), arg0, arg1)
}

// OnStart mocks base method.
func (m *MockHooks) OnStart(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "OnStart", arg0)
}

// OnStart indicates an expected call of OnStart.
func (mr *MockHooksMockRecorder) OnStart(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnStart", reflect.TypeOf((*MockHooks)(nil).OnStart), arg0)
}

// OnStarted mocks base method.
func (m *MockHooks) OnStarted(arg0 string, arg1 error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "OnStarted", arg0, arg1)
}

// OnStarted indicates an expected call of OnStarted.
func (mr *MockHooksMockRecorder) OnStarted(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnStarted", reflect.TypeOf((*MockHooks)(nil).OnStarted), arg0, arg1)
}

// OnStop mocks base method.
func (m *MockHooks) OnStop(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "OnStop", arg0)
}

// OnStop indicates an expected call of OnStop.
func (mr *MockHooksMockRecorder) OnStop(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnStop", reflect.TypeOf((*MockHooks)(nil).OnStop), arg0)
}

// OnStopped mocks base method.
func (m *MockHooks) OnStopped(arg0 string, arg1 error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "OnStopped", arg0, arg1)
}

// OnStopped indicates an expected call of OnStopped.
func (mr *MockHooksMockRecorder) OnStopped(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnStopped", reflect.TypeOf((*MockHooks)(nil).OnStopped), arg0, arg1)
}
