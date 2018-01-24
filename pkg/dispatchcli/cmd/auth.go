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

// GetAuthInfoWriter constructor an ClientAuthInfoWriter
func GetAuthInfoWriter() runtime.ClientAuthInfoWriter {
	// Oauth2Proxy always expects a cookie header even if the server is setup with SkipAuth. Hence, set a dummy default.
	cookie := "unset"
	if dispatchConfig.Cookie != "" {
		cookie = dispatchConfig.Cookie
	}
	return apiclient.APIKeyAuth("cookie", "header", cookie)
}
