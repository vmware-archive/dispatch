// Code generated by mockery v1.0.0
package mocks

import functions "github.com/vmware/dispatch/pkg/functions"
import mock "github.com/stretchr/testify/mock"

// SecretInjector is an autogenerated mock type for the SecretInjector type
type SecretInjector struct {
	mock.Mock
}

// GetMiddleware provides a mock function with given fields: secrets, cookie
func (_m *SecretInjector) GetMiddleware(secrets []string, cookie string) functions.Middleware {
	ret := _m.Called(secrets, cookie)

	var r0 functions.Middleware
	if rf, ok := ret.Get(0).(func([]string, string) functions.Middleware); ok {
		r0 = rf(secrets, cookie)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(functions.Middleware)
		}
	}

	return r0
}
