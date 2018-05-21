///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package client_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/dispatch/pkg/client"
)

const (
	testOrgID = "testOrg"
)

func TestDefaultClient(t *testing.T) {
	transport := client.DefaultHTTPClient("https://example.com", "v1/somepath")
	assert.NotNil(t, transport)
}

func toMap(t *testing.T, val interface{}) map[string]interface{} {
	t.Helper()
	body, err := json.Marshal(val)
	assert.NoError(t, err)
	var ret map[string]interface{}
	assert.NoError(t, json.Unmarshal(body, &ret))
	return ret
}
