///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package backend

import (
	"context"

	knclientset "github.com/knative/serving/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	dapi "github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/utils"
	"github.com/vmware/dispatch/pkg/utils/knaming"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// FunctionType is the function type name
	FunctionType = "Function"
)

// BuildConfig contains build configuration data.
type BuildConfig struct {
	BuildImage     string
	BuildCommand   string
	BuildTemplate  string
	ServiceAccount string
}

// FunctionName returns k8s API name of the Dispatch function
func FunctionName(meta dapi.Meta) string {
	return knaming.GetKnName("Function", meta)
}

type knative struct {
	knClient    knclientset.Interface
	buildConfig *BuildConfig
}

func knClient(kubeconfPath string) knclientset.Interface {
	config, err := utils.KubeClientConfig(kubeconfPath)
	if err != nil {
		log.Fatalf("%+v", errors.Wrap(err, "error configuring k8s API client"))
	}
	return knclientset.NewForConfigOrDie(config)
}

//Knative returns a Knative functions backend
func Knative(kubeconfPath string) Backend {
	// TODO: This all should come from a configmap or some sort of config mechanism
	buildConfig := &BuildConfig{
		BuildImage:     "dispatchframework/dispatch-knative-builder:0.0.1",
		BuildCommand:   "/fetch_source.sh",
		BuildTemplate:  "function-template",
		ServiceAccount: "dispatch-build",
	}
	return &knative{
		knClient:    knClient(kubeconfPath),
		buildConfig: buildConfig,
	}
}

func (h *knative) Add(ctx context.Context, function *dapi.Function) (*dapi.Function, error) {

	service := FromFunction(h.buildConfig, function)

	if err := service.Validate(); err != nil {
		return nil, ValidationError{err}
	}

	services := h.knClient.ServingV1alpha1().Services(function.Meta.Org)

	createdService, err := services.Create(service)
	if err != nil {
		if kerrors.IsAlreadyExists(err) {
			return nil, AlreadyExists{err}
		}
		return nil, errors.Wrap(err, "creating knative service")
	}

	return ToFunction(createdService), nil
}

func (h *knative) Get(ctx context.Context, meta *dapi.Meta) (*dapi.Function, error) {
	services := h.knClient.ServingV1alpha1().Services(meta.Org)

	serviceName := FunctionName(*meta)

	service, err := services.Get(serviceName, v1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil, NotFound{err}
		}
		return nil, errors.Wrapf(err, "getting knative service '%s'", serviceName)
	}

	return ToFunction(service), nil
}

func (h *knative) Delete(ctx context.Context, meta *dapi.Meta) error {
	services := h.knClient.ServingV1alpha1().Services(meta.Org)

	serviceName := FunctionName(*meta)

	err := services.Delete(serviceName, &v1.DeleteOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return NotFound{err}
		}
		return errors.Wrapf(err, "deleting knative service '%s'", serviceName)
	}
	return nil
}

func (h *knative) List(ctx context.Context, meta *dapi.Meta) ([]*dapi.Function, error) {
	services := h.knClient.ServingV1alpha1().Services(meta.Org)

	serviceList, err := services.List(v1.ListOptions{
		LabelSelector: knaming.ToLabelSelector(map[string]string{
			knaming.ProjectLabel: meta.Project,
			knaming.KnTypeLabel:  FunctionType,
		}),
	})
	if err != nil {
		return nil, errors.Wrap(err, "listing knative services")
	}

	var functions []*dapi.Function

	for i := range serviceList.Items {
		objectMeta := &serviceList.Items[i].ObjectMeta
		if objectMeta.Labels[knaming.OrgLabel] != "" && objectMeta.Labels[knaming.KnTypeLabel] == FunctionType {
			functions = append(functions, ToFunction(&serviceList.Items[i]))
		}
	}

	return functions, nil
}

func (h *knative) Update(ctx context.Context, function *dapi.Function) (*dapi.Function, error) {
	service := FromFunction(h.buildConfig, function)

	if err := service.Validate(); err != nil {
		return nil, ValidationError{err}
	}

	services := h.knClient.ServingV1alpha1().Services(function.Meta.Org)

	updatedService, err := services.Update(service)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil, NotFound{err}
		}
		return nil, errors.Wrap(err, "updating knative service")
	}

	return ToFunction(updatedService), nil
}

func (h *knative) RunEndpoint(ctx context.Context, meta *dapi.Meta) (string, error) {
	routes := h.knClient.ServingV1alpha1().Routes(meta.Org)

	serviceName := FunctionName(*meta)

	route, err := routes.Get(serviceName, v1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return "", NotFound{err}
		}
		return "", errors.Wrapf(err, "getting knative route '%s'", serviceName)
	}

	return "http://" + route.Status.DomainInternal, nil
}
