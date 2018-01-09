///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package apimanager

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/vmware/dispatch/pkg/api-manager/gateway"
	"github.com/vmware/dispatch/pkg/api-manager/gen/models"
	"github.com/vmware/dispatch/pkg/api-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/api-manager/gen/restapi/operations/endpoint"
	"github.com/vmware/dispatch/pkg/controller"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/trace"
)

// APIManagerFlags are configuration flags for the function manager
var APIManagerFlags = struct {
	Config          string `long:"config" description:"Path to Config file" default:"./config.dev.json"`
	DbFile          string `long:"db-file" description:"Backend DB URL/Path" default:"./db.bolt"`
	DbBackend       string `long:"db-backend" description:"Backend DB Name" default:"boltdb"`
	DbUser          string `long:"db-username" description:"Backend DB Username" default:"dispatch"`
	DbPassword      string `long:"db-password" description:"Backend DB Password" default:"dispatch"`
	DbDatabase      string `long:"db-database" description:"Backend DB Name" default:"dispatch"`
	OrgID           string `long:"organization" description:"(temporary) Static organization id" default:"dispatch"`
	GatewayHost     string `long:"gateway-host" description:"API Gateway server host" default:"gateway-kong"`
	Gateway         string `long:"gateway" description:"API Gateway Implementation" default:"kong"`
	FunctionManager string `long:"function-manager" description:"Function Manager Host" default:"function-manager"`
	ResyncPeriod    int    `long:"resync-period" description:"The time period (in seconds) to sync with api gateway" default:"60"`
}{}

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

func apiModelOntoEntity(m *models.API) *API {
	defer trace.Tracef("name '%s'", *m.Name)()
	e := API{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: APIManagerFlags.OrgID,
			Name:           *m.Name,
		},
		API: gateway.API{
			Name:           *m.Name,
			Function:       *m.Function,
			Authentication: m.Authentication,
			Enabled:        m.Enabled,
			TLS:            m.TLS,
			Hosts:          m.Hosts,
			Methods:        m.Methods,
			Protocols:      m.Protocols,
			URIs:           m.Uris,
			CORS:           m.Cors,
		},
	}
	return &e
}

func apiEntityToModel(e *API) *models.API {
	defer trace.Tracef("name '%s'", e.Name)()
	m := models.API{
		ID:             strfmt.UUID(e.ID),
		Name:           swag.String(e.Name),
		Function:       swag.String(e.API.Function),
		Authentication: e.API.Authentication,
		Enabled:        e.API.Enabled,
		TLS:            e.API.TLS,
		Hosts:          e.API.Hosts,
		Methods:        e.API.Methods,
		Protocols:      e.API.Protocols,
		Uris:           e.API.URIs,
		Status:         models.Status(e.Status),
		Cors:           e.API.CORS,
	}
	return &m
}

// ConfigureHandlers configure handlers for API Manager
func (h *Handlers) ConfigureHandlers(routableAPI middleware.RoutableAPI) {
	defer trace.Trace("ConfigureHandlers")()
	a, ok := routableAPI.(*operations.APIManagerAPI)
	if !ok {
		panic("Cannot configure API-Manager API")
	}

	a.CookieAuth = func(token string) (interface{}, error) {
		// TODO: be able to retrieve user information from the cookie
		// currently just return the cookie
		log.Printf("cookie auth: %s\n", token)
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
	defer trace.Tracef("name '%s'", *params.Body.Name)()
	e := apiModelOntoEntity(params.Body)

	e.Status = entitystore.StatusCREATING
	if _, err := h.Store.Add(e); err != nil {
		log.Errorf("store error when adding a new api %s: %+v", e.Name, err)
		return endpoint.NewAddAPIInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when storing a new api"),
		})
	}
	if h.watcher != nil {
		h.watcher.OnAction(e)
	} else {
		log.Debugf("note: the watcher is nil")
	}
	m := apiEntityToModel(e)
	return endpoint.NewAddAPIOK().WithPayload(m)
}

func (h *Handlers) deleteAPI(params endpoint.DeleteAPIParams, principal interface{}) middleware.Responder {
	defer trace.Tracef("name '%s'", params.API)()
	name := params.API
	var e API
	if err := h.Store.Get(APIManagerFlags.OrgID, name, &e); err != nil {
		log.Errorf("store error when getting api: %+v", err)
		return endpoint.NewDeleteAPINotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("api not found"),
			})
	}
	e.Status = entitystore.StatusDELETING
	if _, err := h.Store.Update(e.Revision, &e); err != nil {
		log.Errorf("store error when deleting the api %s: %+v", e.Name, err)
		return endpoint.NewDeleteAPIInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when deleting an api"),
		})
	}
	if h.watcher != nil {
		h.watcher.OnAction(&e)
	} else {
		log.Debugf("note: the watcher is nil")
	}
	return endpoint.NewDeleteAPIOK().WithPayload(apiEntityToModel(&e))
}

func (h *Handlers) getAPI(params endpoint.GetAPIParams, principal interface{}) middleware.Responder {
	defer trace.Tracef("name '%s'", params.API)()
	var e API
	err := h.Store.Get(APIManagerFlags.OrgID, params.API, &e)
	if err != nil {
		log.Errorf("store error when getting api: %+v", err)
		return endpoint.NewGetAPINotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("api not found"),
			})
	}
	return endpoint.NewGetAPIOK().WithPayload(apiEntityToModel(&e))
}

func (h *Handlers) getAPIs(params endpoint.GetApisParams, principal interface{}) middleware.Responder {
	defer trace.Trace("")()
	var apis []*API

	var filterNotDeleted = []entitystore.FilterStat{
		entitystore.FilterStat{
			Subject: "Delete", Verb: entitystore.FilterVerbEqual, Object: false,
		},
	}
	err := h.Store.List(APIManagerFlags.OrgID, filterNotDeleted, &apis)
	if err != nil {
		log.Errorf("store error when listing apis: %+v", err)
		return endpoint.NewGetApisDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when getting apis"),
			})
	}
	var apiModels []*models.API
	for _, api := range apis {
		apiModels = append(apiModels, apiEntityToModel(api))
	}
	return endpoint.NewGetApisOK().WithPayload(apiModels)
}

func (h *Handlers) updateAPI(params endpoint.UpdateAPIParams, principal interface{}) middleware.Responder {
	defer trace.Tracef("name '%s'", params.API)()
	name := params.API

	var e API
	err := h.Store.Get(APIManagerFlags.OrgID, name, &e)
	if err != nil {
		log.Errorf("store error when getting api: %+v", err)
		return endpoint.NewUpdateAPINotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("api not found"),
			})
	}
	e.Status = entitystore.StatusUPDATING
	updatedEntity := apiModelOntoEntity(params.Body)
	updatedEntity.API.ID = e.API.ID
	updatedEntity.API.CreatedAt = e.API.CreatedAt
	if _, err := h.Store.Update(e.Revision, updatedEntity); err != nil {
		log.Errorf("store error when updating api: %+v", err)
		return endpoint.NewUpdateAPIInternalServerError().WithPayload(
			&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when updating apis"),
			})
	}
	if h.watcher != nil {
		h.watcher.OnAction(&e)
	} else {
		log.Debugf("note: the watcher is nil")
	}
	return endpoint.NewUpdateAPIOK().WithPayload(apiEntityToModel(updatedEntity))
}
