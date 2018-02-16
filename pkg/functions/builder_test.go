///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package functions

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/rand"
)

func TestImageName(t *testing.T) {
	prefix := rand.String(9)
	faas := rand.String(5)
	fnID := rand.String(6)
	assert.Equal(t, prefix+"/func-"+faas+"-"+fnID, imageName(prefix, faas, fnID))
}

func mockDockerClient(t *testing.T, doer func(*http.Request) (*http.Response, error)) *docker.Client {
	mockHttp := newMockClient(doer)
	if _, ok := mockHttp.Transport.(http.RoundTripper); !ok {
		t.Errorf("mockHttp is not transport: %t", ok)
	}

	// TODO(karol): This line fails because we're using old docker client which expects http.Client.Transport
	// to be of type http.Transport. Newer version of docker client only expects it to be of type http.RoundTripper, which
	// is an interface and allows mocking. After updating the docker client library, we should be able to use mocked version
	// and test BuildImage function.
	client, err := docker.NewClient("http://localhost:2375", "1.0", mockHttp, nil)
	assert.NoError(t, err)
	return client
}

func newMockClient(doer func(*http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{
		Transport: transportFunc(doer),
	}
}

type transportFunc func(*http.Request) (*http.Response, error)

func (tf transportFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return tf(req)
}

func errorMock(statusCode int, message string) func(req *http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		header := http.Header{}
		header.Set("Content-Type", "application/json")

		body, err := json.Marshal(&types.ErrorResponse{
			Message: message,
		})
		if err != nil {
			return nil, err
		}

		return &http.Response{
			StatusCode: statusCode,
			Body:       ioutil.NopCloser(bytes.NewReader(body)),
			Header:     header,
		}, nil
	}
}
