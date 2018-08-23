///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package client

import (
	"context"
	"fmt"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/pkg/errors"

	"github.com/vmware/dispatch/pkg/api/v1"
	swaggerclient "github.com/vmware/dispatch/pkg/function-manager/gen/client"
	"github.com/vmware/dispatch/pkg/function-manager/gen/client/runner"
	"github.com/vmware/dispatch/pkg/function-manager/gen/client/store"
)

// FunctionsClient defines the function client interface
type FunctionsClient interface {
	// Function Runner
	RunFunction(ctx context.Context, organizationID string, run *v1.Run) (*v1.Run, error)
	GetFunctionRun(ctx context.Context, organizationID string, opts FunctionOpts) (*v1.Run, error)
	ListRuns(ctx context.Context, organizationID string, opts FunctionOpts) ([]v1.Run, error)

	// Function store
	CreateFunction(ctx context.Context, organizationID string, function *v1.Function) (*v1.Function, error)
	DeleteFunction(ctx context.Context, organizationID string, functionName string) (*v1.Function, error)
	GetFunction(ctx context.Context, organizationID string, functionName string) (*v1.Function, error)
	ListFunctions(ctx context.Context, organizationID string) ([]v1.Function, error)
	UpdateFunction(ctx context.Context, organizationID string, function *v1.Function) (*v1.Function, error)
}

// FunctionOpts are options for retrieving function runs
type FunctionOpts struct {
	FunctionName *string
	RunName      *string
	Since        time.Time
}

// DefaultFunctionsClient defines the default functions client
type DefaultFunctionsClient struct {
	baseClient

	client *swaggerclient.FunctionManager
	auth   runtime.ClientAuthInfoWriter
}

// NewFunctionsClient is used to create a new functions client
func NewFunctionsClient(host string, auth runtime.ClientAuthInfoWriter, organizationID string) *DefaultFunctionsClient {
	transport := DefaultHTTPClient(host, swaggerclient.DefaultBasePath)
	return &DefaultFunctionsClient{
		baseClient: baseClient{
			organizationID: organizationID,
		},
		client: swaggerclient.New(transport, strfmt.Default),
		auth:   auth,
	}
}

// RunFunction runs a function
func (c *DefaultFunctionsClient) RunFunction(ctx context.Context, organizationID string, run *v1.Run) (*v1.Run, error) {
	functionName := run.FunctionName
	run.FunctionName = ""
	params := runner.RunFunctionParams{
		FunctionName: &functionName,
		Context:      ctx,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
		Body:         run,
	}
	ok, accepted, err := c.client.Runner.RunFunction(&params)
	if err != nil {
		return nil, runSwaggerError(err)
	}
	if ok != nil {
		return ok.Payload, nil
	}
	if accepted != nil {
		return accepted.Payload, nil
	}
	return nil, errors.New("swagger error, returned payload not supported")
}

func runSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *runner.RunFunctionBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *runner.RunFunctionUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *runner.RunFunctionForbidden:
		return NewErrorForbidden(v.Payload)
	case *runner.RunFunctionNotFound:
		return NewErrorNotFound(v.Payload)
	case *runner.RunFunctionUnprocessableEntity:
		return NewErrorInvalidInput(v.Payload)
	case *runner.RunFunctionBadGateway:
		return NewErrorFunctionError(v.Payload)
	case *runner.RunFunctionDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// GetFunctionRun gets the results of a function run
func (c *DefaultFunctionsClient) GetFunctionRun(ctx context.Context, organizationID string, opts FunctionOpts) (*v1.Run, error) {
	s := opts.Since.Unix()
	params := runner.GetRunParams{
		Context:      ctx,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
		FunctionName: opts.FunctionName,
		RunName:      strfmt.UUID(*opts.RunName),
		Since:        &s,
	}
	response, err := c.client.Runner.GetRun(&params)
	if err != nil {
		return nil, getRunSwaggerError(err)
	}
	return response.Payload, nil
}

func getRunSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *runner.GetRunBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *runner.GetRunUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *runner.GetRunForbidden:
		return NewErrorForbidden(v.Payload)
	case *runner.GetRunNotFound:
		return NewErrorNotFound(v.Payload)
	case *runner.GetRunDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// ListRuns lists all the available results from previous function runs filtered by opts
func (c *DefaultFunctionsClient) ListRuns(ctx context.Context, organizationID string, opts FunctionOpts) ([]v1.Run, error) {
	s := opts.Since.Unix()
	params := runner.GetRunsParams{
		Context:      ctx,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
		FunctionName: opts.FunctionName,
		Since:        &s,
	}
	response, err := c.client.Runner.GetRuns(&params)
	if err != nil {
		return nil, listRunsSwaggerError(err)
	}
	runs := []v1.Run{}
	for _, run := range response.Payload {
		runs = append(runs, *run)
	}
	return runs, nil
}

func listRunsSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *runner.GetRunsBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *runner.GetRunsUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *runner.GetRunsForbidden:
		return NewErrorForbidden(v.Payload)
	case *runner.GetRunsDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// CreateFunction creates and adds a new function
func (c *DefaultFunctionsClient) CreateFunction(ctx context.Context, organizationID string, function *v1.Function) (*v1.Function, error) {
	params := store.AddFunctionParams{
		Context:      ctx,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
		Body:         function,
	}
	response, err := c.client.Store.AddFunction(&params)
	if err != nil {
		return nil, createFunctionSwaggerError(err)
	}
	return response.Payload, nil
}

func createFunctionSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *store.AddFunctionBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *store.AddFunctionUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *store.AddFunctionForbidden:
		return NewErrorForbidden(v.Payload)
	case *store.AddFunctionConflict:
		return NewErrorAlreadyExists(v.Payload)
	case *store.AddFunctionDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// DeleteFunction deletes a function
func (c *DefaultFunctionsClient) DeleteFunction(ctx context.Context, organizationID string, functionName string) (*v1.Function, error) {
	params := store.DeleteFunctionParams{
		Context:      ctx,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
		FunctionName: functionName,
	}
	response, err := c.client.Store.DeleteFunction(&params)
	if err != nil {
		return nil, deleteFunctionSwaggerError(err)
	}
	return response.Payload, nil
}

func deleteFunctionSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *store.DeleteFunctionBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *store.DeleteFunctionUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *store.DeleteFunctionForbidden:
		return NewErrorForbidden(v.Payload)
	case *store.DeleteFunctionNotFound:
		return NewErrorNotFound(v.Payload)
	case *store.DeleteFunctionDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// GetFunction gets a function by name
func (c *DefaultFunctionsClient) GetFunction(ctx context.Context, organizationID string, functionName string) (*v1.Function, error) {
	params := store.GetFunctionParams{
		Context:      ctx,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
		FunctionName: functionName,
	}
	response, err := c.client.Store.GetFunction(&params)
	if err != nil {
		return nil, getFunctionSwaggerError(err)
	}
	return response.Payload, nil
}

func getFunctionSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *store.GetFunctionBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *store.GetFunctionUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *store.GetFunctionForbidden:
		return NewErrorForbidden(v.Payload)
	case *store.GetFunctionNotFound:
		return NewErrorNotFound(v.Payload)
	case *store.GetFunctionDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// ListFunctions lists all functions
func (c *DefaultFunctionsClient) ListFunctions(ctx context.Context, organizationID string) ([]v1.Function, error) {
	params := store.GetFunctionsParams{
		Context:      ctx,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.Store.GetFunctions(&params)
	if err != nil {
		return nil, listFunctionsSwaggerError(err)
	}
	functions := []v1.Function{}
	for _, f := range response.Payload {
		functions = append(functions, *f)
	}
	return functions, nil
}

func listFunctionsSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *store.GetFunctionsUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *store.GetFunctionsForbidden:
		return NewErrorForbidden(v.Payload)
	case *store.GetFunctionsDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// UpdateFunction updates a specific function
func (c *DefaultFunctionsClient) UpdateFunction(ctx context.Context, organizationID string, function *v1.Function) (*v1.Function, error) {
	params := store.UpdateFunctionParams{
		Context:      ctx,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
		Body:         function,
		FunctionName: *function.Name,
	}
	response, err := c.client.Store.UpdateFunction(&params)
	if err != nil {
		return nil, updateFunctionSwaggerError(err)
	}
	return response.Payload, nil
}

func updateFunctionSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *store.UpdateFunctionBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *store.UpdateFunctionUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *store.UpdateFunctionForbidden:
		return NewErrorForbidden(v.Payload)
	case *store.UpdateFunctionNotFound:
		return NewErrorNotFound(v.Payload)
	case *store.UpdateFunctionDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}
