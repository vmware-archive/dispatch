///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package service

import (
	"errors"
	"testing"

	"github.com/docker/libkv/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	secretstore "github.com/vmware/dispatch/pkg/secret-store"
	"github.com/vmware/dispatch/pkg/secret-store/builder"
	"github.com/vmware/dispatch/pkg/secret-store/gen/models"
	"github.com/vmware/dispatch/pkg/secret-store/mocks"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func setup() K8sSecretsService {
	return K8sSecretsService{
		EntityStore: &mocks.EntityStore{},
		SecretsAPI:  &mocks.SecretInterface{},
	}
}

func TestGetSecretsSuccess(t *testing.T) {
	organizationId := "vmware"

	entityStore := &mocks.EntityStore{}

	var entities []*secretstore.SecretEntity
	entityStore.On("List", organizationId, mock.Anything, &entities).Return(nil).Run(func(args mock.Arguments) {
		entitySlice := args.Get(2).(*[]*secretstore.SecretEntity)
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

	k8sSecrets := []v1.Secret{v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "000-000-001",
		},
		Data: map[string][]byte{
			"username": []byte("white-rabbit"),
			"password": []byte("iml8_iml8"),
		},
	},
		v1.Secret{
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
		OrgID:       organizationId,
	}

	secrets, _ := secretsService.GetSecrets(entitystore.Options{})

	assert.Equal(t, 2, len(entities), "The number of entities should be 2")
	assert.Equal(t, len(secrets), len(entities), "The number of entities and secrets should be the same")
}

func TestGetSecretByNameSuccess(t *testing.T) {
	organizationId := "vmware"
	secretName := "psql creds"

	entityStore := &mocks.EntityStore{}

	var entities []*secretstore.SecretEntity
	entityStore.On("List", organizationId, mock.Anything, &entities).Return(nil).Run(func(args mock.Arguments) {
		entitySlice := args.Get(2).(*[]*secretstore.SecretEntity)
		*entitySlice = append(*entitySlice, &secretstore.SecretEntity{
			BaseEntity: entitystore.BaseEntity{
				ID:   "000-000-001",
				Name: "psql creds",
			},
		})
		entities = *entitySlice
	})

	k8sSecret := v1.Secret{
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
		OrgID:       organizationId,
	}

	secret, _ := secretsService.GetSecret(secretName, entitystore.Options{})

	assert.NotNil(t, secret, "Received nil expected a secret")
	assert.Equal(t, secretName, *secret.Name, "Returned secret name does not match requested secret name")
}

func TestGetSecretByNameNotFound(t *testing.T) {
	organizationId := "vmware"

	entityStore := &mocks.EntityStore{}

	var entities []*secretstore.SecretEntity
	entityStore.On("List", organizationId, mock.Anything, &entities).Return(nil).Run(func(args mock.Arguments) {
		entitySlice := args.Get(2).(*[]*secretstore.SecretEntity)
		entities = *entitySlice
	})

	secretsAPI := &mocks.SecretInterface{}

	secretsService := K8sSecretsService{
		EntityStore: entityStore,
		SecretsAPI:  secretsAPI,
		OrgID:       organizationId,
	}

	secret, err := secretsService.GetSecret("psql creds", entitystore.Options{})

	secretsAPI.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
	assert.Nil(t, secret, "Was expecting a nil secret")
	assert.Nil(t, err, "Was not expecting an error")
}

func TestGetSecretsEnityStoreKubernetesDiscrepancy(t *testing.T) {

}

func TestAddSecretSuccess(t *testing.T) {
	organizationId := "vmware"
	principalSecretName := "psql cred"
	principal := models.Secret{
		Name: &principalSecretName,
		Secrets: models.SecretValue{
			"username": "white-rabbit",
			"password": "iml8_iml8",
		},
	}

	entityStore := &mocks.EntityStore{}

	secretUUID := "000-000-001"
	entityStore.On("Add", mock.Anything).Return(secretUUID, nil)

	k8sSecret := builder.NewK8sSecretBuilder(principal).Build()

	secretsAPI := &mocks.SecretInterface{}
	secretsAPI.On("Create", mock.Anything).Return(&k8sSecret, nil)

	secretsService := K8sSecretsService{
		EntityStore: entityStore,
		SecretsAPI:  secretsAPI,
		OrgID:       organizationId,
	}

	actualSecret, _ := secretsService.AddSecret(principal)
	assert.Equal(t, principal.Name, actualSecret.Name, "Secret created successfully")
}

func TestAddSecretDuplicateSecret(t *testing.T) {
	organizationId := "vmware"
	principalSecretName := "psql cred"
	principal := models.Secret{
		Name: &principalSecretName,
		Secrets: models.SecretValue{
			"username": "white-rabbit",
			"password": "password",
		},
	}

	entityStore := &mocks.EntityStore{}

	entityStore.On("Add", mock.Anything).Return("", errors.New("Duplicate Name"))

	secretsService := K8sSecretsService{
		EntityStore: entityStore,
		SecretsAPI:  nil,
		OrgID:       organizationId,
	}

	createdSecret, err := secretsService.AddSecret(principal)

	assert.Nil(t, createdSecret, "Got a created secret, expected nil")
	assert.NotEmpty(t, err, "Did not recieve an error when duplicating a name in the entity store.")
}

func TestDeleteSecretSuccess(t *testing.T) {
	organizationId := "vmware"
	secretName := "psql creds"
	entityStore := &mocks.EntityStore{}

	entity := secretstore.SecretEntity{}

	entityStore.On("Get", organizationId, secretName, mock.Anything, &entity).Return(nil).Run(func(args mock.Arguments) {
		entityArg := args.Get(3).(*secretstore.SecretEntity)
		*entityArg = secretstore.SecretEntity{
			BaseEntity: entitystore.BaseEntity{
				ID: "000-000-001",
			},
		}
		entity = *entityArg
	})

	secretsAPI := &mocks.SecretInterface{}
	secretsAPI.On("Delete", "000-000-001", &metav1.DeleteOptions{}).Return(nil)

	entityStore.On("Delete", organizationId, secretName, mock.Anything).Return(nil)

	secretsService := K8sSecretsService{
		OrgID:       organizationId,
		EntityStore: entityStore,
		SecretsAPI:  secretsAPI,
	}

	err := secretsService.DeleteSecret(secretName, entitystore.Options{})

	assert.Nil(t, err, "There was an error deleting the secret")
}

func TestDeleteSecretNotExist(t *testing.T) {
	organizationId := "vmware"
	secretName := "nonexistent"
	entityStore := &mocks.EntityStore{}

	entityStore.On("Get", organizationId, secretName, mock.Anything, mock.Anything).Return(store.ErrKeyNotFound)

	secretsAPI := &mocks.SecretInterface{}

	secretsService := K8sSecretsService{
		OrgID:       organizationId,
		EntityStore: entityStore,
		SecretsAPI:  secretsAPI,
	}

	err := secretsService.DeleteSecret("nonexistent", entitystore.Options{})

	assert.NotNil(t, err, "Should have failed to delete a nonexistent secret")
	secretsAPI.AssertNotCalled(t, "Delete", "Kubernetes secrets Delete was called and should not have been")
	entityStore.AssertNotCalled(t, "Delete", "EntityStore Delete was called and should not have been")
}

func TestUpdateSecretSuccess(t *testing.T) {
	organizationId := "vmware"
	secretName := "psql creds"
	entityStore := &mocks.EntityStore{}

	secretEntity := secretstore.SecretEntity{
		BaseEntity: entitystore.BaseEntity{
			ID:   "000-000-001",
			Name: secretName,
		},
	}

	entityStore.On("Get", organizationId, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		entityInput := args.Get(3).(*secretstore.SecretEntity)
		*entityInput = secretEntity
	})

	secretsAPI := &mocks.SecretInterface{}

	principal := models.Secret{
		Name: &secretEntity.ID,
		Secrets: models.SecretValue{
			"username": "white-rabbit",
			"password": "im_l8_im_l8",
		},
	}

	k8sSecret := builder.NewK8sSecretBuilder(principal).Build()
	secretsAPI.On("Update", mock.Anything).Return(&k8sSecret, nil)

	secretsService := K8sSecretsService{
		OrgID:       organizationId,
		EntityStore: entityStore,
		SecretsAPI:  secretsAPI,
	}

	_, err := secretsService.UpdateSecret(models.Secret{
		Name: &secretName,
		Secrets: models.SecretValue{
			"username": "white-rabbit",
			"password": "im_l8_im_l8",
		},
	}, entitystore.Options{})

	assert.Nil(t, err, "UpdateSecret returned unexpected error")
	entityStore.AssertCalled(t, "Get", organizationId, mock.Anything, mock.Anything, mock.Anything)
	secretsAPI.AssertCalled(t, "Update", mock.Anything)
}

func TestUpdateSecretNotExist(t *testing.T) {
	organizationId := "vmware"
	secretName := "nonexistant"
	es := &mocks.EntityStore{}

	es.On("Get", organizationId, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("some error"))

	secretsAPI := &mocks.SecretInterface{}

	secretsService := K8sSecretsService{
		OrgID:       organizationId,
		EntityStore: es,
		SecretsAPI:  secretsAPI,
	}

	secret := models.Secret{
		Name: &secretName,
	}
	_, err := secretsService.UpdateSecret(secret, entitystore.Options{})

	assert.NotNil(t, err, "Should have failed to update nonexistant secret")
	secretsAPI.AssertNotCalled(t, "Update", "Kubernetes secrets Update was called and should not have been.")
}
