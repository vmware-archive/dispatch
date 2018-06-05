///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package client

import (
	"context"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/pkg/errors"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/client/operations"

	"github.com/vmware/dispatch/pkg/api/v1"
	swaggerclient "github.com/vmware/dispatch/pkg/identity-manager/gen/client"
)

// VersionClient gets the version info from the API
type VersionClient interface {
	// GetVersion returns either *v1.Version or error
	GetVersion(ctx context.Context) (*v1.Version, error)
}

// DefaultVersionClient is the implementation of VersionClient
type DefaultVersionClient struct {
	client *swaggerclient.IdentityManager
	auth   runtime.ClientAuthInfoWriter
}

// NewVersionClient creates an instance of *DefaultVersionClient
func NewVersionClient(host string, auth runtime.ClientAuthInfoWriter) *DefaultVersionClient {
	transport := DefaultHTTPClient(host, swaggerclient.DefaultBasePath)
	return &DefaultVersionClient{
		client: swaggerclient.New(transport, strfmt.Default),
		auth:   auth,
	}
}

// GetVersion returns either *v1.Version or error
func (vc *DefaultVersionClient) GetVersion(ctx context.Context) (*v1.Version, error) {
	params := &operations.GetVersionParams{
		Context: context.Background(),
	}
	resp, err := vc.client.Operations.GetVersion(params, vc.auth)
	if err != nil {
		return nil, errors.Wrapf(err, "error fetching version info")
	}
	return resp.Payload, nil
}
