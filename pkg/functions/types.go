///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package functions

// NO TESTS

type Context map[string]interface{}

type Runnable func(ctx Context, in interface{}) (interface{}, error)
type Middleware func(f Runnable) Runnable

type Exec struct {
	// Code is the function code, either as readable text or base64 encoded (for .zip and .jar archives)
	Code string
	// Main is the function's entry point (aka main function), by default "main"
	Main string
	// Image is the function's docker image
	Image string
	// Language is the function's runtime language
	Language string
}

type Schemas struct {
	// SchemaIn is the function input validation schema data structure. Can be nil.
	SchemaIn interface{}
	// SchemaOut is the function output validation schema data structure. Can be nil.
	SchemaOut interface{}
}

type FunctionExecution struct {
	Context Context

	Name string
	ID   string

	Schemas *Schemas
	Secrets []string
	Cookie  string
}

type FaaSDriver interface {
	// Create creates (or updates, if is already exists) the function in the FaaS implementation.
	// name is the name of the function.
	// exec defines the function implementation.
	Create(f *Function, exec *Exec) error

	// Delete deletes the function in the FaaS implementation.
	// name is the name of the function.
	Delete(f *Function) error

	// GetRunnable returns a callable representation of a function.
	// name is the name of the function.
	GetRunnable(e *FunctionExecution) Runnable
}

type Runner interface {
	Run(fn *FunctionExecution, in interface{}) (interface{}, error)
}

type Validator interface {
	GetMiddleware(schemas *Schemas) Middleware
}

type SecretInjector interface {
	GetMiddleware(secrets []string, cookie string) Middleware
}

type UserError interface {
	AsUserErrorObject() interface{}
}

type FunctionError interface {
	AsFunctionErrorObject() interface{}
}
