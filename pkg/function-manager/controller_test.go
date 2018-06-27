///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functionmanager

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/function-manager/mocks"
	"github.com/vmware/dispatch/pkg/functions"
	fnmocks "github.com/vmware/dispatch/pkg/functions/mocks"
	"github.com/vmware/dispatch/pkg/functions/runner"
	"github.com/vmware/dispatch/pkg/functions/validator"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func TestFuncEntityHandler_Add_ImageNotReady(t *testing.T) {
	imgMgr := &mocks.ImageGetter{}
	imgMgr.On("GetImage", mock.Anything, testOrgID, mock.Anything).Return(
		&v1.Image{
			Language: "python3",
			Status:   v1.StatusINITIALIZED,
		}, nil)
	faas := &fnmocks.FaaSDriver{}
	funcName := "testFunction"
	source := &functions.Source{
		BaseEntity: entitystore.BaseEntity{
			Name:           "sourceName",
			OrganizationID: testOrgID,
		},
		Code:     []byte("some source"),
		Function: funcName,
	}
	function := &functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name:           funcName,
			Status:         entitystore.StatusCREATING,
			OrganizationID: testOrgID,
		},
		ImageName:  "testImage",
		Handler:    "main",
		SourceName: source.Name,
	}

	h := &funcEntityHandler{
		Store:     helpers.MakeEntityStore(t),
		FaaS:      faas,
		ImgClient: imgMgr,
	}

	_, err := h.Store.Add(context.Background(), source)
	require.NoError(t, err)

	_, err = h.Store.Add(context.Background(), function)
	require.NoError(t, err)

	require.NoError(t, h.Add(context.Background(), function))

	faas.AssertNotCalled(t, "Create", function)
	imgMgr.AssertExpectations(t)
}

func TestFuncEntityHandler_Add_ImageReady(t *testing.T) {
	imgMgr := &mocks.ImageGetter{}
	imgMgr.On("GetImage", mock.Anything, mock.Anything, mock.Anything).Return(
		&v1.Image{
			DockerURL: "test/image:latest",
			Language:  "python3",
			Status:    v1.StatusREADY,
		}, nil)
	faas := &fnmocks.FaaSDriver{}
	funcName := "testFunction"
	source := &functions.Source{
		BaseEntity: entitystore.BaseEntity{
			Name:           "sourceName",
			OrganizationID: testOrgID,
		},
		Code:     []byte("some source"),
		Function: funcName,
	}
	function := &functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name:           funcName,
			Status:         entitystore.StatusCREATING,
			OrganizationID: testOrgID,
		},
		ImageName:  "testImage",
		ImageURL:   "test/image:latest",
		Handler:    "main",
		SourceName: source.Name,
	}
	faas.On("Create", mock.Anything, function).Return(nil)

	imageBuilder := &fnmocks.ImageBuilder{}
	imageBuilder.On("BuildImage", mock.Anything, mock.Anything, mock.Anything).Return("fake-image:latest", nil)

	h := &funcEntityHandler{
		Store:        helpers.MakeEntityStore(t),
		FaaS:         faas,
		ImgClient:    imgMgr,
		ImageBuilder: imageBuilder,
	}

	_, err := h.Store.Add(context.Background(), source)
	require.NoError(t, err)

	_, err = h.Store.Add(context.Background(), function)
	require.NoError(t, err)

	require.NoError(t, h.Add(context.Background(), function))

	faas.AssertExpectations(t)
	imgMgr.AssertExpectations(t)
}

func TestFuncEntityHandler_Delete(t *testing.T) {
	faas := &fnmocks.FaaSDriver{}
	testFuncName := "testFunction"
	function := &functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name:           testFuncName,
			Status:         entitystore.StatusDELETING,
			OrganizationID: testOrgID,
		},
		ImageName: "testImage",
		Handler:   "main",
	}
	faas.On("Delete", mock.Anything, function).Return(nil)

	h := &funcEntityHandler{
		Store: helpers.MakeEntityStore(t),
		FaaS:  faas,
	}

	_, err := h.Store.Add(context.Background(), function)
	require.NoError(t, err)

	run := &functions.FnRun{
		BaseEntity: entitystore.BaseEntity{
			Name:           "testRun",
			Status:         entitystore.StatusREADY,
			OrganizationID: testOrgID,
		},
		FunctionName: testFuncName,
	}
	_, err = h.Store.Add(context.Background(), run)
	require.NoError(t, err)

	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter.Add(
		entitystore.FilterStat{
			Scope:   entitystore.FilterScopeExtra,
			Subject: "FunctionName",
			Verb:    entitystore.FilterVerbEqual,
			Object:  testFuncName,
		},
	)
	var runs []*functions.FnRun
	err = h.Store.List(context.Background(), testOrgID, opts, &runs)
	require.NoError(t, err)
	assert.Equal(t, 1, len(runs))
	assert.Equal(t, run.Name, runs[0].Name)

	require.NoError(t, h.Delete(context.Background(), function))

	err = h.Store.List(context.Background(), testOrgID, opts, &runs)
	require.NoError(t, err)
	assert.Equal(t, 0, len(runs))

	faas.AssertExpectations(t)
}

func TestRunEntityHandler_Add(t *testing.T) {
	faas := &fnmocks.FaaSDriver{}
	function := &functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name:           "testFunction",
			Status:         entitystore.StatusDELETING,
			OrganizationID: testOrgID,
		},
		ImageName: "testImage",
		Handler:   "main",
		Schema:    &functions.Schema{},
	}
	fnRun := &functions.FnRun{
		BaseEntity: entitystore.BaseEntity{
			Name:           "testRun",
			OrganizationID: testOrgID,
		},
		FunctionName: "testFunction",
	}

	functionCalled := false
	var runnable functions.Runnable = func(ctx functions.Context, in interface{}) (interface{}, error) {
		functionCalled = true
		return nil, nil
	}
	faas.On("GetRunnable", mock.Anything).Return(runnable)

	var simw functions.Middleware = func(f functions.Runnable) functions.Runnable {
		return f
	}
	secretInjector := &fnmocks.SecretInjector{}
	secretInjector.On("GetMiddleware", testOrgID, mock.Anything, "cookie").Return(simw)
	serviceInjector := &fnmocks.ServiceInjector{}
	serviceInjector.On("GetMiddleware", testOrgID, mock.Anything, "cookie").Return(simw)

	h := &runEntityHandler{
		Store: helpers.MakeEntityStore(t),
		FaaS:  faas,
		Runner: runner.New(&runner.Config{
			Faas:            faas,
			Validator:       validator.NoOp(),
			SecretInjector:  secretInjector,
			ServiceInjector: serviceInjector,
		}),
	}

	_, err := h.Store.Add(context.Background(), function)
	require.NoError(t, err)
	_, err = h.Store.Add(context.Background(), fnRun)
	require.NoError(t, err)

	require.NoError(t, h.Add(context.Background(), fnRun))

	faas.AssertExpectations(t)
	secretInjector.AssertExpectations(t)
	assert.True(t, functionCalled)
}
