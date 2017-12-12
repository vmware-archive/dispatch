///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package service

import (
	"github.com/vmware/dispatch/pkg/secret-store/gen/models"
)

type SecretsService interface {
	AddSecret(secret models.Secret) (*models.Secret, error)
	GetSecrets() ([]*models.Secret, error)
	GetSecret(name string) (*models.Secret, error)
	UpdateSecret(secret models.Secret) (*models.Secret, error)
	DeleteSecret(name string) error
}
