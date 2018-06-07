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

func TestCreateAPI(t *testing.T) {
	fakeServer := fakeserver.NewFakeServer(nil)
	server := httptest.NewServer(fakeServer)
	defer server.Close()

	aclient := client.NewAPIsClient(server.URL, nil, testOrgID)

	apiBody := &v1.API{}

	apiResponse, err := aclient.CreateAPI(context.Background(), testOrgID, apiBody)
	assert.Error(t, err)
	assert.Nil(t, apiResponse)

	apiMap := toMap(t, apiBody)
	fakeServer.AddResponse("POST", "/v1/api", apiMap, apiMap, 200)
	apiResponse, err = aclient.CreateAPI(context.Background(), testOrgID, apiBody)
	assert.NoError(t, err)
	assert.Equal(t, apiResponse, apiBody)

}
