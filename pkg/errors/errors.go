///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package errors

import (
	"net/http"

	"github.com/go-openapi/swag"
	dapi "github.com/vmware/dispatch/pkg/api/v1"
)

// NO TEST

// ErrorReason is a string type
type ErrorReason string

// ReasonDriver represents a [event] driver error
const ReasonDriver ErrorReason = "DriverError"

// ReasonObjectMarshal represents a marshalling error (usually bad request)
const ReasonObjectMarshal ErrorReason = "ObjectMarshallError"

// ReasonObjectNotFound represents a generic not found error
const ReasonObjectNotFound ErrorReason = "ObjectNotFound"

// ReasonRequest represents some user/request error
const ReasonRequest ErrorReason = "RequestError"

// ReasonServer represents a server error (usually unknown)
const ReasonServer ErrorReason = "ServerError"

// ReasonUnknown represents an unknown error
const ReasonUnknown ErrorReason = "UnknownError"

// ReasonedError has an associated ErrorReason
type ReasonedError interface {
	Reason() ErrorReason
}

// DispatchError is an error with a reason and an APIError
type DispatchError struct {
	APIError    dapi.Error  `json:"error"`
	ErrorReason ErrorReason `json:"reason"`
}

// Error implements the Error interface.
func (e *DispatchError) Error() string {
	return *e.APIError.Message
}

// Reason allows access to e's reason without having to know the detailed workings
// of DispatchError.
func (e *DispatchError) Reason() ErrorReason {
	return e.ErrorReason
}

// GetError returns the associated APIError
func GetError(err error) *dapi.Error {
	if derr, ok := err.(*DispatchError); ok {
		return &derr.APIError
	}
	return nil
}

// ReasonForError returns the HTTP status for a particular error.
func ReasonForError(err error) ErrorReason {
	switch t := err.(type) {
	case ReasonedError:
		return t.Reason()
	}
	return ReasonUnknown
}

// NewDriverError creates a new DriverError
func NewDriverError(err error) *DispatchError {
	return &DispatchError{
		APIError: dapi.Error{
			Message: swag.String(err.Error()),
			Code:    http.StatusInternalServerError,
		},
		ErrorReason: ReasonDriver,
	}
}

// IsDriverError returns true if the specified error was created by NewDriverError.
func IsDriverError(err error) bool {
	return ReasonForError(err) == ReasonDriver
}

// NewObjectNotFoundError creates a new ObjectNotFoundError
func NewObjectNotFoundError(err error) *DispatchError {
	return &DispatchError{
		APIError: dapi.Error{
			Message: swag.String(err.Error()),
			Code:    http.StatusNotFound,
		},
		ErrorReason: ReasonObjectNotFound,
	}
}

// IsObjectNotFound returns true if the specified error was created by NewObjectNotFoundError.
func IsObjectNotFound(err error) bool {
	return ReasonForError(err) == ReasonObjectNotFound
}

// NewObjectMarshalError creates a new ObjectMarshalError
func NewObjectMarshalError(err error) *DispatchError {
	return &DispatchError{
		APIError: dapi.Error{
			Message: swag.String(err.Error()),
			Code:    http.StatusBadRequest,
		},
		ErrorReason: ReasonObjectMarshal,
	}
}

// IsObjectMarshall returns true if the specified error was created by NewObjectMarshalError.
func IsObjectMarshall(err error) bool {
	return ReasonForError(err) == ReasonObjectMarshal
}

// NewServerError creates a new ServerError
func NewServerError(err error) *DispatchError {
	return &DispatchError{
		APIError: dapi.Error{
			Message: swag.String(err.Error()),
			Code:    http.StatusInternalServerError,
		},
		ErrorReason: ReasonServer,
	}
}

// IsServer returns true if the specified error was created by NewServerError.
func IsServer(err error) bool {
	return ReasonForError(err) == ReasonServer
}

// NewRequestError creates a new RequestError
func NewRequestError(err error) *DispatchError {
	return &DispatchError{
		APIError: dapi.Error{
			Message: swag.String(err.Error()),
			Code:    http.StatusBadRequest,
		},
		ErrorReason: ReasonRequest,
	}
}

// IsRequest returns true if the specified error was created by NewRequestError.
func IsRequest(err error) bool {
	return ReasonForError(err) == ReasonRequest
}
