///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package injectors

import (
	"context"

	"github.com/vmware/dispatch/pkg/entity-store"

	apiclient "github.com/go-openapi/runtime/client"

	"github.com/pkg/errors"
	"github.com/vmware/dispatch/pkg/functions"
	secretclient "github.com/vmware/dispatch/pkg/secret-store/gen/client"
	serviceclient "github.com/vmware/dispatch/pkg/service-manager/gen/client"
	serviceinstance "github.com/vmware/dispatch/pkg/service-manager/gen/client/service_instance"

	log "github.com/sirupsen/logrus"
)

type serviceInjector struct {
	secretClient  *secretclient.SecretStore
	serviceClient *serviceclient.ServiceManager
}

// NewServiceInjector create a new secret injector
func NewServiceInjector(secretClient *secretclient.SecretStore, serviceClient *serviceclient.ServiceManager) functions.ServiceInjector {
	return &serviceInjector{
		secretClient:  secretClient,
		serviceClient: serviceClient,
	}
}

func getServiceBindings(serviceClient *serviceclient.ServiceManager, secretClient *secretclient.SecretStore, serviceNames []string, cookie string) (map[string]interface{}, error) {
	bindings := make(map[string]interface{})
	apiKeyAuth := apiclient.APIKeyAuth("cookie", "header", cookie)
	for _, name := range serviceNames {
		log.Debugf("getting service instance %s", name)
		resp, err := serviceClient.ServiceInstance.GetServiceInstanceByName(&serviceinstance.GetServiceInstanceByNameParams{
			ServiceInstanceName: name,
			Context:             context.Background(),
		}, apiKeyAuth)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get service instance %s from service manager", name)
		}
		log.Debugf("found service instance %s", name)
		if string(resp.Payload.Binding.Status) != string(entitystore.StatusREADY) {
			return nil, errors.Errorf("failed to get service bindings current status %s", resp.Payload.Binding.Status)
		}
		log.Debugf("getting service binding %s for service %s", resp.Payload.ID, name)
		secrets, err := getSecrets(secretClient, []string{resp.Payload.ID.String()}, cookie)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get service binding secrets for service instance %s", name)
		}
		log.Debugf("found service binding %s for service %s", resp.Payload.ID, name)
		bindings[name] = secrets
	}
	return bindings, nil
}

func (i *serviceInjector) GetMiddleware(serviceNames []string, cookie string) functions.Middleware {
	return func(f functions.Runnable) functions.Runnable {
		return func(ctx functions.Context, in interface{}) (interface{}, error) {
			bindings, err := getServiceBindings(i.serviceClient, i.secretClient, serviceNames, cookie)
			if err != nil {
				log.Errorf("error when getting service bindings from service manager %+v", err)
				return nil, &injectorError{errors.Wrap(err, "error when retrieving bindings from service manager")}
			}
			ctx["serviceBindings"] = bindings
			out, err := f(ctx, in)
			if err != nil {
				return nil, err
			}
			return out, nil
		}
	}
}
