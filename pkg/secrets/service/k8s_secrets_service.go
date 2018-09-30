///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package service

import (
	"context"

	"github.com/pkg/errors"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sv1 "k8s.io/client-go/kubernetes/typed/core/v1"

	dispatchv1 "github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/secrets"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils/knaming"
)

// K8sSecretsService type
type K8sSecretsService struct {
	K8sAPI k8sv1.CoreV1Interface
}

func (secretsService *K8sSecretsService) secretModelToEntity(m *dispatchv1.Secret) *secrets.SecretEntity {
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	tags["label"] = "secret"
	e := secrets.SecretEntity{
		BaseEntity: entitystore.BaseEntity{
			Name: *m.Name,
			Tags: tags,
		},
	}
	return &e
}

// GetSecret gets a specific secret
func (secretsService *K8sSecretsService) GetSecret(ctx context.Context, meta *dispatchv1.Meta) (*dispatchv1.Secret, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	secretName := knaming.SecretName(*meta)
	k8sSecret, err := secretsService.K8sAPI.Secrets(meta.Org).Get(secretName, metav1.GetOptions{})
	if err != nil {
		if errors2.IsNotFound(err) {
			return nil, SecretNotFound{}
		}
		return nil, errors.Wrapf(err, "getting a secret from k8s API: '%s'", secretName)
	}

	return ToSecret(k8sSecret), nil
}

// GetSecrets gets all the secrets
func (secretsService *K8sSecretsService) GetSecrets(ctx context.Context, meta *dispatchv1.Meta) ([]*dispatchv1.Secret, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	k8sSecretList, err := secretsService.K8sAPI.Secrets(meta.Org).List(metav1.ListOptions{
		LabelSelector: knaming.ToLabelSelector(map[string]string{
			knaming.ProjectLabel: meta.Project,
		}),
	})
	if err != nil {
		return nil, errors.Wrap(err, "listing secrets from k8s API")
	}

	var secrets []*dispatchv1.Secret
	for i := range k8sSecretList.Items {
		secrets = append(secrets, ToSecret(&k8sSecretList.Items[i]))
	}

	return secrets, nil
}

// AddSecret adds a secret
func (secretsService *K8sSecretsService) AddSecret(ctx context.Context, secret *dispatchv1.Secret) (*dispatchv1.Secret, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	createdSecret, err := secretsService.K8sAPI.Secrets(secret.Meta.Org).Create(FromSecret(secret))
	if err != nil {
		return nil, errors.Wrapf(err, "creating a k8s secret '%s'", secret.Meta.Name)
	}

	return ToSecret(createdSecret), nil
}

// DeleteSecret deletes a secret
func (secretsService *K8sSecretsService) DeleteSecret(ctx context.Context, meta *dispatchv1.Meta) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	secretName := knaming.SecretName(*meta)
	err := secretsService.K8sAPI.Secrets(meta.Org).Delete(secretName, &metav1.DeleteOptions{})
	if errors2.IsNotFound(err) {
		return SecretNotFound{}
	}

	return errors.Wrapf(err, "deleting secret from k8s API: '%s'", secretName)
}

// UpdateSecret updates a secret
func (secretsService *K8sSecretsService) UpdateSecret(ctx context.Context, secret *dispatchv1.Secret) (*dispatchv1.Secret, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	updatedSecret, err := secretsService.K8sAPI.Secrets(secret.Meta.Org).Update(FromSecret(secret))
	if err != nil {
		if errors2.IsNotFound(err) {
			return nil, SecretNotFound{}
		}
		return nil, errors.Wrapf(err, "creating a k8s secret '%s'", secret.Meta.Name)
	}

	return ToSecret(updatedSecret), nil
}
