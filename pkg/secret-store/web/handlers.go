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
	entityStore    entitystore.EntityStore
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

	a.CookieAuth = func(token string) (interface{}, error) {
		// TODO: be able to retrieve user information from the cookie
		// currently just return the cookie
		return token, nil
	}

	a.BearerAuth = func(token string) (interface{}, error) {
		// TODO: Once IAM issues signed tokens, validate them here.
		return token, nil
	}

	a.SecretGetSecretsHandler = secret.GetSecretsHandlerFunc(h.getSecrets)
	a.SecretAddSecretHandler = secret.AddSecretHandlerFunc(h.addSecret)
	a.SecretGetSecretHandler = secret.GetSecretHandlerFunc(h.getSecret)
	a.SecretDeleteSecretHandler = secret.DeleteSecretHandlerFunc(h.deleteSecret)
	a.SecretUpdateSecretHandler = secret.UpdateSecretHandlerFunc(h.updateSecret)
}

func (h *Handlers) addSecret(params secret.AddSecretParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	vmwSecret, err := h.secretsService.AddSecret(ctx, params.XDispatchOrg, *params.Secret)
	if err != nil {
		if entitystore.IsUniqueViolation(err) {
			return secret.NewAddSecretConflict().WithPayload(&v1.Error{
				Code:    http.StatusConflict,
				Message: utils.ErrorMsgAlreadyExists("secret", *params.Secret.Name),
			})
		}
		log.Errorf("error when creating the secret: %+v", err)
		return secret.NewAddSecretDefault(http.StatusInternalServerError).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("secret", *params.Secret.Name),
		})
	}

	return secret.NewAddSecretCreated().WithPayload(vmwSecret)
}

func (h *Handlers) getSecrets(params secret.GetSecretsParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	vmwSecrets, err := h.secretsService.GetSecrets(ctx, params.XDispatchOrg, entitystore.Options{})
	if err != nil {
		log.Errorf("error when listing secrets: %+v", err)
		return secret.NewGetSecretsDefault(http.StatusInternalServerError).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when listing secrets"),
		})
	}

	return secret.NewGetSecretsOK().WithPayload(vmwSecrets)
}

func (h *Handlers) getSecret(params secret.GetSecretParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	vmwSecret, err := h.secretsService.GetSecret(ctx, params.XDispatchOrg, params.SecretName, entitystore.Options{})
	if err != nil {
		if _, ok := err.(service.SecretNotFound); ok {
			return secret.NewGetSecretNotFound().WithPayload(&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("secret", params.SecretName),
			})
		}

		log.Errorf("error when reading the secret: %+v", err)
		return secret.NewGetSecretDefault(http.StatusInternalServerError).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("secret", params.SecretName),
		})
	}

	return secret.NewGetSecretOK().WithPayload(vmwSecret)
}

func (h *Handlers) updateSecret(params secret.UpdateSecretParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	updatedSecret, err := h.secretsService.UpdateSecret(ctx, params.XDispatchOrg, *params.Secret, entitystore.Options{})

	if err != nil {
		if _, ok := err.(service.SecretNotFound); ok {
			return secret.NewUpdateSecretNotFound().WithPayload(&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("secret", params.SecretName),
			})
		}

		log.Errorf("error when updating secret: %+v", err)
		return secret.NewUpdateSecretDefault(http.StatusInternalServerError).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("secret", params.SecretName),
		})
	}

	return secret.NewUpdateSecretCreated().WithPayload(updatedSecret)
}

func (h *Handlers) deleteSecret(params secret.DeleteSecretParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	err := h.secretsService.DeleteSecret(ctx, params.XDispatchOrg, params.SecretName, entitystore.Options{})
	if err != nil {
		if _, ok := err.(service.SecretNotFound); ok {
			return secret.NewDeleteSecretNotFound().WithPayload(&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("secret", params.SecretName),
			})
		}

		log.Errorf("error when deleting secret: %+v", err)
		return secret.NewDeleteSecretDefault(http.StatusInternalServerError).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("secret", params.SecretName),
		})
	}
	return secret.NewDeleteSecretNoContent()
}
