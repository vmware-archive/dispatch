///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package servicemanager

import (
	"reflect"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/controller"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/service-manager/clients"
	"github.com/vmware/dispatch/pkg/service-manager/entities"
	"github.com/vmware/dispatch/pkg/service-manager/flags"
	"github.com/vmware/dispatch/pkg/trace"
)

// ControllerConfig defines the image manager controller configuration
type ControllerConfig struct {
	ResyncPeriod   time.Duration
	OrganizationID string
}

type serviceClassEntityHandler struct {
	OrganizationID string
	Store          entitystore.EntityStore
	BrokerClient   clients.BrokerClient
}

// Type returns the type of the entity associated to this handler
func (h *serviceClassEntityHandler) Type() reflect.Type {
	defer trace.Trace("")()
	return reflect.TypeOf(&entities.ServiceClass{})
}

// Add creates new service class entities (will change once users fully manage services)
func (h *serviceClassEntityHandler) Add(obj entitystore.Entity) (err error) {
	defer trace.Trace("")()
	_, err = h.Store.Add(obj)
	return
}

// Update updates service class entities
func (h *serviceClassEntityHandler) Update(obj entitystore.Entity) error {
	defer trace.Trace("")()
	return errors.Errorf("ServiceClass is not updateable")
}

// Delete removes service class entities
func (h *serviceClassEntityHandler) Delete(obj entitystore.Entity) error {
	defer trace.Trace("")()
	var deleted entities.ServiceClass
	err := h.Store.Delete(obj.GetOrganizationID(), obj.GetName(), &deleted)
	if err != nil {
		err = errors.Wrapf(err, "error deleting service class entity %s/%s", obj.GetOrganizationID(), obj.GetName())
		log.Error(err)
		return err
	}
	return nil
}

// Sync reconsiles the actual state from the service catalog with the dispatch state
func (h *serviceClassEntityHandler) Sync(organizationID string, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	defer trace.Trace("")()

	classes, err := h.BrokerClient.ListServiceClasses()
	if err != nil {
		return nil, err
	}
	// actualMap maps serviceIDs (OSBAPI service IDs) to service class entities. These entities represent current state.
	actualMap := make(map[string]*entities.ServiceClass)
	for _, class := range classes {
		sc := class.(*entities.ServiceClass)
		actualMap[sc.ServiceID] = sc
	}

	var existing []*entities.ServiceClass
	err = h.Store.List(h.OrganizationID, entitystore.Options{}, &existing)
	if err != nil {
		return nil, errors.Wrap(err, "Sync error listing exising service classes")
	}
	var synced []entitystore.Entity
	// Update any service classes which have been removed.  This is necessary since we are not directly managing the
	// service classes at this time.  We are simply reflecting the current state.
	for _, class := range existing {
		_, ok := actualMap[class.ServiceID]
		if !ok {
			class.SetDelete(true)
			class.SetStatus(entitystore.StatusDELETING)
		} else {
			delete(actualMap, class.ServiceID)
		}
		synced = append(synced, class)
	}
	// Add any service classes which don't exist in the database
	for _, class := range actualMap {
		class.OrganizationID = h.OrganizationID
		err := h.Add(class)
		if err != nil {
			return nil, err
		}
	}
	return synced, err
}

// Error handles service class entities in the error state
func (h *serviceClassEntityHandler) Error(obj entitystore.Entity) error {
	defer trace.Trace("")()

	_, err := h.Store.Update(obj.GetRevision(), obj)
	return err
}

type serviceInstanceEntityHandler struct {
	Store          entitystore.EntityStore
	BrokerClient   clients.BrokerClient
	OrganizationID string
}

// Type returns the type of the entity associated to this handler
func (h *serviceInstanceEntityHandler) Type() reflect.Type {
	defer trace.Trace("")()

	return reflect.TypeOf(&entities.ServiceInstance{})
}

// Add creates new service instance on the kubernetes service catalog according to the plan
// and parameters configured in the service instance entity.  Additionally, we create a binding
// and secrets which can be used by functions
func (h *serviceInstanceEntityHandler) Add(obj entitystore.Entity) (err error) {
	defer trace.Trace("")()

	si := obj.(*entities.ServiceInstance)

	var sc entities.ServiceClass
	if err = h.Store.Get(si.OrganizationID, si.ServiceClass, entitystore.Options{}, &sc); err != nil {
		return
	}

	defer func() { h.Store.UpdateWithError(si, err) }()

	if err = h.BrokerClient.CreateService(&sc, si); err != nil {
		return
	}
	if si.Bind {
		si.Binding.ServiceInstance = si.Name
		log.Debugf("Adding new service binding %s", si.Name)
		_, err = h.Store.Add(si.Binding)
	}
	return
}

// Update updates service instance entities
func (h *serviceInstanceEntityHandler) Update(obj entitystore.Entity) error {
	defer trace.Trace("")()
	si := obj.(*entities.ServiceInstance)
	_, err := h.Store.Update(si.GetRevision(), si)
	return err
}

// Delete deletes service instance entities
func (h *serviceInstanceEntityHandler) Delete(obj entitystore.Entity) error {
	defer trace.Trace("")()

	si := obj.(*entities.ServiceInstance)

	err := h.BrokerClient.DeleteService(si)
	if err != nil {
		log.Error(err)
	}

	var deleted entities.ServiceInstance
	err = h.Store.Delete(si.GetOrganizationID(), si.GetName(), &deleted)
	if err != nil {
		err = errors.Wrapf(err, "error deleting service instance entity %s/%s", si.GetOrganizationID(), si.GetName())
		log.Error(err)
		return err
	}
	return nil
}

// Sync reconsiles the actual state from the service catalog with the dispatch state
func (h *serviceInstanceEntityHandler) Sync(organizationID string, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	defer trace.Trace("")()

	instances, err := h.BrokerClient.ListServiceInstances()
	if err != nil {
		return nil, err
	}
	actualMap := make(map[string]*entities.ServiceInstance)
	for _, instance := range instances {
		si := instance.(*entities.ServiceInstance)
		actualMap[si.ID] = si
	}

	var existing []*entities.ServiceInstance
	err = h.Store.List(h.OrganizationID, entitystore.Options{}, &existing)
	if err != nil {
		return nil, errors.Wrap(err, "Sync error listing exising service instances")
	}
	var synced []entitystore.Entity
	// Update any service instances which have been removed
	for _, instance := range existing {
		if instance.Status == entitystore.StatusINITIALIZED {
			// Hasn't been created yet, so let's do that.
			synced = append(synced, instance)
			continue
		}
		actual, ok := actualMap[instance.ID]
		if !ok {
			instance.SetDelete(true)
			instance.SetStatus(entitystore.StatusDELETING)
		} else if actual.Status == entitystore.StatusUNKNOWN || actual.Status == instance.Status {
			// If status is unknown or hasn't changed, no need to update
			continue
		}
		instance.SetStatus(actual.Status)
		if instance.Status != entitystore.StatusERROR {
			instance.Reason = nil
		}
		log.Debugf("Syncing instance %s with status %s", instance.Name, instance.Status)
		synced = append(synced, instance)
	}
	return synced, err
}

// Error handles service class entities in the error state
func (h *serviceInstanceEntityHandler) Error(obj entitystore.Entity) error {
	defer trace.Trace("")()

	_, err := h.Store.Update(obj.GetRevision(), obj)
	return err
}

type serviceBindingEntityHandler struct {
	OrganizationID string
	Store          entitystore.EntityStore
	BrokerClient   clients.BrokerClient
}

// Type returns the type of the entity associated to this handler
func (h *serviceBindingEntityHandler) Type() reflect.Type {
	defer trace.Trace("")()

	return reflect.TypeOf(&entities.ServiceBinding{})
}

// Add creates new service class entities (will change once users fully manage services)
func (h *serviceBindingEntityHandler) Add(obj entitystore.Entity) (err error) {
	defer trace.Trace("")()
	b := obj.(*entities.ServiceBinding)

	defer func() { h.Store.UpdateWithError(b, err) }()

	var si entities.ServiceInstance
	log.Debugf("Fetching service for name %s", b.Name)
	if err = h.Store.Get(b.OrganizationID, b.Name, entitystore.Options{}, &si); err != nil {
		return
	}
	log.Debugf("Got service %s", si.Name)
	err = h.BrokerClient.CreateBinding(&si, b)
	return
}

// Update updates service class entities
func (h *serviceBindingEntityHandler) Update(obj entitystore.Entity) error {
	defer trace.Trace("")()
	_, err := h.Store.Update(obj.GetRevision(), obj)
	return err
}

// Delete removes service class entities
func (h *serviceBindingEntityHandler) Delete(obj entitystore.Entity) error {
	defer trace.Trace("")()
	var deleted entities.ServiceClass
	err := h.Store.Delete(obj.GetOrganizationID(), obj.GetName(), &deleted)
	if err != nil {
		err = errors.Wrapf(err, "error deleting service class entity %s/%s", obj.GetOrganizationID(), obj.GetName())
		log.Error(err)
		return err
	}
	return nil
}

// Sync reconsiles the actual state from the service catalog with the dispatch state
func (h *serviceBindingEntityHandler) Sync(organizationID string, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	defer trace.Trace("")()

	bindings, err := h.BrokerClient.ListServiceBindings()
	actualMap := make(map[string]*entities.ServiceBinding)
	for _, binding := range bindings {
		b := binding.(*entities.ServiceBinding)
		actualMap[b.BindingID] = b
	}

	var existing []*entities.ServiceBinding
	err = h.Store.List(h.OrganizationID, entitystore.Options{}, &existing)
	if err != nil {
		return nil, errors.Wrap(err, "Sync error listing exising service bindings")
	}

	var synced []entitystore.Entity

	for _, binding := range existing {
		if binding.Status == entitystore.StatusINITIALIZED {
			// Hasn't been created yet, so let's do that.
			synced = append(synced, binding)
			continue
		}
		actual, ok := actualMap[binding.BindingID]
		// If binding isn't present... delete
		// TODO (bjung): would it be better to set the status to INITIALIZED and recreate?
		if !ok {
			binding.SetDelete(true)
			binding.SetStatus(entitystore.StatusDELETING)
			synced = append(synced, binding)
			continue
		} else if actual.Status == entitystore.StatusUNKNOWN || actual.Status == binding.Status {
			// If status is unknown or hasn't changed, no need to update
			continue
		}
		binding.SetStatus(actual.Status)
		if binding.Status != entitystore.StatusERROR {
			binding.Reason = nil
		}
		synced = append(synced, binding)
	}
	return synced, nil
}

// Error handles service class entities in the error state
func (h *serviceBindingEntityHandler) Error(obj entitystore.Entity) error {
	defer trace.Trace("")()

	_, err := h.Store.Update(obj.GetRevision(), obj)
	return err
}

// NewController creates a new service manager controller
func NewController(config *ControllerConfig, store entitystore.EntityStore, brokerClient clients.BrokerClient) controller.Controller {

	defer trace.Trace("")()

	c := controller.NewController(controller.Options{
		OrganizationID: flags.ServiceManagerFlags.OrgID,
		ResyncPeriod:   config.ResyncPeriod,
		Workers:        10, // want more functions concurrently? add more workers // TODO configure workers
	})

	c.AddEntityHandler(&serviceClassEntityHandler{Store: store, BrokerClient: brokerClient, OrganizationID: flags.ServiceManagerFlags.OrgID})
	c.AddEntityHandler(&serviceInstanceEntityHandler{Store: store, BrokerClient: brokerClient, OrganizationID: flags.ServiceManagerFlags.OrgID})
	c.AddEntityHandler(&serviceBindingEntityHandler{Store: store, BrokerClient: brokerClient, OrganizationID: flags.ServiceManagerFlags.OrgID})
	return c
}
