///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package runner

import (
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions"
)

type Config struct {
	Faas           functions.FaaSDriver
	Validator      functions.Validator
	SecretInjector functions.SecretInjector
}

type impl struct {
	Config
}

func New(config *Config) functions.Runner {
	return &impl{*config}
}

func (r *impl) Run(fn *functions.Function, args map[string]interface{}) (map[string]interface{}, error) {
	f := r.Faas.GetRunnable(fn.Name)
	m := Compose(
		r.Validator.GetMiddleware(fn.Schemas),
		r.SecretInjector.GetMiddleware(fn.Secrets, fn.Cookie),
	)
	return m(f)(args)
}

// Compose applies middleware so that:
// the first one is the outermost, the last one is the innermost (calls the actual function).
func Compose(ms ...functions.Middleware) functions.Middleware {
	return func(f functions.Runnable) functions.Runnable {
		for i := range ms {
			f = ms[len(ms)-1-i](f)
		}
		return f
	}
}
