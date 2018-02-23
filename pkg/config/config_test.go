///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_loadConfig(t *testing.T) {
	conf := `{
  "openwhisk": {
    "auth_token": "<redacted>",
    "host": "10.0.10.3"
  },
  "openfaas": {
    "gateway": "http://gateway.openfaas:8080/"
  },
  "riff": {
    "gateway": "http://riff-riff-http-gateway.riff/",
    "funcNamespace": "default"
  },
  "registry": {
    "uri": "some-docker-user",
    "auth": "<redacted>"
  }
}`
	config, err := loadConfig(strings.NewReader(conf))
	require.NoError(t, err)
	assert.Equal(t, "some-docker-user", config.Registry.RegistryURI)
	assert.Equal(t, "http://riff-riff-http-gateway.riff/", config.Riff.Gateway)
	assert.Equal(t, "default", config.Riff.FuncNamespace)
}
