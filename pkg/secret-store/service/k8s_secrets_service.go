///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package service

import (
	"context"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sv1 "k8s.io/client-go/kubernetes/typed/core/v1"

	dispatchv1 "github.com/vmware/dispatch/pkg/api/v1"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	secretstore "github.com/vmware/dispatch/pkg/secret-store"
	"github.com/vmware/dispatch/pkg/secret-store/builder"
	"github.com/vmware/dispatch/pkg/trace"
)

// K8sSecretsService type
type K8sSecretsService struct {
	EntityStore entitystore.EntityStore
	SecretsAPI  k8sv1.SecretInterface
}

func (secretsService *K8sSecretsService) secretModelToEntity(m *dispatchv1.Secret) *secretstore.SecretEntity {
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	tags["label"] = "secret"
	e := secretstore.SecretEntity{
		BaseEntity: entitystore.BaseEntity{
			Name: *m.Name,
			Tags: tags,
		},
	}
	return &e
}

// GetSecret gets a specific secret
func (secretsService *K8sSecretsService) GetSecret(ctx context.Context, organizationID string, name string, opts entitystore.Options) (*dispatchv1.Secret, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	if opts.Filter == nil {
		opts.Filter = entitystore.FilterEverything()
	}
	opts.Filter.Add(entitystore.FilterStat{
		Scope:   entitystore.FilterScopeField,
		Subject: "Name",
		Verb:    entitystore.FilterVerbEqual,
		Object:  name,
	})

	secrets, err := secretsService.GetSecrets(ctx, organizationID, opts)
	if err != nil {
		return nil, err
	} else if len(secrets) < 1 {
		return nil, SecretNotFound{}
	}

	return secrets[0], nil
}

// GetSecrets gets all the secrets
func (secretsService *K8sSecretsService) GetSecrets(ctx context.Context, organizationID string, opts entitystore.Options) ([]*dispatchv1.Secret, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	var entities []*secretstore.SecretEntity

	secretsService.EntityStore.List(ctx, organizationID, opts, &entities)
	if len(entities) == 0 {
		return []*dispatchv1.Secret{}, nil
	}

	var secrets []*dispatchv1.Secret
	for _, entity := range entities {

		k8sSecret, err := secretsService.SecretsAPI.Get(entity.BaseEntity.ID, metav1.GetOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "error retrieve secret from k8s secret apis")
		}
		model := builder.NewDispatchSecretBuilder(*entity, *k8sSecret).Build()
		secrets = append(secrets, &model)
	}
	return secrets, nil
}

// AddSecret adds a secret
func (secretsService *K8sSecretsService) AddSecret(ctx context.Context, organizationID string, secret dispatchv1.Secret) (*dispatchv1.Secret, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	secretEntity := secretsService.secretModelToEntity(&secret)
	secretEntity.OrganizationID = organizationID
	log.Infof("adding secret %s/%s to secret store", organizationID, *secret.Name)
	id, err := secretsService.EntityStore.Add(ctx, secretEntity)
	if err != nil {
		return nil, err
	}

	k8sSecret := builder.NewK8sSecretBuilder(secret).Build()
	k8sSecret.Name = id

	createdSecret, err := secretsService.SecretsAPI.Create(&k8sSecret)
	// TODO: Add goroutine to keep EntityStore and Kubernetes in sync.
	if err != nil {
		secretsService.EntityStore.Delete(ctx, organizationID, id, secretEntity)
	}

	retSecret := builder.NewDispatchSecretBuilder(*secretEntity, *createdSecret).Build()

	return &retSecret, nil
}

// DeleteSecret deletes a secret
func (secretsService *K8sSecretsService) DeleteSecret(ctx context.Context, organizationID string, name string, opts entitystore.Options) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	entity := secretstore.SecretEntity{}

	ok, err := secretsService.EntityStore.Find(ctx, organizationID, name, opts, &entity)
	if err != nil {
		return err
	} else if !ok {
		return SecretNotFound{}
	}

	err = secretsService.SecretsAPI.Delete(entity.ID, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return secretsService.EntityStore.Delete(ctx, organizationID, name, &entity)
}

// UpdateSecret updates a secret
func (secretsService *K8sSecretsService) UpdateSecret(ctx context.Context, organizationID string, secret dispatchv1.Secret, opts entitystore.Options) (*dispatchv1.Secret, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	entity := secretstore.SecretEntity{}
	name := *secret.Name

	// TODO: filter
	ok, err := secretsService.EntityStore.Find(ctx, organizationID, name, opts, &entity)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, SecretNotFound{}
	}

	secret.Name = &entity.ID
	k8sSecret := builder.NewK8sSecretBuilder(secret).Build()

	updatedSecret, err := secretsService.SecretsAPI.Update(&k8sSecret)
	if err != nil {
		return nil, err
	}

	dispatchSecretBuilder := builder.NewDispatchSecretBuilder(entity, *updatedSecret)
	dispatchSecret := dispatchSecretBuilder.Build()

	return &dispatchSecret, nil
}
