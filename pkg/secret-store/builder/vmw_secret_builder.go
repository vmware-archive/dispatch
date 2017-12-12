///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// NO TESTS

package builder

import (
	strfmt "github.com/go-openapi/strfmt"
	"github.com/vmware/dispatch/pkg/secret-store/gen/models"
	"k8s.io/client-go/pkg/api/v1"
)

type VmwSecretBuilder struct {
	k8sSecret v1.Secret
}

func NewVmwSecretBuilder(k8sSecret v1.Secret) *VmwSecretBuilder {
	return &VmwSecretBuilder{
		k8sSecret: k8sSecret,
	}
}

func (builder *VmwSecretBuilder) Build() models.Secret {
	secretValue := models.SecretValue{}
	for k, v := range builder.k8sSecret.Data {
		secretValue[k] = string(v)
	}
	return models.Secret{
		ID:      strfmt.UUID(builder.k8sSecret.UID),
		Name:    &builder.k8sSecret.Name,
		Secrets: secretValue,
	}
}
