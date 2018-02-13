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

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/secret-store/gen/models"
	"github.com/vmware/dispatch/pkg/secret-store/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/secret-store/gen/restapi/operations/secret"
	"github.com/vmware/dispatch/pkg/secret-store/service"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

// SecretStoreFlags are configuration flags for the secret store
var SecretStoreFlags = struct {
	K8sConfig      string `long:"kubeconfig" description:"Path to kubernetes config file"`
	K8sNamespace   string `long:"namespace" description:"Kubernetes namespace" default:"default"`
	OrganizationID string `long:"organization" description:"Organization ID" default:"vmware"`
	DbFile         string `long:"db-file" description:"Backend DB URL/Path" default:"./db.bolt"`
	DbBackend      string `long:"db-backend" description:"Backend DB Name" default:"boltdb"`
	DbUser         string `long:"db-username" description:"Backend DB Username" default:"dispatch"`
	DbPassword     string `long:"db-password" description:"Backend DB Password" default:"dispatch"`
	DbDatabase     string `long:"db-database" description:"Backend DB Name" default:"dispatch"`
}{}

// Handlers encapsulates the secret store handlers
type Handlers struct {
	secretsService service.SecretsService
	entityStore    entitystore.EntityStore
	k8snamespace   string
}

// NewHandlers create new handlers for secret store
func NewHandlers(entityStore entitystore.EntityStore) (*Handlers, error) {
	handlers := new(Handlers)
	var err error
	var config *rest.Config
	if SecretStoreFlags.K8sConfig == "" {
		// creates the in-cluster config
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", SecretStoreFlags.K8sConfig)
	}
	if err != nil {
		return nil, errors.Wrap(err, "Error getting kubernetes config")
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating kubernetes client")
	}

	handlers.secretsService = &service.K8sSecretsService{
		EntityStore: entityStore,
		SecretsAPI:  clientset.CoreV1().Secrets(SecretStoreFlags.K8sNamespace),
		OrgID:       SecretStoreFlags.OrganizationID,
	}

	return handlers, nil
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
		log.Printf("cookie auth: %s\n", token)
		return token, nil
	}

	a.SecretGetSecretsHandler = secret.GetSecretsHandlerFunc(h.getSecrets)
	a.SecretAddSecretHandler = secret.AddSecretHandlerFunc(h.addSecret)
	a.SecretGetSecretHandler = secret.GetSecretHandlerFunc(h.getSecret)
	a.SecretDeleteSecretHandler = secret.DeleteSecretHandlerFunc(h.deleteSecret)
	a.SecretUpdateSecretHandler = secret.UpdateSecretHandlerFunc(h.updateSecret)
}

func (h *Handlers) addSecret(params secret.AddSecretParams, principal interface{}) middleware.Responder {
	trace.Trace("secret.AddSecret." + *params.Secret.Name)()

	vmwSecret, err := h.secretsService.AddSecret(*params.Secret)
	if err != nil {
		log.Errorf("error when creating the secret with k8s APIs: %+v", err)
		return secret.NewAddSecretDefault(http.StatusInternalServerError).WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when creating the secret"),
		})
	}

	return secret.NewAddSecretCreated().WithPayload(vmwSecret)
}

func (h *Handlers) getSecrets(params secret.GetSecretsParams, principal interface{}) middleware.Responder {
	trace.Trace("secret.GetSecrets")()

	filter, err := utils.ParseTags(nil, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return secret.NewGetSecretsBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}

	vmwSecrets, err := h.secretsService.GetSecrets(entitystore.Options{
		Filter: filter,
	})
	if err != nil {
		log.Errorf("error when listing secrets from k8s APIs: %+v", err)
		return secret.NewGetSecretsDefault(http.StatusInternalServerError).WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when listing secrets from k8s APIs"),
		})
	}

	return secret.NewGetSecretsOK().WithPayload(models.GetSecretsOKBody(vmwSecrets))
}

func (h *Handlers) getSecret(params secret.GetSecretParams, principal interface{}) middleware.Responder {
	trace.Trace("secret.GetSecret." + params.SecretName)()

	filter, err := utils.ParseTags(nil, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return secret.NewGetSecretBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	vmwSecret, err := h.secretsService.GetSecret(params.SecretName, entitystore.Options{Filter: filter})
	if err != nil {
		log.Errorf("error when reading the secret from k8s APIs: %+v", err)
		return secret.NewGetSecretDefault(http.StatusInternalServerError).WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when reading the secret"),
		})
	}

	// TODO: create a specific error that can be caught here to send correct error message
	//	if len(vmwSecrets) == 0 {
	//		return secret.NewGetSecretNotFound().WithPayload(&models.Error{
	//			Code:    http.StatusNotFound,
	//			Message: swag.String("secret not found"),
	//		})
	//	}

	return secret.NewGetSecretOK().WithPayload(vmwSecret)
}

func (h *Handlers) updateSecret(params secret.UpdateSecretParams, principal interface{}) middleware.Responder {
	trace.Trace("secret.UpdateSecret." + params.SecretName)()

	filter, err := utils.ParseTags(nil, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return secret.NewUpdateSecretBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	updatedSecret, err := h.secretsService.UpdateSecret(*params.Secret, entitystore.Options{
		Filter: filter,
	})

	if err != nil {
		if _, ok := err.(service.SecretNotFound); ok {
			return secret.NewUpdateSecretNotFound()
		}

		log.Errorf("error when updating secret from k8s APIs: %+v", err)
		return secret.NewUpdateSecretDefault(http.StatusInternalServerError).WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when updating secret"),
		})
	}

	return secret.NewUpdateSecretCreated().WithPayload(updatedSecret)
}

func (h *Handlers) deleteSecret(params secret.DeleteSecretParams, principal interface{}) middleware.Responder {
	trace.Trace("secret.DeleteSecret." + params.SecretName)()

	filter, err := utils.ParseTags(nil, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return secret.NewDeleteSecretBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	err = h.secretsService.DeleteSecret(params.SecretName, entitystore.Options{
		Filter: filter,
	})
	if err != nil {
		log.Errorf("error when deleting secret from k8s APIs: %+v", err)
		return secret.NewDeleteSecretDefault(http.StatusInternalServerError).WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when deleting the secret"),
		})
	}
	return secret.NewDeleteSecretNoContent()
}
