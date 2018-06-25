///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package service

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/vmware/dispatch/pkg/utils"

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
func (s *DBSecretsService) GetSecret(ctx context.Context, organizationID string, name string, opts entitystore.Options) (*v1.Secret, error) {
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

	secrets, err := s.GetSecrets(ctx, organizationID, opts)
	if err != nil {
		return nil, err
	} else if len(secrets) < 1 {
		return nil, SecretNotFound{}
	}

	return secrets[0], nil
}

// GetSecrets gets all the secrets
func (s *DBSecretsService) GetSecrets(ctx context.Context, organizationID string, opts entitystore.Options) ([]*v1.Secret, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	var entities []*secretstore.SecretEntity

	s.EntityStore.List(ctx, organizationID, opts, &entities)
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
func (s *DBSecretsService) AddSecret(ctx context.Context, organizationID string, secret v1.Secret) (*v1.Secret, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	// TODO: encrypt secret's content
	e := s.secretModelToEntity(&secret)
	e.OrganizationID = organizationID
	_, err := s.EntityStore.Add(ctx, e)
	if err != nil {
		return nil, err
	}
	return s.secretEntityToModel(e), nil
}

// DeleteSecret deletes a secret
func (s *DBSecretsService) DeleteSecret(ctx context.Context, organizationID string, name string, opts entitystore.Options) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	entity := secretstore.SecretEntity{}

	ok, err := s.EntityStore.Find(ctx, organizationID, name, opts, &entity)
	if err != nil {
		return err
	} else if !ok {
		return SecretNotFound{}
	}

	return s.EntityStore.Delete(ctx, organizationID, name, &entity)
}

// UpdateSecret updates a secret
func (s *DBSecretsService) UpdateSecret(ctx context.Context, organizationID string, secret v1.Secret, opts entitystore.Options) (*v1.Secret, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	entity := secretstore.SecretEntity{}
	name := *secret.Name

	// TODO: filter
	ok, err := s.EntityStore.Find(ctx, organizationID, name, opts, &entity)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, SecretNotFound{}
	}

	// TODO: encrypt secret's content
	entity.Secrets = secret.Secrets
	_, err = s.EntityStore.Update(ctx, entity.Revision, &entity)
	if err != nil {
		return nil, err
	}

	return s.secretEntityToModel(&entity), nil
}

func (s *DBSecretsService) secretModelToEntity(m *v1.Secret) *secretstore.SecretEntity {
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	e := secretstore.SecretEntity{
		BaseEntity: entitystore.BaseEntity{
			Name: *m.Name,
			Tags: tags,
		},
		Secrets: m.Secrets,
	}
	return &e
}

// Build converts a DispatchSecretBuilder to a swagger model Secret
func (s *DBSecretsService) secretEntityToModel(e *secretstore.SecretEntity) *v1.Secret {
	var tags []*v1.Tag
	for k, v := range e.Tags {
		tags = append(tags, &v1.Tag{Key: k, Value: v})
	}
	return &v1.Secret{
		ID:      strfmt.UUID(e.ID),
		Name:    &e.Name,
		Kind:    utils.SecretKind,
		Secrets: e.Secrets,
		Tags:    tags,
	}
}
