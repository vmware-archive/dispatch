///////////////////////////////////////////////////////////////////////
// Copyright (C) 2017 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package secretstore

import (
	"log"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/secret-store/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/secret-store/gen/restapi/operations"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/secret-store/gen/restapi/operations/secret"
)

const NameFieldSelector = "metadata.name="

var SecretStoreFlags = struct {
	K8sConfig    string `long:"kubeconfig" description:"Path to kubernetes config file"`
	K8sNamespace string `long:"namespace" description:"Kubernetes namespace" default:"default"`
}{}

type Handlers struct {
	secretsAPI   v1.SecretInterface
	k8snamespace string
}

func NewHandlers() (*Handlers, error) {
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

	handlers.secretsAPI = clientset.Secrets(SecretStoreFlags.K8sNamespace)

	return handlers, nil
}

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

	a.SecretGetSecretsHandler = secret.GetSecretsHandlerFunc(h.readSecret)
	a.SecretAddSecretHandler = secret.AddSecretHandlerFunc(h.createSecret)
	a.SecretGetSecretHandler = secret.GetSecretHandlerFunc(h.readSecretByName)
	a.SecretDeleteSecretHandler = secret.DeleteSecretHandlerFunc(h.deleteSecret)
	a.SecretUpdateSecretHandler = secret.UpdateSecretHandlerFunc(h.updateSecret)
}

func (h *Handlers) createSecret(params secret.AddSecretParams, principal interface{}) middleware.Responder {
	// TODO: implement creation logic.
	return secret.NewAddSecretDefault(500).WithPayload(
		&models.Error{Message: swag.String("addSecret not implemented")})
}

func (h *Handlers) readSecret(params secret.GetSecretsParams, principal interface{}) middleware.Responder {
	listOptions := metav1.ListOptions{}

	vmwSecrets, err := h.readSecrets(listOptions)

	if err != nil {
		return secret.NewGetSecretsDefault(500)
	}

	return secret.NewGetSecretsOK().WithPayload(models.GetSecretsOKBody(vmwSecrets))
}

func (h *Handlers) readSecretByName(params secret.GetSecretParams, principal interface{}) middleware.Responder {
	listOptions := metav1.ListOptions{
		FieldSelector: NameFieldSelector + params.SecretName,
	}

	vmwSecrets, err := h.readSecrets(listOptions)
	if err != nil {
		return secret.NewGetSecretNotFound()
	}

	if len(vmwSecrets) == 0 {
		return secret.NewGetSecretDefault(500)
	}

	return secret.NewGetSecretOK().WithPayload(vmwSecrets[0])
}

func (h *Handlers) readSecrets(listOptions metav1.ListOptions) ([]*models.Secret, error) {
	secrets, err := h.secretsAPI.List(listOptions)

	if err != nil {
		return nil, err
	}

	var vmwSecrets []*models.Secret
	for _, v := range secrets.Items {
		secretValue := models.SecretValue{}
		for k, mv := range v.Data {
			secretValue[k] = string(mv)
		}
		secretName := v.Name
		vmwSecret := models.Secret{
			Name:    &secretName,
			Secrets: secretValue,
		}
		vmwSecrets = append(vmwSecrets, &vmwSecret)
	}

	return vmwSecrets, nil
}

func (h *Handlers) deleteSecret(params secret.DeleteSecretParams, principal interface{}) middleware.Responder {
	return secret.NewDeleteSecretDefault(500).WithPayload(
		&models.Error{Message: swag.String("deleteSecret has not yet been implemented")})

}

func (h *Handlers) updateSecret(params secret.UpdateSecretParams, principal interface{}) middleware.Responder {
	return secret.NewUpdateSecretDefault(500).WithPayload(
		&models.Error{Message: swag.String("updateSecret has not yet been implemented")})
}
