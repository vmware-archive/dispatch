package eventmanager

import (
	"fmt"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/event-manager/gen/models"
	driverapi "github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/drivers"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

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

	t := driverTypeModelToEntity(params.Body)

	t.Status = entitystore.StatusREADY
	if _, err := h.Store.Add(t); err != nil {
		log.Errorf("store error when adding a new driver type %s: %+v", t.Name, err)
		return driverapi.NewAddDriverTypeInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when storing a new event driver type"),
		})
	}

	return driverapi.NewAddDriverTypeCreated().WithPayload(driverTypeEntityToModel(t))
}

func (h *Handlers) getDriverType(params driverapi.GetDriverTypeParams, principal interface{}) middleware.Responder {
	defer trace.Trace("getDriverType")()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.getDriverType")
	defer sp.Finish()

	t := DriverType{}
	var err error

	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return driverapi.NewGetDriverTypeBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	err = h.Store.Get(EventManagerFlags.OrgID, params.DriverTypeName, opts, &t)
	if err != nil {
		log.Warnf("Received GET for non-existent driver type %s", params.DriverTypeName)
		log.Debugf("store error when getting driver type: %+v", err)
		return driverapi.NewGetDriverNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String(fmt.Sprintf("driver type %s not found", params.DriverTypeName)),
			})
	}
	return driverapi.NewGetDriverTypeOK().WithPayload(driverTypeEntityToModel(&t))
}

func (h *Handlers) getDriverTypes(params driverapi.GetDriverTypesParams, principal interface{}) middleware.Responder {
	defer trace.Trace("getDriverTypes")()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.getDriverTypes")
	defer sp.Finish()

	var driverTypes []*DriverType
	var err error

	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return driverapi.NewGetDriverTypeBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	// delete filter
	err = h.Store.List(EventManagerFlags.OrgID, opts, &driverTypes)
	if err != nil {
		log.Errorf("store error when listing driver types: %+v", err)
		return driverapi.NewGetDriverTypeDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when getting driver types"),
			})
	}
	var driverTypeModels []*models.DriverType
	for _, driverType := range driverTypes {
		driverTypeModels = append(driverTypeModels, driverTypeEntityToModel(driverType))
	}
	return driverapi.NewGetDriverTypesOK().WithPayload(driverTypeModels)
}

func (h *Handlers) deleteDriverType(params driverapi.DeleteDriverTypeParams, principal interface{}) middleware.Responder {
	defer trace.Tracef("name '%s'", params.DriverTypeName)()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.deleteDriverType")
	defer sp.Finish()

	name := params.DriverTypeName
	var t DriverType
	var err error

	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return driverapi.NewDeleteDriverTypeBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	if err := h.Store.Get(EventManagerFlags.OrgID, name, opts, &t); err != nil {
		log.Errorf("store error when getting driver type: %+v", err)
		return driverapi.NewDeleteDriverTypeNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("driver type not found"),
			})
	}
	if err := h.Store.Delete(EventManagerFlags.OrgID, t.ID, &t); err != nil {
		log.Errorf("store error when deleting the event driver type %s: %+v", t.Name, err)
		return driverapi.NewDeleteDriverTypeInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when deleting an event driver type"),
		})
	}
	return driverapi.NewDeleteDriverTypeOK().WithPayload(driverTypeEntityToModel(&t))
}
