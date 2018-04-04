///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functions

// NO TESTS

// Context provides function context
type Context map[string]interface{}

// Message contains context and payload for function invocations and events
type Message struct {
	Context Context     `json:"context"`
	Payload interface{} `json:"payload"`
}

// Runnable is a runnable representation of a function
type Runnable func(ctx Context, in interface{}) (interface{}, error)

// Middleware allows injecting extra steps for each function execution
type Middleware func(f Runnable) Runnable

// Exec includes data required to execute a function
type Exec struct {
	// Code is the function code, either as readable text or base64 encoded (for .zip and .jar archives)
	Code string
	// Main is the function's entry point (aka main function), by default "main"
	Main string
	// Image is the function's docker image
	Image string
	// Name is the function's name
	Name string
}

// Schemas represent function validation schemas
type Schemas struct {
	// SchemaIn is the function input validation schema data structure. Can be nil.
	SchemaIn interface{}
	// SchemaOut is the function output validation schema data structure. Can be nil.
	SchemaOut interface{}
}

// FunctionExecution represents single instance of function execution
type FunctionExecution struct {
	Context Context
	RunID   string

	FunctionID string
	FaasID     string

	Schemas  *Schemas
	Secrets  []string
	Services []string
	Cookie   string
}

// FaaSDriver manages Serverless functions and allows to create or delete function,
// as well as to retrieve runnable representation of the function.
type FaaSDriver interface {
	// Create creates (or updates, if is already exists) the function in the FaaS implementation.
	// name is the name of the function.
	// exec defines the function implementation.
	Create(f *Function, exec *Exec) error

	// Delete deletes the function in the FaaS implementation.
	// f is a reference to function defition.
	Delete(f *Function) error

	// GetRunnable returns a callable representation of a function.
	// e is a reference to FunctionExecution.
	GetRunnable(e *FunctionExecution) Runnable
}

//go:generate mockery -name ImageBuilder -case underscore -dir .

// ImageBuilder builds a docker image for a serverless function.
type ImageBuilder interface {
	// BuildImage builds a function image and pushes it to the docker registry.
	// Returns image full name.
	BuildImage(faas, fnID string, e *Exec) (string, error)
}

// Runner knows how to execute a function
type Runner interface {
	Run(fn *FunctionExecution, in interface{}) (interface{}, error)
}

// Validator validates function input/output
type Validator interface {
	GetMiddleware(schemas *Schemas) Middleware
}

// SecretInjector injects secrets into function execution
type SecretInjector interface {
	GetMiddleware(secrets []string, services []string, cookie string) Middleware
}

// UserError represents user error
type UserError interface {
	AsUserErrorObject() interface{}
}

// FunctionError represents error caused by the function
type FunctionError interface {
	AsFunctionErrorObject() interface{}
}
