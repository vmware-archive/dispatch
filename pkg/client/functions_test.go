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

func TestCreateFunction(t *testing.T) {
	fakeServer := fakeserver.NewFakeServer(nil)
	server := httptest.NewServer(fakeServer)
	defer server.Close()

	fclient := client.NewFunctionsClient(server.URL, nil)

	functionBody := &v1.Function{}

	functionResponse, err := fclient.CreateFunction(context.Background(), testOrgID, functionBody)
	assert.Error(t, err)
	assert.Nil(t, functionResponse)

	functionMap := toMap(t, functionBody)
	fakeServer.AddResponse("POST", "/v1/function", functionMap, functionMap, 201)
	functionResponse, err = fclient.CreateFunction(context.Background(), testOrgID, functionBody)
	assert.NoError(t, err)
	assert.Equal(t, functionResponse, functionBody)

}
