///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package runner

import (
	"github.com/vmware/dispatch/pkg/functions"
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

func (r *impl) Run(fn *functions.Function, in interface{}) (interface{}, error) {
	f := r.Faas.GetRunnable(fn.Name)
	m := Compose(
		r.Validator.GetMiddleware(fn.Schemas),
		r.SecretInjector.GetMiddleware(fn.Secrets, fn.Cookie),
	)
	return m(f)(fn.Context, in)
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
