///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package drivers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/event-manager/drivers/entities"
	"github.com/vmware/dispatch/pkg/event-manager/gen/models"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations"
	driverapi "github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/drivers"
	"github.com/vmware/dispatch/pkg/secret-store/gen/client/secret"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

var builtInDrivers = map[string]map[string]bool{
	"vcenter": {
		"vcenterurl": true,
	},
}

// Handlers is a base struct for event manager drivers API handlers.
type Handlers struct {
	store   entitystore.EntityStore
	watcher controller.Watcher
	config  ConfigOpts
}

// ConfigOpts configures driver Handlers
type ConfigOpts struct {
	DriverImage     string
	SidecarImage    string
	TransportType   string
	RabbitMQURL     string
	KafkaBrokers    []string
	TracerURL       string
	K8sConfig       string
	DriverNamespace string
	SecretStoreURL  string
	OrgID           string
}

// NewHandlers Creates new instance of driver handlers
func NewHandlers(store entitystore.EntityStore, watcher controller.Watcher, config ConfigOpts) *Handlers {
	return &Handlers{
		watcher: watcher,
		store:   store,
		config:  config,
	}
}

// ConfigureHandlers configures API handlers for driver endpoints
func (h *Handlers) ConfigureHandlers(api middleware.RoutableAPI) {
	defer trace.Trace("subscriptions.ConfigureHandlers")()
	a, ok := api.(*operations.EventManagerAPI)
	if !ok {
		panic("Cannot configure api")
	}

	a.DriversAddDriverHandler = driverapi.AddDriverHandlerFunc(h.addDriver)
	a.DriversGetDriverHandler = driverapi.GetDriverHandlerFunc(h.getDriver)
	a.DriversGetDriversHandler = driverapi.GetDriversHandlerFunc(h.getDrivers)
	a.DriversDeleteDriverHandler = driverapi.DeleteDriverHandlerFunc(h.deleteDriver)
	a.DriversAddDriverTypeHandler = driverapi.AddDriverTypeHandlerFunc(h.addDriverType)
	a.DriversGetDriverTypeHandler = driverapi.GetDriverTypeHandlerFunc(h.getDriverType)
	a.DriversGetDriverTypesHandler = driverapi.GetDriverTypesHandlerFunc(h.getDriverTypes)
	a.DriversDeleteDriverTypeHandler = driverapi.DeleteDriverTypeHandlerFunc(h.deleteDriverType)
}

func (h *Handlers) addDriver(params driverapi.AddDriverParams, principal interface{}) middleware.Responder {
	defer trace.Tracef("name: %s", *params.Body.Name)()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.addDriver")
	defer sp.Finish()

	if err := params.Body.Validate(strfmt.Default); err != nil {
		return driverapi.NewAddDriverBadRequest().WithPayload(&models.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(fmt.Sprintf("invalid event driver payload: %s", err)),
		})
	}

	d := &entities.Driver{}
	d.FromModel(params.Body, h.config.OrgID)

	// If driver
	if _, ok := builtInDrivers[d.Type]; ok {
		d.Image = h.config.DriverImage
	} else {
		driverType := h.getDT(d.Type)
		if driverType == nil {
			return driverapi.NewAddDriverBadRequest().WithPayload(&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(fmt.Sprintf("Specified driver type %s does not exist", d.Type)),
			})
		}
		d.Image = driverType.Image
	}

	// validate the driver config
	// TODO: find a better way to do the validation
	if err := h.validateEventDriver(d); err != nil {
		log.Errorln(err)
		return driverapi.NewAddDriverBadRequest().WithPayload(&models.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(fmt.Sprintf("invalid event driver type or configuration: %s", err)),
		})
	}

	d.Status = entitystore.StatusCREATING
	if _, err := h.store.Add(d); err != nil {
		if entitystore.IsUniqueViolation(err) {
			return driverapi.NewAddDriverConflict().WithPayload(&models.Error{
				Code:    http.StatusConflict,
				Message: swag.String("error creating driver: non-unique name"),
			})
		}
		log.Errorf("store error when adding a new driver %s: %+v", d.Name, err)
		return driverapi.NewAddDriverInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when storing a new event driver"),
		})
	}
	if h.watcher != nil {
		h.watcher.OnAction(d)
	} else {
		log.Debugf("note: the watcher is nil")
	}
	return driverapi.NewAddDriverCreated().WithPayload(d.ToModel())
}

func (h *Handlers) getDriver(params driverapi.GetDriverParams, principal interface{}) middleware.Responder {
	defer trace.Trace("getDriver")()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.getDriver")
	defer sp.Finish()

	d := &entities.Driver{}

	filter, err := utils.ParseTags(entitystore.FilterEverything(), params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return driverapi.NewDeleteDriverBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	opts := entitystore.Options{Filter: filter}

	err = h.store.Get(h.config.OrgID, params.DriverName, opts, d)
	if err != nil {
		log.Warnf("Received GET for non-existent driver %s", params.DriverName)
		log.Debugf("store error when getting driver: %+v", err)
		return driverapi.NewGetDriverNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String(fmt.Sprintf("driver %s not found", params.DriverName)),
			})
	}
	return driverapi.NewGetDriverOK().WithPayload(d.ToModel())
}

func (h *Handlers) getDrivers(params driverapi.GetDriversParams, principal interface{}) middleware.Responder {
	defer trace.Trace("getDrivers")()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.getDrivers")
	defer sp.Finish()

	var drivers []*entities.Driver

	filter, err := utils.ParseTags(entitystore.FilterEverything(), params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return driverapi.NewDeleteDriverBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	opts := entitystore.Options{Filter: filter}

	// delete filter
	err = h.store.List(h.config.OrgID, opts, &drivers)
	if err != nil {
		log.Errorf("store error when listing drivers: %+v", err)
		return driverapi.NewGetDriverDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when getting drivers"),
			})
	}
	var driverModels []*models.Driver
	for _, driver := range drivers {
		driverModels = append(driverModels, driver.ToModel())
	}
	return driverapi.NewGetDriversOK().WithPayload(driverModels)
}

func (h *Handlers) deleteDriver(params driverapi.DeleteDriverParams, principal interface{}) middleware.Responder {
	defer trace.Tracef("name '%s'", params.DriverName)()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.deleteDriver")
	defer sp.Finish()

	name := params.DriverName

	filter, err := utils.ParseTags(entitystore.FilterEverything(), params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return driverapi.NewDeleteDriverBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	opts := entitystore.Options{Filter: filter}

	d := &entities.Driver{}

	if err = h.store.Get(h.config.OrgID, name, opts, d); err != nil {
		log.Errorf("store error when getting driver: %+v", err)
		return driverapi.NewDeleteDriverNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("driver not found"),
			})
	}
	d.Status = entitystore.StatusDELETING
	if _, err = h.store.Update(d.Revision, d); err != nil {
		log.Errorf("store error when deleting the event driver %s: %+v", d.Name, err)
		return driverapi.NewDeleteDriverInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when deleting an event driver"),
		})
	}
	if h.watcher != nil {
		h.watcher.OnAction(d)
	} else {
		log.Debugf("note: the watcher is nil")
	}
	return driverapi.NewDeleteDriverOK().WithPayload(d.ToModel())
}

func (h *Handlers) getDT(driverTypeName string) *entities.DriverType {
	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}

	t := entities.DriverType{}

	err := h.store.Get(h.config.OrgID, driverTypeName, opts, &t)
	if err != nil {
		log.Debugf("store error when getting driver type %s: %+v", driverTypeName, err)
		return nil
	}
	return &t
}

func (h *Handlers) addDriverType(params driverapi.AddDriverTypeParams, principal interface{}) middleware.Responder {
	defer trace.Trace("addDriverType")()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.addDriverType")
	defer sp.Finish()

	if err := params.Body.Validate(strfmt.Default); err != nil {
		return driverapi.NewAddDriverTypeBadRequest().WithPayload(&models.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(fmt.Sprintf("invalid driver type payload: %s", err)),
		})
	}

	name := *params.Body.Name
	if _, ok := builtInDrivers[name]; ok {
		return driverapi.NewGetDriverTypeBadRequest().WithPayload(&models.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(fmt.Sprintf("Built-in event driver type %s already exists", name)),
		})
	}

	dt := &entities.DriverType{}
	dt.FromModel(params.Body, h.config.OrgID)
	dt.Status = entitystore.StatusREADY
	if _, err := h.store.Add(dt); err != nil {
		if entitystore.IsUniqueViolation(err) {
			return driverapi.NewAddDriverTypeConflict().WithPayload(&models.Error{
				Code:    http.StatusConflict,
				Message: swag.String("error creating driver type: non-unique name"),
			})
		}
		log.Errorf("store error when adding a new driver type %s: %+v", dt.Name, err)
		return driverapi.NewAddDriverTypeInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when storing a new event driver type"),
		})
	}

	return driverapi.NewAddDriverTypeCreated().WithPayload(dt.ToModel())
}

func (h *Handlers) getDriverType(params driverapi.GetDriverTypeParams, principal interface{}) middleware.Responder {
	defer trace.Trace("getDriverType")()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.getDriverType")
	defer sp.Finish()

	if _, ok := builtInDrivers[params.DriverTypeName]; ok {
		// Return built-in driver type
		// TODO: See if there is a better way to handle built-in driver types
		tm := models.DriverType{
			Image:   swag.String(h.config.DriverImage),
			Name:    swag.String(params.DriverTypeName),
			BuiltIn: swag.Bool(true),
		}
		return driverapi.NewGetDriverTypeOK().WithPayload(&tm)
	}

	dt := &entities.DriverType{}

	filter, err := utils.ParseTags(entitystore.FilterEverything(), params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return driverapi.NewGetDriverTypeBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	opts := entitystore.Options{Filter: filter}

	if err = h.store.Get(h.config.OrgID, params.DriverTypeName, opts, dt); err != nil {
		log.Warnf("Received GET for non-existent driver type %s", params.DriverTypeName)
		log.Debugf("store error when getting driver type: %+v", err)
		return driverapi.NewGetDriverNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String(fmt.Sprintf("driver type %s not found", params.DriverTypeName)),
			})
	}
	return driverapi.NewGetDriverTypeOK().WithPayload(dt.ToModel())
}

func (h *Handlers) getDriverTypes(params driverapi.GetDriverTypesParams, principal interface{}) middleware.Responder {
	defer trace.Trace("getDriverTypes")()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.getDriverTypes")
	defer sp.Finish()

	var driverTypes []*entities.DriverType

	filter, err := utils.ParseTags(entitystore.FilterEverything(), params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return driverapi.NewGetDriverTypeBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	opts := entitystore.Options{Filter: filter}

	// delete filter
	err = h.store.List(h.config.OrgID, opts, &driverTypes)
	if err != nil {
		log.Errorf("store error when listing driver types: %+v", err)
		return driverapi.NewGetDriverTypesDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when getting driver types"),
			})
	}
	var driverTypeModels []*models.DriverType
	for _, dt := range driverTypes {
		driverTypeModels = append(driverTypeModels, dt.ToModel())
	}
	for typeName := range builtInDrivers {
		// Include built-in driver types.
		// TODO: See if there is a better way to handle built-in driver types
		d := models.DriverType{
			Image:   swag.String(h.config.DriverImage),
			Name:    swag.String(typeName),
			BuiltIn: swag.Bool(true),
		}
		driverTypeModels = append(driverTypeModels, &d)
	}
	return driverapi.NewGetDriverTypesOK().WithPayload(driverTypeModels)
}

func (h *Handlers) deleteDriverType(params driverapi.DeleteDriverTypeParams, principal interface{}) middleware.Responder {
	defer trace.Tracef("name '%s'", params.DriverTypeName)()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.deleteDriverType")
	defer sp.Finish()

	filter, err := utils.ParseTags(entitystore.FilterEverything(), params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return driverapi.NewDeleteDriverTypeBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	opts := entitystore.Options{Filter: filter}

	dt := &entities.DriverType{}

	if err = h.store.Get(h.config.OrgID, params.DriverTypeName, opts, dt); err != nil {
		log.Errorf("store error when getting driver type: %+v", err)
		return driverapi.NewDeleteDriverTypeNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("driver type not found"),
			})
	}
	if err = h.store.Delete(h.config.OrgID, dt.Name, dt); err != nil {
		log.Errorf("store error when deleting the event driver type %s: %+v", dt.Name, err)
		return driverapi.NewDeleteDriverTypeInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when deleting an event driver type"),
		})
	}
	return driverapi.NewDeleteDriverTypeOK().WithPayload(dt.ToModel())
}

// make sure the input includes all required config values
func (h *Handlers) validateEventDriver(driver *entities.Driver) error {
	template, ok := builtInDrivers[driver.Type]
	if !ok {
		// custom driver, no validation
		return nil
	}
	apiKeyAuth := client.APIKeyAuth("cookie", "header", "cookie")
	secrets := make(map[string]string)
	for _, name := range driver.Secrets {
		resp, err := SecretStoreClient(h.config.SecretStoreURL).Secret.GetSecret(&secret.GetSecretParams{
			SecretName: name,
			Context:    context.Background(),
		}, apiKeyAuth)
		if err != nil {
			return errors.Wrapf(err, "failed to get secret %s from secret store", name)
		}
		for key, value := range resp.Payload.Secrets {
			secrets[key] = value
		}
	}

	for k := range template {
		if _, ok := driver.Config[k]; ok {
			continue
		}
		if _, ok := secrets[k]; ok {
			continue
		}
		return fmt.Errorf("no configuration field %s in config or secrets", k)
	}

	return nil
}
