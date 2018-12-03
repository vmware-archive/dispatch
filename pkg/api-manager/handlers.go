///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package apimanager

import (
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	"github.com/vmware/dispatch/pkg/api-manager/gateway"
	"github.com/vmware/dispatch/pkg/api-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/api-manager/gen/restapi/operations/endpoint"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/controller"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

// Handlers define a set of handlers for API Manager
type Handlers struct {
	Store   entitystore.EntityStore
	watcher controller.Watcher
}

// NewHandlers create a new API Manager Handler
func NewHandlers(watcher controller.Watcher, store entitystore.EntityStore) *Handlers {
	return &Handlers{
		Store:   store,
		watcher: watcher,
	}
}

func apiModelOntoEntity(organizationID string, m *v1.API) *API {
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	// If hosts are missing then API path (URI) will be namespaced by the org name since all org's share an API-Gateway and the path
	// needs to be unique. If hosts is specified then we don't enforce this. TODO: enforce uniqueness of specified hosts when
	// sharing a gateway.
	var uris []string
	if len(m.Hosts) == 0 {
		for _, uri := range m.Uris {
			uris = append(uris, fmt.Sprintf("/%s/%s", organizationID, strings.TrimPrefix(uri, "/")))
		}
	} else {
		uris = m.Uris
	}
	var methods []string
	for _, method := range m.Methods {
		methods = append(methods, strings.ToUpper(method))
	}
	e := API{
		BaseEntity: entitystore.BaseEntity{
			Name:           *m.Name,
			OrganizationID: organizationID,
			Tags:           tags,
		},
		API: gateway.API{
			Name:           fmt.Sprintf("%s-%s", organizationID, *m.Name),
			OrganizationID: organizationID,
			Function:       *m.Function,
			Authentication: m.Authentication,
			Enabled:        m.Enabled,
			TLS:            m.TLS,
			Hosts:          m.Hosts,
			Methods:        methods,
			Protocols:      m.Protocols,
			URIs:           uris,
			CORS:           m.Cors,
		},
	}
	return &e
}

func apiEntityToModel(e *API) *v1.API {
	var tags []*v1.Tag
	for k, v := range e.Tags {
		tags = append(tags, &v1.Tag{Key: k, Value: v})
	}
	m := v1.API{
		ID:             strfmt.UUID(e.ID),
		Name:           swag.String(e.Name),
		Kind:           utils.APIKind,
		Function:       swag.String(e.API.Function),
		Authentication: e.API.Authentication,
		Enabled:        e.API.Enabled,
		TLS:            e.API.TLS,
		Hosts:          e.API.Hosts,
		Methods:        e.API.Methods,
		Protocols:      e.API.Protocols,
		Uris:           e.API.URIs,
		Status:         v1.Status(e.Status),
		Cors:           e.API.CORS,
		Tags:           tags,
	}
	return &m
}

// ConfigureHandlers configure handlers for API Manager
func (h *Handlers) ConfigureHandlers(routableAPI middleware.RoutableAPI) {
	a, ok := routableAPI.(*operations.APIManagerAPI)
	if !ok {
		panic("Cannot configure API-Manager API")
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
	a.EndpointAddAPIHandler = endpoint.AddAPIHandlerFunc(h.addAPI)
	a.EndpointDeleteAPIHandler = endpoint.DeleteAPIHandlerFunc(h.deleteAPI)
	a.EndpointGetAPIHandler = endpoint.GetAPIHandlerFunc(h.getAPI)
	a.EndpointGetApisHandler = endpoint.GetApisHandlerFunc(h.getAPIs)
	a.EndpointUpdateAPIHandler = endpoint.UpdateAPIHandlerFunc(h.updateAPI)
}

func (h *Handlers) addAPI(params endpoint.AddAPIParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	e := apiModelOntoEntity(params.XDispatchOrg, params.Body)

	e.Status = entitystore.StatusCREATING
	if _, err := h.Store.Add(ctx, e); err != nil {
		if entitystore.IsUniqueViolation(err) {
			return endpoint.NewAddAPIConflict().WithPayload(&v1.Error{
				Code:    http.StatusConflict,
				Message: utils.ErrorMsgAlreadyExists("API", e.Name),
			})
		}
		log.Errorf("store error when adding a new api %s: %+v", e.Name, err)
		return endpoint.NewAddAPIDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("API", e.Name),
		})
	}
	if h.watcher != nil {
		h.watcher.OnAction(ctx, e)
	} else {
		log.Debugf("note: the watcher is nil")
	}
	m := apiEntityToModel(e)
	return endpoint.NewAddAPIOK().WithPayload(m)
}

func (h *Handlers) deleteAPI(params endpoint.DeleteAPIParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	name := params.API

	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}

	var e API
	if err := h.Store.Get(ctx, params.XDispatchOrg, name, opts, &e); err != nil {
		log.Errorf("store error when getting api: %+v", err)
		return endpoint.NewDeleteAPINotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("API", name),
			})
	}
	e.Status = entitystore.StatusDELETING
	if _, err := h.Store.Update(ctx, e.Revision, &e); err != nil {
		log.Errorf("store error when deleting the api %s: %+v", e.Name, err)
		return endpoint.NewDeleteAPIDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("API", name),
		})
	}
	if h.watcher != nil {
		h.watcher.OnAction(ctx, &e)
	} else {
		log.Debugf("note: the watcher is nil")
	}
	return endpoint.NewDeleteAPIOK().WithPayload(apiEntityToModel(&e))
}

func (h *Handlers) getAPI(params endpoint.GetAPIParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()
	log.Debugf("Trying to get api with params: %+v", params)

	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}

	var e API
	log.Debugln("Getting from store")
	err := h.Store.Get(ctx, params.XDispatchOrg, params.API, opts, &e)
	if err != nil {
		log.Errorf("store error when getting api: %+v", err)
		return endpoint.NewGetAPINotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("API", params.API),
			})
	}
	return endpoint.NewGetAPIOK().WithPayload(apiEntityToModel(&e))
}

func (h *Handlers) getAPIs(params endpoint.GetApisParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()
	log.Debugf("Getting apis: %+v", params)

	var apis []*API

	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}

	err := h.Store.List(ctx, params.XDispatchOrg, opts, &apis)
	if err != nil {
		log.Errorf("store error when listing apis: %+v", err)
		return endpoint.NewGetApisDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when getting apis"),
			})
	}
	var apiModels []*v1.API
	for _, api := range apis {
		apiModels = append(apiModels, apiEntityToModel(api))
	}
	return endpoint.NewGetApisOK().WithPayload(apiModels)
}

func (h *Handlers) updateAPI(params endpoint.UpdateAPIParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	name := params.API

	log.Infof("Updating api: %+v", params)

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}

	var e API
	err = h.Store.Get(ctx, params.XDispatchOrg, name, opts, &e)
	log.Infof("Got api: %+v", e)
	if err != nil {
		log.Errorf("store error when getting api: %+v", err)
		return endpoint.NewUpdateAPINotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("API", name),
			})
	}

	updatedEntity := apiModelOntoEntity(params.XDispatchOrg, params.Body)
	updatedEntity.Status = entitystore.StatusUPDATING
	updatedEntity.API.ID = e.API.ID
	updatedEntity.API.CreatedAt = e.API.CreatedAt
	updatedEntity.ID = e.ID
	log.Infof("Going to update entity")
	if _, err := h.Store.Update(ctx, e.Revision, updatedEntity); err != nil {
		log.Errorf("store error when updating api: %+v", err)
		return endpoint.NewUpdateAPIDefault(500).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: utils.ErrorMsgInternalError("API", name),
			})
	}
	if h.watcher != nil {
		h.watcher.OnAction(ctx, updatedEntity)
	} else {
		log.Debugf("note: the watcher is nil")
	}
	return endpoint.NewUpdateAPIOK().WithPayload(apiEntityToModel(updatedEntity))
}
