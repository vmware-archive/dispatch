///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package client

import (
	"context"
	"strings"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/pkg/errors"
	"github.com/vmware/dispatch/pkg/function-manager/gen/client/runner"
	"github.com/vmware/dispatch/pkg/function-manager/gen/client/store"

	swaggerclient "github.com/vmware/dispatch/pkg/function-manager/gen/client"
	"github.com/vmware/dispatch/pkg/function-manager/gen/models"
)

// NO TESTS

type FunctionsClient interface {
	// Function Runner
	RunFunction(context.Context, *FunctionRun) (*FunctionRun, error)
	GetFunctionRun(ctx context.Context, functionName string, runName string) (*FunctionRun, error)
	ListRuns(context.Context) ([]FunctionRun, error)
	ListFunctionRuns(context.Context, string) ([]FunctionRun, error)

	// Function store
	CreateFunction(context.Context, *Function) (*Function, error)
	DeleteFunction(context.Context, string) (*Function, error)
	GetFunction(context.Context, string) (*Function, error)
	ListFunctions(context.Context) ([]Function, error)
	UpdateFunction(context.Context, *Function) (*Function, error)
}

type Function struct {
	models.Function
}

type FunctionRun struct {
	models.Run
}

type DefaultFunctionsClient struct {
	client *swaggerclient.FunctionManager
	auth   runtime.ClientAuthInfoWriter
}

func NewFunctionsClient(path string, auth runtime.ClientAuthInfoWriter) FunctionsClient {
	schemas := []string{"http"}
	if idx := strings.Index(path, "://"); idx != -1 {
		// http schema included in path
		schemas = []string{path[:idx]}
		path = path[idx+3:]

	}
	transport := httptransport.New(path, swaggerclient.DefaultBasePath, schemas)
	return &DefaultFunctionsClient{
		client: swaggerclient.New(transport, strfmt.Default),
		auth:   auth,
	}
}

func (c *DefaultFunctionsClient) RunFunction(ctx context.Context, run *FunctionRun) (*FunctionRun, error) {
	functionName := run.FunctionName
	run.FunctionName = ""
	params := runner.RunFunctionParams{
		FunctionName: functionName,
		Context:      ctx,
		Body:         &run.Run,
	}
	ok, accepted, err := c.client.Runner.RunFunction(&params, c.auth)
	if err != nil {
		return nil, errors.Wrapf(err, "error when running the function %s", functionName)
	}
	if ok != nil {
		return &FunctionRun{Run: *ok.Payload}, nil
	}
	if accepted != nil {
		return &FunctionRun{Run: *accepted.Payload}, nil
	}
	return nil, errors.New("swagger error, returned payload not supported")
}

func (c *DefaultFunctionsClient) GetFunctionRun(ctx context.Context, functionName string, runName string) (*FunctionRun, error) {
	params := runner.GetRunParams{
		Context:      ctx,
		FunctionName: functionName,
		RunName:      strfmt.UUID(runName),
	}
	response, err := c.client.Runner.GetRun(&params, c.auth)
	if err != nil {
		return nil, errors.Wrapf(err, "error when retrieving the run %s for function %s", runName, functionName)
	}
	return &FunctionRun{Run: *response.Payload}, nil
}

func (c *DefaultFunctionsClient) ListRuns(ctx context.Context) ([]FunctionRun, error) {
	params := runner.GetRunsParams{
		Context: ctx,
	}
	response, err := c.client.Runner.GetRuns(&params, c.auth)
	if err != nil {
		return nil, errors.Wrapf(err, "error when retrieving runs")
	}
	var runs []FunctionRun
	for _, run := range response.Payload {
		runs = append(runs, FunctionRun{Run: *run})
	}
	return runs, nil
}

func (c *DefaultFunctionsClient) ListFunctionRuns(ctx context.Context, functionName string) ([]FunctionRun, error) {
	params := runner.GetFunctionRunsParams{
		Context:      ctx,
		FunctionName: functionName,
	}
	response, err := c.client.Runner.GetFunctionRuns(&params, c.auth)
	if err != nil {
		return nil, errors.Wrapf(err, "error when retrieving runs for function %s", functionName)
	}
	var runs []FunctionRun
	for _, run := range response.Payload {
		runs = append(runs, FunctionRun{Run: *run})
	}
	return runs, nil
}

func (c *DefaultFunctionsClient) CreateFunction(ctx context.Context, function *Function) (*Function, error) {
	params := store.AddFunctionParams{
		Context: ctx,
		Body:    &function.Function,
	}
	response, err := c.client.Store.AddFunction(&params, c.auth)
	if err != nil {
		return nil, errors.Wrap(err, "error when creating a function")
	}
	return &Function{Function: *response.Payload}, nil
}

func (c *DefaultFunctionsClient) DeleteFunction(ctx context.Context, functionName string) (*Function, error) {
	params := store.DeleteFunctionParams{
		Context:      ctx,
		FunctionName: functionName,
	}
	response, err := c.client.Store.DeleteFunction(&params, c.auth)
	if err != nil {
		return nil, errors.Wrapf(err, "error when deleting the function %s", functionName)
	}
	return &Function{Function: *response.Payload}, nil
}

func (c *DefaultFunctionsClient) GetFunction(ctx context.Context, functionName string) (*Function, error) {
	params := store.GetFunctionParams{
		Context:      ctx,
		FunctionName: functionName,
	}
	response, err := c.client.Store.GetFunction(&params, c.auth)
	if err != nil {
		return nil, errors.Wrapf(err, "error when retrieving the function %s", functionName)
	}
	return &Function{Function: *response.Payload}, nil
}

func (c *DefaultFunctionsClient) ListFunctions(ctx context.Context) ([]Function, error) {
	params := store.GetFunctionsParams{
		Context: ctx,
	}
	response, err := c.client.Store.GetFunctions(&params, c.auth)
	if err != nil {
		return nil, errors.Wrap(err, "error when retrieving the functions")
	}
	var functions []Function
	for _, f := range response.Payload {
		functions = append(functions, Function{Function: *f})
	}
	return functions, nil
}

func (c *DefaultFunctionsClient) UpdateFunction(ctx context.Context, function *Function) (*Function, error) {
	params := store.UpdateFunctionParams{
		Context:      ctx,
		Body:         &function.Function,
		FunctionName: *function.Name,
	}
	response, err := c.client.Store.UpdateFunction(&params, c.auth)
	if err != nil {
		return nil, errors.Wrapf(err, "error when updating the function %s", *function.Name)
	}
	return &Function{Function: *response.Payload}, nil
}
