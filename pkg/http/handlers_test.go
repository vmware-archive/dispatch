///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package http

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/http/mocks"
)

func TestAllInOneRouter(t *testing.T) {

	type testCase struct {
		url           string
		expectedCode  int
		expectedBody  string
		expectedError bool
	}

	cases := []testCase{
		{
			url:           "http://localhost:8080/v1/",
			expectedCode:  http.StatusNotFound,
			expectedError: true,
		},
		{
			url:           "http://localhost:8080/v1/nonexistent",
			expectedCode:  http.StatusNotFound,
			expectedError: true,
		},
		{
			url:           "http://localhost:8080/",
			expectedCode:  http.StatusNotFound,
			expectedError: true,
		},
		{
			url:           "http://localhost:8080/v1/iam",
			expectedCode:  http.StatusNotImplemented,
			expectedError: true,
		},
		{
			url:           "http://localhost:8080/v1/api",
			expectedCode:  http.StatusNotImplemented,
			expectedError: true,
		},
		{
			url:          "http://localhost:8080/v1/image",
			expectedCode: http.StatusOK,
			expectedBody: "imageHandler",
		},
		{
			url:          "http://localhost:8080/v1/function",
			expectedCode: http.StatusOK,
			expectedBody: "functionHandler",
		},
		{
			url:          "http://localhost:8080/v1/event",
			expectedCode: http.StatusOK,
			expectedBody: "eventsHandler",
		},
		{
			url:          "http://localhost:8080/v1/secret",
			expectedCode: http.StatusOK,
			expectedBody: "secretsHandler",
		},
	}

	for _, c := range cases {
		r := createRouter()
		var buf bytes.Buffer
		req := httptest.NewRequest("POST", c.url, &buf)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		resp := w.Result()

		if c.expectedCode != resp.StatusCode {
			t.Errorf("url: %s, expected status code: %d, received: %d", c.url, c.expectedCode, resp.StatusCode)
		}
		body, _ := ioutil.ReadAll(resp.Body)
		if c.expectedError {
			err := unmarshallErrorHelper(t, body)
			if c.expectedCode != resp.StatusCode {
				t.Errorf("url: %s, expected error code: %d, received: %d", c.url, c.expectedCode, err.Code)
			}
		} else {
			if c.expectedBody != string(body) {
				t.Errorf("url: %s, expected body: %s, received: %s", c.url, c.expectedBody, string(body))

			}
		}
	}

}

func createRouter() *AllInOneRouter {
	r := &AllInOneRouter{}
	imageMock := &mocks.HandlerMock{}
	imageMock.On("ServeHTTP", mock.Anything, mock.Anything).Run(testHandler("imageHandler"))
	r.ImagesHandler = imageMock
	functionMock := &mocks.HandlerMock{}
	functionMock.On("ServeHTTP", mock.Anything, mock.Anything).Run(testHandler("functionHandler"))
	r.FunctionsHandler = functionMock
	eventsMock := &mocks.HandlerMock{}
	eventsMock.On("ServeHTTP", mock.Anything, mock.Anything).Run(testHandler("eventsHandler"))
	r.EventsHandler = eventsMock
	secretsMock := &mocks.HandlerMock{}
	secretsMock.On("ServeHTTP", mock.Anything, mock.Anything).Run(testHandler("secretsHandler"))
	r.SecretsHandler = secretsMock
	return r
}

func unmarshallErrorHelper(t *testing.T, body []byte) *v1.Error {
	t.Helper()
	errorObj := new(v1.Error)
	err := json.Unmarshal(body, errorObj)
	assert.NoError(t, err)
	return errorObj
}
