///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functions

import (
	"context"

	"github.com/pkg/errors"
)

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

// Schemas represent function validation schemas
type Schemas struct {
	// SchemaIn is the function input validation schema data structure. Can be nil.
	SchemaIn interface{}
	// SchemaOut is the function output validation schema data structure. Can be nil.
	SchemaOut interface{}
}

// FunctionExecution represents single instance of function execution
type FunctionExecution struct {
	Context        Context
	OrganizationID string
	RunID          string

	FunctionID string
	FaasID     string

	Schemas  *Schemas
	Secrets  []string
	Services []string
	Cookie   string
}

//go:generate mockery -name FaaSDriver -case underscore -dir . -note "CLOSE THIS FILE AS QUICKLY AS POSSIBLE"

// FaaSDriver manages Serverless functions and allows to create or delete function,
// as well as to retrieve runnable representation of the function.
type FaaSDriver interface {
	// Create creates (or updates, if is already exists) the function in the FaaS implementation.
	// name is the name of the function.
	// exec defines the function implementation.
	Create(ctx context.Context, f *Function) error

	// Delete deletes the function in the FaaS implementation.
	// f is a reference to function defition.
	Delete(ctx context.Context, f *Function) error

	// GetRunnable returns a callable representation of a function.
	// e is a reference to FunctionExecution.
	GetRunnable(e *FunctionExecution) Runnable
}

// FunctionResources Memory and CPU
type FunctionResources struct {
	Memory string `json:"memory"`
	CPU    string `json:"cpu"`
}

//go:generate mockery -name ImageBuilder -case underscore -dir . -note "CLOSE THIS FILE AS QUICKLY AS POSSIBLE"

// ImageBuilder builds or removes a docker image for a serverless function.
type ImageBuilder interface {
	// BuildImage builds a function image and pushes it to the docker registry.
	// Returns image full name.
	BuildImage(ctx context.Context, f *Function, code []byte) (string, error)

	// RemoveImage removes a function image from the docker host
	RemoveImage(ctx context.Context, f *Function) error
}

// Runner knows how to execute a function
type Runner interface {
	Run(fn *FunctionExecution, in interface{}) (interface{}, error)
}

// Validator validates function input/output
type Validator interface {
	GetMiddleware(schemas *Schemas) Middleware
}

//go:generate mockery -name SecretInjector -case underscore -dir . -note "CLOSE THIS FILE AS QUICKLY AS POSSIBLE"

// SecretInjector injects secrets into function execution
type SecretInjector interface {
	GetMiddleware(organizationID string, secrets []string, cookie string) Middleware
}

//go:generate mockery -name ServiceInjector -case underscore -dir . -note "CLOSE THIS FILE AS QUICKLY AS POSSIBLE"

// ServiceInjector injects service bindings into function execution
type ServiceInjector interface {
	GetMiddleware(organizationID string, services []string, cookie string) Middleware
}

// InputError represents user/input error
type InputError interface {
	AsInputErrorObject() interface{}
}

// FunctionError represents error caused by the function
type FunctionError interface {
	AsFunctionErrorObject() interface{}
}

// SystemError represents error in the Dispatch infrastructure
type SystemError interface {
	AsSystemErrorObject() interface{}
}

// StackTracer is part of the errors pkg public API and returns the error stacktrace
type StackTracer interface {
	StackTrace() errors.StackTrace
}
