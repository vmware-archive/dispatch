///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package endpoints

import (
	"net/http"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"

	dapi "github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/endpoints/backend"
	"github.com/vmware/dispatch/pkg/endpoints/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/endpoints/gen/restapi/operations/endpoint"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

// EndpointHandlers is the interface for endpoints
type EndpointHandlers interface {
	AddEndpoint(params endpoint.AddEndpointParams, principal interface{}) middleware.Responder
	DeleteEndpoint(params endpoint.DeleteEndpointParams, principal interface{}) middleware.Responder
	UpdateEndpoint(params endpoint.UpdateEndpointParams, principal interface{}) middleware.Responder
	GetEndpoint(params endpoint.GetEndpointParams, principal interface{}) middleware.Responder
	GetEndpoints(params endpoint.GetEndpointsParams, principal interface{}) middleware.Responder
}

type defaultHandlers struct {
	backend   backend.Backend
	namespace string
}

// NewHandlers is the constructor for the endpoint handlers
func NewHandlers(kubeconfPath, namespace, internalGateway, sharedGateway, dispatchHost string) EndpointHandlers {
	return &defaultHandlers{
		backend:   backend.Knative(kubeconfPath, internalGateway, sharedGateway, dispatchHost),
		namespace: namespace,
	}
}

// ConfigureHandlers configure handlers for Endpoint
func ConfigureHandlers(routableAPI middleware.RoutableAPI, handlers EndpointHandlers) {
	a, ok := routableAPI.(*operations.EndpointsAPI)
	if !ok {
		panic("Cannot configure endpoints handlers")
	}

	a.CookieAuth = func(token string) (interface{}, error) {
		// TODO: be able to retrieve user information from the cookie
		// currently just return the cookie
		return token, nil
	}

	a.BearerAuth = func(token string) (interface{}, error) {
		// TODO: Once IAM issues signed tokens, validate them here.
		return token, nil
	}

	a.Logger = log.Printf
	a.EndpointAddEndpointHandler = endpoint.AddEndpointHandlerFunc(handlers.AddEndpoint)
	a.EndpointDeleteEndpointHandler = endpoint.DeleteEndpointHandlerFunc(handlers.DeleteEndpoint)
	a.EndpointGetEndpointHandler = endpoint.GetEndpointHandlerFunc(handlers.GetEndpoint)
	a.EndpointGetEndpointsHandler = endpoint.GetEndpointsHandlerFunc(handlers.GetEndpoints)
	a.EndpointUpdateEndpointHandler = endpoint.UpdateEndpointHandlerFunc(handlers.UpdateEndpoint)
}

// AddEndpoint adds a new endpoint route
func (h *defaultHandlers) AddEndpoint(params endpoint.AddEndpointParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := h.namespace
	project := *params.XDispatchProject

	model := params.Body
	utils.AdjustMeta(&model.Meta, dapi.Meta{Org: org, Project: project})

	createdEndpoint, err := h.backend.Add(ctx, model)
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "creating endpoint"))
		return endpoint.NewAddEndpointDefault(http.StatusInternalServerError).WithPayload(
			&dapi.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String(err.Error()),
			})
	}
	log.Infof("created endpoint: %+v", createdEndpoint)
	return endpoint.NewAddEndpointOK().WithPayload(createdEndpoint)
}

// GetEndpoint gets an endpoint
func (h *defaultHandlers) GetEndpoint(params endpoint.GetEndpointParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := h.namespace
	project := *params.XDispatchProject

	name := params.Endpoint
	log.Debugf("getting endpoint %s in %s:%s", name, org, project)
	model, err := h.backend.Get(ctx, &dapi.Meta{Name: name, Org: org, Project: project})
	if err != nil {
		if _, ok := err.(backend.NotFound); ok {
			return endpoint.NewGetEndpointNotFound().WithPayload(&dapi.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("endpoint", name),
			})
		}
		errors.Wrapf(err, "getting endpoint '%s'", name)
	}

	return endpoint.NewGetEndpointOK().WithPayload(model)
}

// GetEndpoints gets endpoints
func (h *defaultHandlers) GetEndpoints(params endpoint.GetEndpointsParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := h.namespace
	project := *params.XDispatchProject

	log.Debugf("getting endpoints in %s:%s", org, project)
	models, err := h.backend.List(ctx, &dapi.Meta{Org: org, Project: project})
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "listing endpoints"))
		return endpoint.NewGetEndpointsDefault(http.StatusInternalServerError).WithPayload(
			&dapi.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String(err.Error()),
			})
	}

	return endpoint.NewGetEndpointsOK().WithPayload(models)
}

// UpdateEndpoint updates an endpoint route
func (h *defaultHandlers) UpdateEndpoint(params endpoint.UpdateEndpointParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := h.namespace
	project := *params.XDispatchProject

	model := params.Body
	utils.AdjustMeta(&model.Meta, dapi.Meta{Org: org, Project: project})

	updatedEndpoint, err := h.backend.Update(ctx, model)
	log.Infof("Updated Endpoint: %+v", updatedEndpoint)
	if err != nil {
		log.Errorf("cannot update endpoint: %v", err)
		return endpoint.NewUpdateEndpointDefault(http.StatusInternalServerError).WithPayload(
			&dapi.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String(err.Error()),
			})
	}
	return endpoint.NewUpdateEndpointOK().WithPayload(updatedEndpoint)
}

// DeleteEndpoint deletes an endpoint
func (h *defaultHandlers) DeleteEndpoint(params endpoint.DeleteEndpointParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := h.namespace
	project := *params.XDispatchProject

	name := params.Endpoint
	log.Debugf("deleting endpoint %s in %s:%s", name, org, project)
	err := h.backend.Delete(ctx, &dapi.Meta{Name: name, Org: org, Project: project})
	if err != nil {
		if _, ok := err.(backend.NotFound); ok {
			return endpoint.NewDeleteEndpointNotFound().WithPayload(&dapi.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("endpoint", name),
			})
		}
		errors.Wrapf(err, "deleting endpoint '%s'", name)
	}

	return endpoint.NewDeleteEndpointOK()
}
