///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package applicationmanager

// NO TEST

import (
	"net/http"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/application-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/application-manager/gen/restapi/operations/application"
	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

// ApplicationManagerFlags are configuration flags for the function manager
var ApplicationManagerFlags = struct {
	Config       string `long:"config" description:"Path to Config file" default:"./config.dev.json"`
	DbFile       string `long:"db-file" description:"Backend DB URL/Path" default:"./db.bolt"`
	DbBackend    string `long:"db-backend" description:"Backend DB Name" default:"boltdb"`
	DbUser       string `long:"db-username" description:"Backend DB Username" default:"dispatch"`
	DbPassword   string `long:"db-password" description:"Backend DB Password" default:"dispatch"`
	DbDatabase   string `long:"db-database" description:"Backend DB Name" default:"dispatch"`
	OrgID        string `long:"organization" description:"(temporary) Static organization id" default:"dispatch"`
	ResyncPeriod int    `long:"resync-period" description:"The time period (in seconds) to sync with api gateway" default:"60"`
	Tracer       string `long:"tracer" description:"Open Tracing Tracer endpoint" default:""`
}{}

// Handlers define a set of handlers for Application Manager
type Handlers struct {
	store   entitystore.EntityStore
	watcher controller.Watcher
}

// NewHandlers create a new Application Manager Handler
func NewHandlers(watcher controller.Watcher, store entitystore.EntityStore) *Handlers {
	return &Handlers{
		store:   store,
		watcher: watcher,
	}
}

// ConfigureHandlers configure handlers for Application Manager
func (h *Handlers) ConfigureHandlers(routableAPI middleware.RoutableAPI) {
	a, ok := routableAPI.(*operations.ApplicationManagerAPI)
	if !ok {
		panic("Cannot configure Application-Manager API")
	}

	a.CookieAuth = func(token string) (interface{}, error) {
		// TODO: be able to retrieve user information from the cookie
		// currently just return the cookie
		log.Printf("cookie auth: %s\n", token)
		return token, nil
	}

	a.BearerAuth = func(token string) (interface{}, error) {
		// TODO: Once IAM issues signed tokens, validate them here.
		log.Printf("bearer auth: %s\n", token)
		return token, nil
	}

	a.Logger = log.Printf
	a.ApplicationAddAppHandler = application.AddAppHandlerFunc(h.addApp)
	a.ApplicationDeleteAppHandler = application.DeleteAppHandlerFunc(h.deleteApp)
	a.ApplicationGetAppHandler = application.GetAppHandlerFunc(h.getApp)
	a.ApplicationGetAppsHandler = application.GetAppsHandlerFunc(h.getApps)
	a.ApplicationUpdateAppHandler = application.UpdateAppHandlerFunc(h.updateApp)
}

func applicationModelOntoEntity(m *v1.Application) *Application {
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	e := Application{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: ApplicationManagerFlags.OrgID,
			Name:           *m.Name,
			Tags:           tags,
		},
	}
	return &e
}

func applicationEntityToModel(e *Application) *v1.Application {
	var tags []*v1.Tag
	for k, v := range e.Tags {
		tags = append(tags, &v1.Tag{Key: k, Value: v})
	}
	m := v1.Application{
		ID:           strfmt.UUID(e.ID),
		Name:         swag.String(e.Name),
		Kind:         utils.ApplicationKind,
		Status:       v1.Status(e.Status),
		CreatedTime:  e.CreatedTime.Unix(),
		ModifiedTime: e.ModifiedTime.Unix(),
		Tags:         tags,
	}
	return &m
}

func (h *Handlers) addApp(params application.AddAppParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	e := applicationModelOntoEntity(params.Body)

	e.Status = entitystore.StatusREADY
	if _, err := h.store.Add(ctx, e); err != nil {
		if entitystore.IsUniqueViolation(err) {
			return application.NewAddAppConflict().WithPayload(&v1.Error{
				Code:    http.StatusConflict,
				Message: swag.String("error creating application: non-unique name"),
			})
		}
		log.Errorf("store error when adding a new application %s: %+v", e.Name, err)
		return application.NewAddAppInternalServerError().WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when storing a new api"),
		})
	}
	m := applicationEntityToModel(e)
	return application.NewAddAppOK().WithPayload(m)
}

func (h *Handlers) deleteApp(params application.DeleteAppParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	name := params.Application

	var app Application
	if err := h.store.Get(ctx, ApplicationManagerFlags.OrgID, name, entitystore.Options{}, &app); err != nil {
		log.Errorf("store error when getting application: %+v", err)
		return application.NewDeleteAppNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("application not found"),
			})
	}

	if err := h.store.Delete(ctx, app.OrganizationID, app.Name, &app); err != nil {
		return application.NewDeleteAppInternalServerError().WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String(errors.Wrap(err, "store error when deleting application").Error()),
			})
	}
	return application.NewDeleteAppOK().WithPayload(applicationEntityToModel(&app))
}

func (h *Handlers) getApp(params application.GetAppParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	var e Application
	err := h.store.Get(ctx, ApplicationManagerFlags.OrgID, params.Application, entitystore.Options{}, &e)
	if err != nil {
		log.Errorf("store error when getting application: %+v", err)
		return application.NewGetAppNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("application not found"),
			})
	}
	return application.NewGetAppOK().WithPayload(applicationEntityToModel(&e))
}

func (h *Handlers) getApps(params application.GetAppsParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	var apps []*Application

	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}
	err := h.store.List(ctx, ApplicationManagerFlags.OrgID, opts, &apps)
	if err != nil {
		log.Errorf("store error when listing applications: %+v", err)
		return application.NewGetAppsDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when getting apis"),
			})
	}
	var appModels []*v1.Application
	for _, app := range apps {
		appModels = append(appModels, applicationEntityToModel(app))
	}
	return application.NewGetAppsOK().WithPayload(appModels)
}

func (h *Handlers) updateApp(params application.UpdateAppParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	name := params.Application

	var e Application
	err := h.store.Get(ctx, ApplicationManagerFlags.OrgID, name, entitystore.Options{}, &e)
	if err != nil {
		log.Errorf("store error when getting application: %+v", err)
		return application.NewUpdateAppNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("application not found"),
			})
	}
	e.Status = entitystore.StatusREADY
	updatedEntity := applicationModelOntoEntity(params.Body)
	if _, err := h.store.Update(ctx, e.Revision, updatedEntity); err != nil {
		log.Errorf("store error when updating application: %+v", err)
		return application.NewUpdateAppInternalServerError().WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when updating application"),
			})
	}
	return application.NewUpdateAppOK().WithPayload(applicationEntityToModel(updatedEntity))
}
