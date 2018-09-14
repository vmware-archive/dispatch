///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmware/dispatch/pkg/utils/knaming"
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/typed/core/v1"

	dispatchv1 "github.com/vmware/dispatch/pkg/api/v1"
)

//const testOrg = "vmware"

const (
	n1      = "000-000-001"
	n2      = "000-000-002"
	project = "project"
)

var (
	m1 = dispatchv1.Meta{Org: testOrg, Project: project, Name: n1}
	m2 = dispatchv1.Meta{Org: testOrg, Project: project, Name: n2}
)

func newFakeCoreV1() v1.CoreV1Interface {
	return fake.NewSimpleClientset().CoreV1()
}

func k8sSecretsService(k8sCoreV1 v1.CoreV1Interface) *K8sSecretsService {
	return &K8sSecretsService{
		K8sAPI: k8sCoreV1,
	}
}

func setupFakeCoreV1() v1.CoreV1Interface {

	k8sSecrets := []k8sv1.Secret{{
		ObjectMeta: metav1.ObjectMeta{
			Name: knaming.SecretName(m1),
			Labels: map[string]string{
				knaming.NameLabel:    n1,
				knaming.ProjectLabel: project,
			},
			Annotations: map[string]string{
				knaming.InitialObjectAnnotation: knaming.ToJSONString(dispatchv1.Secret{
					Meta: m1,
				}),
			},
		},
		Data: map[string][]byte{
			knaming.SecretKey: knaming.ToJSONBytes(map[string]string{
				"username": "white-rabbit",
				"password": "iml8_iml8",
			}),
		},
	},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: knaming.SecretName(m2),
				Labels: map[string]string{
					knaming.NameLabel:    n2,
					knaming.ProjectLabel: project,
				},
				Annotations: map[string]string{
					knaming.InitialObjectAnnotation: knaming.ToJSONString(dispatchv1.Secret{
						Meta: m2,
					}),
				},
			},
			Data: map[string][]byte{
				knaming.SecretKey: knaming.ToJSONBytes(map[string]string{
					"apiKey": "df1e8004-55f5-4aa8-8e89-73bd0f21a4de",
				}),
			},
		},
	}

	fakeCoreV1 := newFakeCoreV1()
	fakeSecrets := fakeCoreV1.Secrets(testOrg)
	for i := range k8sSecrets {
		fakeSecrets.Create(&k8sSecrets[i])
	}

	return fakeCoreV1
}

func TestGetSecretsSuccess(t *testing.T) {
	secretsService := k8sSecretsService(setupFakeCoreV1())
	secrets, err := secretsService.GetSecrets(context.Background(), &dispatchv1.Meta{Org: testOrg, Project: project})
	require.NoError(t, err)
	assert.Equal(t, 2, len(secrets), "The number of secrets should be the same as stored in k8s")
}

func TestGetSecretByNameSuccess(t *testing.T) {
	secretsService := k8sSecretsService(setupFakeCoreV1())
	secret, err := secretsService.GetSecret(context.Background(), &dispatchv1.Meta{Org: testOrg, Project: project, Name: n1})
	require.NoError(t, err)

	require.NotNil(t, secret, "Received nil expected a secret")
	assert.Equal(t, n1, secret.Meta.Name, "Returned secret name does not match requested secret name")
	assert.Equal(t, "white-rabbit", secret.Secrets["username"])
}

func TestGetSecretByNameNotFound(t *testing.T) {

	secretsService := k8sSecretsService(setupFakeCoreV1())
	secret, err := secretsService.GetSecret(context.Background(), &dispatchv1.Meta{Org: testOrg, Project: project, Name: "psql-creds"})

	assert.Equal(t, SecretNotFound{}, err, "Was expecting a SecretNotFound error")
	assert.Nil(t, secret, "Was expecting a nil secret")
}

func TestAddSecretSuccess(t *testing.T) {
	secretName := "psql cred"
	secret := &dispatchv1.Secret{
		Meta: dispatchv1.Meta{
			Org:     testOrg,
			Project: project,
			Name:    secretName,
		},
		Secrets: dispatchv1.SecretValue{
			"username": "white-rabbit",
			"password": "iml8_iml8",
		},
	}

	secretsService := k8sSecretsService(setupFakeCoreV1())

	actualSecret, err := secretsService.AddSecret(context.Background(), secret)
	require.NoError(t, err)
	assert.Equal(t, secret.Meta.Name, actualSecret.Meta.Name, "Secret created successfully")
}

func TestAddSecretDuplicateSecret(t *testing.T) {
	secretName := n1
	secret := &dispatchv1.Secret{
		Meta: dispatchv1.Meta{
			Org:     testOrg,
			Project: project,
			Name:    secretName,
		},
		Secrets: dispatchv1.SecretValue{
			"username": "white-rabbit",
			"password": "iml8_iml8",
		},
	}

	secretsService := k8sSecretsService(setupFakeCoreV1())

	createdSecret, err := secretsService.AddSecret(context.Background(), secret)
	assert.Error(t, err)
	assert.Nil(t, createdSecret, "Got a created secret, expected nil")
	//assert.NotEmpty(t, err, "Did not recieve an error when duplicating a name in the entity store.")
}

func TestDeleteSecretSuccess(t *testing.T) {
	secretsService := k8sSecretsService(setupFakeCoreV1())
	err := secretsService.DeleteSecret(context.Background(), &dispatchv1.Meta{Name: n1, Org: testOrg, Project: project})
	require.NoError(t, err)

	secret, err := secretsService.GetSecret(context.Background(), &dispatchv1.Meta{Name: n1, Org: testOrg, Project: project})
	require.Nil(t, secret)
	assert.Equal(t, SecretNotFound{}, err)
}

func TestDeleteSecretNotExist(t *testing.T) {
	secretsService := k8sSecretsService(setupFakeCoreV1())
	secretName := "nonexistent"
	err := secretsService.DeleteSecret(context.Background(), &dispatchv1.Meta{Name: secretName, Org: testOrg, Project: project})
	assert.Equal(t, SecretNotFound{}, err, "Should have returned SecretNotFound error")
}

func TestUpdateSecretSuccess(t *testing.T) {

	usernameField := "username"
	yellowRabbit := "yellow-rabbit"

	secret := &dispatchv1.Secret{
		Meta: dispatchv1.Meta{
			Name:    n1,
			Org:     testOrg,
			Project: project,
		},
		Secrets: dispatchv1.SecretValue{
			usernameField: yellowRabbit,
			"password":    "im_l8_im_l8",
		},
	}

	secretsService := k8sSecretsService(setupFakeCoreV1())

	updatedSecret, err := secretsService.UpdateSecret(context.Background(), secret)
	require.NoError(t, err, "UpdateSecret returned unexpected error")
	assert.Equal(t, yellowRabbit, updatedSecret.Secrets[usernameField])
}

func TestUpdateSecretNotExist(t *testing.T) {
	secretName := "nonexistent"

	secretsService := k8sSecretsService(setupFakeCoreV1())

	secret := &dispatchv1.Secret{
		Meta: dispatchv1.Meta{
			Org:     testOrg,
			Project: project,
			Name:    secretName,
		},
	}
	_, err := secretsService.UpdateSecret(context.Background(), secret)

	assert.Equal(t, SecretNotFound{}, err, "Should have returned SecretNotFound error")
}
