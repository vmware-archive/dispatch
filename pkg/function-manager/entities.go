///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functionmanager

// NO TESTS

import (
	"github.com/go-openapi/spec"

	entitystore "github.com/vmware/dispatch/pkg/entity-store"
)

// Function struct represents function entity that is stored in entity store
type Function struct {
	entitystore.BaseEntity
	Code      string   `json:"code"`
	Main      string   `json:"main"`
	ImageName string   `json:"image"`
	Schema    *Schema  `json:"schema,omitempty"`
	Secrets   []string `json:"secrets,omitempty"`
}

// Schema struct stores input and output validation schemas
type Schema struct {
	In  *spec.Schema `json:"in,omitempty"`
	Out *spec.Schema `json:"out,omitempty"`
}

// FnRun struct represents single function run
type FnRun struct {
	entitystore.BaseEntity
	FunctionName string      `json:"functionName"`
	Blocking     bool        `json:"blocking"`
	Input        interface{} `json:"input,omitempty"`
	Output       interface{} `json:"output,omitempty"`
	Secrets      []string    `json:"secrets,omitempty"`
	Logs         []string    `json:"logs,omitempty"`
}
