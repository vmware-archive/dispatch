///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package api

// NO TESTS

import (
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/dispatch/pkg/api"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/testing/dev"
)

var (
	postgresConfig = entitystore.BackendConfig{
		Address:  "192.168.99.100:5432",
		Username: "testuser",
		Password: "testpasswd",
		Bucket:   "testdb",
	}
)

// MakeEntityStore returns an EntityStore for test
func MakeEntityStore(t *testing.T) entitystore.EntityStore {

	if dev.Local() {
		// test with postgres db locally only if a postgres db is set up
		es, err := entitystore.NewFromBackend(postgresConfig)
		assert.NoError(t, err)
		return es
	}

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
	w := httptest.NewRecorder()

	responder.WriteResponse(w, runtime.JSONProducer())
	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading back response: %v", err)
	}
	assert.Equal(t, statusCode, resp.StatusCode)
	if len(body) > 0 {
		err = json.Unmarshal(body, responseObject)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
	}
}
