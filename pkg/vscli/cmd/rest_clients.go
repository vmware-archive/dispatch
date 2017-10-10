///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"crypto/tls"
	"fmt"
	"net/http"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	fnclient "gitlab.eng.vmware.com/serverless/serverless/pkg/function-manager/gen/client"
	imageclient "gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/client"
)

// NO TEST

func functionManagerClient() *fnclient.FunctionManager {
	host := fmt.Sprintf("%s:%d", vsConfig.Host, vsConfig.Port)
	tlsClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	transport := httptransport.NewWithClient(host, fnclient.DefaultBasePath, []string{"https"}, tlsClient)
	return fnclient.New(transport, strfmt.Default)
}

func imageManagerClient() *imageclient.ImageManager {
	host := fmt.Sprintf("%s:%d", vsConfig.Host, vsConfig.Port)
	tlsClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	transport := httptransport.NewWithClient(host, imageclient.DefaultBasePath, []string{"https"}, tlsClient)
	return imageclient.New(transport, strfmt.Default)
}
