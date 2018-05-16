///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package injectors

import (
	"context"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/functions"
)

type secretInjector struct {
	secretClient client.SecretsClient
}

// NewSecretInjector create a new secret injector
func NewSecretInjector(secretClient client.SecretsClient) functions.SecretInjector {
	return &secretInjector{
		secretClient: secretClient,
	}
}

func getSecrets(client client.SecretsClient, secretNames []string) (map[string]interface{}, error) {

	secrets := make(map[string]interface{})
	for _, name := range secretNames {
		resp, err := client.GetSecret(context.Background(), name)
		if err != nil {
			return secrets, errors.Wrapf(err, "failed to get secrets from secret store")
		}
		if resp.Name == nil {
			err := errors.Errorf("%s", name)

			return secrets, err
		}

		for key, value := range resp.Secrets {
			secrets[key] = value
		}
	}
	return secrets, nil
}

func (i *secretInjector) GetMiddleware(secretNames []string, cookie string) functions.Middleware {
	return func(f functions.Runnable) functions.Runnable {
		return func(ctx functions.Context, in interface{}) (interface{}, error) {
			secrets, err := getSecrets(i.secretClient, secretNames)
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
