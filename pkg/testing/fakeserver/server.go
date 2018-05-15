///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package fakeserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"reflect"
)

// NO TESTS

// Payload defines combination of expected request body, response body and response code
type Payload struct {
	RequestBody  map[string]interface{}
	ResponseBody map[string]interface{}
	ResponseCode int
}

// FakeServer allows setting up HTTP handlers for testing purposes
type FakeServer struct {
	payloadMap map[string]Payload

	debug bool
}

// NewFakeServer creates new instance of fake server
func NewFakeServer(responses map[string]Payload) *FakeServer {
	server := FakeServer{}
	if responses != nil {
		server.payloadMap = responses
	} else {
		server.payloadMap = make(map[string]Payload)
	}
	return &server
}

// EnableDebug enables debugging
func (s *FakeServer) EnableDebug() {
	s.debug = true
}

// DisableDebug disables debugging
func (s *FakeServer) DisableDebug() {
	s.debug = false
}

// AddResponse adds a response mapping
func (s *FakeServer) AddResponse(method string, path string, requestBody map[string]interface{}, responseBody map[string]interface{}, responseCode int) {
	if method == "" {
		method = "GET"
	}
	if responseCode == 0 {
		responseCode = 200
	}
	resp := Payload{
		RequestBody:  requestBody,
		ResponseBody: responseBody,
		ResponseCode: responseCode,
	}
	s.payloadMap[method+path] = resp
}

// ServeHTTP implements http.Handler interface
func (s *FakeServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if reqRaw, err := httputil.DumpRequest(req, true); err != nil {
		fmt.Printf("Error formating request: %s", err)
	} else if s.debug {
		fmt.Println(string(reqRaw))
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	var requestBody map[string]interface{}
	if len(body) > 0 {
		json.Unmarshal(body, &requestBody)
	}

	path := req.URL.Path
	if req.URL.RawQuery != "" {
		path = path + "?" + req.URL.RawQuery
	}
	if s.debug {
		fmt.Printf("Searching for %+v\n", req.Method+path)
	}
	expectedPayload, ok := s.payloadMap[req.Method+path]
	if !ok {
		w.WriteHeader(404)
		if s.debug {
			fmt.Println("No mapping found for this payload.")
			fmt.Println("Mapping content:")
			for key := range s.payloadMap {
				fmt.Printf("Key: %+v\n", key)
			}
		}
		return
	}
	if !reflect.DeepEqual(expectedPayload.RequestBody, requestBody) {
		w.WriteHeader(404)
		if s.debug {
			fmt.Println("Received body does not match expected body.")
			fmt.Println("Mapping content:")
			fmt.Printf("Expected:\n %+v\n", expectedPayload.RequestBody)
			fmt.Printf("Received:\n %+v\n", requestBody)
		}
		return
	}
	if s.debug {
		fmt.Printf("Returning code: %d\n", expectedPayload.ResponseCode)
		fmt.Printf("Returning body: %s\n", expectedPayload.ResponseBody)
	}

	respJSON, err := json.Marshal(expectedPayload.ResponseBody)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(expectedPayload.ResponseCode)
	w.Write(respJSON)
}
