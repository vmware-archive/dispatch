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

func TestCreateBaseImage(t *testing.T) {
	fakeServer := fakeserver.NewFakeServer(nil)
	server := httptest.NewServer(fakeServer)
	defer server.Close()

	iclient := client.NewBaseImagesClient(server.URL, nil, testOrgID)

	imageBody := &v1.BaseImage{}

	imageResponse, err := iclient.CreateBaseImage(context.Background(), testOrgID, imageBody)
	assert.Error(t, err)
	assert.Nil(t, imageResponse)

	imageMap := toMap(t, imageBody)
	fakeServer.AddResponse("POST", "/v1/baseimage", imageMap, imageMap, 201)
	imageResponse, err = iclient.CreateBaseImage(context.Background(), testOrgID, imageBody)
	assert.NoError(t, err)
	assert.Equal(t, imageResponse, imageBody)

}
