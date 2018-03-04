///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// NO TESTS

package builder

import (
	strfmt "github.com/go-openapi/strfmt"
	secretstore "github.com/vmware/dispatch/pkg/secret-store"
	"github.com/vmware/dispatch/pkg/secret-store/gen/models"
	"k8s.io/api/core/v1"
)

// VmwSecretBuilder type
type VmwSecretBuilder struct {
	k8sSecret v1.Secret
	entity    secretstore.SecretEntity
}

// NewVmwSecretBuilder creates a new VmwSecretBuilder
func NewVmwSecretBuilder(entity secretstore.SecretEntity, k8sSecret v1.Secret) *VmwSecretBuilder {
	return &VmwSecretBuilder{
		k8sSecret: k8sSecret,
		entity:    entity,
	}
}

// Build converts a VmwSecretBuilder to a swagger model Secret
func (builder *VmwSecretBuilder) Build() models.Secret {
	secretValue := models.SecretValue{}
	for k, v := range builder.k8sSecret.Data {
		secretValue[k] = string(v)
	}
	tags := []*models.Tag{}
	for k, v := range builder.entity.Tags {
		tags = append(tags, &models.Tag{Key: k, Value: v})
	}
	return models.Secret{
		ID:   strfmt.UUID(builder.k8sSecret.UID),
		Name: &builder.entity.Name,
		// Name:    &builder.k8sSecret.Name,
		Secrets: secretValue,
		Tags:    tags,
	}
}
