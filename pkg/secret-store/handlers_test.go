///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package secretstore

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/secret-store/mocks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func TestReadSecretsSuccess(t *testing.T) {
	secretsAPI := &mocks.SecretInterface{}

	handlers := Handlers{
		secretsAPI: secretsAPI,
	}

	k8sSecrets := apiv1.SecretList{
		Items: []apiv1.Secret{
			apiv1.Secret{},
		},
	}

	listOptions := metav1.ListOptions{}
	secretsAPI.On("List", listOptions).Return(&k8sSecrets, nil)

	vmwSecrets, err := handlers.readSecrets(listOptions)

	assert.Equal(t, nil, err)
	assert.Equal(t, len(k8sSecrets.Items), len(vmwSecrets))
}

func TestReadSecretsError(t *testing.T) {
	secretsAPI := &mocks.SecretInterface{}

	handlers := Handlers{
		secretsAPI: secretsAPI,
	}

	listOptions := metav1.ListOptions{}
	expectedErr := errors.New("error")
	secretsAPI.On("List", listOptions).Return(nil, expectedErr)

	vmwSecrets, err := handlers.readSecrets(listOptions)

	assert.Equal(t, expectedErr, err)
	assert.Empty(t, vmwSecrets)
}
