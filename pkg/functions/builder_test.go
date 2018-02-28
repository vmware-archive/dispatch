///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package functions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
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
	assert.Equal(t, prefix+"/func-"+faas+"-"+fnID+":latest", imageName(prefix, faas, fnID))
}

func TestWriteFunctionDockerfile(t *testing.T) {
	wd, err := os.Getwd()
	assert.NoError(t, err)
	tmpDir, err := ioutil.TempDir("", "func-build")
	assert.NoError(t, err)
	exec := Exec{
		Name:     "testFunc",
		Code:     "def hello(*args, **kwargs): pass",
		Image:    "not/a/real/image:test",
		Language: "python3",
	}
	err = writeFunctionDockerfile(tmpDir, filepath.Join(wd, "../../images/function-manager/templates"), "openfaas", &exec)
	assert.NoError(t, err)
	b, err := ioutil.ReadFile(filepath.Join(tmpDir, "Dockerfile"))
	assert.NoError(t, err)
	target := fmt.Sprintf(`FROM vmware/dispatch-openfaas-watchdog:revbf667b8 AS watchdog
FROM %s
COPY --from=watchdog /go/src/github.com/openfaas/faas/watchdog/watchdog /usr/bin/fwatchdog

WORKDIR /root/

COPY index.py .
RUN pip3 install -U setuptools

RUN mkdir function && touch function/__init__.py
COPY %s function/handler.py

ENV fprocess="python3 index.py"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]`, exec.Image, "function.txt")
	assert.Equal(t, target, string(b))

	b, err = ioutil.ReadFile(filepath.Join(tmpDir, "function.txt"))
	assert.NoError(t, err)
	assert.Equal(t, exec.Code, string(b))
}

func TestWriteFunctionDockerfileUnsupportedLanguage(t *testing.T) {
	wd, err := os.Getwd()
	assert.NoError(t, err)
	tmpDir, err := ioutil.TempDir("", "func-build")
	assert.NoError(t, err)
	exec := Exec{
		Name:     "testFunc",
		Code:     "def hello(*args, **kwargs): pass",
		Image:    "not/a/real/image:test",
		Language: "unsupported-language",
	}
	err = writeFunctionDockerfile(tmpDir, filepath.Join(wd, "../../images/function-manager/templates"), "openfaas", &exec)
	assert.Error(t, err)
	assert.Equal(t, "faas driver openfaas does not support language unsupported-language", err.Error())
}

func mockDockerClient(t *testing.T, doer func(*http.Request) (*http.Response, error)) *docker.Client {
	mockHTTP := newMockClient(doer)
	if _, ok := mockHTTP.Transport.(http.RoundTripper); !ok {
		t.Errorf("mockHTTP is not transport: %t", ok)
	}

	// TODO(karol): This line fails because we're using old docker client which expects http.Client.Transport
	// to be of type http.Transport. Newer version of docker client only expects it to be of type http.RoundTripper, which
	// is an interface and allows mocking. After updating the docker client library, we should be able to use mocked version
	// and test BuildImage function.
	client, err := docker.NewClient("http://localhost:2375", "1.0", mockHTTP, nil)
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
