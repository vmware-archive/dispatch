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
	imgMgr.On("GetImage", mock.Anything, mock.Anything).Return(
		&v1.Image{
			Language: "python3",
			Status:   v1.StatusINITIALIZED,
		}, nil)
	faas := &fnmocks.FaaSDriver{}
	function := &functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name:   "testFunction",
			Status: entitystore.StatusCREATING,
		},
		ImageName: "testImage",
		Code:      "some code",
		Main:      "main",
	}
	exec := &functions.Exec{
		Code: "some code", Main: "main", Image: "test/image:latest",
	}

	h := &funcEntityHandler{
		Store:     helpers.MakeEntityStore(t),
		FaaS:      faas,
		ImgClient: imgMgr,
	}

	_, err := h.Store.Add(context.Background(), function)
	require.NoError(t, err)

	require.NoError(t, h.Add(context.Background(), function))

	faas.AssertNotCalled(t, "Create", function, exec)
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
	function := &functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name:   "testFunction",
			Status: entitystore.StatusCREATING,
		},
		ImageName: "testImage",
		Code:      "some code",
		Main:      "main",
	}
	exec := &functions.Exec{
		Code: "some code", Main: "main", Image: "test/image:latest",
	}
	faas.On("Create", mock.Anything, function, exec).Return(nil)

	h := &funcEntityHandler{
		Store:     helpers.MakeEntityStore(t),
		FaaS:      faas,
		ImgClient: imgMgr,
	}

	_, err := h.Store.Add(context.Background(), function)
	require.NoError(t, err)

	require.NoError(t, h.Add(context.Background(), function))

	faas.AssertExpectations(t)
	imgMgr.AssertExpectations(t)
}

func TestFuncEntityHandler_Delete(t *testing.T) {
	faas := &fnmocks.FaaSDriver{}
	function := &functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name:   "testFunction",
			Status: entitystore.StatusDELETING,
		},
		ImageName: "testImage",
		Code:      "some code",
		Main:      "main",
	}
	faas.On("Delete", mock.Anything, function).Return(nil)

	h := &funcEntityHandler{
		Store: helpers.MakeEntityStore(t),
		FaaS:  faas,
	}

	_, err := h.Store.Add(context.Background(), function)
	require.NoError(t, err)

	require.NoError(t, h.Delete(context.Background(), function))

	faas.AssertExpectations(t)
}

func TestRunEntityHandler_Add(t *testing.T) {
	faas := &fnmocks.FaaSDriver{}
	function := &functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name:   "testFunction",
			Status: entitystore.StatusDELETING,
		},
		ImageName: "testImage",
		Code:      "some code",
		Main:      "main",
		Schema:    &functions.Schema{},
	}
	fnRun := &functions.FnRun{
		BaseEntity: entitystore.BaseEntity{
			Name: "testRun",
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
	secretInjector.On("GetMiddleware", mock.Anything, "cookie").Return(simw)
	serviceInjector := &fnmocks.ServiceInjector{}
	serviceInjector.On("GetMiddleware", mock.Anything, "cookie").Return(simw)

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
