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
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/dispatch/pkg/api"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/testing/store"
)

func MakeEntityStore(t *testing.T) entitystore.EntityStore {
	_, kv := store.MakeKVStore(t)
	return entitystore.New(kv)
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
