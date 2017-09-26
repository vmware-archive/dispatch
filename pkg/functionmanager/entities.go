///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package functionmanager

// NO TESTS

import (
	entitystore "gitlab.eng.vmware.com/serverless/serverless/pkg/entity-store"
)

// Function struct represents function entity that is stored in entity store
type Function struct {
	entitystore.BaseEntity
	Code      string `json:"code"`
	ImageName string `json:"image"`
	Schema    Schema `json:"schema"`
}

// Schema struct stores input and output validation schemas
type Schema struct {
	In  string `json:"in"`
	Out string `json:"out"`
}

// FnRun struct represents single function run
type FnRun struct {
	entitystore.BaseEntity
	FunctionName string                 `json:"functionName"`
	Blocking     bool                   `json:"blocking"`
	Input        map[string]interface{} `json:"input"`
	Output       map[string]interface{} `json:"output"`
	Secrets      []string               `json:"secrets"`
}
