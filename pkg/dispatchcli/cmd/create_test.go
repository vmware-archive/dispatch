///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client/mocks"
	"github.com/vmware/dispatch/pkg/utils"
)

func TestCmdCreate(t *testing.T) {
	var buf bytes.Buffer
	path := createConfig(t, "")
	defer os.Remove(path) // clean up

	cli := NewCLI(os.Stdin, &buf, &buf)
	cli.SetOutput(&buf)
	cli.SetArgs([]string{fmt.Sprintf("--config=%s", path), "create"})
	err := cli.Execute()
	assert.Nil(t, err)
	assert.True(t, strings.Contains(buf.String(), "Create a resource"))
}

var secretSeed = `kind: Secret
name: open-sesame
secrets:
  password: OpenSesame
tags:
  - key: role
    value: test`

func TestCreateBatchSecret(t *testing.T) {
	var stdout, stderr bytes.Buffer

	cli := NewCLI(os.Stdin, &stdout, &stderr)

	tmpfile, err := ioutil.TempFile("", "seed")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name()) // clean up

	_, err = tmpfile.Write([]byte(secretSeed))
	assert.NoError(t, err)

	sc := &mocks.SecretsClient{}

	createMap := map[string]ModelAction{
		utils.SecretKind: CallCreateSecret(sc),
	}

	secret := &v1.Secret{
		Kind:    utils.SecretKind,
		Name:    swag.String("open-sesame"),
		Secrets: map[string]string{"password": "OpenSesame"},
		Tags:    []*v1.Tag{&v1.Tag{Key: "role", Value: "test"}},
	}

	sc.On("CreateSecret", mock.Anything, mock.Anything, secret).Once().Return(secret, nil)

	file = tmpfile.Name()
	err = importFile(&stdout, &stderr, cli, nil, createMap, "Created")
	assert.Nil(t, err)
}
