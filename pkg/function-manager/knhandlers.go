///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functionmanager

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	fnrunner "github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations/runner"
	fnstore "github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations/store"
)

// NewHandlers is the constructor for the function manager API knHandlers
func NewHandlers(kubeconfPath string) Handlers {
	config, err := kubeClientConfig(kubeconfPath)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error configuring k8s API client"))
	}
	return &knHandlers{
		config,
	}
}

func kubeClientConfig(kubeconfPath string) (*rest.Config, error) {
	if kubeconfPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfPath)
	}
	return rest.InClusterConfig()
}

type knHandlers struct {
	k8sConfig *rest.Config
}

func (*knHandlers) addFunction(params fnstore.AddFunctionParams, principal interface{}) middleware.Responder {
	panic("implement me")
}

func (*knHandlers) getFunction(params fnstore.GetFunctionParams, principal interface{}) middleware.Responder {
	panic("implement me")
}

func (*knHandlers) deleteFunction(params fnstore.DeleteFunctionParams, principal interface{}) middleware.Responder {
	panic("implement me")
}

func (*knHandlers) getFunctions(params fnstore.GetFunctionsParams, principal interface{}) middleware.Responder {
	panic("implement me")
}

func (*knHandlers) updateFunction(params fnstore.UpdateFunctionParams, principal interface{}) middleware.Responder {
	panic("implement me")
}

func (*knHandlers) runFunction(params fnrunner.RunFunctionParams, principal interface{}) middleware.Responder {
	panic("implement me")
}

func (*knHandlers) getRun(params fnrunner.GetRunParams, principal interface{}) middleware.Responder {
	panic("implement me")
}

func (*knHandlers) getRuns(params fnrunner.GetRunsParams, principal interface{}) middleware.Responder {
	panic("implement me")
}
