///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package validator

// NO TESTS

import "gitlab.eng.vmware.com/serverless/serverless/pkg/functions"

type noOp struct {
}

func NoOp() functions.Validator {
	return &noOp{}
}

func (*noOp) GetMiddleware(schemas *functions.Schemas) functions.Middleware {
	return func(f functions.Runnable) functions.Runnable {
		return f
	}
}
