///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"crypto/tls"
	"fmt"
	"net/http"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	apiclient "github.com/vmware/dispatch/pkg/api-manager/gen/client"
	applicationclient "github.com/vmware/dispatch/pkg/application-manager/gen/client"
	eventclient "github.com/vmware/dispatch/pkg/event-manager/gen/client"
	fnclient "github.com/vmware/dispatch/pkg/function-manager/gen/client"
	identitymanager "github.com/vmware/dispatch/pkg/identity-manager/gen/client"
	imageclient "github.com/vmware/dispatch/pkg/image-manager/gen/client"
	secretclient "github.com/vmware/dispatch/pkg/secret-store/gen/client"
)

// NO TEST

func tlsClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: dispatchConfig.Insecure,
			},
		},
	}
}

func httpTransport(path string) *httptransport.Runtime {
	host := fmt.Sprintf("%s:%d", dispatchConfig.Host, dispatchConfig.Port)
	if dispatchConfig.Scheme == "http" {
		return httptransport.NewWithClient(host, path, []string{"http"}, &http.Client{})
	}
	return httptransport.NewWithClient(host, path, []string{"https"}, tlsClient())
}

func functionManagerClient() *fnclient.FunctionManager {
	return fnclient.New(httpTransport(fnclient.DefaultBasePath), strfmt.Default)
}

func imageManagerClient() *imageclient.ImageManager {
	return imageclient.New(httpTransport(imageclient.DefaultBasePath), strfmt.Default)
}

func secretStoreClient() *secretclient.SecretStore {
	return secretclient.New(httpTransport(secretclient.DefaultBasePath), strfmt.Default)
}

func apiManagerClient() *apiclient.APIManager {
	return apiclient.New(httpTransport(apiclient.DefaultBasePath), strfmt.Default)
}

func applicationManagerClient() *applicationclient.ApplicationManager {
	return applicationclient.New(httpTransport(applicationclient.DefaultBasePath), strfmt.Default)
}

func eventManagerClient() *eventclient.EventManager {
	return eventclient.New(httpTransport(eventclient.DefaultBasePath), strfmt.Default)
}

func identityManagerClient() *identitymanager.IdentityManager {
	return identitymanager.New(httpTransport(identitymanager.DefaultBasePath), strfmt.Default)
}
