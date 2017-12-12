// Code generated by mockery v1.0.0
package mocks

import mock "github.com/stretchr/testify/mock"
import models "github.com/vmware/dispatch/pkg/secret-store/gen/models"

// SecretsService is an autogenerated mock type for the SecretsService type
type SecretsService struct {
	mock.Mock
}

// AddSecret provides a mock function with given fields: secret
func (_m *SecretsService) AddSecret(secret models.Secret) (*models.Secret, error) {
	ret := _m.Called(secret)

	var r0 *models.Secret
	if rf, ok := ret.Get(0).(func(models.Secret) *models.Secret); ok {
		r0 = rf(secret)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(models.Secret) error); ok {
		r1 = rf(secret)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteSecret provides a mock function with given fields: name
func (_m *SecretsService) DeleteSecret(name string) error {
	ret := _m.Called(name)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetSecret provides a mock function with given fields: name
func (_m *SecretsService) GetSecret(name string) (*models.Secret, error) {
	ret := _m.Called(name)

	var r0 *models.Secret
	if rf, ok := ret.Get(0).(func(string) *models.Secret); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSecrets provides a mock function with given fields:
func (_m *SecretsService) GetSecrets() ([]*models.Secret, error) {
	ret := _m.Called()

	var r0 []*models.Secret
	if rf, ok := ret.Get(0).(func() []*models.Secret); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateSecret provides a mock function with given fields: secret
func (_m *SecretsService) UpdateSecret(secret models.Secret) (*models.Secret, error) {
	ret := _m.Called(secret)

	var r0 *models.Secret
	if rf, ok := ret.Get(0).(func(models.Secret) *models.Secret); ok {
		r0 = rf(secret)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(models.Secret) error); ok {
		r1 = rf(secret)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
