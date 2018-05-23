///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package injectors

import (
	"github.com/vmware/dispatch/pkg/functions"
)

// StripContextMiddleware creates a middleware that cleans context from keys that should not be passed to the function.
func StripContextMiddleware() functions.Middleware {
	return func(f functions.Runnable) functions.Runnable {
		return func(ctx functions.Context, in interface{}) (interface{}, error) {
			delete(ctx, functions.GoContext)
			return f(ctx, in)
		}
	}
}
