///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package backend

import (
	"context"
	"encoding/json"
	"fmt"

	knclientset "github.com/knative/serving/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	dapi "github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/functions/config"
	"github.com/vmware/dispatch/pkg/utils"
	"github.com/vmware/dispatch/pkg/utils/knaming"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BuildConfig contains build configuration data.
type BuildConfig struct {
	BuildImage     string
	BuildCommand   string
	Env            map[string]string
	BuildTemplate  string
	ServiceAccount string
	StorageConfig  *config.StorageConfig
}

type knative struct {
	knClient      knclientset.Interface
	ingressGatway string
	buildConfig   *BuildConfig
}

func knClient(kubeconfPath string) knclientset.Interface {
	config, err := utils.KubeClientConfig(kubeconfPath)
	if err != nil {
		log.Fatalf("%+v", errors.Wrap(err, "error configuring k8s API client"))
	}
	return knclientset.NewForConfigOrDie(config)
}

//Knative returns a Knative functions backend
func Knative(kubeconfPath, ingressGateway, buildImage string, storageConfig *config.StorageConfig) Backend {
	// TODO: This all should come from a configmap or some sort of config mechanism
	buildConfig := &BuildConfig{
		BuildImage:     buildImage,
		BuildCommand:   "/fetch_source.sh",
		BuildTemplate:  "function-template",
		ServiceAccount: "dispatch-build",
		StorageConfig:  storageConfig,
	}
	return &knative{
		knClient:      knClient(kubeconfPath),
		ingressGatway: ingressGateway,
		buildConfig:   buildConfig,
	}
}

func (h *knative) Add(ctx context.Context, function *dapi.Function) (*dapi.Function, error) {

	service := FromFunction(h.buildConfig, function)
	if err := service.Validate(); err != nil {
		fmt.Println(err.Message)
		return nil, ValidationError{err}
	}
	services := h.knClient.ServingV1alpha1().Services(function.Meta.Org)

	serviceString, _ := json.MarshalIndent(service, "", "\t")
	log.Debugf("creating service for function %s:\n", function.Name, string(serviceString))

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

	serviceName := knaming.FunctionName(*meta)

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

	serviceName := knaming.FunctionName(*meta)

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
	log.Debugf("fetching kservices for namespace: %s", meta.Org)
	services := h.knClient.ServingV1alpha1().Services(meta.Org)

	serviceList, err := services.List(v1.ListOptions{
		LabelSelector: knaming.ToLabelSelector(map[string]string{
			knaming.ProjectLabel: meta.Project,
			knaming.KnTypeLabel:  knaming.FunctionKnType,
		}),
	})
	if err != nil {
		return nil, errors.Wrap(err, "listing knative services")
	}

	var functions []*dapi.Function

	for i := range serviceList.Items {
		objectMeta := &serviceList.Items[i].ObjectMeta
		if objectMeta.Labels[knaming.OrgLabel] != "" && objectMeta.Labels[knaming.KnTypeLabel] == knaming.FunctionKnType {
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

func (h *knative) RunEndpoint(ctx context.Context, meta *dapi.Meta) (string, string, error) {
	routes := h.knClient.ServingV1alpha1().Routes(meta.Org)

	serviceName := knaming.FunctionName(*meta)

	route, err := routes.Get(serviceName, v1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return "", "", NotFound{err}
		}
		return "", "", errors.Wrapf(err, "getting knative route '%s'", serviceName)
	}

	if h.ingressGatway == "" {
		return "http://" + route.Status.DomainInternal, route.Status.DomainInternal, nil
	}

	return "http://" + h.ingressGatway, route.Status.DomainInternal, nil
}
