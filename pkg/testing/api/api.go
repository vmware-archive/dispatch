///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package api

// NO TESTS

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/dispatch/pkg/api"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
)

// MakeEntityStore returns an EntityStore for test
func MakeEntityStore(t *testing.T) entitystore.EntityStore {
	// not local, use libkv
	file, err := ioutil.TempFile(os.TempDir(), "test")
	assert.NoError(t, err, "Cannot create temp file")
	libkvConfig := entitystore.BackendConfig{
		Backend: "boltdb",
		Address: file.Name(),
		Bucket:  "test",
	}
	es, err := entitystore.NewFromBackend(libkvConfig)
	assert.NoError(t, err, "Cannot create store")
	return es
}

// MakeAPI returns an API for testing
func MakeAPI(t *testing.T, registrar api.HandlerRegistrar, api api.SwaggerAPI) {
	registrar(api)
}

// HandlerRequest is a convenience function for testing API handlers
func HandlerRequest(t *testing.T, responder middleware.Responder, responseObject interface{}, statusCode int) {
	HandlerRequestWithResponse(t, responder, responseObject, statusCode)
}

// HandlerRequestWithResponse is a convenience function for testing API handlers that additionally returns the response object
func HandlerRequestWithResponse(t *testing.T, responder middleware.Responder, responseObject interface{}, statusCode int) *http.Response {
	w := httptest.NewRecorder()

	responder.WriteResponse(w, runtime.JSONProducer())
	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading back response: %v", err)
	}
	t.Log(string(body))
	assert.Equal(t, statusCode, resp.StatusCode)
	if len(body) > 0 {
		err = json.Unmarshal(body, responseObject)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
	}
	return resp
}
