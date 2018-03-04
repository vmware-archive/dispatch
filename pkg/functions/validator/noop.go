///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package validator

// NO TESTS

import "github.com/vmware/dispatch/pkg/functions"

type noOp struct {
}

// NoOp is a NoOp validator
func NoOp() functions.Validator {
	return &noOp{}
}

func (*noOp) GetMiddleware(schemas *functions.Schemas) functions.Middleware {
	return func(f functions.Runnable) functions.Runnable {
		return f
	}
}
