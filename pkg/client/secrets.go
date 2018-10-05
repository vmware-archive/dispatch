///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package client

import (
	"context"
	"fmt"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	"github.com/vmware/dispatch/pkg/api/v1"
	swaggerclient "github.com/vmware/dispatch/pkg/secrets/gen/client"
	secretclient "github.com/vmware/dispatch/pkg/secrets/gen/client/secret"
)

// SecretsClient defines the secrets client interface
type SecretsClient interface {
	CreateSecret(ctx context.Context, organizationID string, secret *v1.Secret) (*v1.Secret, error)
	DeleteSecret(ctx context.Context, organizationID string, secretName string) error
	UpdateSecret(ctx context.Context, organizationID string, secret *v1.Secret) (*v1.Secret, error)
	GetSecret(ctx context.Context, organizationID string, secretName string) (*v1.Secret, error)
	ListSecrets(ctx context.Context, organizationID string) ([]v1.Secret, error)
}

// NewSecretsClient is used to create a new secrets client
func NewSecretsClient(host string, auth runtime.ClientAuthInfoWriter, organizationID, project string) SecretsClient {
	transport := DefaultHTTPClient(host, swaggerclient.DefaultBasePath)
	return &DefaultSecretsClient{
		baseClient: baseClient{
			organizationID: organizationID,
			projectName:    project,
		},
		client: swaggerclient.New(transport, strfmt.Default),
		auth:   auth,
	}
}

// DefaultSecretsClient defines the default secrets client
type DefaultSecretsClient struct {
	baseClient

	client *swaggerclient.Secrets
	auth   runtime.ClientAuthInfoWriter
}

// CreateSecret creates a secret
func (c *DefaultSecretsClient) CreateSecret(ctx context.Context, organizationID string, secret *v1.Secret) (*v1.Secret, error) {
	params := secretclient.AddSecretParams{
		Context:      ctx,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
		Secret:       secret,
	}
	response, err := c.client.Secret.AddSecret(&params)
	if err != nil {
		return nil, createSecretSwaggerError(err)
	}
	return response.Payload, nil
}

func createSecretSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *secretclient.AddSecretBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *secretclient.AddSecretUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *secretclient.AddSecretForbidden:
		return NewErrorForbidden(v.Payload)
	case *secretclient.AddSecretConflict:
		return NewErrorAlreadyExists(v.Payload)
	case *secretclient.AddSecretDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// DeleteSecret deletes a secret
func (c *DefaultSecretsClient) DeleteSecret(ctx context.Context, organizationID string, secretName string) error {
	params := secretclient.DeleteSecretParams{
		Context:      ctx,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
		SecretName:   secretName,
	}
	_, err := c.client.Secret.DeleteSecret(&params)
	if err != nil {
		return deleteSecretSwaggerError(err)
	}
	return nil
}

func deleteSecretSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *secretclient.DeleteSecretBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *secretclient.DeleteSecretUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *secretclient.DeleteSecretForbidden:
		return NewErrorForbidden(v.Payload)
	case *secretclient.DeleteSecretNotFound:
		return NewErrorNotFound(v.Payload)
	case *secretclient.DeleteSecretDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// UpdateSecret updates a secret
func (c *DefaultSecretsClient) UpdateSecret(ctx context.Context, organizationID string, secret *v1.Secret) (*v1.Secret, error) {
	params := secretclient.UpdateSecretParams{
		Context:      ctx,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
		Secret:       secret,
		SecretName:   *secret.Name,
	}
	response, err := c.client.Secret.UpdateSecret(&params)
	if err != nil {
		return nil, updateSecretSwaggerError(err)
	}
	return response.Payload, nil
}

func updateSecretSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *secretclient.UpdateSecretBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *secretclient.UpdateSecretUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *secretclient.UpdateSecretForbidden:
		return NewErrorForbidden(v.Payload)
	case *secretclient.UpdateSecretNotFound:
		return NewErrorNotFound(v.Payload)
	case *secretclient.UpdateSecretDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// GetSecret retrieves a secret
func (c *DefaultSecretsClient) GetSecret(ctx context.Context, organizationID string, secretName string) (*v1.Secret, error) {
	params := secretclient.GetSecretParams{
		Context:      ctx,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
		SecretName:   secretName,
	}
	response, err := c.client.Secret.GetSecret(&params)
	if err != nil {
		return nil, getSecretSwaggerError(err)
	}
	return response.Payload, nil
}

func getSecretSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *secretclient.GetSecretBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *secretclient.GetSecretUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *secretclient.GetSecretForbidden:
		return NewErrorForbidden(v.Payload)
	case *secretclient.GetSecretNotFound:
		return NewErrorNotFound(v.Payload)
	case *secretclient.GetSecretDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// ListSecrets lists secrets
func (c *DefaultSecretsClient) ListSecrets(ctx context.Context, organizationID string) ([]v1.Secret, error) {
	params := secretclient.GetSecretsParams{
		Context:      ctx,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.Secret.GetSecrets(&params)
	if err != nil {
		return nil, listSecretsSwaggerError(err)
	}
	secrets := []v1.Secret{}
	for _, secret := range response.Payload {
		secrets = append(secrets, *secret)
	}
	return secrets, nil
}

func listSecretsSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *secretclient.GetSecretsUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *secretclient.GetSecretsForbidden:
		return NewErrorForbidden(v.Payload)
	case *secretclient.GetSecretsDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}
