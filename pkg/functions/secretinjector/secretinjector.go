///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package secretinjector

import (
	"context"

	apiclient "github.com/go-openapi/runtime/client"

	"github.com/pkg/errors"
	"github.com/vmware/dispatch/pkg/functions"
	secretclient "github.com/vmware/dispatch/pkg/secret-store/gen/client"
	"github.com/vmware/dispatch/pkg/secret-store/gen/client/secret"

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
		if resp.Payload.Name == nil {
			err := errors.Errorf("%s", name)

			return secrets, err
		}

		for key, value := range resp.Payload.Secrets {
			secrets[key] = value
		}
		//secrets[*resp.Payload.Name] = resp.Payload.Secrets
	}
	return secrets, nil
}

func (injector *secretInjector) GetMiddleware(secretNames []string, cookie string) functions.Middleware {
	return func(f functions.Runnable) functions.Runnable {
		return func(ctx functions.Context, in interface{}) (interface{}, error) {
			secrets, err := injector.getSecrets(secretNames, cookie)
			if err != nil {
				log.Errorf("error when get secrets from secret store %+v", err)
				return nil, &injectorError{errors.Wrap(err, "error when retrieving secrets from secret store")}
			}
			ctx["secrets"] = secrets
			out, err := f(ctx, in)
			if err != nil {
				return nil, err
			}
			return out, nil
		}
	}
}
