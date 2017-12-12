///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package client

import (
	"github.com/go-openapi/runtime"
	swaggerclient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NO TESTS

const TokenHeaderName = "cookie"

func AuthWithUserPassword(username string, password string) runtime.ClientAuthInfoWriter {
	return swaggerclient.BasicAuth(username, password)
}

func AuthWithToken(token string) runtime.ClientAuthInfoWriter {
	return swaggerclient.APIKeyAuth(TokenHeaderName, "header", token)
}

func AuthWithMulti(writers ...runtime.ClientAuthInfoWriter) runtime.ClientAuthInfoWriter {
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
