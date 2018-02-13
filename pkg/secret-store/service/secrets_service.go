///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package service

import (
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/secret-store/gen/models"
)

type SecretNotFound struct {
	error
}

type SecretsService interface {
	AddSecret(models.Secret) (*models.Secret, error)
	GetSecrets(opts entitystore.Options) ([]*models.Secret, error)
	GetSecret(name string, opts entitystore.Options) (*models.Secret, error)
	UpdateSecret(secret models.Secret, opts entitystore.Options) (*models.Secret, error)
	DeleteSecret(name string, opts entitystore.Options) error
}
