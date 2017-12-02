///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package service

import (
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	secretstore "github.com/vmware/dispatch/pkg/secret-store"
	"github.com/vmware/dispatch/pkg/secret-store/builder"
	"github.com/vmware/dispatch/pkg/secret-store/gen/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/core/v1"
)

type K8sSecretsService struct {
	EntityStore entitystore.EntityStore
	SecretsAPI  v1.SecretInterface
	OrgID       string
}

func (secretsService *K8sSecretsService) GetSecret(name string) (*models.Secret, error) {
	nameFilter := func(entity entitystore.Entity) bool {
		return entity.GetName() == name
	}

	listOptions := metav1.ListOptions{}

	secrets, err := secretsService.getSecrets(nameFilter, listOptions)

	if len(secrets) < 1 {
		return nil, err
	}

	return secrets[0], nil
}

func (secretsService *K8sSecretsService) GetSecrets() ([]*models.Secret, error) {
	return secretsService.getSecrets(func(entity entitystore.Entity) bool { return true }, metav1.ListOptions{})
}

func (secretsService *K8sSecretsService) getSecrets(filter entitystore.Filter, listOptions metav1.ListOptions) ([]*models.Secret, error) {
	var entities []*secretstore.SecretEntity
	secretsService.EntityStore.List(secretsService.OrgID, filter, &entities)

	if len(entities) == 0 {
		return []*models.Secret{}, nil
	}

	secrets := []*models.Secret{}

	k8sSecretList, err := secretsService.SecretsAPI.List(listOptions)

	if err != nil {
		return nil, err
	}

	for _, k8sSecret := range k8sSecretList.Items {
		for _, secretEntity := range entities {
			if k8sSecret.Name != secretEntity.BaseEntity.ID {
				continue
			}

			builder := builder.NewVmwSecretBuilder(k8sSecret)
			secretValue := models.SecretValue{}
			for k, v := range k8sSecret.Data {
				secretValue[k] = string(v)
			}

			secret := builder.Build()
			secretName := secretEntity.Name
			secret.Name = &secretName
			secrets = append(secrets, &secret)
		}
	}

	return secrets, nil
}

func (secretsService *K8sSecretsService) AddSecret(secret models.Secret) (*models.Secret, error) {
	secretEntity := secretstore.SecretEntity{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: secretsService.OrgID,
			Name:           *secret.Name,
			Tags: entitystore.Tags{
				"label": "secret",
			},
		},
	}

	id, err := secretsService.EntityStore.Add(&secretEntity)

	if err != nil {
		return nil, err
	}

	k8sSecret := builder.NewK8sSecretBuilder(secret).Build()
	k8sSecret.Name = id

	createdSecret, err := secretsService.SecretsAPI.Create(&k8sSecret)

	// TODO: Add goroutine to keep EntityStore and Kubernetes in sync.
	if err != nil {
		secretsService.EntityStore.Delete(secretsService.OrgID, id, &secretEntity)
	}

	retSecret := builder.NewVmwSecretBuilder(*createdSecret).Build()

	return &retSecret, nil
}

func (secretsService *K8sSecretsService) DeleteSecret(name string) error {
	entity := secretstore.SecretEntity{}

	err := secretsService.EntityStore.Get(secretsService.OrgID, name, &entity)

	if err != nil {
		return err
	}

	err = secretsService.SecretsAPI.Delete(entity.ID, &metav1.DeleteOptions{})

	if err != nil {
		return err
	}

	return secretsService.EntityStore.Delete(secretsService.OrgID, name, &entity)
}

func (secretsService *K8sSecretsService) UpdateSecret(secret models.Secret) (*models.Secret, error) {
	var entities []*secretstore.SecretEntity

	filter := entitystore.Filter(func(entity entitystore.Entity) bool {
		return entity.GetName() == *secret.Name
	})

	err := secretsService.EntityStore.List(secretsService.OrgID, filter, entities)

	if len(entities) == 0 {
		return nil, nil
	}

	k8sSecret := builder.NewK8sSecretBuilder(secret).Build()

	updatedSecret, err := secretsService.SecretsAPI.Update(&k8sSecret)

	if err != nil {
		return nil, err
	}

	vmwSecretBuilder := builder.NewVmwSecretBuilder(*updatedSecret)
	vmwSecret := vmwSecretBuilder.Build()

	return &vmwSecret, nil
}
