///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package client

import (
	"fmt"

	"github.com/vmware/dispatch/pkg/api/v1"
)

// Error is an interface implemented by all errors declared here.
type Error interface {
	Error() string
	Message() string
	Code() int
}

type baseError struct {
	code    int
	message string
}

func (b baseError) Error() string {
	return fmt.Sprintf("[Code: %d] %s", b.code, b.message)
}

func (b baseError) Message() string {
	return b.message
}

func (b baseError) Code() int {
	return b.code
}

// ErrorServerUnknownError represents unknown server error
type ErrorServerUnknownError struct {
	baseError
}

// NewErrorServerUnknownError creates new instance of ErrorServerUnknownError based on Error Model
func NewErrorServerUnknownError(apiError *v1.Error) *ErrorServerUnknownError {
	return &ErrorServerUnknownError{
		baseError: baseErrFromModel(apiError),
	}
}

// ErrorBadRequest represents client-side request error
type ErrorBadRequest struct {
	baseError
}

// NewErrorBadRequest creates new instance of ErrorBadRequest based on Error Model
func NewErrorBadRequest(apiError *v1.Error) *ErrorBadRequest {
	return &ErrorBadRequest{
		baseError: baseErrFromModel(apiError),
	}
}

// ErrorAlreadyExists represents error when resource already exists on the server
type ErrorAlreadyExists struct {
	baseError
}

// NewErrorAlreadyExists creates new instance of  ErrorAlreadyExists based on Error Model
func NewErrorAlreadyExists(apiError *v1.Error) *ErrorAlreadyExists {
	return &ErrorAlreadyExists{
		baseError: baseErrFromModel(apiError),
	}
}

// ErrorNotFound represents error of missing resource
type ErrorNotFound struct {
	baseError
}

// NewErrorNotFound creates new instance of ErrorNotFound based on Error Model
func NewErrorNotFound(apiError *v1.Error) *ErrorNotFound {
	return &ErrorNotFound{
		baseError: baseErrFromModel(apiError),
	}
}

// ErrorForbidden represents authz error
type ErrorForbidden struct {
	baseError
}

// NewErrorForbidden creates new instance of ErrorForbidden based on Error Model
func NewErrorForbidden(apiError *v1.Error) *ErrorForbidden {
	return &ErrorForbidden{
		baseError: baseErrFromModel(apiError),
	}
}

// ErrorUnauthorized represents authn error
type ErrorUnauthorized struct {
	baseError
}

// NewErrorUnauthorized creates new instance of ErrorUnauthorized based on Error Model
func NewErrorUnauthorized(apiError *v1.Error) *ErrorUnauthorized {
	return &ErrorUnauthorized{
		baseError: baseErrFromModel(apiError),
	}
}

// ErrorInvalidInput represents error of request input being invalid
type ErrorInvalidInput struct {
	baseError
}

// NewErrorInvalidInput creates new instance of ErrorInvalidInput based on Error Model
func NewErrorInvalidInput(apiError *v1.Error) *ErrorInvalidInput {
	return &ErrorInvalidInput{
		baseError: baseErrFromModel(apiError),
	}
}

// ErrorFunctionError represents error that happened when executing the function
type ErrorFunctionError struct {
	baseError
}

// NewErrorFunctionError creates new instance of ErrorFunctionError based on Error Model
func NewErrorFunctionError(apiError *v1.Error) *ErrorFunctionError {
	return &ErrorFunctionError{
		baseError: baseErrFromModel(apiError),
	}
}

func baseErrFromModel(apiError *v1.Error) baseError {
	message := ""
	if apiError.Message != nil {
		message = *apiError.Message
	}
	return baseError{
		code:    int(apiError.Code),
		message: message,
	}
}
