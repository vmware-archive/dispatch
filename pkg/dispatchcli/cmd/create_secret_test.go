///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client/mocks"
)

func TestCmdCreateSecret(t *testing.T) {
	var buf bytes.Buffer

	cli := NewCLI(os.Stdin, &buf, &buf)
	cli.SetOutput(&buf)
	cli.SetArgs([]string{"create", "secret", "--help"})
	err := cli.Execute()
	assert.Nil(t, err)
	assert.True(t, strings.Contains(buf.String(), "Create a dispatch secret"))
}

func TestCreateSecret(t *testing.T) {
	var stdout, stderr bytes.Buffer

	cli := NewCLI(os.Stdin, &stdout, &stderr)

	sc := &mocks.SecretsClient{}

	body := map[string]string{"secretKey": "secretValue"}

	tmpfile, err := ioutil.TempFile("", "createSecret")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name()) // clean up

	enc := json.NewEncoder(tmpfile)
	err = enc.Encode(body)
	require.NoError(t, err)

	args := []string{"test", tmpfile.Name()}
	dispatchConfig.JSON = true

	secret := &v1.Secret{
		Meta: v1.Meta{
			Name: args[0],
		},
		Secrets: body,
	}

	sc.On("CreateSecret", mock.Anything, mock.Anything, secret).Once().Return(secret, nil)
	err = createSecret(&stdout, &stderr, cli, args, sc)
	assert.NoError(t, err)

	var secObj v1.Secret
	err = json.Unmarshal(stdout.Bytes(), &secObj)
	assert.NoError(t, err)
	assert.Equal(t, "test", secObj.Meta.Name)
	assert.Equal(t, v1.SecretValue{"secretKey": "secretValue"}, secObj.Secrets)
}
