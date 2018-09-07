///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// NO TESTS

package web

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"

	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/secret-store/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/secret-store/gen/restapi/operations/secret"
	"github.com/vmware/dispatch/pkg/secret-store/service"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

// Handlers encapsulates the secret store handlers
type Handlers struct {
	secretsService service.SecretsService
}

// NewHandlers create new handlers for secret store
func NewHandlers(secretsService service.SecretsService) *Handlers {
	handlers := new(Handlers)

	handlers.secretsService = secretsService

	return handlers
}

// ConfigureHandlers registers secret store handlers to the API
func ConfigureHandlers(api middleware.RoutableAPI, h *Handlers) {
	a, ok := api.(*operations.SecretStoreAPI)
	if !ok {
		panic("Cannot configure api")
	}

	a.SecretGetSecretsHandler = secret.GetSecretsHandlerFunc(h.getSecrets)
	a.SecretAddSecretHandler = secret.AddSecretHandlerFunc(h.addSecret)
	a.SecretGetSecretHandler = secret.GetSecretHandlerFunc(h.getSecret)
	a.SecretDeleteSecretHandler = secret.DeleteSecretHandlerFunc(h.deleteSecret)
	a.SecretUpdateSecretHandler = secret.UpdateSecretHandlerFunc(h.updateSecret)
}

func (h *Handlers) addSecret(params secret.AddSecretParams) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := *params.XDispatchOrg
	project := *params.XDispatchProject

	utils.AdjustMeta(&params.Secret.Meta, v1.Meta{Org: org, Project: project})

	vmwSecret, err := h.secretsService.AddSecret(ctx, params.Secret)
	if err != nil {
		if entitystore.IsUniqueViolation(err) {
			return secret.NewAddSecretConflict().WithPayload(&v1.Error{
				Code:    http.StatusConflict,
				Message: utils.ErrorMsgAlreadyExists("secret", *params.Secret.Name),
			})
		}
		log.Errorf("error when creating the secret with k8s APIs: %+v", err)
		return secret.NewAddSecretDefault(http.StatusInternalServerError).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("secret", *params.Secret.Name),
		})
	}

	return secret.NewAddSecretCreated().WithPayload(vmwSecret)
}

func (h *Handlers) getSecrets(params secret.GetSecretsParams) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := *params.XDispatchOrg
	project := *params.XDispatchProject

	vmwSecrets, err := h.secretsService.GetSecrets(ctx, &v1.Meta{Org: org, Project: project})
	if err != nil {
		log.Errorf("error when listing secrets from k8s APIs: %+v", err)
		return secret.NewGetSecretsDefault(http.StatusInternalServerError).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when listing secrets from k8s APIs"),
		})
	}

	return secret.NewGetSecretsOK().WithPayload(vmwSecrets)
}

func (h *Handlers) getSecret(params secret.GetSecretParams) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := *params.XDispatchOrg
	project := *params.XDispatchProject

	vmwSecret, err := h.secretsService.GetSecret(ctx, &v1.Meta{Org: org, Project: project, Name: params.SecretName})
	if err != nil {
		if _, ok := err.(service.SecretNotFound); ok {
			return secret.NewGetSecretNotFound().WithPayload(&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("secret", params.SecretName),
			})
		}

		log.Errorf("error when reading the secret from k8s APIs: %+v", err)
		return secret.NewGetSecretDefault(http.StatusInternalServerError).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("secret", params.SecretName),
		})
	}

	return secret.NewGetSecretOK().WithPayload(vmwSecret)
}

func (h *Handlers) updateSecret(params secret.UpdateSecretParams) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := *params.XDispatchOrg
	project := *params.XDispatchProject

	utils.AdjustMeta(&params.Secret.Meta, v1.Meta{Org: org, Project: project})

	updatedSecret, err := h.secretsService.UpdateSecret(ctx, params.Secret)

	if err != nil {
		if _, ok := err.(service.SecretNotFound); ok {
			return secret.NewUpdateSecretNotFound().WithPayload(&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("secret", params.SecretName),
			})
		}

		log.Errorf("error when updating secret from k8s APIs: %+v", err)
		return secret.NewUpdateSecretDefault(http.StatusInternalServerError).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("secret", params.SecretName),
		})
	}

	return secret.NewUpdateSecretCreated().WithPayload(updatedSecret)
}

func (h *Handlers) deleteSecret(params secret.DeleteSecretParams) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := *params.XDispatchOrg
	project := *params.XDispatchProject

	err := h.secretsService.DeleteSecret(ctx, &v1.Meta{Org: org, Project: project, Name: params.SecretName})
	if err != nil {
		if _, ok := err.(service.SecretNotFound); ok {
			return secret.NewDeleteSecretNotFound().WithPayload(&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("secret", params.SecretName),
			})
		}

		log.Errorf("error when deleting secret from k8s APIs: %+v", err)
		return secret.NewDeleteSecretDefault(http.StatusInternalServerError).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("secret", params.SecretName),
		})
	}
	return secret.NewDeleteSecretNoContent()
}
