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

func (err *DriverError) Error() string {
	return err.Err.Error()
}

// ObjectNotFoundError declares an object-not-found error
type ObjectNotFoundError struct {
	Err error `json:"err"`
}

func (err *ObjectNotFoundError) Error() string {
	return err.Err.Error()
}

// ObjectMarshalError declares an object marshal error
type ObjectMarshalError struct {
	Err error `json:"err"`
}

func (err *ObjectMarshalError) Error() string {
	return err.Err.Error()
}

// RequestError indicates a user/client error
type RequestError struct {
	Err error `json:"err"`
}

func (err *RequestError) Error() string {
	return err.Err.Error()
}

// NewRequestError creates a new RequestError
func NewRequestError(err error) *RequestError {
	return &RequestError{Err: err}
}

// ServerError indicates a unexpected error on the server side
type ServerError struct {
	Err error `json:"err"`
}

func (err *ServerError) Error() string {
	return err.Err.Error()
}

// NewServerError creates a new ServerError
func NewServerError(err error) *ServerError {
	return &ServerError{Err: err}
}
