///////////////////////////////////////////////////////////////////////
// Copyright (C) 2017 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package secretstore

import (
	"fmt"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	strfmt "github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/secret-store/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/secret-store/gen/restapi/operations"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/secret-store/gen/restapi/operations/secret"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/trace"
)

// SecretStoreFlags are configuration flags for the secret store
var SecretStoreFlags = struct {
	K8sConfig    string `long:"kubeconfig" description:"Path to kubernetes config file"`
	K8sNamespace string `long:"namespace" description:"Kubernetes namespace" default:"default"`
}{}

// Handlers encapsulates the secret store handlers
type Handlers struct {
	secretsAPI   v1.SecretInterface
	k8snamespace string
}

// NewHandlers create new handlers for secret store
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
	inputSecret := transformVmwToK8s(*params.Secret)

	k8sSecret, err := h.secretsAPI.Create(inputSecret)
	if err != nil {
		log.Errorf("error when creating the secret with k8s APIs: %+v", err)
		return secret.NewAddSecretDefault(http.StatusInternalServerError).WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when creating the secret"),
		})
	}

	vmwSecret := transformK8sToVmw(*k8sSecret)

	return secret.NewAddSecretCreated().WithPayload(vmwSecret)
}

func (h *Handlers) getSecrets(params secret.GetSecretsParams, principal interface{}) middleware.Responder {
	trace.Trace("secret.GetSecrets")()
	listOptions := metav1.ListOptions{}
	vmwSecrets, err := h.readSecrets(listOptions)
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
	listOptions := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", params.SecretName),
	}
	vmwSecrets, err := h.readSecrets(listOptions)
	if err != nil {
		log.Errorf("error when reading the secret from k8s APIs: %+v", err)
		return secret.NewGetSecretDefault(http.StatusInternalServerError).WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when reading the secret"),
		})
	}
	if len(vmwSecrets) == 0 {
		return secret.NewGetSecretNotFound().WithPayload(&models.Error{
			Code:    http.StatusNotFound,
			Message: swag.String("secret not found"),
		})
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
			ID:      strfmt.UUID(v.UID),
			Name:    &secretName,
			Secrets: secretValue,
		}
		vmwSecrets = append(vmwSecrets, &vmwSecret)
	}
	return vmwSecrets, nil
}

func (h *Handlers) updateSecret(params secret.UpdateSecretParams, principal interface{}) middleware.Responder {
	trace.Trace("secret.UpdateSecret." + params.SecretName)()
	k8sSecret := transformVmwToK8s(*params.Secret)
	updatedSecret, err := h.secretsAPI.Update(k8sSecret)
	if err != nil {
		log.Errorf("error when updating secret from k8s APIs: %+v", err)
		return secret.NewUpdateSecretDefault(http.StatusInternalServerError).WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when updating secret"),
		})
	}
	return secret.NewUpdateSecretCreated().WithPayload(transformK8sToVmw(*updatedSecret))
}

func (h *Handlers) deleteSecret(params secret.DeleteSecretParams, principal interface{}) middleware.Responder {
	trace.Trace("secret.DeleteSecret." + params.SecretName)()
	err := h.secretsAPI.Delete(params.SecretName, &metav1.DeleteOptions{})
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
