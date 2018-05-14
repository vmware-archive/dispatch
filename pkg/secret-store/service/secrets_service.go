///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package service

import (
	"github.com/vmware/dispatch/pkg/api/v1"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
)

// SecretNotFound is the error type when the secret is not found
type SecretNotFound struct {
	error
}

// SecretsService defines the secrets service interface
type SecretsService interface {
	AddSecret(v1.Secret) (*v1.Secret, error)
	GetSecrets(opts entitystore.Options) ([]*v1.Secret, error)
	GetSecret(name string, opts entitystore.Options) (*v1.Secret, error)
	UpdateSecret(secret v1.Secret, opts entitystore.Options) (*v1.Secret, error)
	DeleteSecret(name string, opts entitystore.Options) error
}
