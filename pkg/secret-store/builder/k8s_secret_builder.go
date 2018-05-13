///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// NO TESTS

package builder

import (
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	dispatchv1 "github.com/vmware/dispatch/pkg/api/v1"
)

// K8sSecretBuilder type
type K8sSecretBuilder struct {
	Secret dispatchv1.Secret
}

// NewK8sSecretBuilder creates a new K8sSecretBuilder
func NewK8sSecretBuilder(secret dispatchv1.Secret) *K8sSecretBuilder {
	return &K8sSecretBuilder{
		Secret: secret,
	}
}

// Build converts a K8sSecretBuilder to a k8s secret
func (builder *K8sSecretBuilder) Build() k8sv1.Secret {

	data := make(map[string][]byte)
	for k, v := range builder.Secret.Secrets {
		data[k] = []byte(v)
	}

	return k8sv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: *builder.Secret.Name,
		},
		Data: data,
	}
}
