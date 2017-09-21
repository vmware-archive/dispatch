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
	Active   bool   `json:"enabled"`
	Code     string `json:"code"`
	Image    string `json:"image"`
	Language string `json:"language"`
	Schema   Schema `json:"schema"`
}

// Schema struct stores input and output validation schemas
type Schema struct {
	In  string `json:"in"`
	Out string `json:"out"`
}

// FnRun struct represents single function run
type FnRun struct {
	entitystore.BaseEntity
	FunctionID string   `json:"function_id"`
	Blocking   bool     `json:"blocking"`
	Output     string   `json:"output"`
	Arguments  string   `json:"arguments"`
	Secrets    []string `json:"secrets"`
}
