///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"github.com/go-openapi/runtime"
	apiclient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

func multiAuth(writers ...runtime.ClientAuthInfoWriter) runtime.ClientAuthInfoWriter {
	return runtime.ClientAuthInfoWriterFunc(func(r runtime.ClientRequest, registry strfmt.Registry) error {
		for _, w := range writers {
			err := w.AuthenticateRequest(r, registry)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// GetAuthInfoWriter constructor an ClientAuthInfoWriter based on the SkipAuth flag
func GetAuthInfoWriter() runtime.ClientAuthInfoWriter {
	if dispatchConfig.SkipAuth == true {
		return multiAuth(apiclient.BasicAuth("devuser", "vmware"), apiclient.APIKeyAuth("cookie", "header", "cookie"))
	}
	return apiclient.APIKeyAuth("cookie", "header", dispatchConfig.Cookie)
}
