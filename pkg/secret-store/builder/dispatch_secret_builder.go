///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// NO TESTS

package builder

import (
	"github.com/go-openapi/strfmt"
	k8sv1 "k8s.io/api/core/v1"

	dispatchv1 "github.com/vmware/dispatch/pkg/api/v1"
	secretstore "github.com/vmware/dispatch/pkg/secret-store"
	"github.com/vmware/dispatch/pkg/utils"
)

// DispatchSecretBuilder type
type DispatchSecretBuilder struct {
	k8sSecret k8sv1.Secret
	entity    secretstore.SecretEntity
}

// NewDispatchSecretBuilder creates a new DispatchSecretBuilder
func NewDispatchSecretBuilder(entity secretstore.SecretEntity, k8sSecret k8sv1.Secret) *DispatchSecretBuilder {
	return &DispatchSecretBuilder{
		k8sSecret: k8sSecret,
		entity:    entity,
	}
}

// Build converts a DispatchSecretBuilder to a swagger model Secret
func (builder *DispatchSecretBuilder) Build() dispatchv1.Secret {
	secretValue := dispatchv1.SecretValue{}
	for k, v := range builder.k8sSecret.Data {
		secretValue[k] = string(v)
	}
	tags := []*dispatchv1.Tag{}
	for k, v := range builder.entity.Tags {
		tags = append(tags, &dispatchv1.Tag{Key: k, Value: v})
	}
	return dispatchv1.Secret{
		ID:   strfmt.UUID(builder.k8sSecret.UID),
		Name: &builder.entity.Name,
		Kind: utils.SecretKind,
		// Name:    &builder.k8sSecret.Name,
		Secrets: secretValue,
		Tags:    tags,
	}
}
