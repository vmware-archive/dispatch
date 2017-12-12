///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package errors

// NO TEST

// DriverError declares a driver error
// it may caused by
//	- network connection error
// 	- driver adaptor system error
//	- driver side error
type DriverError struct {
	Err error `json:"err"`
}

// ObjectNotFoundError declares an object-not-found error
type ObjectNotFoundError struct {
	Err error `json:"err"`
}

// ObjectMarshalError declares an object marshal error
type ObjectMarshalError struct {
	Err error `json:"err"`
}

func (err *DriverError) Error() string {
	return err.Err.Error()
}

func (err *ObjectNotFoundError) Error() string {
	return err.Err.Error()
}

func (err *ObjectMarshalError) Error() string {
	return err.Err.Error()
}
