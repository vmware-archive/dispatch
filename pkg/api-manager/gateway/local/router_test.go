///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package local

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/dispatch/pkg/api-manager/gateway"
)

func TestMatchAPIAgainst(t *testing.T) {
	// empty API matches everything
	emptyAPI := &gateway.API{}

	hostsAPI := &gateway.API{Hosts: []string{"example.com", "my.host.test"}}
	methodsAPI := &gateway.API{Methods: []string{"GET", "POST"}}
	pathsAPI := &gateway.API{URIs: []string{"/hello", "/", "/some/multi/level/path"}}
	methodPathAPI := &gateway.API{
		Methods: []string{"GET", "POST"},
		URIs:    []string{"/hello", "/", "/some/multi/level/path"},
	}

	testCases := []struct {
		InAPI  *gateway.API
		Host   string
		Method string
		Path   string
		Out    bool
	}{
		{emptyAPI, "not", "important", "at all", true},
		{hostsAPI, "missing", "notimportant", "/", false},
		{hostsAPI, "example.com", "notimportant", "/", true},
		{methodsAPI, "something", "GET", "/", true},
		{methodsAPI, "missing", "OPTIONS", "/", true},
		{methodsAPI, "missing", "DELETE", "/", false},
		{pathsAPI, "missing", "notimportant", "/", true},
		{pathsAPI, "missing", "notimportant", "/notexistent", false},
		{methodPathAPI, "missing", "GET", "/some/multi/level/path", true},
		{methodPathAPI, "missing", "PUT", "/some/multi/level/path", false},
	}

	for _, c := range testCases {
		result := matchAPIAgainst(c.InAPI, c.Host, c.Method, c.Path)
		assert.Equal(t, result, c.Out)
	}
}

func TestGetInput(t *testing.T) {
	req1 := httptest.NewRequest("POST", "http://example.com/hello", bytes.NewBuffer([]byte(`{"key":"value"}`)))
	req1.Host = "example.com"
	req1.Header.Add("Content-type", "application/json")
	body, err := getInput(req1)
	assert.NoError(t, err)
	assert.Equal(t, "value", body.(map[string]interface{})["key"])

	req2 := httptest.NewRequest("POST", "http://example.com/hello", bytes.NewBuffer(nil))
	req2.Host = "example.com"
	body, err = getInput(req2)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, body)

	req3 := httptest.NewRequest("POST", "http://example.com/hello", bytes.NewBuffer([]byte(`name=VMware&place=Palo Alto`)))
	req3.Host = "example.com"
	req3.Header.Add("Content-type", "application/x-www-form-urlencoded")
	body, err = getInput(req3)
	assert.NoError(t, err)
	assert.Equal(t, "Palo Alto", body.(map[string]interface{})["place"])

	req4 := httptest.NewRequest("POST", "http://example.com/hello", bytes.NewBuffer([]byte(`blabla`)))
	req4.Host = "example.com"
	req4.Header.Add("Content-type", "text/plain")
	body, err = getInput(req4)
	assert.Error(t, err)

}

func TestMatchString(t *testing.T) {
	testCases := []struct {
		InSlice  []string
		InNeedle string
		Out      bool
	}{
		{[]string{"one", "two", "three"}, "one", true},
		{[]string{}, "not important", true},
		{[]string{"not", "important"}, "", true},
		{[]string{"one", "two"}, "missing", false},
	}

	for _, c := range testCases {
		result := matchString(c.InSlice, c.InNeedle)
		assert.Equal(t, result, c.Out)
	}
}

func TestIsJSONContentType(t *testing.T) {
	testCases := []struct {
		In  string
		Out bool
	}{
		{"", false},
		{"multipart/form-data; boundary=---------------------------974767299852498929531610575", false},
		{"text/html; charset=utf-8", false},
		{"application/json", true},
		{"application/cloudevents+json", true},
	}

	for _, c := range testCases {
		result := isJSONContentType(c.In)
		assert.Equal(t, result, c.Out)
	}
}

func TestCleanHost(t *testing.T) {
	testCases := []struct {
		In  string
		Out string
	}{
		{"", ""},
		{"host", "host"},
		{"host:12345", "host"},
		{":1245", ""},
	}

	for _, c := range testCases {
		result := cleanHost(c.In)
		assert.Equal(t, result, c.Out)
	}
}
