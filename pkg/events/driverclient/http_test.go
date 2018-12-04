///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package driverclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithNoOpts(t *testing.T) {
	client, err := newHTTPClient()
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}
	assert.Equal(t, "http://localhost:8080/v1/event/ingest", client.getURL())
}

func TestWithGateway(t *testing.T) {
	opt := WithGateway("172.17.0.1:8080")
	client, err := newHTTPClient(opt)
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}
	assert.Equal(t, "http://172.17.0.1:8080/v1/event/ingest", client.getURL())
}

func TestWithHost(t *testing.T) {
	opt := WithHost("example.com")
	client, err := newHTTPClient(opt)
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}
	assert.Equal(t, "http://example.com:8080/v1/event/ingest", client.getURL())
}

func TestWithDefaultPort(t *testing.T) {
	opt := WithPort("")
	client, err := newHTTPClient(opt)
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}
	assert.Equal(t, "http://localhost:8080/v1/event/ingest", client.getURL())
}

func TestWithPort(t *testing.T) {
	opt := WithPort("8081")
	client, err := newHTTPClient(opt)
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}
	assert.Equal(t, "http://localhost:8081/v1/event/ingest", client.getURL())
}

func TestWithURL(t *testing.T) {
	opt := WithURL("http://example.com/my/path")
	client, err := newHTTPClient(opt)
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}
	assert.Equal(t, "http://example.com/my/path", client.getURL())
}
