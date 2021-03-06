// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/go-connector/services/decryption (interfaces: Decryptor)

// Package mockdecryptor is a generated GoMock package.
package mockdecryptor

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	go_chainscript "github.com/stratumn/go-chainscript"
	decryption "github.com/stratumn/go-connector/services/decryption"
	reflect "reflect"
)

// MockDecryptor is a mock of Decryptor interface
type MockDecryptor struct {
	ctrl     *gomock.Controller
	recorder *MockDecryptorMockRecorder
}

// MockDecryptorMockRecorder is the mock recorder for MockDecryptor
type MockDecryptorMockRecorder struct {
	mock *MockDecryptor
}

// NewMockDecryptor creates a new mock instance
func NewMockDecryptor(ctrl *gomock.Controller) *MockDecryptor {
	mock := &MockDecryptor{ctrl: ctrl}
	mock.recorder = &MockDecryptorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDecryptor) EXPECT() *MockDecryptorMockRecorder {
	return m.recorder
}

// DecryptLink mocks base method
func (m *MockDecryptor) DecryptLink(arg0 context.Context, arg1 *go_chainscript.Link) error {
	ret := m.ctrl.Call(m, "DecryptLink", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DecryptLink indicates an expected call of DecryptLink
func (mr *MockDecryptorMockRecorder) DecryptLink(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DecryptLink", reflect.TypeOf((*MockDecryptor)(nil).DecryptLink), arg0, arg1)
}

// DecryptLinkData mocks base method
func (m *MockDecryptor) DecryptLinkData(arg0 context.Context, arg1 []byte, arg2 []*decryption.Recipient) ([]byte, error) {
	ret := m.ctrl.Call(m, "DecryptLinkData", arg0, arg1, arg2)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DecryptLinkData indicates an expected call of DecryptLinkData
func (mr *MockDecryptorMockRecorder) DecryptLinkData(arg0, arg1, arg2 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DecryptLinkData", reflect.TypeOf((*MockDecryptor)(nil).DecryptLinkData), arg0, arg1, arg2)
}

// DecryptLinks mocks base method
func (m *MockDecryptor) DecryptLinks(arg0 context.Context, arg1 []*go_chainscript.Link) error {
	ret := m.ctrl.Call(m, "DecryptLinks", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DecryptLinks indicates an expected call of DecryptLinks
func (mr *MockDecryptorMockRecorder) DecryptLinks(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DecryptLinks", reflect.TypeOf((*MockDecryptor)(nil).DecryptLinks), arg0, arg1)
}
