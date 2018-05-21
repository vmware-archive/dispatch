///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package servicemanager

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/vmware/dispatch/pkg/entity-store"

	"github.com/vmware/dispatch/pkg/service-manager/entities"
	"github.com/vmware/dispatch/pkg/service-manager/mocks"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func TestServiceClassSyncReady(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	client := &mocks.BrokerClient{}

	handler := serviceClassEntityHandler{
		OrganizationID: "test",
		Store:          es,
		BrokerClient:   client,
	}

	ready := entities.ServiceClass{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			Name:           "test",
			Status:         entitystore.StatusREADY,
		},
		ServiceID: "deadbeef",
	}

	client.On("ListServiceClasses").Return([]entitystore.Entity{&ready}, nil).Times(2)
	classes, err := handler.Sync(context.Background(), time.Duration(1))
	assert.NoError(t, err)
	// First time through the entity is created
	assert.Len(t, classes, 0)
	// Second time through it's found, though not returned because in ready state
	assert.Len(t, classes, 0)
	sc := entities.ServiceClass{}
	found, err := es.Find(context.Background(), "test", "test", entitystore.Options{}, &sc)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, entitystore.StatusREADY, sc.Status)
}

func TestServiceClassSyncRemoved(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	client := &mocks.BrokerClient{}

	handler := serviceClassEntityHandler{
		OrganizationID: "test",
		Store:          es,
		BrokerClient:   client,
	}

	ready := entities.ServiceClass{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			Name:           "test",
			Status:         entitystore.StatusREADY,
		},
		ServiceID: "deadbeef",
	}

	_, err := es.Add(context.Background(), &ready)
	assert.NoError(t, err)

	client.On("ListServiceClasses").Return([]entitystore.Entity{}, nil).Once()
	classes, err := handler.Sync(context.Background(), time.Duration(1))
	assert.NoError(t, err)
	// First time through the entity is marked for deletion
	assert.Len(t, classes, 1)
	assert.Equal(t, entitystore.StatusDELETING, classes[0].GetStatus())
}

func TestServiceInstanceAdd(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	client := &mocks.BrokerClient{}

	handler := serviceInstanceEntityHandler{
		OrganizationID: "test",
		Store:          es,
		BrokerClient:   client,
	}

	class := entities.ServiceClass{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			Name:           "class",
			Status:         entitystore.StatusREADY,
		},
		ServiceID: "deadbeef",
	}
	_, err := es.Add(context.Background(), &class)
	assert.NoError(t, err)

	missingClass := entities.ServiceInstance{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			Name:           "missing",
			Status:         entitystore.StatusINITIALIZED,
		},
		ServiceClass: "class-missing",
	}
	_, err = es.Add(context.Background(), &missingClass)
	assert.NoError(t, err)

	client.On("CreateService", mock.Anything, &missingClass).Return(nil).Once()
	err = handler.Add(context.Background(), &missingClass)
	assert.Error(t, err)

	instance := entities.ServiceInstance{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			Name:           "instance",
			Status:         entitystore.StatusINITIALIZED,
		},
		ServiceClass: "class",
	}
	_, err = es.Add(context.Background(), &instance)
	assert.NoError(t, err)

	client.On("CreateService", mock.Anything, &instance).Return(nil).Once()
	err = handler.Add(context.Background(), &instance)
	assert.NoError(t, err)
}

func TestServiceInstanceDelete(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	client := &mocks.BrokerClient{}

	handler := serviceInstanceEntityHandler{
		OrganizationID: "test",
		Store:          es,
		BrokerClient:   client,
	}

	instance := entities.ServiceInstance{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			Name:           "instance",
			Status:         entitystore.StatusINITIALIZED,
		},
		ServiceClass: "class",
	}
	_, err := es.Add(context.Background(), &instance)
	assert.NoError(t, err)

	i := entities.ServiceInstance{}
	found, err := es.Find(context.Background(), instance.OrganizationID, instance.Name, entitystore.Options{}, &i)
	assert.NoError(t, err)
	assert.True(t, found)

	client.On("DeleteService", &instance).Return(nil).Once()
	err = handler.Delete(context.Background(), &instance)
	assert.NoError(t, err)

	i = entities.ServiceInstance{}
	found, err = es.Find(context.Background(), instance.OrganizationID, instance.Name, entitystore.Options{}, &i)
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestServiceInstanceSyncInitialized(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	client := &mocks.BrokerClient{}

	handler := serviceInstanceEntityHandler{
		OrganizationID: "test",
		Store:          es,
		BrokerClient:   client,
	}

	initialized := entities.ServiceInstance{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			Name:           "instance",
			Status:         entitystore.StatusINITIALIZED,
		},
		ServiceClass: "class",
	}
	_, err := es.Add(context.Background(), &initialized)
	assert.NoError(t, err)

	// Return no actual entities, but one waiting to be created
	client.On("ListServiceInstances").Return([]entitystore.Entity{}, nil).Once()
	instances, err := handler.Sync(context.Background(), time.Duration(1))
	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, entitystore.StatusINITIALIZED, instances[0].GetStatus())
}

func TestServiceInstanceSyncReady(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	client := &mocks.BrokerClient{}

	handler := serviceInstanceEntityHandler{
		OrganizationID: "test",
		Store:          es,
		BrokerClient:   client,
	}

	ready := entities.ServiceInstance{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			Name:           "instance",
			Status:         entitystore.StatusREADY,
		},
		ServiceClass: "class",
	}
	_, err := es.Add(context.Background(), &ready)
	assert.NoError(t, err)

	// Return no actual entities, but one ready... will be marked for deletion
	client.On("ListServiceInstances").Return([]entitystore.Entity{}, nil).Once()
	instances, err := handler.Sync(context.Background(), time.Duration(1))
	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, entitystore.StatusDELETING, instances[0].GetStatus())

	// Return a ready entity, status is equal... nothing returned
	client.On("ListServiceInstances").Return([]entitystore.Entity{&ready}, nil).Once()
	instances, err = handler.Sync(context.Background(), time.Duration(1))
	assert.NoError(t, err)
	assert.Len(t, instances, 0)

	// Remove the entity, creating an orphan
	err = es.Delete(context.Background(), ready.OrganizationID, ready.Name, &ready)
	assert.NoError(t, err)
	// Return a ready, but orphaned entity
	orphan := entities.ServiceInstance{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			Name:           "orphan",
			Status:         entitystore.StatusREADY,
		},
		ServiceClass: "class",
	}
	client.On("ListServiceInstances").Return([]entitystore.Entity{&orphan}, nil).Once()
	instances, err = handler.Sync(context.Background(), time.Duration(1))
	assert.Len(t, instances, 1)
	assert.Equal(t, entitystore.StatusDELETING, instances[0].GetStatus())
	assert.Equal(t, orphan.Name, instances[0].GetName())
}

func TestServiceInstanceSyncDeleting(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	client := &mocks.BrokerClient{}

	handler := serviceInstanceEntityHandler{
		OrganizationID: "test",
		Store:          es,
		BrokerClient:   client,
	}

	deleting := entities.ServiceInstance{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			Name:           "instance-stored",
			Status:         entitystore.StatusDELETING,
			Delete:         true,
		},
		ServiceClass: "class",
	}
	id, err := es.Add(context.Background(), &deleting)
	assert.NoError(t, err)

	// Return a ready entity, will be marked marked for deletion
	ready := entities.ServiceInstance{
		BaseEntity: entitystore.BaseEntity{
			ID:             id,
			OrganizationID: "test",
			Name:           "instance-actual",
			Status:         entitystore.StatusREADY,
		},
		ServiceClass: "class",
	}
	client.On("ListServiceInstances").Return([]entitystore.Entity{&ready}, nil).Once()
	instances, err := handler.Sync(context.Background(), time.Duration(1))
	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, entitystore.StatusDELETING, instances[0].GetStatus())
	assert.True(t, instances[0].GetDelete())
}

func TestServiceBindingAdd(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	client := &mocks.BrokerClient{}

	handler := serviceBindingEntityHandler{
		OrganizationID: "test",
		Store:          es,
		BrokerClient:   client,
	}

	initializedService := entities.ServiceInstance{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			Name:           "instance",
			Status:         entitystore.StatusINITIALIZED,
		},
		ServiceClass: "class",
	}
	_, err := es.Add(context.Background(), &initializedService)
	assert.NoError(t, err)
	initializedBinding := entities.ServiceBinding{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			Name:           "binding",
			Status:         entitystore.StatusINITIALIZED,
		},
		ServiceInstance: "instance",
	}
	_, err = es.Add(context.Background(), &initializedBinding)
	assert.NoError(t, err)

	// binding not called as the service is not ready
	err = handler.Add(context.Background(), &initializedBinding)
	assert.NoError(t, err)

	initializedService.SetStatus(entitystore.StatusREADY)
	_, err = es.Update(context.Background(), initializedService.Revision, &initializedService)
	assert.NoError(t, err)
	// binding called this time
	client.On("CreateBinding", mock.AnythingOfType("*entities.ServiceInstance"), &initializedBinding).Return(nil).Once()
	err = handler.Add(context.Background(), &initializedBinding)
	assert.NoError(t, err)
}

func TestServiceBindingDelete(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	client := &mocks.BrokerClient{}

	handler := serviceBindingEntityHandler{
		OrganizationID: "test",
		Store:          es,
		BrokerClient:   client,
	}

	readyBinding := entities.ServiceBinding{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			Name:           "instance",
			Status:         entitystore.StatusREADY,
		},
		ServiceInstance: "instance",
	}
	_, err := es.Add(context.Background(), &readyBinding)
	assert.NoError(t, err)
	client.On("DeleteBinding", &readyBinding).Return(nil).Once()

	err = handler.Delete(context.Background(), &readyBinding)
	assert.NoError(t, err)

	b := entities.ServiceBinding{}
	found, err := es.Find(context.Background(), readyBinding.OrganizationID, readyBinding.Name, entitystore.Options{}, &b)
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestServiceBindingSyncInitialized(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	client := &mocks.BrokerClient{}

	handler := serviceBindingEntityHandler{
		OrganizationID: "test",
		Store:          es,
		BrokerClient:   client,
	}

	initializedService := entities.ServiceInstance{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			Name:           "instance",
			Status:         entitystore.StatusINITIALIZED,
		},
		ServiceClass: "class",
	}
	_, err := es.Add(context.Background(), &initializedService)
	assert.NoError(t, err)
	initializedBinding := entities.ServiceBinding{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			Name:           "binding",
			Status:         entitystore.StatusINITIALIZED,
		},
		ServiceInstance: "instance",
	}
	_, err = es.Add(context.Background(), &initializedBinding)
	assert.NoError(t, err)

	client.On("ListServiceBindings").Return([]entitystore.Entity{}, nil).Once()
	bindings, err := handler.Sync(context.Background(), time.Duration(1))
	assert.NoError(t, err)
	assert.Len(t, bindings, 1)
	assert.Equal(t, entitystore.StatusINITIALIZED, bindings[0].GetStatus())
}

func TestServiceBindingSyncMissingService(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	client := &mocks.BrokerClient{}

	handler := serviceBindingEntityHandler{
		OrganizationID: "test",
		Store:          es,
		BrokerClient:   client,
	}

	initializedBinding := entities.ServiceBinding{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			Name:           "binding",
			Status:         entitystore.StatusINITIALIZED,
		},
		ServiceInstance: "instance",
	}
	_, err := es.Add(context.Background(), &initializedBinding)
	assert.NoError(t, err)

	client.On("ListServiceBindings").Return([]entitystore.Entity{}, nil).Once()
	// Delete bindings which are missing services
	bindings, err := handler.Sync(context.Background(), time.Duration(1))
	assert.NoError(t, err)
	assert.Len(t, bindings, 1)
	assert.True(t, bindings[0].GetDelete())
	assert.Equal(t, entitystore.StatusDELETING, bindings[0].GetStatus())
}

func TestServiceBindingSyncReady(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	client := &mocks.BrokerClient{}

	handler := serviceBindingEntityHandler{
		OrganizationID: "test",
		Store:          es,
		BrokerClient:   client,
	}

	readyService := entities.ServiceInstance{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			Name:           "instance",
			Status:         entitystore.StatusREADY,
		},
		ServiceClass: "class",
	}
	_, err := es.Add(context.Background(), &readyService)
	assert.NoError(t, err)
	readyBinding := entities.ServiceBinding{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			Name:           "binding",
			Status:         entitystore.StatusREADY,
		},
		ServiceInstance: "instance",
	}
	_, err = es.Add(context.Background(), &readyBinding)
	assert.NoError(t, err)

	client.On("ListServiceBindings").Return([]entitystore.Entity{&readyBinding}, nil).Once()
	// Binding status matches actual... return nothing
	bindings, err := handler.Sync(context.Background(), time.Duration(1))
	assert.NoError(t, err)
	assert.Len(t, bindings, 0)

	client.On("ListServiceBindings").Return([]entitystore.Entity{}, nil).Once()
	// Actual binding missing... delete binding entity (should we recreate?)
	bindings, err = handler.Sync(context.Background(), time.Duration(1))
	assert.NoError(t, err)
	assert.Len(t, bindings, 1)
	assert.True(t, bindings[0].GetDelete())
	assert.Equal(t, entitystore.StatusDELETING, bindings[0].GetStatus())

	// Remove the entity, creating an orphan
	err = es.Delete(context.Background(), readyBinding.OrganizationID, readyBinding.Name, &readyBinding)
	assert.NoError(t, err)
	// Return a ready, but orphaned entity
	orphan := entities.ServiceBinding{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			Name:           "orphan",
			Status:         entitystore.StatusREADY,
		},
		ServiceInstance: "instance",
	}
	client.On("ListServiceBindings").Return([]entitystore.Entity{&orphan}, nil).Once()
	bindings, err = handler.Sync(context.Background(), time.Duration(1))
	assert.Len(t, bindings, 1)
	assert.Equal(t, entitystore.StatusDELETING, bindings[0].GetStatus())
	assert.Equal(t, orphan.Name, bindings[0].GetName())
}
