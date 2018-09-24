///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package backend

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-openapi/strfmt"
	"github.com/knative/pkg/apis/istio/common/v1alpha1"
	"github.com/knative/pkg/apis/istio/v1alpha3"
	sharedclientset "github.com/knative/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	dapi "github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/utils"
	"github.com/vmware/dispatch/pkg/utils/knaming"
)

type knativeEndpointsConfig struct {
	InternalGateway string
	SharedGateway   string
	DispatchHost    string
}

type knative struct {
	knClient sharedclientset.Interface
	config   knativeEndpointsConfig
}

func knClient(kubeconfPath string) sharedclientset.Interface {
	config, err := utils.KubeClientConfig(kubeconfPath)
	if err != nil {
		log.Fatalf("%+v", errors.Wrap(err, "error configuring k8s API client"))
	}
	return sharedclientset.NewForConfigOrDie(config)
}

//Knative returns a Knative functions backend
func Knative(kubeconfPath, internalGateway, sharedGateway, dispatchHost string) Backend {
	return &knative{
		knClient: knClient(kubeconfPath),
		config: knativeEndpointsConfig{
			InternalGateway: internalGateway,
			SharedGateway:   sharedGateway,
			DispatchHost:    dispatchHost,
		},
	}
}

func (h *knative) Add(ctx context.Context, endpoint *dapi.Endpoint) (*dapi.Endpoint, error) {
	virtualService := h.fromEndpoint(endpoint)

	newVirtualService, err := h.knClient.NetworkingV1alpha3().VirtualServices(endpoint.Org).Create(virtualService)
	if err != nil {
		if kerrors.IsAlreadyExists(err) {
			return nil, AlreadyExists{err}
		}
		return nil, errors.Wrap(err, "creating istio virtualservice")
	}

	return h.toEndpoint(newVirtualService)
}

func (h *knative) Get(ctx context.Context, meta *dapi.Meta) (*dapi.Endpoint, error) {
	virtualServices := h.knClient.NetworkingV1alpha3().VirtualServices(meta.Org)
	endpointName := knaming.EndpointName(*meta)

	virtualService, err := virtualServices.Get(endpointName, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil, NotFound{err}
		}
		return nil, errors.Wrapf(err, "getting istio virtualservice '%s'", endpointName)
	}

	return h.toEndpoint(virtualService)
}

func (h *knative) Delete(ctx context.Context, meta *dapi.Meta) error {
	virtualServices := h.knClient.NetworkingV1alpha3().VirtualServices(meta.Org)
	endpointName := knaming.EndpointName(*meta)

	err := virtualServices.Delete(endpointName, &metav1.DeleteOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return NotFound{err}
		}
		return errors.Wrapf(err, "deleting istio virtualservice '%s'", endpointName)
	}
	return nil
}

func (h *knative) List(ctx context.Context, meta *dapi.Meta) ([]*dapi.Endpoint, error) {
	virtualServices := h.knClient.NetworkingV1alpha3().VirtualServices(meta.Org)

	log.Debugf("listing virtualservices for namespace: %s [%s]", meta.Org, meta.Project)
	virtualServiceList, err := virtualServices.List(metav1.ListOptions{
		LabelSelector: knaming.ToLabelSelector(map[string]string{
			knaming.ProjectLabel: meta.Project,
		}),
	})
	if err != nil {
		return nil, errors.Wrap(err, "listing istio virtualservices")
	}

	var endpoints []*dapi.Endpoint

	for i := range virtualServiceList.Items {
		objectMeta := &virtualServiceList.Items[i].ObjectMeta
		if objectMeta.Labels[knaming.OrgLabel] != "" {
			e, err := h.toEndpoint(&virtualServiceList.Items[i])
			if err != nil {
				return nil, err
			}
			endpoints = append(endpoints, e)
		}
	}
	log.Debugf("found %d virtualservices for namespace: %s [%s]", len(endpoints), meta.Org, meta.Project)
	return endpoints, nil
}

func (h *knative) Update(ctx context.Context, endpoint *dapi.Endpoint) (*dapi.Endpoint, error) {
	virtualService := h.fromEndpoint(endpoint)
	virtualServices := h.knClient.NetworkingV1alpha3().VirtualServices(endpoint.Org)

	updatedVirtualService, err := virtualServices.Update(virtualService)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil, NotFound{err}
		}
		return nil, errors.Wrap(err, "updating istio virtualservice")
	}

	return h.toEndpoint(updatedVirtualService)
}

func (h *knative) fromEndpoint(model *dapi.Endpoint) *v1alpha3.VirtualService {
	virtualService := &v1alpha3.VirtualService{
		ObjectMeta: knaming.ToObjectMeta(model.Meta, *model),
	}

	hosts := model.Hosts
	if len(model.Hosts) == 0 {
		hosts = append(hosts, fmt.Sprintf("%s.%s.%s", model.Project, model.Org, h.config.DispatchHost))
	}

	virtualService.Spec.Hosts = hosts
	virtualService.Spec.Gateways = []string{h.config.SharedGateway, "mesh"}

	var routes []v1alpha3.HTTPRoute
	for _, prefix := range model.Uris {
		// TODO: Check for conflicts/duplicate paths
		fName := knaming.FunctionName(dapi.Meta{Name: model.Function, Project: model.Project, Org: model.Org})
		log.Debugf("creating route for prefix %s to %s.%s.svc.cluster.local", prefix, fName, model.Org)
		var matches []v1alpha3.HTTPMatchRequest
		for _, method := range model.Methods {
			match := v1alpha3.HTTPMatchRequest{
				Uri:    &v1alpha1.StringMatch{Exact: prefix},
				Method: &v1alpha1.StringMatch{Exact: strings.ToUpper(method)},
			}
			matches = append(matches, match)
		}
		route := v1alpha3.HTTPRoute{
			Match: matches,
			Route: []v1alpha3.DestinationWeight{
				v1alpha3.DestinationWeight{
					Destination: v1alpha3.Destination{
						Host: h.config.InternalGateway,
						Port: v1alpha3.PortSelector{
							Number: 80,
						},
					},
					Weight: 100,
				},
			},
			Rewrite: &v1alpha3.HTTPRewrite{
				Authority: fmt.Sprintf("%s.%s.svc.cluster.local", fName, model.Org),
			},
		}
		routes = append(routes, route)
	}
	log.Infof("Routes: %+v", routes)
	virtualService.Spec.Http = routes
	return virtualService
}

func (h *knative) toEndpoint(virtualService *v1alpha3.VirtualService) (*dapi.Endpoint, error) {
	if virtualService == nil {
		return nil, nil
	}
	objMeta := &virtualService.ObjectMeta
	endpoint := dapi.NewEndpoint()
	if err := knaming.FromJSONString(objMeta.Annotations[knaming.InitialObjectAnnotation], endpoint); err != nil {
		return nil, fmt.Errorf("error decoding endpoint from virtualservice: %v", err)
	}
	endpoint.CreatedTime = virtualService.CreationTimestamp.Unix()
	endpoint.ID = strfmt.UUID(objMeta.UID)
	endpoint.ModifiedTime = virtualService.CreationTimestamp.Unix()
	endpoint.Hosts = virtualService.Spec.Hosts

	endpoint.BackingObject = virtualService
	return endpoint, nil
}
