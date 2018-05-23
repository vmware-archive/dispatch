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
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/trace"
)

type serviceInjector struct {
	secretClient  client.SecretsClient
	serviceClient client.ServicesClient
}

// NewServiceInjector create a new secret injector
func NewServiceInjector(secretClient client.SecretsClient, serviceClient client.ServicesClient) functions.ServiceInjector {
	return &serviceInjector{
		secretClient:  secretClient,
		serviceClient: serviceClient,
	}
}

func getServiceBindings(ctx context.Context, serviceClient client.ServicesClient, secretClient client.SecretsClient, serviceNames []string) (map[string]interface{}, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()
	bindings := make(map[string]interface{})
	for _, name := range serviceNames {
		log.Debugf("getting service instance %s", name)
		resp, err := serviceClient.GetServiceInstance(ctx, name)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get service instance %s from service manager", name)
		}
		log.Debugf("found service instance %s", name)
		if string(resp.Binding.Status) != string(entitystore.StatusREADY) {
			return nil, errors.Errorf("failed to get service bindings current status %s", resp.Binding.Status)
		}
		log.Debugf("getting service binding %s for service %s", resp.ID, name)
		secrets, err := getSecrets(ctx, secretClient, []string{resp.ID.String()})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get service binding secrets for service instance %s", name)
		}
		log.Debugf("found service binding %s for service %s", resp.ID, name)
		bindings[name] = secrets
	}
	return bindings, nil
}

func (i *serviceInjector) GetMiddleware(serviceNames []string, cookie string) functions.Middleware {
	return func(f functions.Runnable) functions.Runnable {
		return func(ctx functions.Context, in interface{}) (interface{}, error) {
			gctx := context.Background()
			if ctxValue, ok := ctx[functions.GoContext]; ok {
				gctx = ctxValue.(context.Context)
				span, newCtx := trace.Trace(gctx, "Service Injector")
				defer span.Finish()
				gctx = newCtx
				ctx[functions.GoContext] = newCtx
			}
			bindings, err := getServiceBindings(gctx, i.serviceClient, i.secretClient, serviceNames)
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
