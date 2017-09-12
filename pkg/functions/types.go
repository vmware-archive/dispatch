///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package functions

// NO TESTS

type F func(params map[string]interface{}) (map[string]interface{}, error)

type FuncProps struct {
	// Code is the function code, either as readable text or base64 encoded (for .zip and .jar archives)
	Code string
	// Main is the function's entry point (aka main function), by default "main"
	Main string
	// Image is the function's docker image
	Image string
}

type FaaSDriver interface {
	// Create creates (or updates, if is already exists) the function in the FaaS implementation.
	// name is an FQN of the function.
	// props completely define the function (Code is the function source code ).
	Create(name string, props FuncProps)

	// Delete deletes the function in the FaaS implementation.
	// name is the FQN of the function.
	Delete(name string)

	// GetRunnable returns a callable representation of a function.
	// name is the FQN of the function.
	GetRunnable(name string) F
}
