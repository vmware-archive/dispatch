///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package listener

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vmware/dispatch/pkg/events/mocks"
	"github.com/vmware/dispatch/pkg/events/parser"
	"github.com/vmware/dispatch/pkg/events/validator"
)

func TestHTTPHandlerSuccess(t *testing.T) {

	m := mockSharedListener()
	m.parser = &parser.JSONEventParser{}
	m.validator = validator.NewDefaultValidator()
	m.transport.(*mocks.Transport).On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	listener, err := NewHTTP(m, 8080)
	assert.NoError(t, err)

	buf1 := bytes.NewBuffer(eventJSON(&testEvent1))
	req1 := httptest.NewRequest("POST", "http://localhost:8080/foo", buf1)
	w1 := httptest.NewRecorder()
	listener.ServeHTTP(w1, req1)
	resp1 := w1.Result()
	body1, _ := ioutil.ReadAll(resp1.Body)
	assert.Contains(t,
		strings.TrimSpace(string(body1)),
		"",
	)
	assert.Equal(t, http.StatusCreated, resp1.StatusCode)

}

func TestHTTPHandlerBadMethod(t *testing.T) {
	req := httptest.NewRequest("PUT", "http://localhost:8080/foo", nil)
	w := httptest.NewRecorder()

	listener, err := NewHTTP(emptySharedListener(), 8080)
	assert.NoError(t, err)

	listener.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "Incorrect method PUT, POST must be used", strings.TrimSpace(string(body)))
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestHTTPHandlerEmptyPayload(t *testing.T) {
	req := httptest.NewRequest("POST", "http://localhost:8080/foo", nil)
	w := httptest.NewRecorder()

	m := mockSharedListener()
	m.parser.(*mocks.StreamParser).On("Parse", mock.Anything).Return(nil, errors.New("error parsing"))
	listener, err := NewHTTP(m, 8080)
	assert.NoError(t, err)

	listener.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "Error parsing input: error parsing", strings.TrimSpace(string(body)))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHTTPHandlerEmptyList(t *testing.T) {
	m := mockSharedListener()
	m.parser = &parser.JSONEventParser{}
	listener, err := NewHTTP(m, 8080)
	assert.NoError(t, err)

	buf1 := bytes.NewBufferString("[]")
	req1 := httptest.NewRequest("POST", "http://localhost:8080/foo", buf1)
	w1 := httptest.NewRecorder()
	listener.ServeHTTP(w1, req1)
	resp1 := w1.Result()
	body1, _ := ioutil.ReadAll(resp1.Body)
	assert.Equal(t, "No events parsed", strings.TrimSpace(string(body1)))
	assert.Equal(t, http.StatusBadRequest, resp1.StatusCode)
}

func TestHTTPHandlerMalformedPayload(t *testing.T) {

	m := mockSharedListener()
	m.parser = &parser.JSONEventParser{}
	listener, err := NewHTTP(m, 8080)
	assert.NoError(t, err)

	buf1 := bytes.NewBufferString("{error_json}")
	req1 := httptest.NewRequest("POST", "http://localhost:8080/foo", buf1)
	w1 := httptest.NewRecorder()
	listener.ServeHTTP(w1, req1)
	resp1 := w1.Result()
	body1, _ := ioutil.ReadAll(resp1.Body)
	assert.Equal(t,
		"Error parsing input: invalid character 'e' looking for beginning of object key string",
		strings.TrimSpace(string(body1)))
	assert.Equal(t, http.StatusBadRequest, resp1.StatusCode)

	buf2 := bytes.NewBufferString("[{dsgdsgsg}]")
	req2 := httptest.NewRequest("POST", "http://localhost:8080/foo", buf2)
	w2 := httptest.NewRecorder()
	listener.ServeHTTP(w2, req2)
	resp2 := w2.Result()
	body2, _ := ioutil.ReadAll(resp2.Body)
	assert.Equal(t,
		"Error parsing input: invalid character 'd' looking for beginning of object key string",
		strings.TrimSpace(string(body2)))
	assert.Equal(t, http.StatusBadRequest, resp2.StatusCode)
}

func TestHTTPHandlerInvalidPayload(t *testing.T) {

	m := mockSharedListener()
	m.parser = &parser.JSONEventParser{}
	m.validator = validator.NewDefaultValidator()
	listener, err := NewHTTP(m, 8080)
	assert.NoError(t, err)

	invalidEvent := testEvent1
	invalidEvent.EventID = ""
	buf1 := bytes.NewBuffer(eventJSON(&invalidEvent))
	req1 := httptest.NewRequest("POST", "http://localhost:8080/foo", buf1)
	w1 := httptest.NewRecorder()
	listener.ServeHTTP(w1, req1)
	resp1 := w1.Result()
	body1, _ := ioutil.ReadAll(resp1.Body)
	assert.Contains(t,
		strings.TrimSpace(string(body1)),
		"Field validation for 'EventID' failed on the 'required' tag",
	)
	assert.Equal(t, http.StatusBadRequest, resp1.StatusCode)

}

func TestHTTPHandlerErrorPublish(t *testing.T) {

	m := mockSharedListener()
	m.parser = &parser.JSONEventParser{}
	m.validator = validator.NewDefaultValidator()
	m.transport.(*mocks.Transport).On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error publishing"))
	listener, err := NewHTTP(m, 8080)
	assert.NoError(t, err)

	buf1 := bytes.NewBuffer(eventJSON(&testEvent1))
	req1 := httptest.NewRequest("POST", "http://localhost:8080/foo", buf1)
	w1 := httptest.NewRecorder()
	listener.ServeHTTP(w1, req1)
	resp1 := w1.Result()
	body1, _ := ioutil.ReadAll(resp1.Body)
	assert.Contains(t,
		strings.TrimSpace(string(body1)),
		"error publishing",
	)
	assert.Equal(t, http.StatusInternalServerError, resp1.StatusCode)

}
