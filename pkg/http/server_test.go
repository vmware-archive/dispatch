///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package http

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vmware/dispatch/pkg/http/mocks"
)

func TestServer_Serve(t *testing.T) {
	mockHandler := &mocks.HandlerMock{}
	mockHandler.On("ServeHTTP", mock.Anything, mock.Anything).Run(testHandler("testmessage"))
	server := NewServer(mockHandler)
	server.Port = 0
	go server.Serve()
	server.Wait()

	resp, err := http.Get(server.HTTPURL())
	assert.NoError(t, err)
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "testmessage", string(body))

	server.Shutdown()
}

func testHandler(body string) func(mock.Arguments) {
	return func(arguments mock.Arguments) {
		rw := arguments[0].(http.ResponseWriter)
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(body))
	}
}
