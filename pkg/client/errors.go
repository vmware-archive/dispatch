///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package client

import "fmt"

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

// ErrorBadRequest represents client-side request error
type ErrorBadRequest struct {
	baseError
}

// ErrorAlreadyExists represents error when resource already exists on the server
type ErrorAlreadyExists struct {
	baseError
}

// ErrorNotFound represents error of missing resource
type ErrorNotFound struct {
	baseError
}

// ErrorForbidden represents auth error
type ErrorForbidden struct {
	baseError
}
