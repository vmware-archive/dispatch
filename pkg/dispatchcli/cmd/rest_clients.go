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
	"github.com/vmware/dispatch/pkg/client"

	applicationclient "github.com/vmware/dispatch/pkg/application-manager/gen/client"
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

func getDispatchHost() string {
	if dispatchConfig.Scheme == "" {
		dispatchConfig.Scheme = "https"
	}

	// TODO(karols): this is a hack as it changes the global http.DefaultTransport.
	// Instead, Client constructor should accept a flag (or custom transport)
	if dispatchConfig.Insecure {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
			InsecureSkipVerify: dispatchConfig.Insecure,
		}
	}

	return fmt.Sprintf("%s://%s:%d", dispatchConfig.Scheme, dispatchConfig.Host, dispatchConfig.Port)
}

func functionManagerClient() client.FunctionsClient {
	return client.NewFunctionsClient(getDispatchHost(), GetAuthInfoWriter(), getOrganization())
}

func imageManagerClient() client.ImagesClient {
	return client.NewImagesClient(getDispatchHost(), GetAuthInfoWriter(), getOrganization())
}

func secretStoreClient() client.SecretsClient {
	return client.NewSecretsClient(getDispatchHost(), GetAuthInfoWriter(), getOrganization())
}

func serviceManagerClient() client.ServicesClient {
	return client.NewServicesClient(getDispatchHost(), GetAuthInfoWriter(), getOrganization())
}

func apiManagerClient() client.APIsClient {
	return client.NewAPIsClient(getDispatchHost(), GetAuthInfoWriter(), getOrganization())
}

func applicationManagerClient() *applicationclient.ApplicationManager {
	return applicationclient.New(httpTransport(applicationclient.DefaultBasePath), strfmt.Default)
}

func eventManagerClient() client.EventsClient {
	return client.NewEventsClient(getDispatchHost(), GetAuthInfoWriter(), getOrganization())
}

func identityManagerClient() client.IdentityClient {
	return client.NewIdentityClient(getDispatchHost(), GetAuthInfoWriter(), getOrganization())
}
