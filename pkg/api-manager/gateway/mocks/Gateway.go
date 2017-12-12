// Code generated by mockery v1.0.0
package mocks

import gateway "github.com/vmware/dispatch/pkg/api-manager/gateway"
import mock "github.com/stretchr/testify/mock"

// Gateway is an autogenerated mock type for the Gateway type
type Gateway struct {
	mock.Mock
}

// AddAPI provides a mock function with given fields: api
func (_m *Gateway) AddAPI(api *gateway.API) (*gateway.API, error) {
	ret := _m.Called(api)

	var r0 *gateway.API
	if rf, ok := ret.Get(0).(func(*gateway.API) *gateway.API); ok {
		r0 = rf(api)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gateway.API)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*gateway.API) error); ok {
		r1 = rf(api)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteAPI provides a mock function with given fields: api
func (_m *Gateway) DeleteAPI(api *gateway.API) error {
	ret := _m.Called(api)

	var r0 error
	if rf, ok := ret.Get(0).(func(*gateway.API) error); ok {
		r0 = rf(api)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAPI provides a mock function with given fields: name
func (_m *Gateway) GetAPI(name string) (*gateway.API, error) {
	ret := _m.Called(name)

	var r0 *gateway.API
	if rf, ok := ret.Get(0).(func(string) *gateway.API); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gateway.API)
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

// Initialize provides a mock function with given fields:
func (_m *Gateway) Initialize() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateAPI provides a mock function with given fields: name, api
func (_m *Gateway) UpdateAPI(name string, api *gateway.API) (*gateway.API, error) {
	ret := _m.Called(name, api)

	var r0 *gateway.API
	if rf, ok := ret.Get(0).(func(string, *gateway.API) *gateway.API); ok {
		r0 = rf(name, api)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gateway.API)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, *gateway.API) error); ok {
		r1 = rf(name, api)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
