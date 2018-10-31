///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"github.com/go-openapi/runtime"
	"github.com/vmware/dispatch/pkg/client"
)

func imagesClient(config *serverConfig) client.ImagesClient {
	if config.ImageManager != "" {
		return client.NewImagesClient(config.ImageManager, getAuth(), "")
	}
	return client.NewImagesClient(getLocalEndpoint(config), getAuth(), "")

}

func functionsClient(config *serverConfig) client.FunctionsClient {
	if config.FunctionManager != "" {
		return client.NewFunctionsClient(config.FunctionManager, getAuth(), "")
	}
	return client.NewFunctionsClient(getLocalEndpoint(config), getAuth(), "")

}

func secretsClient(config *serverConfig) client.SecretsClient {
	if config.SecretsStore != "" {
		return client.NewSecretsClient(config.SecretsStore, getAuth(), "")
	}
	return client.NewSecretsClient(getLocalEndpoint(config), getAuth(), "")

}

func getAuth() runtime.ClientAuthInfoWriter {
	return client.AuthWithToken("cookie")
}
