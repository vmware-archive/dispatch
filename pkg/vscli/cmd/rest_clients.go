///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"fmt"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	fnclient "gitlab.eng.vmware.com/serverless/serverless/pkg/function-manager/gen/client"
	imageclient "gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/client"
)

// NO TEST

func functionManagerClient() *fnclient.FunctionManager {
	host := fmt.Sprintf("%s:%d", vsConfig.Host, vsConfig.Port)
	transport := httptransport.New(host, fnclient.DefaultBasePath, []string{"http"})
	return fnclient.New(transport, strfmt.Default)
}

func imageManagerClient() *imageclient.ImageManager {
	host := fmt.Sprintf("%s:%d", vsConfig.Host, vsConfig.Port)
	transport := httptransport.New(host, imageclient.DefaultBasePath, []string{"http"})
	return imageclient.New(transport, strfmt.Default)
}
