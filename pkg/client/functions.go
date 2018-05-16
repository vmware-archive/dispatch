///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package client

import (
	"context"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/pkg/errors"

	"github.com/vmware/dispatch/pkg/api/v1"
	swaggerclient "github.com/vmware/dispatch/pkg/function-manager/gen/client"
	"github.com/vmware/dispatch/pkg/function-manager/gen/client/runner"
	"github.com/vmware/dispatch/pkg/function-manager/gen/client/store"
)

// NO TESTS

// FunctionsClient defines the function client interface
type FunctionsClient interface {
	// Function Runner
	RunFunction(ctx context.Context, run *v1.Run) (*v1.Run, error)
	GetFunctionRun(ctx context.Context, functionName string, runName string) (*v1.Run, error)
	ListRuns(ctx context.Context) ([]v1.Run, error)
	ListFunctionRuns(ctx context.Context, functionName string) ([]v1.Run, error)

	// Function store
	CreateFunction(ctx context.Context, function *v1.Function) (*v1.Function, error)
	DeleteFunction(ctx context.Context, functionName string) (*v1.Function, error)
	GetFunction(ctx context.Context, functionName string) (*v1.Function, error)
	ListFunctions(ctx context.Context) ([]v1.Function, error)
	UpdateFunction(ctx context.Context, function *v1.Function) (*v1.Function, error)
}

// DefaultFunctionsClient defines the default functions client
type DefaultFunctionsClient struct {
	client *swaggerclient.FunctionManager
	auth   runtime.ClientAuthInfoWriter
}

// NewFunctionsClient is used to create a new functions client
func NewFunctionsClient(host string, auth runtime.ClientAuthInfoWriter) FunctionsClient {
	transport := DefaultHTTPClient(host, swaggerclient.DefaultBasePath)
	return &DefaultFunctionsClient{
		client: swaggerclient.New(transport, strfmt.Default),
		auth:   auth,
	}
}

// RunFunction runs a function
func (c *DefaultFunctionsClient) RunFunction(ctx context.Context, run *v1.Run) (*v1.Run, error) {
	functionName := run.FunctionName
	run.FunctionName = ""
	params := runner.RunFunctionParams{
		FunctionName: &functionName,
		Context:      ctx,
		Body:         run,
	}
	ok, accepted, err := c.client.Runner.RunFunction(&params, c.auth)
	if err != nil {
		return nil, errors.Wrapf(err, "error when running the function %s", functionName)
	}
	if ok != nil {
		return ok.Payload, nil
	}
	if accepted != nil {
		return accepted.Payload, nil
	}
	return nil, errors.New("swagger error, returned payload not supported")
}

// GetFunctionRun gets the results of a function run
func (c *DefaultFunctionsClient) GetFunctionRun(ctx context.Context, functionName string, runName string) (*v1.Run, error) {
	params := runner.GetRunParams{
		Context:      ctx,
		FunctionName: &functionName,
		RunName:      strfmt.UUID(runName),
	}
	response, err := c.client.Runner.GetRun(&params, c.auth)
	if err != nil {
		return nil, errors.Wrapf(err, "error when retrieving the run %s for function %s", runName, functionName)
	}
	return response.Payload, nil
}

// ListRuns lists all the available results from previous function runs
func (c *DefaultFunctionsClient) ListRuns(ctx context.Context) ([]v1.Run, error) {
	params := runner.GetRunsParams{
		Context: ctx,
	}
	response, err := c.client.Runner.GetRuns(&params, c.auth)
	if err != nil {
		return nil, errors.Wrapf(err, "error when retrieving runs")
	}
	var runs []v1.Run
	for _, run := range response.Payload {
		runs = append(runs, *run)
	}
	return runs, nil
}

// ListFunctionRuns lists the available results from specific function runs
func (c *DefaultFunctionsClient) ListFunctionRuns(ctx context.Context, functionName string) ([]v1.Run, error) {
	params := runner.GetRunsParams{
		Context:      ctx,
		FunctionName: &functionName,
	}
	response, err := c.client.Runner.GetRuns(&params, c.auth)
	if err != nil {
		return nil, errors.Wrapf(err, "error when retrieving runs for function %s", functionName)
	}
	var runs []v1.Run
	for _, run := range response.Payload {
		runs = append(runs, *run)
	}
	return runs, nil
}

// CreateFunction creates and adds a new function
func (c *DefaultFunctionsClient) CreateFunction(ctx context.Context, function *v1.Function) (*v1.Function, error) {
	params := store.AddFunctionParams{
		Context: ctx,
		Body:    function,
	}
	response, err := c.client.Store.AddFunction(&params, c.auth)
	if err != nil {
		return nil, errors.Wrap(err, "error when creating a function")
	}
	return response.Payload, nil
}

// DeleteFunction deletes a function
func (c *DefaultFunctionsClient) DeleteFunction(ctx context.Context, functionName string) (*v1.Function, error) {
	params := store.DeleteFunctionParams{
		Context:      ctx,
		FunctionName: functionName,
	}
	response, err := c.client.Store.DeleteFunction(&params, c.auth)
	if err != nil {
		return nil, errors.Wrapf(err, "error when deleting the function %s", functionName)
	}
	return response.Payload, nil
}

// GetFunction gets a function by name
func (c *DefaultFunctionsClient) GetFunction(ctx context.Context, functionName string) (*v1.Function, error) {
	params := store.GetFunctionParams{
		Context:      ctx,
		FunctionName: functionName,
	}
	response, err := c.client.Store.GetFunction(&params, c.auth)
	if err != nil {
		return nil, errors.Wrapf(err, "error when retrieving the function %s", functionName)
	}
	return response.Payload, nil
}

// ListFunctions lists all functions
func (c *DefaultFunctionsClient) ListFunctions(ctx context.Context) ([]v1.Function, error) {
	params := store.GetFunctionsParams{
		Context: ctx,
	}
	response, err := c.client.Store.GetFunctions(&params, c.auth)
	if err != nil {
		return nil, errors.Wrap(err, "error when retrieving the functions")
	}
	var functions []v1.Function
	for _, f := range response.Payload {
		functions = append(functions, *f)
	}
	return functions, nil
}

// UpdateFunction updates a specific function
func (c *DefaultFunctionsClient) UpdateFunction(ctx context.Context, function *v1.Function) (*v1.Function, error) {
	params := store.UpdateFunctionParams{
		Context:      ctx,
		Body:         function,
		FunctionName: *function.Name,
	}
	response, err := c.client.Store.UpdateFunction(&params, c.auth)
	if err != nil {
		return nil, errors.Wrapf(err, "error when updating the function %s", *function.Name)
	}
	return response.Payload, nil
}
