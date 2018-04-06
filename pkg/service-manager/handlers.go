///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package servicemanager

import (
	"fmt"
	"net/http"

	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/utils"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	log "github.com/sirupsen/logrus"

	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/service-manager/entities"
	"github.com/vmware/dispatch/pkg/service-manager/flags"
	"github.com/vmware/dispatch/pkg/service-manager/gen/models"
	"github.com/vmware/dispatch/pkg/service-manager/gen/restapi/operations"
	serviceclass "github.com/vmware/dispatch/pkg/service-manager/gen/restapi/operations/service_class"
	serviceinstance "github.com/vmware/dispatch/pkg/service-manager/gen/restapi/operations/service_instance"
	"github.com/vmware/dispatch/pkg/trace"
)

// Handlers encapsulates the service manager handlers
type Handlers struct {
	Store   entitystore.EntityStore
	Watcher controller.Watcher
}

// NewHandlers is the constructor for the Handlers type
func NewHandlers(watcher controller.Watcher, store entitystore.EntityStore) *Handlers {
	defer trace.Trace("")()
	return &Handlers{
		Store:   store,
		Watcher: watcher,
	}
}

// ConfigureHandlers registers the service manager handlers to the API
func (h *Handlers) ConfigureHandlers(api middleware.RoutableAPI) {
	defer trace.Trace("")()
	a, ok := api.(*operations.ServiceManagerAPI)
	if !ok {
		panic("Cannot configure api")
	}

	entities.InitializeStatusMap()

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

	a.ServiceClassGetServiceClassByNameHandler = serviceclass.GetServiceClassByNameHandlerFunc(h.getServiceClassByName)
	a.ServiceClassGetServiceClassesHandler = serviceclass.GetServiceClassesHandlerFunc(h.getServiceClasses)

	a.ServiceInstanceAddServiceInstanceHandler = serviceinstance.AddServiceInstanceHandlerFunc(h.addServiceInstance)
	a.ServiceInstanceGetServiceInstanceByNameHandler = serviceinstance.GetServiceInstanceByNameHandlerFunc(h.getServiceInstanceByName)
	a.ServiceInstanceGetServiceInstancesHandler = serviceinstance.GetServiceInstancesHandlerFunc(h.getServiceInstances)
	a.ServiceInstanceDeleteServiceInstanceByNameHandler = serviceinstance.DeleteServiceInstanceByNameHandlerFunc(h.deleteServiceInstanceByName)
}

func (h *Handlers) getServiceClassByName(params serviceclass.GetServiceClassByNameParams, principal interface{}) middleware.Responder {
	defer trace.Trace("")()
	e := entities.ServiceClass{}
	err := h.Store.Get(flags.ServiceManagerFlags.OrgID, params.ServiceClassName, entitystore.Options{}, &e)
	if err != nil {
		log.Warnf("Received GET for non-existent service class %s", params.ServiceClassName)
		log.Debugf("store error when getting service class: %+v", err)
		return serviceclass.NewGetServiceClassByNameNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String(fmt.Sprintf("service class %s not found", params.ServiceClassName)),
			})
	}
	m := entities.ServiceClassEntityToModel(&e)
	return serviceclass.NewGetServiceClassByNameOK().WithPayload(m)
}

func (h *Handlers) getServiceClasses(params serviceclass.GetServiceClassesParams, principal interface{}) middleware.Responder {
	defer trace.Trace("")()
	var classes []*entities.ServiceClass

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}
	err = h.Store.List(flags.ServiceManagerFlags.OrgID, opts, &classes)
	if err != nil {
		log.Errorf("store error when listing service classes: %+v", err)
		return serviceclass.NewGetServiceClassesDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when getting service classes"),
			})
	}
	var classModels []*models.ServiceClass
	for _, class := range classes {
		classModels = append(classModels, entities.ServiceClassEntityToModel(class))
	}
	return serviceclass.NewGetServiceClassesOK().WithPayload(classModels)
}

func (h *Handlers) addServiceInstance(params serviceinstance.AddServiceInstanceParams, principal interface{}) middleware.Responder {
	defer trace.Trace("")()
	serviceRequest := params.Body
	e, b := entities.ServiceInstanceModelToEntity(serviceRequest)
	e.Status = entitystore.StatusINITIALIZED

	var sc entities.ServiceClass
	exists, err := h.Store.Find(e.OrganizationID, e.ServiceClass, entitystore.Options{}, &sc)
	if !exists {
		log.Debugf("service class %s does not exist", e.ServiceClass)
		return serviceinstance.NewAddServiceInstanceBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(fmt.Sprintf("Service class %s does not exist", e.ServiceClass)),
			},
		)
	}
	if err != nil {
		log.Debugf("store error when fetching service broker: %+v", err)
		return serviceinstance.NewAddServiceInstanceDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String(fmt.Sprintf("Error fetching service class %s", e.ServiceClass)),
			},
		)
	}
	// Get Plan and determine if bindable.  Plan "Bindable" is optional and trumps class setting.
	for _, p := range sc.Plans {
		if p.Name == e.ServicePlan && p.Bindable {
			e.Bind = true
			b.Status = entitystore.StatusINITIALIZED
			e.Binding = b
		}
	}
	// TODO (bjung): actually validate the binding/update/add schema against the parameters
	_, err = h.Store.Add(e)
	if err != nil {
		if entitystore.IsUniqueViolation(err) {
			return serviceinstance.NewAddServiceInstanceConflict().WithPayload(&models.Error{
				Code:    http.StatusConflict,
				Message: swag.String("error creating service instance: non-unique name"),
			})
		}
		log.Debugf("store error when adding service instance: %+v", err)
		return serviceinstance.NewAddServiceInstanceBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String("store error when adding service instance"),
			})
	}

	h.Watcher.OnAction(e)

	m := entities.ServiceInstanceEntityToModel(e, b)
	return serviceinstance.NewAddServiceInstanceCreated().WithPayload(m)
}

func (h *Handlers) getServiceInstanceByName(params serviceinstance.GetServiceInstanceByNameParams, principal interface{}) middleware.Responder {
	defer trace.Trace("")()
	si := entities.ServiceInstance{}

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}
	err = h.Store.Get(flags.ServiceManagerFlags.OrgID, params.ServiceInstanceName, opts, &si)
	if err != nil {
		log.Warnf("Received GET for non-existent service instance %s", params.ServiceInstanceName)
		log.Debugf("store error when getting service instance: %+v", err)
		return serviceinstance.NewGetServiceInstanceByNameNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String(fmt.Sprintf("service instance %s not found", params.ServiceInstanceName)),
			})
	}
	b := entities.ServiceBinding{}
	err = h.Store.Get(flags.ServiceManagerFlags.OrgID, params.ServiceInstanceName, opts, &b)
	if err != nil {
		log.Warnf("Received GET for non-existent service binding %s", params.ServiceInstanceName)
		log.Debugf("store error when getting service binding: %+v", err)
		return serviceinstance.NewGetServiceInstanceByNameNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String(fmt.Sprintf("service binding %s not found", params.ServiceInstanceName)),
			})
	}
	m := entities.ServiceInstanceEntityToModel(&si, &b)
	return serviceinstance.NewGetServiceInstanceByNameOK().WithPayload(m)
}

func (h *Handlers) getServiceInstances(params serviceinstance.GetServiceInstancesParams, principal interface{}) middleware.Responder {
	defer trace.Trace("")()
	var services []*entities.ServiceInstance

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return serviceinstance.NewGetServiceInstancesBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}

	err = h.Store.List(flags.ServiceManagerFlags.OrgID, opts, &services)
	if err != nil {
		log.Errorf("store error when listing service instances: %+v", err)
		return serviceinstance.NewGetServiceInstancesDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error while listing service instances"),
			})
	}
	var bindings []*entities.ServiceBinding
	err = h.Store.List(flags.ServiceManagerFlags.OrgID, opts, &bindings)
	if err != nil {
		log.Errorf("store error when listing service bindings: %+v", err)
		return serviceinstance.NewGetServiceInstancesDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error while listing service bindings"),
			})
	}
	bindingsMap := make(map[string]*entities.ServiceBinding)
	for _, binding := range bindings {
		bindingsMap[binding.Name] = binding
	}
	var serviceModels []*models.ServiceInstance
	for _, service := range services {
		binding := bindingsMap[service.Name]
		serviceModels = append(serviceModels, entities.ServiceInstanceEntityToModel(service, binding))
	}
	return serviceinstance.NewGetServiceInstancesOK().WithPayload(serviceModels)
}

func (h *Handlers) deleteServiceInstanceByName(params serviceinstance.DeleteServiceInstanceByNameParams, principal interface{}) middleware.Responder {
	defer trace.Trace("")()
	e := entities.ServiceInstance{}

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}
	err = h.Store.Get(flags.ServiceManagerFlags.OrgID, params.ServiceInstanceName, opts, &e)
	if err != nil {
		return serviceinstance.NewDeleteServiceInstanceByNameNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("service instance not found"),
			})
	}
	err = h.Store.Delete(flags.ServiceManagerFlags.OrgID, params.ServiceInstanceName, &entities.ServiceInstance{})
	if err != nil {
		return serviceinstance.NewDeleteServiceInstanceByNameNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("service instance not found while deleting"),
			})
	}
	e.Delete = true
	e.Status = entitystore.StatusDELETED

	h.Watcher.OnAction(&e)

	m := entities.ServiceInstanceEntityToModel(&e, nil)
	return serviceinstance.NewDeleteServiceInstanceByNameOK().WithPayload(m)
}
