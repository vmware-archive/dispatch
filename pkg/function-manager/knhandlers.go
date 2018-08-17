///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functionmanager

import (
	"github.com/go-openapi/runtime/middleware"
	knclientset "github.com/knative/serving/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/dispatch/pkg/utils/knaming"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	fnrunner "github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations/runner"
	fnstore "github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations/store"
)

func kubeClientConfig(kubeconfPath string) (*rest.Config, error) {
	if kubeconfPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfPath)
	}
	return rest.InClusterConfig()
}

func knClient(kubeconfPath string) knclientset.Interface {
	config, err := kubeClientConfig(kubeconfPath)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error configuring k8s API client"))
	}
	return knclientset.NewForConfigOrDie(config)
}

type knHandlers struct {
	knClient knclientset.Interface
}

// NewHandlers is the constructor for the function manager API knHandlers
func NewHandlers(kubeconfPath string) Handlers {
	return &knHandlers{knClient: knClient(kubeconfPath)}
}

func (h *knHandlers) addFunction(params fnstore.AddFunctionParams, principal interface{}) middleware.Responder {
	org := params.XDispatchOrg
	project := params.XDispatchProject

	function := params.Body
	knaming.AdjustMeta(&function.Meta, org, project)

	service := ToKnService(function)

	if err := service.Validate(); err != nil {
		// TODO handle validation error
		panic(errors.Wrap(err, "knative service validation"))
	}

	services := h.knClient.ServingV1alpha1().Services(org)

	createdService, err := services.Create(service)
	if err != nil {
		// TODO handler knative service creation error
		panic(errors.Wrap(err, "creating knative service"))
	}

	return fnstore.NewAddFunctionCreated().WithPayload(FromKnService(createdService))
}

func (*knHandlers) getFunction(params fnstore.GetFunctionParams, principal interface{}) middleware.Responder {
	//params.FunctionName
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
