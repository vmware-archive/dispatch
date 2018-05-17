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

func TestCreateServiceInstance(t *testing.T) {
	fakeServer := fakeserver.NewFakeServer(nil)
	server := httptest.NewServer(fakeServer)
	defer server.Close()

	sclient := client.NewServicesClient(server.URL, nil, testOrgID)

	serviceInstanceBody := &v1.ServiceInstance{}

	serviceInstanceResponse, err := sclient.CreateServiceInstance(context.Background(), serviceInstanceBody)
	assert.Error(t, err)
	assert.Nil(t, serviceInstanceResponse)

	serviceInstanceMap := toMap(t, serviceInstanceBody)
	fakeServer.AddResponse("POST", "/v1/serviceinstance", serviceInstanceMap, serviceInstanceMap, 201)
	serviceInstanceResponse, err = sclient.CreateServiceInstance(context.Background(), serviceInstanceBody)
	assert.NoError(t, err)
	assert.Equal(t, serviceInstanceResponse, serviceInstanceBody)

}
