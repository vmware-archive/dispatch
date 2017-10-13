///////////////////////////////////////////////////////////////////////
// Copyright (C) 2017 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package secretstore

import (
	"log"

	"github.com/go-openapi/runtime/middleware"
	strfmt "github.com/go-openapi/strfmt"

	"github.com/pkg/errors"

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
	trace.Trace("secret.Create." + *params.Secret.Name)()
	inputSecret := transformVmwToK8s(*params.Secret)

	k8sSecret, err := h.secretsAPI.Create(inputSecret)

	if err != nil {
		errorString := err.Error()
		return secret.NewAddSecretDefault(500).WithPayload(&models.Error{Code: 500, Message: &errorString})
	}

	vmwSecret := transformK8sToVmw(*k8sSecret)

	return secret.NewAddSecretCreated().WithPayload(vmwSecret)
}

func (h *Handlers) readSecret(params secret.GetSecretsParams, principal interface{}) middleware.Responder {
	trace.Trace("secret.GetSecrets")()
	listOptions := metav1.ListOptions{}

	vmwSecrets, err := h.readSecrets(listOptions)

	if err != nil {
		errorString := err.Error()
		return secret.NewGetSecretsDefault(500).WithPayload(&models.Error{Code: 500, Message: &errorString})
	}

	return secret.NewGetSecretsOK().WithPayload(models.GetSecretsOKBody(vmwSecrets))
}

func (h *Handlers) readSecretByName(params secret.GetSecretParams, principal interface{}) middleware.Responder {
	trace.Trace("secret.GetSecret." + params.SecretName)()
	listOptions := metav1.ListOptions{
		FieldSelector: NameFieldSelector + params.SecretName,
	}

	vmwSecrets, err := h.readSecrets(listOptions)
	if err != nil {
		return secret.NewGetSecretNotFound()
	}

	if len(vmwSecrets) == 0 {
		// TODO: Add a meaningful error message
		return secret.NewGetSecretDefault(404)
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
		errorString := err.Error()
		return secret.NewUpdateSecretDefault(500).WithPayload(&models.Error{Code: 500, Message: &errorString})
	}

	return secret.NewUpdateSecretCreated().WithPayload(transformK8sToVmw(*updatedSecret))
}

func (h *Handlers) deleteSecret(params secret.DeleteSecretParams, principal interface{}) middleware.Responder {
	trace.Trace("secret.DeleteSecret." + params.SecretName)()
	err := h.secretsAPI.Delete(params.SecretName, &metav1.DeleteOptions{})

	if err != nil {
		errorString := err.Error()
		return secret.NewDeleteSecretDefault(500).WithPayload(&models.Error{Code: 500, Message: &errorString})
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
