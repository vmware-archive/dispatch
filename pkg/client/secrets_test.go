///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package client_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/testing/fakeserver"
)

func TestCreateSecret(t *testing.T) {
	fakeServer := fakeserver.NewFakeServer(nil)
	server := httptest.NewServer(fakeServer)
	defer server.Close()

	sclient := client.NewSecretsClient(server.URL, nil)

	secretBody := &v1.Secret{}

	secretResponse, err := sclient.CreateSecret(context.Background(), secretBody)
	assert.Error(t, err)
	assert.Nil(t, secretResponse)

	secretMap := toMap(t, secretBody)
	fakeServer.AddResponse("POST", "/v1/secret", secretMap, secretMap, 201)
	secretResponse, err = sclient.CreateSecret(context.Background(), secretBody)
	assert.NoError(t, err)
	assert.Equal(t, secretResponse, secretBody)

}
