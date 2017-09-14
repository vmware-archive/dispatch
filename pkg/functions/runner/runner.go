///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package runner

import (
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions"
)

type impl struct {
	faas      functions.FaaSDriver
	validator functions.Validator
}

func New(faas functions.FaaSDriver, validator functions.Validator) functions.Runner {
	return &impl{
		faas:      faas,
		validator: validator,
	}
}

func (r *impl) RunFunction(fn *functions.Function, args map[string]interface{}) (map[string]interface{}, error) {
	f := r.faas.GetRunnable(fn.Name)
	m := Compose(r.validator.GetMiddleware(fn.Schemas))
	return m(f)(args)
}

// Compose applies middleware so that:
// the first one is the outermost, the last one is the innermost (calls the actual function).
func Compose(ms ...functions.Middleware) functions.Middleware {
	return func(f functions.F) functions.F {
		for i := range ms {
			f = ms[len(ms)-1-i](f)
		}
		return f
	}
}
