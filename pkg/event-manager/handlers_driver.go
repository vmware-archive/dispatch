///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"context"
	"fmt"
	"net/http"

	apiclient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/event-manager/gen/models"
	driverapi "github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/drivers"
	"github.com/vmware/dispatch/pkg/secret-store/gen/client/secret"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

var builtInDrivers = map[string]map[string]bool{
	"vcenter": map[string]bool{
		"vcenterurl": true,
	},
}

// make sure the input includes all required config values
func validateEventDriver(driver *Driver) error {
	template, ok := builtInDrivers[driver.Type]
	if !ok {
		// custom driver, no validation
		return nil
	}

	apiKeyAuth := apiclient.APIKeyAuth("cookie", "header", "cookie")
	secrets := make(map[string]string)
	for _, name := range driver.Secrets {
		resp, err := SecretStoreClient().Secret.GetSecret(&secret.GetSecretParams{
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

	e := driverModelToEntity(params.Body)
	if _, ok := builtInDrivers[e.Type]; ok {
		e.Image = EventManagerFlags.EventDriverImage
		e.Mode = "http"
	} else {
		driverType := h.getDT(e.Type)
		if driverType == nil {
			return driverapi.NewAddDriverBadRequest().WithPayload(&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(fmt.Sprintf("Specified driver type %s does not exist", e.Type)),
			})
		}
		e.Image = driverType.Image
		e.Mode = driverType.Mode
	}

	// validate the driver config
	// TODO: find a better way to do the validation
	if err := validateEventDriver(e); err != nil {
		log.Errorln(err)
		return driverapi.NewAddDriverBadRequest().WithPayload(&models.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(fmt.Sprintf("invalid event driver type or configuration: %s", err)),
		})
	}

	e.Status = entitystore.StatusCREATING
	if _, err := h.Store.Add(e); err != nil {
		log.Errorf("store error when adding a new driver %s: %+v", e.Name, err)
		return driverapi.NewAddDriverInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when storing a new event driver"),
		})
	}
	if h.Watcher != nil {
		h.Watcher.OnAction(e)
	} else {
		log.Debugf("note: the watcher is nil")
	}
	return driverapi.NewAddDriverCreated().WithPayload(driverEntityToModel(e))
}

func (h *Handlers) getDriver(params driverapi.GetDriverParams, principal interface{}) middleware.Responder {
	defer trace.Trace("getDriver")()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.getDriver")
	defer sp.Finish()

	e := Driver{}
	var err error

	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return driverapi.NewGetDriverBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	err = h.Store.Get(EventManagerFlags.OrgID, params.DriverName, opts, &e)
	if err != nil {
		log.Warnf("Received GET for non-existent driver %s", params.DriverName)
		log.Debugf("store error when getting driver: %+v", err)
		return driverapi.NewGetDriverNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String(fmt.Sprintf("driver %s not found", params.DriverName)),
			})
	}
	return driverapi.NewGetDriverOK().WithPayload(driverEntityToModel(&e))
}

func (h *Handlers) getDrivers(params driverapi.GetDriversParams, principal interface{}) middleware.Responder {
	defer trace.Trace("getDrivers")()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.getDrivers")
	defer sp.Finish()

	var drivers []*Driver
	var err error

	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return driverapi.NewGetDriverBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	// delete filter
	err = h.Store.List(EventManagerFlags.OrgID, opts, &drivers)
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
		driverModels = append(driverModels, driverEntityToModel(driver))
	}
	return driverapi.NewGetDriversOK().WithPayload(driverModels)
}

func (h *Handlers) deleteDriver(params driverapi.DeleteDriverParams, principal interface{}) middleware.Responder {
	defer trace.Tracef("name '%s'", params.DriverName)()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.deleteDriver")
	defer sp.Finish()

	name := params.DriverName
	var e Driver
	var err error

	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return driverapi.NewDeleteDriverBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	if err := h.Store.Get(EventManagerFlags.OrgID, name, opts, &e); err != nil {
		log.Errorf("store error when getting driver: %+v", err)
		return driverapi.NewDeleteDriverNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("driver not found"),
			})
	}
	e.Status = entitystore.StatusDELETING
	if _, err := h.Store.Update(e.Revision, &e); err != nil {
		log.Errorf("store error when deleting the event driver %s: %+v", e.Name, err)
		return driverapi.NewDeleteDriverInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when deleting an event driver"),
		})
	}
	if h.Watcher != nil {
		h.Watcher.OnAction(&e)
	} else {
		log.Debugf("note: the watcher is nil")
	}
	return driverapi.NewDeleteDriverOK().WithPayload(driverEntityToModel(&e))
}

func (h *Handlers) getDT(driverTypeName string) *DriverType {
	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}

	t := DriverType{}

	err := h.Store.Get(EventManagerFlags.OrgID, driverTypeName, opts, &t)
	if err != nil {
		log.Debugf("store error when getting driver type %s: %+v", driverTypeName, err)
		return nil
	}
	return &t
}
