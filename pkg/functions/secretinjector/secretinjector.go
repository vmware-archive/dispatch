///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package secretinjector

import (
	"context"

	apiclient "github.com/go-openapi/runtime/client"

	"github.com/pkg/errors"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions"
	secretclient "gitlab.eng.vmware.com/serverless/serverless/pkg/secret-store/gen/client"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/secret-store/gen/client/secret"

	log "github.com/sirupsen/logrus"
)

type secretInjector struct {
	client *secretclient.SecretStore
}

// New create a new secret injector
func New(secretClient *secretclient.SecretStore) functions.SecretInjector {
	return &secretInjector{client: secretClient}
}

type injectorError struct {
	Err error `json:"err"`
}

func (err *injectorError) Error() string {
	return err.Err.Error()
}

func (err *injectorError) AsUserErrorObject() interface{} {
	return err
}

func (injector *secretInjector) getSecrets(secretNames []string, cookie string) (map[string]interface{}, error) {

	secrets := make(map[string]interface{})
	apiKeyAuth := apiclient.APIKeyAuth("cookie", "header", cookie)
	for _, name := range secretNames {
		resp, err := injector.client.Secret.GetSecret(&secret.GetSecretParams{
			SecretName: name,
			Context:    context.Background(),
		}, apiKeyAuth)
		if err != nil {
			return secrets, errors.Wrapf(err, "failed to get secrets from secret store")
		}
		secrets[*resp.Payload.Name] = resp.Payload.Secrets
	}
	return secrets, nil
}

func (injector *secretInjector) GetMiddleware(secretNames []string, cookie string) functions.Middleware {
	return func(f functions.Runnable) functions.Runnable {
		return func(input map[string]interface{}) (map[string]interface{}, error) {
			secrets, err := injector.getSecrets(secretNames, cookie)
			if err != nil {
				log.Errorf("error when get secrets from secret store %+v", err)
				return nil, &injectorError{err}
			}
			input["_meta"] = map[string]interface{}{"secrets": secrets}
			output, err := f(input)
			if err != nil {
				return nil, err
			}
			delete(output, "_meta")
			return output, nil
		}
	}
}
