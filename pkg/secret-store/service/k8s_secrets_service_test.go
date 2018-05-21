///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	dispatchv1 "github.com/vmware/dispatch/pkg/api/v1"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	secretstore "github.com/vmware/dispatch/pkg/secret-store"
	"github.com/vmware/dispatch/pkg/secret-store/builder"
	"github.com/vmware/dispatch/pkg/secret-store/mocks"
)

func setup() K8sSecretsService {
	return K8sSecretsService{
		EntityStore: &mocks.EntityStore{},
		SecretsAPI:  &mocks.SecretInterface{},
	}
}

func TestGetSecretsSuccess(t *testing.T) {
	organizationID := "vmware"

	entityStore := &mocks.EntityStore{}

	var entities []*secretstore.SecretEntity
	entityStore.On("List", mock.Anything, organizationID, mock.Anything, &entities).Return(nil).Run(func(args mock.Arguments) {
		entitySlice := args.Get(3).(*[]*secretstore.SecretEntity)
		*entitySlice = append(*entitySlice, &secretstore.SecretEntity{
			BaseEntity: entitystore.BaseEntity{
				ID:   "000-000-001",
				Name: "psql creds",
			},
		})
		*entitySlice = append(*entitySlice, &secretstore.SecretEntity{
			BaseEntity: entitystore.BaseEntity{
				ID:   "000-000-002",
				Name: "twitter-api-key",
			},
		})
		entities = *entitySlice
	})

	k8sSecrets := []k8sv1.Secret{k8sv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "000-000-001",
		},
		Data: map[string][]byte{
			"username": []byte("white-rabbit"),
			"password": []byte("iml8_iml8"),
		},
	},
		k8sv1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "000-000-002",
			},
			Data: map[string][]byte{
				"apiKey": []byte("df1e8004-55f5-4aa8-8e89-73bd0f21a4de"),
			},
		},
	}

	secretsAPI := &mocks.SecretInterface{}

	secretsAPI.On("Get", "000-000-001", metav1.GetOptions{}).Return(&k8sSecrets[0], nil)
	secretsAPI.On("Get", "000-000-002", metav1.GetOptions{}).Return(&k8sSecrets[1], nil)

	secretsService := K8sSecretsService{
		EntityStore: entityStore,
		SecretsAPI:  secretsAPI,
	}

	secrets, _ := secretsService.GetSecrets(context.Background(), organizationID, entitystore.Options{})

	assert.Equal(t, 2, len(entities), "The number of entities should be 2")
	assert.Equal(t, len(secrets), len(entities), "The number of entities and secrets should be the same")
}

func TestGetSecretByNameSuccess(t *testing.T) {
	organizationID := "vmware"
	secretName := "psql creds"

	entityStore := &mocks.EntityStore{}

	var entities []*secretstore.SecretEntity
	entityStore.On("List", mock.Anything, organizationID, mock.Anything, &entities).Return(nil).Run(func(args mock.Arguments) {
		entitySlice := args.Get(3).(*[]*secretstore.SecretEntity)
		*entitySlice = append(*entitySlice, &secretstore.SecretEntity{
			BaseEntity: entitystore.BaseEntity{
				ID:   "000-000-001",
				Name: "psql creds",
			},
		})
		entities = *entitySlice
	})

	k8sSecret := k8sv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "000-000-001",
		},
		Data: map[string][]byte{
			"username": []byte("white-rabbit"),
			"password": []byte("iml8_iml8"),
		},
	}
	secretsAPI := &mocks.SecretInterface{}

	secretsAPI.On("Get", "000-000-001", metav1.GetOptions{}).Return(&k8sSecret, nil)

	secretsService := K8sSecretsService{
		EntityStore: entityStore,
		SecretsAPI:  secretsAPI,
	}

	secret, _ := secretsService.GetSecret(context.Background(), organizationID, secretName, entitystore.Options{})

	assert.NotNil(t, secret, "Received nil expected a secret")
	assert.Equal(t, secretName, *secret.Name, "Returned secret name does not match requested secret name")
}

func TestGetSecretByNameNotFound(t *testing.T) {
	organizationID := "vmware"

	entityStore := &mocks.EntityStore{}

	var entities []*secretstore.SecretEntity
	entityStore.On("List", mock.Anything, organizationID, mock.Anything, &entities).Return(nil).Run(func(args mock.Arguments) {
		entitySlice := args.Get(3).(*[]*secretstore.SecretEntity)
		entities = *entitySlice
	})

	secretsAPI := &mocks.SecretInterface{}

	secretsService := K8sSecretsService{
		EntityStore: entityStore,
		SecretsAPI:  secretsAPI,
	}

	secret, err := secretsService.GetSecret(context.Background(), organizationID, "psql creds", entitystore.Options{})

	secretsAPI.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
	assert.Nil(t, secret, "Was expecting a nil secret")
	assert.Equal(t, err, SecretNotFound{}, "Was expecting a SecretNotFound error")
}

func TestGetSecretsEnityStoreKubernetesDiscrepancy(t *testing.T) {

}

func TestAddSecretSuccess(t *testing.T) {
	organizationID := "vmware"
	principalSecretName := "psql cred"
	principal := dispatchv1.Secret{
		Name: &principalSecretName,
		Secrets: dispatchv1.SecretValue{
			"username": "white-rabbit",
			"password": "iml8_iml8",
		},
	}

	entityStore := &mocks.EntityStore{}

	secretUUID := "000-000-001"
	entityStore.On("Add", mock.Anything, mock.Anything).Return(secretUUID, nil)

	k8sSecret := builder.NewK8sSecretBuilder(principal).Build()

	secretsAPI := &mocks.SecretInterface{}
	secretsAPI.On("Create", mock.Anything).Return(&k8sSecret, nil)

	secretsService := K8sSecretsService{
		EntityStore: entityStore,
		SecretsAPI:  secretsAPI,
	}

	actualSecret, _ := secretsService.AddSecret(context.Background(), organizationID, principal)
	assert.Equal(t, principal.Name, actualSecret.Name, "Secret created successfully")
}

func TestAddSecretDuplicateSecret(t *testing.T) {
	organizationID := "vmware"
	principalSecretName := "psql cred"
	principal := dispatchv1.Secret{
		Name: &principalSecretName,
		Secrets: dispatchv1.SecretValue{
			"username": "white-rabbit",
			"password": "password",
		},
	}

	entityStore := &mocks.EntityStore{}

	entityStore.On("Add", mock.Anything, mock.Anything).Return("", errors.New("Duplicate Name"))

	secretsService := K8sSecretsService{
		EntityStore: entityStore,
		SecretsAPI:  nil,
	}

	createdSecret, err := secretsService.AddSecret(context.Background(), organizationID, principal)

	assert.Nil(t, createdSecret, "Got a created secret, expected nil")
	assert.NotEmpty(t, err, "Did not recieve an error when duplicating a name in the entity store.")
}

func TestDeleteSecretSuccess(t *testing.T) {
	organizationID := "vmware"
	secretName := "psql creds"
	entityStore := &mocks.EntityStore{}

	entity := secretstore.SecretEntity{}

	entityStore.On("Find", mock.Anything, organizationID, secretName, mock.Anything, &entity).Return(true, nil).Run(func(args mock.Arguments) {
		entityArg := args.Get(4).(*secretstore.SecretEntity)
		*entityArg = secretstore.SecretEntity{
			BaseEntity: entitystore.BaseEntity{
				ID: "000-000-001",
			},
		}
		entity = *entityArg
	})

	secretsAPI := &mocks.SecretInterface{}
	secretsAPI.On("Delete", "000-000-001", &metav1.DeleteOptions{}).Return(nil)

	entityStore.On("Delete", mock.Anything, organizationID, secretName, mock.Anything).Return(nil)

	secretsService := K8sSecretsService{
		EntityStore: entityStore,
		SecretsAPI:  secretsAPI,
	}

	err := secretsService.DeleteSecret(context.Background(), organizationID, secretName, entitystore.Options{})

	assert.Nil(t, err, "There was an error deleting the secret")
}

func TestDeleteSecretNotExist(t *testing.T) {
	organizationID := "vmware"
	secretName := "nonexistent"
	entityStore := &mocks.EntityStore{}

	entityStore.On("Find", mock.Anything, organizationID, secretName, mock.Anything, mock.Anything).Return(false, nil)

	secretsAPI := &mocks.SecretInterface{}

	secretsService := K8sSecretsService{
		EntityStore: entityStore,
		SecretsAPI:  secretsAPI,
	}

	err := secretsService.DeleteSecret(context.Background(), organizationID, "nonexistent", entitystore.Options{})

	assert.Equal(t, SecretNotFound{}, err, "Should have returned SecretNotFound error")
	secretsAPI.AssertNotCalled(t, "Delete", "Kubernetes secrets Delete was called and should not have been")
	entityStore.AssertNotCalled(t, "Delete", "EntityStore Delete was called and should not have been")
}

func TestUpdateSecretSuccess(t *testing.T) {
	organizationID := "vmware"
	secretName := "psql creds"
	entityStore := &mocks.EntityStore{}

	secretEntity := secretstore.SecretEntity{
		BaseEntity: entitystore.BaseEntity{
			ID:   "000-000-001",
			Name: secretName,
		},
	}

	entityStore.On("Find", mock.Anything, organizationID, mock.Anything, mock.Anything, mock.Anything).Return(true, nil).Run(func(args mock.Arguments) {
		entityInput := args.Get(4).(*secretstore.SecretEntity)
		*entityInput = secretEntity
	})

	secretsAPI := &mocks.SecretInterface{}

	principal := dispatchv1.Secret{
		Name: &secretEntity.ID,
		Secrets: dispatchv1.SecretValue{
			"username": "white-rabbit",
			"password": "im_l8_im_l8",
		},
	}

	k8sSecret := builder.NewK8sSecretBuilder(principal).Build()
	secretsAPI.On("Update", mock.Anything).Return(&k8sSecret, nil)

	secretsService := K8sSecretsService{
		EntityStore: entityStore,
		SecretsAPI:  secretsAPI,
	}

	_, err := secretsService.UpdateSecret(context.Background(), organizationID, dispatchv1.Secret{
		Name: &secretName,
		Secrets: dispatchv1.SecretValue{
			"username": "white-rabbit",
			"password": "im_l8_im_l8",
		},
	}, entitystore.Options{})

	assert.Nil(t, err, "UpdateSecret returned unexpected error")
	entityStore.AssertCalled(t, "Find", mock.Anything, organizationID, mock.Anything, mock.Anything, mock.Anything)
	secretsAPI.AssertCalled(t, "Update", mock.Anything)
}

func TestUpdateSecretNotExist(t *testing.T) {
	organizationID := "vmware"
	secretName := "nonexistant"
	es := &mocks.EntityStore{}

	es.On("Find", mock.Anything, organizationID, mock.Anything, mock.Anything, mock.Anything).Return(false, nil)

	secretsAPI := &mocks.SecretInterface{}

	secretsService := K8sSecretsService{
		EntityStore: es,
		SecretsAPI:  secretsAPI,
	}

	secret := dispatchv1.Secret{
		Name: &secretName,
	}
	_, err := secretsService.UpdateSecret(context.Background(), organizationID, secret, entitystore.Options{})

	assert.Equal(t, SecretNotFound{}, err, "Should have returned SecretNotFound error")
	secretsAPI.AssertNotCalled(t, "Update", "Kubernetes secrets Update was called and should not have been.")
}
