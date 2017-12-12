///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// NO TESTS

package builder

import (
	"github.com/vmware/dispatch/pkg/secret-store/gen/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

type K8sSecretBuilder struct {
	Secret models.Secret
}

func NewK8sSecretBuilder(secret models.Secret) *K8sSecretBuilder {
	return &K8sSecretBuilder{
		Secret: secret,
	}
}

func (builder *K8sSecretBuilder) Build() v1.Secret {

	data := make(map[string][]byte)
	for k, v := range builder.Secret.Secrets {
		data[k] = []byte(v)
	}

	return v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: *builder.Secret.Name,
		},
		Data: data,
	}
}
