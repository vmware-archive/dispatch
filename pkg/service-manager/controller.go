///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package servicemanager

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/controller"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/service-manager/clients"
	"github.com/vmware/dispatch/pkg/service-manager/entities"
	"github.com/vmware/dispatch/pkg/trace"
)

var serviceClassOrganizationID = "___global___"

// ControllerConfig defines the image manager controller configuration
type ControllerConfig struct {
	ResyncPeriod      time.Duration
	ZookeeperLocation string
}

type serviceClassEntityHandler struct {
	Store        entitystore.EntityStore
	BrokerClient clients.BrokerClient
}

// Type returns the type of the entity associated to this handler
func (h *serviceClassEntityHandler) Type() reflect.Type {
	return reflect.TypeOf(&entities.ServiceClass{})
}

// Add creates new service class entities (will change once users fully manage services)
func (h *serviceClassEntityHandler) Add(ctx context.Context, obj entitystore.Entity) (err error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	_, err = h.Store.Add(ctx, obj)
	return
}

// Update updates service class entities
func (h *serviceClassEntityHandler) Update(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	sc := obj.(*entities.ServiceClass)
	_, err := h.Store.Update(ctx, sc.Revision, sc)
	return err
}

// Delete removes service class entities
func (h *serviceClassEntityHandler) Delete(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	var deleted entities.ServiceClass
	err := h.Store.Delete(ctx, obj.GetOrganizationID(), obj.GetName(), &deleted)
	if err != nil {
		err = errors.Wrapf(err, "error deleting service class entity %s/%s", obj.GetOrganizationID(), obj.GetName())
		log.Error(err)
		return err
	}
	return nil
}

func (h *serviceClassEntityHandler) needsUpdate(actual *entities.ServiceClass, existing *entities.ServiceClass) (*entities.ServiceClass, bool) {
	if actual.Status == entitystore.StatusUNKNOWN {
		return nil, false
	}
	// Keys are sorted, so encoding should produce a comparable result
	actualJSON, _ := json.Marshal(actual.Plans)
	existingJSON, _ := json.Marshal(existing.Plans)
	if string(actualJSON) != string(existingJSON) ||
		actual.Status != existing.Status ||
		actual.Bindable != existing.Bindable {
		existing.Status = actual.Status
		existing.Bindable = actual.Bindable
		existing.Plans = actual.Plans
		return existing, true
	}
	return nil, false
}

// Sync reconciles the actual state from the service catalog with the dispatch state
func (h *serviceClassEntityHandler) Sync(ctx context.Context, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

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
	err = h.Store.List(ctx, serviceClassOrganizationID, entitystore.Options{}, &existing)
	if err != nil {
		return nil, errors.Wrap(err, "Sync error listing existing service classes")
	}
	var synced []entitystore.Entity
	// Update any service classes which have been removed.  This is necessary since we are not directly managing the
	// service classes at this time.  We are simply reflecting the current state.
	for _, class := range existing {
		actual, ok := actualMap[class.ServiceID]
		if !ok {
			class.SetDelete(true)
			class.SetStatus(entitystore.StatusDELETING)
		} else {
			delete(actualMap, class.ServiceID)
			var ok bool
			class, ok = h.needsUpdate(actual, class)
			if !ok {
				continue
			}
		}
		synced = append(synced, class)
	}
	// Add any service classes which don't exist in the database
	for _, class := range actualMap {
		class.OrganizationID = serviceClassOrganizationID
		err := h.Add(ctx, class)
		if err != nil {
			return nil, err
		}
	}
	return synced, err
}

// Error handles service class entities in the error state
func (h *serviceClassEntityHandler) Error(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	_, err := h.Store.Update(ctx, obj.GetRevision(), obj)
	return err
}

type serviceInstanceEntityHandler struct {
	Store        entitystore.EntityStore
	BrokerClient clients.BrokerClient
}

// Type returns the type of the entity associated to this handler
func (h *serviceInstanceEntityHandler) Type() reflect.Type {
	return reflect.TypeOf(&entities.ServiceInstance{})
}

// Add creates new service instance on the kubernetes service catalog according to the plan
// and parameters configured in the service instance entity.  Additionally, we create a binding
// and secrets which can be used by functions
func (h *serviceInstanceEntityHandler) Add(ctx context.Context, obj entitystore.Entity) (err error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	si := obj.(*entities.ServiceInstance)

	var sc entities.ServiceClass
	if err = h.Store.Get(ctx, serviceClassOrganizationID, si.ServiceClass, entitystore.Options{}, &sc); err != nil {
		return
	}

	defer func() { h.Store.UpdateWithError(ctx, si, err) }()

	if err = h.BrokerClient.CreateService(&sc, si); err != nil {
		return
	}
	return
}

// Update updates service instance entities
func (h *serviceInstanceEntityHandler) Update(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	si := obj.(*entities.ServiceInstance)
	_, err := h.Store.Update(ctx, si.GetRevision(), si)
	return err
}

// Delete deletes service instance entities
func (h *serviceInstanceEntityHandler) Delete(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	si := obj.(*entities.ServiceInstance)

	var b entities.ServiceBinding
	found, err := h.Store.Find(ctx, si.GetOrganizationID(), si.GetID(), entitystore.Options{}, &b)
	if found {
		log.Debugf("waiting to delete service instance %s, binding still exists")
		return nil
	}

	err = h.BrokerClient.DeleteService(si)
	if err != nil {
		log.Error(err)
	}

	// TODO (bjung): We really shoudn't actually delete the entity until the the resource
	// is actually deleted.  As-is it works, but we are repeatedly calling delete as the controller
	// thinks the resource has been orphaned (which it has)
	var deleted entities.ServiceInstance
	err = h.Store.Delete(ctx, si.GetOrganizationID(), si.GetName(), &deleted)
	if err != nil {
		err = errors.Wrapf(err, "error deleting service instance entity %s/%s", si.GetOrganizationID(), si.GetName())
		log.Error(err)
		return err
	}
	return nil
}

// Sync reconsiles the actual state from the service catalog with the dispatch state
func (h *serviceInstanceEntityHandler) Sync(ctx context.Context, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

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
	err = h.Store.ListGlobal(ctx, entitystore.Options{}, &existing)
	if err != nil {
		return nil, errors.Wrap(err, "Sync error listing existing service instances")
	}
	var synced []entitystore.Entity
	// Update any service instances which have been removed
	for _, instance := range existing {
		log.Debugf("Processing service instance %s [%s]", instance.Name, instance.Status)
		if instance.Status == entitystore.StatusINITIALIZED {
			// Hasn't been created yet, so let's do that.
			synced = append(synced, instance)
			continue
		}
		if instance.Delete {
			// Marked for deletion... ignore actual status - though we need to start tracking
			// actual state separately from desired stated (i.e. marked for delete, but is currently
			// in ready state)
			delete(actualMap, instance.ID)
			synced = append(synced, instance)
			continue
		}
		actual, ok := actualMap[instance.ID]
		if !ok {
			instance.SetDelete(true)
			instance.SetStatus(entitystore.StatusDELETING)
			log.Debugf("Setting service instance %s for deletion", instance.Name)
			synced = append(synced, instance)
			continue
		} else {
			delete(actualMap, instance.ID)
		}
		if actual.Status == entitystore.StatusUNKNOWN || actual.Status == instance.Status {
			// If status is unknown or hasn't changed, no need to update
			continue
		}
		instance.SetStatus(actual.Status)
		if instance.Status != entitystore.StatusERROR {
			instance.Reason = nil
		} else {
			instance.SetReason(actual.Reason)
		}
		log.Debugf("Syncing instance %s with status %s", instance.Name, instance.Status)
		synced = append(synced, instance)
	}
	// Clean up any orphaned bindings
	for _, s := range actualMap {
		s.SetDelete(true)
		s.SetStatus(entitystore.StatusDELETING)
		synced = append(synced, s)
	}
	return synced, err
}

// Error handles service class entities in the error state
func (h *serviceInstanceEntityHandler) Error(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	_, err := h.Store.Update(ctx, obj.GetRevision(), obj)
	return err
}

type serviceBindingEntityHandler struct {
	Store        entitystore.EntityStore
	BrokerClient clients.BrokerClient
}

// Type returns the type of the entity associated to this handler
func (h *serviceBindingEntityHandler) Type() reflect.Type {
	return reflect.TypeOf(&entities.ServiceBinding{})
}

// Add creates new service class entities (will change once users fully manage services)
func (h *serviceBindingEntityHandler) Add(ctx context.Context, obj entitystore.Entity) (err error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	b := obj.(*entities.ServiceBinding)

	var si entities.ServiceInstance
	log.Debugf("Fetching service for name %s", b.ServiceInstance)
	if err = h.Store.Get(ctx, b.OrganizationID, b.ServiceInstance, entitystore.Options{}, &si); err != nil {
		return
	}
	if si.Status != entitystore.StatusREADY {
		log.Debugf("Service %s not ready for binding %s", si.Name, si.Status)
		return
	}
	defer func() { h.Store.UpdateWithError(ctx, b, err) }()

	err = h.BrokerClient.CreateBinding(&si, b)
	return
}

// Update updates service class entities
func (h *serviceBindingEntityHandler) Update(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	_, err := h.Store.Update(ctx, obj.GetRevision(), obj)
	return err
}

// Delete removes service binding entities
func (h *serviceBindingEntityHandler) Delete(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	b := obj.(*entities.ServiceBinding)

	log.Debugf("Deleting service binding %s", b.Name)
	err := h.BrokerClient.DeleteBinding(b)
	if err != nil {
		log.Error(err)
		return err
	}

	// TODO (bjung): We really shoudn't actually delete the entity until the the resource
	// is actually deleted.  As-is it works, but we are repeatedly calling delete as the controller
	// thinks the resource has been orphaned (which it has)
	var deleted entities.ServiceBinding
	err = h.Store.Delete(ctx, obj.GetOrganizationID(), obj.GetName(), &deleted)
	if err != nil {
		err = errors.Wrapf(err, "error deleting service binding entity %s [%s]", b.Name, b.BindingID)
		log.Error(err)
		return err
	}
	return nil
}

// Sync reconciles the actual state from the service catalog with the dispatch state
func (h *serviceBindingEntityHandler) Sync(ctx context.Context, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	bindings, err := h.BrokerClient.ListServiceBindings()
	actualMap := make(map[string]*entities.ServiceBinding)
	for _, binding := range bindings {
		b := binding.(*entities.ServiceBinding)
		actualMap[b.BindingID] = b
	}

	var existing []*entities.ServiceBinding
	err = h.Store.ListGlobal(ctx, entitystore.Options{}, &existing)
	if err != nil {
		return nil, errors.Wrap(err, "Sync error listing existing service bindings")
	}

	var existingServices []*entities.ServiceInstance
	err = h.Store.ListGlobal(ctx, entitystore.Options{}, &existingServices)
	if err != nil {
		return nil, errors.Wrap(err, "Sync error listing existing services")
	}
	serviceMap := make(map[string]*entities.ServiceInstance)
	for _, service := range existingServices {
		serviceMap[service.Name] = service
	}

	var synced []entitystore.Entity

	for _, binding := range existing {
		log.Debugf("Processing service binding %s [%s]", binding.Name, binding.Status)
		if _, ok := serviceMap[binding.ServiceInstance]; !ok {
			log.Debugf("Service for binding %s missing, delete", binding.Name)
			// No matching service exists... delete
			binding.SetDelete(true)
			binding.SetStatus(entitystore.StatusDELETING)
			synced = append(synced, binding)
			continue
		}
		if binding.Status == entitystore.StatusINITIALIZED {
			// Hasn't been created yet, so let's do that.
			synced = append(synced, binding)
			continue
		}
		if binding.Delete {
			// Marked for deletion... ignore actual status - though we need to start tracking
			// actual state separately from desired stated (i.e. marked for delete, but is currently
			// in ready state)
			delete(actualMap, binding.BindingID)
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
		} else {
			delete(actualMap, binding.BindingID)
		}
		if actual.Status == entitystore.StatusUNKNOWN || actual.Status == binding.Status {
			// If status is unknown or hasn't changed, no need to update
			continue
		}
		binding.SetStatus(actual.Status)
		if binding.Status != entitystore.StatusERROR {
			binding.Reason = nil
		} else {
			binding.SetReason(actual.Reason)
		}
		synced = append(synced, binding)
	}
	// Clean up any orphaned bindings
	for _, b := range actualMap {
		b.SetDelete(true)
		b.SetStatus(entitystore.StatusDELETING)
		synced = append(synced, b)
	}
	return synced, nil
}

// Error handles service class entities in the error state
func (h *serviceBindingEntityHandler) Error(ctx context.Context, obj entitystore.Entity) error {
	_, err := h.Store.Update(ctx, obj.GetRevision(), obj)
	return err
}

// NewController creates a new service manager controller
func NewController(config *ControllerConfig, store entitystore.EntityStore, brokerClient clients.BrokerClient) controller.Controller {
	log.Debugf("Configuration for service manager: %v", config)
	c := controller.NewController(controller.Options{
		ResyncPeriod:      config.ResyncPeriod,
		Workers:           10, // want more functions concurrently? add more workers // TODO configure workers
		ZookeeperLocation: config.ZookeeperLocation,
	})

	c.AddEntityHandler(&serviceClassEntityHandler{Store: store, BrokerClient: brokerClient})
	c.AddEntityHandler(&serviceInstanceEntityHandler{Store: store, BrokerClient: brokerClient})
	c.AddEntityHandler(&serviceBindingEntityHandler{Store: store, BrokerClient: brokerClient})
	return c
}
