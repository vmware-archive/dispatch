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
  "function": {
    "openwhisk": {
      "authToken": "<redacted>",
      "host": "10.0.10.3"
    },
    "openfaas": {
      "gateway": "http://gateway.openfaas:8080/"
    },
    "riff": {
      "kafkaBrokers": ["transport-kafka.riff-system:9092"],
      "funcNamespace": "default"
    }
  },
  "registry": {
    "uri": "some-docker-user",
    "auth": "<redacted>"
  }
}`
	config, err := loadConfig(strings.NewReader(conf))
	require.NoError(t, err)
	assert.Equal(t, "some-docker-user", config.Registry.RegistryURI)
	assert.Equal(t, []string{"transport-kafka.riff-system:9092"}, config.Function.Riff.KafkaBrokers)
	assert.Equal(t, "default", config.Function.Riff.FuncNamespace)
}
