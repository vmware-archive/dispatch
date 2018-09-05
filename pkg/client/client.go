///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package client

import (
	"net/http"
	"strings"

	"github.com/go-openapi/runtime"
	swaggerclient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// TokenHeaderName defines the cookie token
const TokenHeaderName = "cookie"

// AuthWithUserPassword authenticates with username and password
func AuthWithUserPassword(username string, password string) runtime.ClientAuthInfoWriter {
	return swaggerclient.BasicAuth(username, password)
}

// AuthWithToken authenticates with a token
func AuthWithToken(token string) runtime.ClientAuthInfoWriter {
	return swaggerclient.APIKeyAuth(TokenHeaderName, "header", token)
}

// AuthWithMulti writes authentication info to a request
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

// DefaultHTTPClient Creates a default HTTP transport for all clients
func DefaultHTTPClient(host, basePath string) *swaggerclient.Runtime {
	schemas := []string{"http"}
	if idx := strings.Index(host, "://"); idx != -1 {
		// http schema included in path
		schemas = []string{host[:idx]}
		host = host[idx+3:]

	}
	transport := swaggerclient.New(host, basePath, schemas)
	transport.Transport = NewTracingRoundTripper(http.DefaultTransport)
	return transport
}

// baseClient represents fields & methods common for all Dispatch services.
type baseClient struct {
	organizationID string
	projectName    string
}

func (c *baseClient) getOrgID(organizationID string) string {
	if organizationID != "" {
		return organizationID
	}
	return c.organizationID
}
