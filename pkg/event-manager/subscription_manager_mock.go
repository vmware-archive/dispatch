///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by mockery v1.0.0

package eventmanager

import (
	mock "github.com/stretchr/testify/mock"
)

// SubscriptionManagerMock is an autogenerated mock type for the SubscriptionManagerMock type
type SubscriptionManagerMock struct {
	mock.Mock
}

// Create provides a mock function with given fields: _a0
func (_m *SubscriptionManagerMock) Create(_a0 *Subscription) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*Subscription) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: _a0
func (_m *SubscriptionManagerMock) Delete(_a0 *Subscription) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*Subscription) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Run provides a mock function with given fields: _a0
func (_m *SubscriptionManagerMock) Run(_a0 []*Subscription) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func([]*Subscription) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
