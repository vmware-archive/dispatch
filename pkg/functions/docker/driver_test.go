///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package docker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vmware/dispatch/pkg/api/v1"
	mocks "github.com/vmware/dispatch/pkg/mocks/docker"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/functions"
)

func startHTTPServer() (*httptest.Server, string) {
	server := httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusOK)
		}),
	)

	port := strings.Split(server.URL, ":")[2]

	return server, port
}

func TestDriverCreate(t *testing.T) {
	f := functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name: "hello",
			ID:   "deadbeef",
		},
	}
	dockerMock := &mocks.CommonAPIClient{}
	d := New(dockerMock)
	d.RetryTimeout = 0
	server, port := startHTTPServer()
	defer server.Close()

	dockerMock.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		container.ContainerCreateCreatedBody{}, error(nil),
	)

	dockerMock.On("ContainerStart", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	c := types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			State: &types.ContainerState{
				Running: true,
			},
		},
		NetworkSettings: &types.NetworkSettings{
			NetworkSettingsBase: types.NetworkSettingsBase{
				Ports: nat.PortMap{
					functionAPIPort: []nat.PortBinding{{
						HostIP:   "0.0.0.0",
						HostPort: port,
					}},
				},
			},
		},
	}

	dockerMock.On("ContainerInspect", mock.Anything, mock.Anything).Return(
		c, nil,
	)

	dockerMock.On("ContainerList", mock.Anything, mock.Anything).Return(
		[]types.Container{}, nil,
	)

	err := d.Create(context.Background(), &f)
	assert.NoError(t, err)

	server.Config.Handler = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
	})
	err = d.Create(context.Background(), &f)
	assert.Error(t, err)
}

func TestDriverGetRunnableMissing(t *testing.T) {
	dockerMock := &mocks.CommonAPIClient{}
	d := New(dockerMock)

	dockerMock.On("ContainerList", mock.Anything, mock.Anything).Return(
		[]types.Container{}, nil,
	)

	f := d.GetRunnable(&functions.FunctionExecution{FunctionID: "deadbeef"})
	ctx := functions.Context{}
	_, err := f(ctx, map[string]interface{}{"name": "Me", "place": "Here"})

	assert.Error(t, err)
}

func TestDriverGetRunnable(t *testing.T) {
	dockerMock := &mocks.CommonAPIClient{}
	d := New(dockerMock)

	dockerMock.On("ContainerList", mock.Anything, mock.Anything).Return(
		[]types.Container{{}}, nil,
	)

	server, port := startHTTPServer()
	defer server.Close()

	server.Config.Handler = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		result, _ := json.Marshal(&functions.Message{
			Context: functions.Context{
				functions.LogsKey: v1.Logs{Stdout: []string{"log log log", "log log log"}},
			},
			Payload: map[string]interface{}{"myField": "Hello, Me from Here"},
		})
		rw.Write(result)
	})

	c := types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			State: &types.ContainerState{
				Running: true,
			},
		},
		NetworkSettings: &types.NetworkSettings{
			NetworkSettingsBase: types.NetworkSettingsBase{
				Ports: nat.PortMap{
					functionAPIPort: []nat.PortBinding{{
						HostIP:   "0.0.0.0",
						HostPort: port,
					}},
				},
			},
		},
		Config: &container.Config{},
	}

	dockerMock.On("ContainerInspect", mock.Anything, mock.Anything).Return(
		c, nil,
	)

	f := d.GetRunnable(&functions.FunctionExecution{FunctionID: "deadbeef"})
	ctx := functions.Context{}

	r, err := f(ctx, map[string]interface{}{"name": "Me", "place": "Here"})
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"myField": "Hello, Me from Here"}, r)
	assert.Equal(t, v1.Logs{Stdout: []string{"log log log", "log log log"}}, ctx["logs"])
}

func TestOfDriverDelete(t *testing.T) {
	f := functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name: "hello",
			ID:   "deadbeef",
		},
	}
	dockerMock := &mocks.CommonAPIClient{}
	d := New(dockerMock)

	dockerMock.On("ContainerList", mock.Anything, mock.Anything).Return(
		[]types.Container{{}}, nil,
	)

	dockerMock.On("ContainerRemove", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := d.Delete(context.Background(), &f)
	assert.NoError(t, err)

}
