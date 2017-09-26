///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package functions

// NO TESTS

type Runnable func(args map[string]interface{}) (map[string]interface{}, error)
type Middleware func(f Runnable) Runnable

type Exec struct {
	// Code is the function code, either as readable text or base64 encoded (for .zip and .jar archives)
	Code string
	// Main is the function's entry point (aka main function), by default "main"
	Main string
	// Image is the function's docker image
	Image string
}

type Schemas struct {
	// SchemaIn is the function's input object validation schema FQN. It is optional.
	SchemaIn string
	// SchemaOut is the function's input object validation schema FQN. It is optional.
	SchemaOut string
}

type Function struct {
	Name string

	Exec    *Exec
	Schemas *Schemas
}

type FaaSDriver interface {
	// Create creates (or updates, if is already exists) the function in the FaaS implementation.
	// name is the name of the function.
	// exec defines the function implementation.
	Create(name string, exec *Exec) error

	// Delete deletes the function in the FaaS implementation.
	// name is the name of the function.
	Delete(name string) error

	// GetRunnable returns a callable representation of a function.
	// name is the name of the function.
	GetRunnable(name string) Runnable
}

type Runner interface {
	Run(fn *Function, args map[string]interface{}) (map[string]interface{}, error)
}

type Validator interface {
	GetMiddleware(schemas *Schemas) Middleware
}
