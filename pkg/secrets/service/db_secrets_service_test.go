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
	"github.com/stretchr/testify/require"
	dispatchv1 "github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/secrets"
	"github.com/vmware/dispatch/pkg/secrets/mocks"
)

const (
	testOrg = "vmware"
)

func TestDBGetSecretsSuccess(t *testing.T) {
	entityStore := &mocks.EntityStore{}

	var entities []*secrets.SecretEntity
	entityStore.On("List", mock.Anything, testOrg, mock.Anything, &entities).Return(nil).Run(func(args mock.Arguments) {
		entitySlice := args.Get(3).(*[]*secrets.SecretEntity)
		*entitySlice = append(*entitySlice, &secrets.SecretEntity{
			BaseEntity: entitystore.BaseEntity{
				ID:   "000-000-001",
				Name: "psql creds",
			},
			Secrets: map[string]string{
				"username": "white-rabbit",
				"password": "iml8_iml8",
			},
		})
		*entitySlice = append(*entitySlice, &secrets.SecretEntity{
			BaseEntity: entitystore.BaseEntity{
				ID:   "000-000-002",
				Name: "twitter-api-key",
			},
			Secrets: map[string]string{
				"apiKey": "df1e8004-55f5-4aa8-8e89-73bd0f21a4de",
			},
		})
		entities = *entitySlice
	})

	secretsService := DBSecretsService{
		EntityStore: entityStore,
	}

	secrets, _ := secretsService.GetSecrets(context.Background(), &dispatchv1.Meta{Org: testOrg})

	assert.Equal(t, 2, len(entities), "The number of entities should be 2")
	assert.Equal(t, len(secrets), len(entities), "The number of entities and secrets should be the same")
}

func TestDBGetSecretByNameSuccess(t *testing.T) {
	secretName := "psql creds"

	entityStore := &mocks.EntityStore{}
	entityStore.On("Find", mock.Anything, testOrg, secretName, mock.Anything, mock.Anything).Return(true, nil).Run(func(args mock.Arguments) {
		entity := args.Get(4).(*secrets.SecretEntity)
		*entity = secrets.SecretEntity{
			BaseEntity: entitystore.BaseEntity{
				OrganizationID: testOrg,

				ID:   "000-000-001",
				Name: "psql creds",
			},
			Secrets: map[string]string{
				"username": "white-rabbit",
				"password": "iml8_iml8",
			},
		}
	})

	secretsService := DBSecretsService{
		EntityStore: entityStore,
	}

	secret, err := secretsService.GetSecret(context.Background(), &dispatchv1.Meta{Org: testOrg, Name: secretName})
	require.NoError(t, err)

	assert.NotNil(t, secret, "Received nil expected a secret")
	assert.Equal(t, secretName, secret.Meta.Name, "Returned secret name does not match requested secret name")
}

func TestDBGetSecretByNameNotFound(t *testing.T) {
	entityStore := &mocks.EntityStore{}
	entityStore.On("Find", mock.Anything, testOrg, mock.Anything, mock.Anything, mock.Anything).Return(false, nil)

	secretsService := DBSecretsService{
		EntityStore: entityStore,
	}

	secret, err := secretsService.GetSecret(context.Background(), &dispatchv1.Meta{Org: testOrg, Name: "psql creds"})

	assert.Nil(t, secret, "Was expecting a nil secret")
	assert.Equal(t, err, SecretNotFound{}, "Was expecting a SecretNotFound error")
}

func TestADBddSecretSuccess(t *testing.T) {
	principalSecretName := "psql cred"
	principal := &dispatchv1.Secret{
		Meta: dispatchv1.Meta{
			Org:  testOrg,
			Name: principalSecretName,
		},
		Secrets: dispatchv1.SecretValue{
			"username": "white-rabbit",
			"password": "iml8_iml8",
		},
	}

	entityStore := &mocks.EntityStore{}

	secretUUID := "000-000-001"
	entityStore.On("Add", mock.Anything, mock.Anything).Return(secretUUID, nil)

	secretsService := DBSecretsService{
		EntityStore: entityStore,
	}

	actualSecret, err := secretsService.AddSecret(context.Background(), principal)
	require.NoError(t, err)
	assert.Equal(t, principal.Name, actualSecret.Name, "Secret created successfully")
}

func TestDBAddSecretDuplicateSecret(t *testing.T) {
	principalSecretName := "psql cred"
	principal := &dispatchv1.Secret{
		Meta: dispatchv1.Meta{
			Org:  testOrg,
			Name: principalSecretName,
		},
		Secrets: dispatchv1.SecretValue{
			"username": "white-rabbit",
			"password": "password",
		},
	}

	entityStore := &mocks.EntityStore{}

	entityStore.On("Add", mock.Anything, mock.Anything).Return("", errors.New("Duplicate Name"))

	secretsService := DBSecretsService{
		EntityStore: entityStore,
	}

	createdSecret, err := secretsService.AddSecret(context.Background(), principal)

	assert.Nil(t, createdSecret, "Got a created secret, expected nil")
	assert.NotEmpty(t, err, "Did not recieve an error when duplicating a name in the entity store.")
}

func TestDBDeleteSecretSuccess(t *testing.T) {
	secretName := "psql creds"
	entityStore := &mocks.EntityStore{}

	entity := secrets.SecretEntity{}

	entityStore.On("Find", mock.Anything, testOrg, secretName, mock.Anything, &entity).Return(true, nil).Run(func(args mock.Arguments) {
		entityArg := args.Get(4).(*secrets.SecretEntity)
		*entityArg = secrets.SecretEntity{
			BaseEntity: entitystore.BaseEntity{
				ID: "000-000-001",
			},
		}
		entity = *entityArg
	})

	entityStore.On("Delete", mock.Anything, testOrg, secretName, mock.Anything).Return(nil)

	secretsService := DBSecretsService{
		EntityStore: entityStore,
	}

	err := secretsService.DeleteSecret(context.Background(), &dispatchv1.Meta{Org: testOrg, Name: secretName})
	require.NoError(t, err, "There was an error deleting the secret")
}

func TestDBDeleteSecretNotExist(t *testing.T) {
	secretName := "nonexistent"
	entityStore := &mocks.EntityStore{}

	entityStore.On("Find", mock.Anything, testOrg, secretName, mock.Anything, mock.Anything).Return(false, nil)

	secretsService := DBSecretsService{
		EntityStore: entityStore,
	}

	err := secretsService.DeleteSecret(context.Background(), &dispatchv1.Meta{Org: testOrg, Name: "nonexistent"})

	assert.Equal(t, SecretNotFound{}, err, "Should have returned SecretNotFound error")
	entityStore.AssertNotCalled(t, "Delete", "EntityStore Delete was called and should not have been")
}

func TestDBUpdateSecretSuccess(t *testing.T) {
	secretName := "psql creds"
	entityStore := &mocks.EntityStore{}

	secretEntity := secrets.SecretEntity{
		BaseEntity: entitystore.BaseEntity{
			ID:   "000-000-001",
			Name: secretName,
		},
	}

	entityStore.On("Find", mock.Anything, testOrg, mock.Anything, mock.Anything, mock.Anything).Return(true, nil).Run(func(args mock.Arguments) {
		entityInput := args.Get(4).(*secrets.SecretEntity)
		*entityInput = secretEntity
	})

	entityStore.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(int64(0), nil)

	secretsService := DBSecretsService{
		EntityStore: entityStore,
	}

	_, err := secretsService.UpdateSecret(context.Background(), &dispatchv1.Secret{
		Meta: dispatchv1.Meta{Org: testOrg, Name: secretName},
		Secrets: dispatchv1.SecretValue{
			"username": "white-rabbit",
			"password": "im_l8_im_l8",
		},
	})

	assert.Nil(t, err, "UpdateSecret returned unexpected error")
	entityStore.AssertCalled(t, "Find", mock.Anything, testOrg, mock.Anything, mock.Anything, mock.Anything)
	entityStore.AssertCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
}

func TestDBUpdateSecretNotExist(t *testing.T) {
	secretName := "nonexistant"
	es := &mocks.EntityStore{}

	es.On("Find", mock.Anything, testOrg, mock.Anything, mock.Anything, mock.Anything).Return(false, nil)

	secretsService := DBSecretsService{
		EntityStore: es,
	}

	_, err := secretsService.UpdateSecret(context.Background(), &dispatchv1.Secret{
		Meta: dispatchv1.Meta{Org: testOrg, Name: secretName},
	})

	assert.Equal(t, SecretNotFound{}, err, "Should have returned SecretNotFound error")
}
