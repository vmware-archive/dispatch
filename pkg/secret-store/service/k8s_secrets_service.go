///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package service

import (
	"github.com/pkg/errors"

	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	secretstore "github.com/vmware/dispatch/pkg/secret-store"
	"github.com/vmware/dispatch/pkg/secret-store/builder"
	"github.com/vmware/dispatch/pkg/secret-store/gen/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/core/v1"
)

// K8sSecretsService type
type K8sSecretsService struct {
	EntityStore entitystore.EntityStore
	SecretsAPI  v1.SecretInterface
	OrgID       string
}

func (secretsService *K8sSecretsService) secretModelToEntity(m *models.Secret) *secretstore.SecretEntity {
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	tags["label"] = "secret"
	e := secretstore.SecretEntity{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: secretsService.OrgID,
			Name:           *m.Name,
			Tags:           tags,
		},
	}
	return &e
}

// GetSecret gets a specific secret
func (secretsService *K8sSecretsService) GetSecret(name string, opts entitystore.Options) (*models.Secret, error) {

	if opts.Filter == nil {
		opts.Filter = entitystore.FilterEverything()
	}
	opts.Filter.Add(entitystore.FilterStat{
		Scope:   entitystore.FilterScopeField,
		Subject: "Name",
		Verb:    entitystore.FilterVerbEqual,
		Object:  name,
	})

	secrets, err := secretsService.getSecrets(opts)
	if len(secrets) < 1 {
		return nil, err
	}

	return secrets[0], nil
}

// GetSecrets gets all the secrets
func (secretsService *K8sSecretsService) GetSecrets(opts entitystore.Options) ([]*models.Secret, error) {
	return secretsService.getSecrets(opts)
}

func (secretsService *K8sSecretsService) getSecrets(opts entitystore.Options) ([]*models.Secret, error) {
	var entities []*secretstore.SecretEntity

	secretsService.EntityStore.List(secretsService.OrgID, opts, &entities)
	if len(entities) == 0 {
		return []*models.Secret{}, nil
	}

	secrets := []*models.Secret{}
	for _, entity := range entities {

		k8sSecret, err := secretsService.SecretsAPI.Get(entity.BaseEntity.ID, metav1.GetOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "error retrieve secret from k8s secret apis")
		}
		model := builder.NewVmwSecretBuilder(*entity, *k8sSecret).Build()
		secrets = append(secrets, &model)
	}
	return secrets, nil
}

// AddSecret adds a secret
func (secretsService *K8sSecretsService) AddSecret(secret models.Secret) (*models.Secret, error) {

	secretEntity := secretsService.secretModelToEntity(&secret)
	id, err := secretsService.EntityStore.Add(secretEntity)
	if err != nil {
		return nil, err
	}

	k8sSecret := builder.NewK8sSecretBuilder(secret).Build()
	k8sSecret.Name = id

	createdSecret, err := secretsService.SecretsAPI.Create(&k8sSecret)
	// TODO: Add goroutine to keep EntityStore and Kubernetes in sync.
	if err != nil {
		secretsService.EntityStore.Delete(secretsService.OrgID, id, secretEntity)
	}

	retSecret := builder.NewVmwSecretBuilder(*secretEntity, *createdSecret).Build()

	return &retSecret, nil
}

// DeleteSecret deletes a secret
func (secretsService *K8sSecretsService) DeleteSecret(name string, opts entitystore.Options) error {
	entity := secretstore.SecretEntity{}

	err := secretsService.EntityStore.Get(secretsService.OrgID, name, opts, &entity)
	if err != nil {
		return err
	}

	err = secretsService.SecretsAPI.Delete(entity.ID, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return secretsService.EntityStore.Delete(secretsService.OrgID, name, &entity)
}

// UpdateSecret updates a secret
func (secretsService *K8sSecretsService) UpdateSecret(secret models.Secret, opts entitystore.Options) (*models.Secret, error) {
	entity := secretstore.SecretEntity{}
	name := *secret.Name

	// TODO: filter
	err := secretsService.EntityStore.Get(secretsService.OrgID, name, opts, &entity)
	// assumes any entity store error means entity not found. updates to entity store will fix this.
	if err != nil {
		return nil, SecretNotFound{}
	}

	secret.Name = &entity.ID
	k8sSecret := builder.NewK8sSecretBuilder(secret).Build()

	updatedSecret, err := secretsService.SecretsAPI.Update(&k8sSecret)
	if err != nil {
		return nil, err
	}

	vmwSecretBuilder := builder.NewVmwSecretBuilder(entity, *updatedSecret)
	vmwSecret := vmwSecretBuilder.Build()

	return &vmwSecret, nil
}
