///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package injectors

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
	secretClient *secretclient.SecretStore
}

// NewSecretInjector create a new secret injector
func NewSecretInjector(secretClient *secretclient.SecretStore) functions.SecretInjector {
	return &secretInjector{
		secretClient: secretClient,
	}
}

func getSecrets(client *secretclient.SecretStore, secretNames []string, cookie string) (map[string]interface{}, error) {

	secrets := make(map[string]interface{})
	apiKeyAuth := apiclient.APIKeyAuth("cookie", "header", cookie)
	for _, name := range secretNames {
		resp, err := client.Secret.GetSecret(&secret.GetSecretParams{
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
	}
	return secrets, nil
}

func (i *secretInjector) GetMiddleware(secretNames []string, cookie string) functions.Middleware {
	return func(f functions.Runnable) functions.Runnable {
		return func(ctx functions.Context, in interface{}) (interface{}, error) {
			secrets, err := getSecrets(i.secretClient, secretNames, cookie)
			if err != nil {
				log.Errorf("error when getting secrets from secret store %+v", err)
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
