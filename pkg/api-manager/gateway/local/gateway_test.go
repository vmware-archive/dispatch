///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package local

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vmware/dispatch/pkg/api-manager/gateway"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client/mocks"
)

func TestGatewayGetRequest(t *testing.T) {
	// empty API matches everything
	fnClient := &mocks.FunctionsClient{}
	fnClient.On("RunFunction", mock.Anything, mock.Anything, mock.Anything).Return(
		&v1.Run{}, nil,
	)
	gw, err := NewGateway(nil, fnClient)
	assert.NoError(t, err)

	api1 := &gateway.API{
		ID:        uuid.NewV4().String(),
		CreatedAt: int(time.Now().Unix()),
		Name:      "api1",
		Function:  "function1",
		URIs:      []string{"/hello"},
		Methods:   []string{"GET"},
		Enabled:   true,
	}

	req1 := httptest.NewRequest("GET", "http://localhost:8080/hello", nil)

	rec1 := httptest.NewRecorder()
	gw.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusNotFound, rec1.Code)

	gw.AddAPI(context.Background(), api1)
	rec2 := httptest.NewRecorder()
	gw.ServeHTTP(rec2, req1)
	assert.Equal(t, http.StatusOK, rec2.Code)

	apiResult, err := gw.GetAPI(context.Background(), api1.Name)
	assert.NoError(t, err)
	assert.Equal(t, api1, apiResult)

	gw.DeleteAPI(context.Background(), apiResult)
	rec3 := httptest.NewRecorder()
	gw.ServeHTTP(rec3, req1)
	assert.Equal(t, http.StatusNotFound, rec3.Code)

	_, err = gw.GetAPI(context.Background(), api1.Name)
	assert.Error(t, err)
}

func TestGatewayPostRequest(t *testing.T) {
	// empty API matches everything
	fnClient := &mocks.FunctionsClient{}
	fnClient.On("RunFunction", mock.Anything, mock.Anything, mock.Anything).Return(
		&v1.Run{}, nil,
	)
	gw, err := NewGateway(nil, fnClient)
	assert.NoError(t, err)

	api1 := &gateway.API{
		ID:        uuid.NewV4().String(),
		CreatedAt: int(time.Now().Unix()),
		Name:      "api",
		Function:  "function1",
		Hosts:     []string{"example.com"},
		Methods:   []string{"POST"},
		Enabled:   true,
	}

	api2 := &gateway.API{
		ID:        uuid.NewV4().String(),
		CreatedAt: int(time.Now().Unix()),
		Name:      "api",
		Function:  "function1",
		Hosts:     []string{"example.com"},
		Enabled:   true,
	}

	req1 := httptest.NewRequest("POST", "http://example.com/hello", bytes.NewBuffer([]byte(`{"key":"value"}`)))
	req1.Host = "example.com"
	req1.Header.Add("Content-type", "application/json")

	rec1 := httptest.NewRecorder()
	gw.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusNotFound, rec1.Code)

	gw.AddAPI(context.Background(), api1)
	rec2 := httptest.NewRecorder()
	gw.ServeHTTP(rec2, req1)
	fmt.Printf(rec2.Body.String())
	assert.Equal(t, http.StatusOK, rec2.Code)

	apiResult, err := gw.GetAPI(context.Background(), api1.Name)
	assert.NoError(t, err)
	assert.Equal(t, api1, apiResult)

	req2 := httptest.NewRequest("POST", "http://example.com/hello", bytes.NewBuffer([]byte(`{"key":"value"}`)))
	req2.Host = "example.com"
	req2.Header.Add("Content-type", "application/json")
	gw.UpdateAPI(context.Background(), apiResult.Name, api2)
	rec3 := httptest.NewRecorder()
	gw.ServeHTTP(rec3, req2)
	fmt.Printf(rec3.Body.String())
	assert.Equal(t, http.StatusOK, rec3.Code)

	apiResult, err = gw.GetAPI(context.Background(), api1.Name)
	assert.NoError(t, err)
	assert.Equal(t, api2, apiResult)

	gw.DeleteAPI(context.Background(), api2)
	rec4 := httptest.NewRecorder()
	gw.ServeHTTP(rec4, req1)
	assert.Equal(t, http.StatusNotFound, rec4.Code)

}
