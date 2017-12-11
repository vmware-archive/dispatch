///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////// NO TESTS
package web

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	strfmt "github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/secret-store/gen/models"
	"github.com/vmware/dispatch/pkg/secret-store/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/secret-store/gen/restapi/operations/secret"
	"github.com/vmware/dispatch/pkg/secret-store/service"
	"github.com/vmware/dispatch/pkg/trace"
)

// SecretStoreFlags are configuration flags for the secret store
var SecretStoreFlags = struct {
	K8sConfig      string `long:"kubeconfig" description:"Path to kubernetes config file"`
	K8sNamespace   string `long:"namespace" description:"Kubernetes namespace" default:"default"`
	OrganizationID string `long:"organization" description:"Organization ID" default:"vmware"`
	DbFile         string `long:"database file" description:"File to use to write to database" default:"./db.bolt"`
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
		SecretsAPI:  clientset.Secrets(SecretStoreFlags.K8sNamespace),
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

	vmwSecrets, err := h.secretsService.GetSecrets()
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

	vmwSecret, err := h.secretsService.GetSecret(params.SecretName)
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
	updatedSecret, err := h.secretsService.UpdateSecret(*params.Secret)

	if err != nil {
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
	err := h.secretsService.DeleteSecret(params.SecretName)
	if err != nil {
		log.Errorf("error when deleting secret from k8s APIs: %+v", err)
		return secret.NewDeleteSecretDefault(http.StatusInternalServerError).WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when deleting the secret"),
		})
	}
	return secret.NewDeleteSecretNoContent()
}

func transformVmwToK8s(secret models.Secret) *apiv1.Secret {
	data := make(map[string][]byte)
	for k, v := range secret.Secrets {
		data[k] = []byte(v)
	}
	return &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: *secret.Name,
		},
		Data: data,
	}
}

func transformK8sToVmw(secret apiv1.Secret) *models.Secret {
	secretValue := models.SecretValue{}
	for k, v := range secret.Data {
		secretValue[k] = string(v)
	}
	return &models.Secret{
		ID:      strfmt.UUID(secret.UID),
		Name:    &secret.Name,
		Secrets: secretValue,
	}
}
