///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functions

// NO TESTS

import (
	"time"

	"github.com/go-openapi/spec"
	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/function-manager/gen/models"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/trace"
)

// Function struct represents function entity that is stored in entity store
type Function struct {
	entitystore.BaseEntity
	FaasID    string   `json:"faasId"`
	Code      string   `json:"code"`
	Main      string   `json:"main"`
	ImageName string   `json:"image"`
	Schema    *Schema  `json:"schema,omitempty"`
	Secrets   []string `json:"secrets,omitempty"`
	Services  []string `json:"services,omitempty"`
}

// Schema struct stores input and output validation schemas
type Schema struct {
	In  *spec.Schema `json:"in,omitempty"`
	Out *spec.Schema `json:"out,omitempty"`
}

// FnRun struct represents single function run
type FnRun struct {
	entitystore.BaseEntity
	FunctionName string                 `json:"functionName"`
	FunctionID   string                 `json:"functionID"`
	FaasID       string                 `json:"faasId"`
	Blocking     bool                   `json:"blocking"`
	Input        interface{}            `json:"input,omitempty"`
	Output       interface{}            `json:"output,omitempty"`
	Secrets      []string               `json:"secrets,omitempty"`
	Services     []string               `json:"services,omitempty"`
	HTTPContext  map[string]interface{} `json:"httpContext,omitempty"`
	Event        *events.CloudEvent     `json:"event,omitempty"`
	Logs         *models.Logs           `json:"logs,omitempty"`
	FinishedTime time.Time              `json:"finishedTime,omitempty"`

	WaitChan chan struct{} `json:"-"`
}

// Wait waits for function execution to finish
func (r *FnRun) Wait() {
	defer trace.Trace("")()

	if r.WaitChan != nil {
		<-r.WaitChan
	}
}

// Done reports completion of function execution
func (r *FnRun) Done() {
	defer trace.Trace("")()

	defer func() {
		recover()
	}()
	close(r.WaitChan)
}
