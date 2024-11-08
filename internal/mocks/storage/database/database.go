// Code generated by MockGen. DO NOT EDIT.
// Source: F:\Projects\Go\src\goph-keeper\internal\core\keeper\keeper.go
//
// Generated by this command:
//
//	mockgen -source=F:\Projects\Go\src\goph-keeper\internal\core\keeper\keeper.go -package=database
//

// Package database is a generated GoMock package.
package database

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"

	models "github.com/playmixer/secret-keeper/internal/adapter/models"
)

// MockStorage is a mock of Storage interface.
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage.
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance.
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// DelSecret mocks base method.
func (m *MockStorage) DelSecret(ctx context.Context, id uint) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DelSecret", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// DelSecret indicates an expected call of DelSecret.
func (mr *MockStorageMockRecorder) DelSecret(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DelSecret", reflect.TypeOf((*MockStorage)(nil).DelSecret), ctx, id)
}

// GetMetaDatasByUserID mocks base method.
func (m *MockStorage) GetMetaDatasByUserID(ctx context.Context, userID uint) (*[]models.Secret, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMetaDatasByUserID", ctx, userID)
	ret0, _ := ret[0].(*[]models.Secret)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMetaDatasByUserID indicates an expected call of GetMetaDatasByUserID.
func (mr *MockStorageMockRecorder) GetMetaDatasByUserID(ctx, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMetaDatasByUserID", reflect.TypeOf((*MockStorage)(nil).GetMetaDatasByUserID), ctx, userID)
}

// GetSecret mocks base method.
func (m *MockStorage) GetSecret(ctx context.Context, id uint) (*models.Secret, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSecret", ctx, id)
	ret0, _ := ret[0].(*models.Secret)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSecret indicates an expected call of GetSecret.
func (mr *MockStorageMockRecorder) GetSecret(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSecret", reflect.TypeOf((*MockStorage)(nil).GetSecret), ctx, id)
}

// GetUserByLogin mocks base method.
func (m *MockStorage) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByLogin", ctx, login)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByLogin indicates an expected call of GetUserByLogin.
func (mr *MockStorageMockRecorder) GetUserByLogin(ctx, login any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByLogin", reflect.TypeOf((*MockStorage)(nil).GetUserByLogin), ctx, login)
}

// NewSecret mocks base method.
func (m *MockStorage) NewSecret(ctx context.Context, secret *models.Secret) (*models.Secret, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewSecret", ctx, secret)
	ret0, _ := ret[0].(*models.Secret)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewSecret indicates an expected call of NewSecret.
func (mr *MockStorageMockRecorder) NewSecret(ctx, secret any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewSecret", reflect.TypeOf((*MockStorage)(nil).NewSecret), ctx, secret)
}

// Registration mocks base method.
func (m *MockStorage) Registration(ctx context.Context, login, passwordHash string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Registration", ctx, login, passwordHash)
	ret0, _ := ret[0].(error)
	return ret0
}

// Registration indicates an expected call of Registration.
func (mr *MockStorageMockRecorder) Registration(ctx, login, passwordHash any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Registration", reflect.TypeOf((*MockStorage)(nil).Registration), ctx, login, passwordHash)
}

// UpdSecret mocks base method.
func (m *MockStorage) UpdSecret(ctx context.Context, secret *models.Secret) (*models.Secret, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdSecret", ctx, secret)
	ret0, _ := ret[0].(*models.Secret)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdSecret indicates an expected call of UpdSecret.
func (mr *MockStorageMockRecorder) UpdSecret(ctx, secret any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdSecret", reflect.TypeOf((*MockStorage)(nil).UpdSecret), ctx, secret)
}
