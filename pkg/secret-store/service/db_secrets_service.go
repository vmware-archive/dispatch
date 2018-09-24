///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package service

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/pkg/errors"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/secret-store"
	"github.com/vmware/dispatch/pkg/trace"
)

// DBSecretsService implements service which stores all secrets data in entity store.
type DBSecretsService struct {
	EntityStore entitystore.EntityStore
}

// GetSecret gets a specific secret
func (s *DBSecretsService) GetSecret(ctx context.Context, meta *v1.Meta) (*v1.Secret, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	opts := entitystore.Options{Filter: entitystore.FilterEverything()}

	var secretEntity secretstore.SecretEntity
	found, err := s.EntityStore.Find(ctx, meta.Org, meta.Name, opts, &secretEntity)
	if err != nil {
		return nil, errors.Wrapf(err, "getting a secret from EntityStore: '%s'", meta.Name)
	}
	if !found {
		return nil, SecretNotFound{}
	}

	return s.secretEntityToModel(&secretEntity), nil
}

// GetSecrets gets all the secrets
func (s *DBSecretsService) GetSecrets(ctx context.Context, meta *v1.Meta) ([]*v1.Secret, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	var entities []*secretstore.SecretEntity

	opts := entitystore.Options{Filter: entitystore.FilterEverything()}

	s.EntityStore.List(ctx, meta.Org, opts, &entities)
	if len(entities) == 0 {
		return []*v1.Secret{}, nil
	}

	var secrets []*v1.Secret
	for i := range entities {
		secrets = append(secrets, s.secretEntityToModel(entities[i]))
	}
	return secrets, nil
}

// AddSecret adds a secret
func (s *DBSecretsService) AddSecret(ctx context.Context, secret *v1.Secret) (*v1.Secret, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	// TODO: encrypt secret's content
	e := s.secretModelToEntity(secret)
	_, err := s.EntityStore.Add(ctx, e)
	if err != nil {
		return nil, err
	}
	return s.secretEntityToModel(e), nil
}

// DeleteSecret deletes a secret
func (s *DBSecretsService) DeleteSecret(ctx context.Context, meta *v1.Meta) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	entity := secretstore.SecretEntity{}

	opts := entitystore.Options{Filter: entitystore.FilterEverything()}

	ok, err := s.EntityStore.Find(ctx, meta.Org, meta.Name, opts, &entity)
	if err != nil {
		return errors.Wrapf(err, "finding a secret in EntityStore: '%s'", meta.Name)
	} else if !ok {
		return SecretNotFound{}
	}

	return s.EntityStore.Delete(ctx, meta.Org, meta.Name, &entity)
}

// UpdateSecret updates a secret
func (s *DBSecretsService) UpdateSecret(ctx context.Context, secret *v1.Secret) (*v1.Secret, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	entity := secretstore.SecretEntity{}

	opts := entitystore.Options{Filter: entitystore.FilterEverything()}

	ok, err := s.EntityStore.Find(ctx, secret.Meta.Org, secret.Meta.Name, opts, &entity)
	if err != nil {
		return nil, errors.Wrapf(err, "finding a secret in EntityStore: '%s'", secret.Meta.Name)
	} else if !ok {
		return nil, SecretNotFound{}
	}

	// TODO: encrypt secret's content
	entity.Secrets = secret.Secrets
	_, err = s.EntityStore.Update(ctx, entity.Revision, &entity)
	if err != nil {
		return nil, errors.Wrapf(err, "updating a secret in EntityStore: '%s'", secret.Meta.Name)
	}

	return s.secretEntityToModel(&entity), nil
}

func (s *DBSecretsService) secretModelToEntity(m *v1.Secret) *secretstore.SecretEntity {
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	return &secretstore.SecretEntity{
		BaseEntity: entitystore.BaseEntity{
			Name:           m.Meta.Name,
			OrganizationID: m.Meta.Org,
			Tags:           tags,
		},
		Secrets: m.Secrets,
	}
}

// Build converts a DispatchSecretBuilder to a swagger model Secret
func (s *DBSecretsService) secretEntityToModel(e *secretstore.SecretEntity) *v1.Secret {
	var tags []*v1.Tag
	for k, v := range e.Tags {
		tags = append(tags, &v1.Tag{Key: k, Value: v})
	}
	return &v1.Secret{
		Meta: v1.Meta{
			Org:  e.OrganizationID,
			Name: e.Name,
		},
		ID:      strfmt.UUID(e.ID),
		Kind:    v1.SecretKind,
		Secrets: e.Secrets,
		Tags:    tags,
	}
}
