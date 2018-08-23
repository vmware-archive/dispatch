///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functionmanager

import (
	"os"
	"path/filepath"

	"github.com/go-openapi/runtime/middleware"
	knclientset "github.com/knative/serving/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	dapi "github.com/vmware/dispatch/pkg/api/v1"
	fnrunner "github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations/runner"
	fnstore "github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations/store"
	"github.com/vmware/dispatch/pkg/utils/knaming"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func kubeClientConfig(kubeconfPath string) (*rest.Config, error) {
	if kubeconfPath == "" {
		userKubeConfig := filepath.Join(os.Getenv("HOME"), ".kube/config")
		if _, err := os.Stat(userKubeConfig); err == nil {
			kubeconfPath = userKubeConfig
		}
	}
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

func (h *knHandlers) addFunction(params fnstore.AddFunctionParams) middleware.Responder {
	org := *params.XDispatchOrg
	project := *params.XDispatchProject

	function := params.Body
	knaming.AdjustMeta(&function.Meta, org, project)

	service := FromFunction(function)

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

	return fnstore.NewAddFunctionCreated().WithPayload(ToFunction(createdService))
}

func (h *knHandlers) getFunction(params fnstore.GetFunctionParams) middleware.Responder {
	org := *params.XDispatchOrg
	project := *params.XDispatchProject

	name := params.FunctionName

	services := h.knClient.ServingV1alpha1().Services(org)

	serviceName := knaming.FunctionName(project, name)

	service, err := services.Get(serviceName, v1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return fnstore.NewGetFunctionNotFound()
		}
		// TODO the right thing
		panic(errors.Wrapf(err, "getting knative service '%s'", serviceName))
	}

	return fnstore.NewGetFunctionOK().WithPayload(ToFunction(service))
}

func (h *knHandlers) deleteFunction(params fnstore.DeleteFunctionParams) middleware.Responder {
	org := *params.XDispatchOrg
	project := *params.XDispatchProject

	name := params.FunctionName

	services := h.knClient.ServingV1alpha1().Services(org)

	serviceName := knaming.FunctionName(project, name)

	err := services.Delete(serviceName, &v1.DeleteOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return fnstore.NewDeleteFunctionOK()
		}
		// TODO the right thing
		panic(errors.Wrapf(err, "deleting knative service '%s'", serviceName))
	}

	return fnstore.NewDeleteFunctionOK()
}

func (h *knHandlers) getFunctions(params fnstore.GetFunctionsParams) middleware.Responder {
	org := *params.XDispatchOrg
	project := *params.XDispatchProject

	services := h.knClient.ServingV1alpha1().Services(org)

	serviceList, err := services.List(v1.ListOptions{
		LabelSelector: knaming.ToLabelSelector(map[string]string{
			knaming.ProjectLabel: project,
			knaming.KnTypeLabel:  knaming.FunctionKnType,
		}),
	})
	if err != nil {
		// TODO the right thing
		panic(errors.Wrap(err, "listing knative services"))
	}

	var functions []*dapi.Function

	for i := range serviceList.Items {
		objectMeta := &serviceList.Items[i].ObjectMeta
		if objectMeta.Labels[knaming.OrgLabel] != "" && objectMeta.Labels[knaming.KnTypeLabel] == knaming.FunctionKnType {
			functions = append(functions, ToFunction(&serviceList.Items[i]))
		}
	}

	return fnstore.NewGetFunctionsOK().WithPayload(functions)
}

func (h *knHandlers) updateFunction(params fnstore.UpdateFunctionParams) middleware.Responder {
	org := *params.XDispatchOrg
	project := *params.XDispatchProject

	function := params.Body
	knaming.AdjustMeta(&function.Meta, org, project)

	service := FromFunction(function)

	if err := service.Validate(); err != nil {
		// TODO handle validation error
		panic(errors.Wrap(err, "knative service validation"))
	}

	services := h.knClient.ServingV1alpha1().Services(org)

	updatedService, err := services.Update(service)
	if err != nil {
		// TODO handler knative service creation error
		panic(errors.Wrap(err, "updating knative service"))
	}

	return fnstore.NewUpdateFunctionOK().WithPayload(ToFunction(updatedService))
}

func (*knHandlers) runFunction(params fnrunner.RunFunctionParams) middleware.Responder {
	panic("implement me")
}

func (*knHandlers) getRun(params fnrunner.GetRunParams) middleware.Responder {
	panic("implement me")
}

func (*knHandlers) getRuns(params fnrunner.GetRunsParams) middleware.Responder {
	panic("implement me")
}
