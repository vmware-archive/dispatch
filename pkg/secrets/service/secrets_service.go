///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package service

import (
	"context"

	"github.com/vmware/dispatch/pkg/api/v1"
)

// SecretNotFound is the error type when the secret is not found
type SecretNotFound struct {
	error
}

// SecretsService defines the secrets service interface
type SecretsService interface {
	AddSecret(ctx context.Context, secret *v1.Secret) (*v1.Secret, error)
	GetSecrets(ctx context.Context, meta *v1.Meta) ([]*v1.Secret, error)
	GetSecret(ctx context.Context, meta *v1.Meta) (*v1.Secret, error)
	UpdateSecret(ctx context.Context, secret *v1.Secret) (*v1.Secret, error)
	DeleteSecret(ctx context.Context, meta *v1.Meta) error
}
