///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package api

// NO TESTS

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-swagger/go-swagger/examples/generated/restapi"
	"github.com/stretchr/testify/assert"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/api"
	entitystore "gitlab.eng.vmware.com/serverless/serverless/pkg/entity-store"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/testing/store"
)

// MakeAPI returns an API for testing
func MakeAPI(t *testing.T, registrar api.HandlerRegistrar, api api.SwaggerAPI) {
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}
	_, kv := store.MakeKVStore(t)
	es := entitystore.New(kv)
	api.SetSpec(swaggerSpec)
	registrar(api, es)
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
