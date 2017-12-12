// Code generated by mockery v1.0.0
package mocks

import context "context"

import mock "github.com/stretchr/testify/mock"
import oauth2 "golang.org/x/oauth2"

// Config is an autogenerated mock type for the Config type
type Config struct {
	mock.Mock
}

// AuthCodeURL provides a mock function with given fields: _a0, _a1
func (_m *Config) AuthCodeURL(_a0 string, _a1 ...oauth2.AuthCodeOption) string {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, ...oauth2.AuthCodeOption) string); ok {
		r0 = rf(_a0, _a1...)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Exchange provides a mock function with given fields: _a0, _a1
func (_m *Config) Exchange(_a0 context.Context, _a1 string) (*oauth2.Token, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *oauth2.Token
	if rf, ok := ret.Get(0).(func(context.Context, string) *oauth2.Token); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*oauth2.Token)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
