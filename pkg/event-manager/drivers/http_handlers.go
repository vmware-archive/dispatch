///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package drivers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/event-manager/drivers/entities"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations"
	driverapi "github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/drivers"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

// Handlers is a base struct for event manager drivers API handlers.
type Handlers struct {
	store         entitystore.EntityStore
	watcher       controller.Watcher
	secretsClient client.SecretsClient
}

// NewHandlers Creates new instance of driver handlers
func NewHandlers(store entitystore.EntityStore, watcher controller.Watcher, secretsClient client.SecretsClient) *Handlers {
	return &Handlers{
		watcher:       watcher,
		store:         store,
		secretsClient: secretsClient,
	}
}

// ConfigureHandlers configures API handlers for driver endpoints
func (h *Handlers) ConfigureHandlers(api middleware.RoutableAPI) {
	a, ok := api.(*operations.EventManagerAPI)
	if !ok {
		panic("Cannot configure api")
	}

	a.DriversAddDriverHandler = driverapi.AddDriverHandlerFunc(h.addDriver)
	a.DriversGetDriverHandler = driverapi.GetDriverHandlerFunc(h.getDriver)
	a.DriversGetDriversHandler = driverapi.GetDriversHandlerFunc(h.getDrivers)
	a.DriversUpdateDriverHandler = driverapi.UpdateDriverHandlerFunc(h.updateDriver)
	a.DriversDeleteDriverHandler = driverapi.DeleteDriverHandlerFunc(h.deleteDriver)
	a.DriversAddDriverTypeHandler = driverapi.AddDriverTypeHandlerFunc(h.addDriverType)
	a.DriversGetDriverTypeHandler = driverapi.GetDriverTypeHandlerFunc(h.getDriverType)
	a.DriversGetDriverTypesHandler = driverapi.GetDriverTypesHandlerFunc(h.getDriverTypes)
	a.DriversUpdateDriverTypeHandler = driverapi.UpdateDriverTypeHandlerFunc(h.updateDriverType)
	a.DriversDeleteDriverTypeHandler = driverapi.DeleteDriverTypeHandlerFunc(h.deleteDriverType)
}

func (h *Handlers) addDriver(params driverapi.AddDriverParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	if err := params.Body.Validate(strfmt.Default); err != nil {
		return driverapi.NewAddDriverBadRequest().WithPayload(&v1.Error{
			Code:    http.StatusBadRequest,
			Message: utils.ErrorMsgBadRequest("event driver", "", err),
		})
	}

	d := &entities.Driver{}
	d.FromModel(params.Body, params.XDispatchOrg)

	driverType := h.getDT(ctx, d.OrganizationID, d.Type)
	if driverType == nil {
		return driverapi.NewAddDriverBadRequest().WithPayload(&v1.Error{
			Code:    http.StatusBadRequest,
			Message: utils.ErrorMsgBadRequest("event driver", d.Name, fmt.Errorf("driver type %s does not exist", d.Type)),
		})
	}
	d.Image = driverType.Image
	d.Expose = driverType.Expose

	d.Status = entitystore.StatusCREATING
	if _, err := h.store.Add(ctx, d); err != nil {
		if entitystore.IsUniqueViolation(err) {
			return driverapi.NewAddDriverConflict().WithPayload(&v1.Error{
				Code:    http.StatusConflict,
				Message: utils.ErrorMsgAlreadyExists("event driver", d.Name),
			})
		}
		log.Errorf("store error when adding a new driver %s: %+v", d.Name, err)
		return driverapi.NewAddDriverDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("event driver", d.Name),
		})
	}
	if h.watcher != nil {
		h.watcher.OnAction(ctx, d)
	} else {
		log.Debugf("note: the watcher is nil")
	}
	return driverapi.NewAddDriverCreated().WithPayload(d.ToModel())
}

func (h *Handlers) getDriver(params driverapi.GetDriverParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	d := &entities.Driver{}

	err := h.store.Get(ctx, params.XDispatchOrg, params.DriverName, entitystore.Options{}, d)
	if err != nil {
		log.Warnf("Received GET for non-existent driver %s", params.DriverName)
		log.Debugf("store error when getting driver: %+v", err)
		return driverapi.NewGetDriverNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("event driver", params.DriverName),
			})
	}
	return driverapi.NewGetDriverOK().WithPayload(d.ToModel())
}

func (h *Handlers) getDrivers(params driverapi.GetDriversParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	var drivers []*entities.Driver

	// delete filter
	err := h.store.List(ctx, params.XDispatchOrg, entitystore.Options{}, &drivers)
	if err != nil {
		log.Errorf("store error when listing drivers: %+v", err)
		return driverapi.NewGetDriverDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when getting drivers"),
			})
	}
	var driverModels []*v1.EventDriver
	for _, driver := range drivers {
		driverModels = append(driverModels, driver.ToModel())
	}
	return driverapi.NewGetDriversOK().WithPayload(driverModels)
}

func (h *Handlers) updateDriver(params driverapi.UpdateDriverParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	name := params.DriverName

	d := &entities.Driver{}

	if err := h.store.Get(ctx, params.XDispatchOrg, name, entitystore.Options{}, d); err != nil {
		log.Errorf("store error when getting driver: %+v", err)
		return driverapi.NewUpdateDriverNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("event driver", params.DriverName),
			})
	}
	d.FromModel(params.Body, d.OrganizationID)
	d.Status = entitystore.StatusUPDATING
	if _, err := h.store.Update(ctx, d.Revision, d); err != nil {
		log.Errorf("store error when updating the event driver %s: %+v", d.Name, err)
		return driverapi.NewUpdateDriverDefault(500).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: utils.ErrorMsgInternalError("event driver", params.DriverName),
			})
	}

	if h.watcher != nil {
		h.watcher.OnAction(ctx, d)
	} else {
		log.Debugf("note: the watcher is nil")
	}

	return driverapi.NewUpdateDriverOK().WithPayload(d.ToModel())
}

func (h *Handlers) deleteDriver(params driverapi.DeleteDriverParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	name := params.DriverName

	d := &entities.Driver{}

	if err := h.store.Get(ctx, params.XDispatchOrg, name, entitystore.Options{}, d); err != nil {
		log.Errorf("store error when getting driver: %+v", err)
		return driverapi.NewDeleteDriverNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("event driver", params.DriverName),
			})
	}
	d.Status = entitystore.StatusDELETING
	d.SetDelete(true)
	if _, err := h.store.Update(ctx, d.Revision, d); err != nil {
		log.Errorf("store error when deleting the event driver %s: %+v", d.Name, err)
		return driverapi.NewDeleteDriverDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("event driver", params.DriverName),
		})
	}
	if h.watcher != nil {
		h.watcher.OnAction(ctx, d)
	} else {
		log.Debugf("note: the watcher is nil")
	}
	return driverapi.NewDeleteDriverOK().WithPayload(d.ToModel())
}

func (h *Handlers) getDT(ctx context.Context, orgID string, driverTypeName string) *entities.DriverType {
	t := entities.DriverType{}

	err := h.store.Get(ctx, orgID, driverTypeName, entitystore.Options{}, &t)
	if err != nil {
		log.Debugf("store error when getting driver type %s: %+v", driverTypeName, err)
		return nil
	}
	return &t
}

func (h *Handlers) addDriverType(params driverapi.AddDriverTypeParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	if err := params.Body.Validate(strfmt.Default); err != nil {
		return driverapi.NewAddDriverTypeBadRequest().WithPayload(&v1.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(fmt.Sprintf("invalid driver type payload: %s", err)),
		})
	}

	name := *params.Body.Name
	dt := &entities.DriverType{}
	dt.FromModel(params.Body, params.XDispatchOrg)
	dt.Status = entitystore.StatusREADY
	if _, err := h.store.Add(ctx, dt); err != nil {
		if entitystore.IsUniqueViolation(err) {
			return driverapi.NewAddDriverTypeConflict().WithPayload(&v1.Error{
				Code:    http.StatusConflict,
				Message: utils.ErrorMsgAlreadyExists("event driver type", name),
			})
		}
		log.Errorf("store error when adding a new driver type %s: %+v", dt.Name, err)
		return driverapi.NewAddDriverTypeDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("event driver type", dt.Name),
		})
	}

	return driverapi.NewAddDriverTypeCreated().WithPayload(dt.ToModel())
}

func (h *Handlers) getDriverType(params driverapi.GetDriverTypeParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	dt := &entities.DriverType{}

	if err := h.store.Get(ctx, params.XDispatchOrg, params.DriverTypeName, entitystore.Options{}, dt); err != nil {
		log.Warnf("Received GET for non-existent driver type %s", params.DriverTypeName)
		log.Debugf("store error when getting driver type: %+v", err)
		return driverapi.NewGetDriverNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("event driver type", params.DriverTypeName),
			})
	}
	return driverapi.NewGetDriverTypeOK().WithPayload(dt.ToModel())
}

func (h *Handlers) getDriverTypes(params driverapi.GetDriverTypesParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	var driverTypes []*entities.DriverType

	// delete filter
	err := h.store.List(ctx, params.XDispatchOrg, entitystore.Options{}, &driverTypes)
	if err != nil {
		log.Errorf("store error when listing driver types: %+v", err)
		return driverapi.NewGetDriverTypesDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when getting driver types"),
			})
	}
	var driverTypeModels []*v1.EventDriverType
	for _, dt := range driverTypes {
		driverTypeModels = append(driverTypeModels, dt.ToModel())
	}

	return driverapi.NewGetDriverTypesOK().WithPayload(driverTypeModels)
}

func (h *Handlers) updateDriverType(params driverapi.UpdateDriverTypeParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	dt := &entities.DriverType{}

	if err := h.store.Get(ctx, params.XDispatchOrg, params.DriverTypeName, entitystore.Options{}, dt); err != nil {
		log.Errorf("store error when getting driver type: %+v", err)
		return driverapi.NewUpdateDriverTypeNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("event driver type", params.DriverTypeName),
			})
	}

	dt.FromModel(params.Body, params.XDispatchOrg)

	if _, err := h.store.Update(ctx, dt.Revision, dt); err != nil {
		log.Errorf("store error when updating the event driver type %s: %+v", dt.Name, err)
		return driverapi.NewUpdateDriverTypeDefault(500).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: utils.ErrorMsgInternalError("event driver type", dt.Name),
			})
	}

	return driverapi.NewUpdateDriverTypeOK().WithPayload(dt.ToModel())
}

func (h *Handlers) deleteDriverType(params driverapi.DeleteDriverTypeParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	dt := &entities.DriverType{}

	if err := h.store.Get(ctx, params.XDispatchOrg, params.DriverTypeName, entitystore.Options{}, dt); err != nil {
		log.Errorf("store error when getting driver type: %+v", err)
		return driverapi.NewDeleteDriverTypeNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("event driver type", params.DriverTypeName),
			})
	}
	if err := h.store.Delete(ctx, params.XDispatchOrg, dt.Name, dt); err != nil {
		log.Errorf("store error when deleting the event driver type %s: %+v", dt.Name, err)
		return driverapi.NewDeleteDriverTypeDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("event driver type", dt.Name),
		})
	}
	return driverapi.NewDeleteDriverTypeOK().WithPayload(dt.ToModel())
}
