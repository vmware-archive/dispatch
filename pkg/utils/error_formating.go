///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package utils

import (
	"fmt"

	"github.com/go-openapi/swag"
)

// ErrorMsgAlreadyExists creates an error message for resource that already exists
func ErrorMsgAlreadyExists(kind, name string) *string {
	return swag.String(fmt.Sprintf("%s %s already exists", kind, name))
}

// ErrorMsgNotFound creates an error message for resource that was not found
func ErrorMsgNotFound(kind, name string) *string {
	return swag.String(fmt.Sprintf("%s %s not found", kind, name))
}

// ErrorMsgInternalError creates an error message for internal error
func ErrorMsgInternalError(kind, name string) *string {
	return swag.String(fmt.Sprintf("internal error when processing %s %s", kind, name))
}

// ErrorMsgBadRequest creates an error message for bad request error
func ErrorMsgBadRequest(kind, name string, err error) *string {
	return swag.String(fmt.Sprintf("bad request when processing %s %s: %s", kind, name, err.Error()))
}
