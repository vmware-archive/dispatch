///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package servicemanager

import (
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
	classes, err := handler.Sync("test", time.Duration(1))
	assert.NoError(t, err)
	// First time through the entity is created
	assert.Len(t, classes, 0)
	// Second time through it's found, though not returned because in ready state
	assert.Len(t, classes, 0)
	sc := entities.ServiceClass{}
	found, err := es.Find("test", "test", entitystore.Options{}, &sc)
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

	_, err := es.Add(&ready)
	assert.NoError(t, err)

	client.On("ListServiceClasses").Return([]entitystore.Entity{}, nil).Once()
	classes, err := handler.Sync("test", time.Duration(1))
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
	_, err := es.Add(&class)
	assert.NoError(t, err)

	missingClass := entities.ServiceInstance{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			Name:           "missing",
			Status:         entitystore.StatusINITIALIZED,
		},
		ServiceClass: "class-missing",
	}
	_, err = es.Add(&missingClass)
	assert.NoError(t, err)

	client.On("CreateService", mock.Anything, &missingClass).Return(nil).Once()
	err = handler.Add(&missingClass)
	assert.Error(t, err)

	instance := entities.ServiceInstance{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			Name:           "instance",
			Status:         entitystore.StatusINITIALIZED,
		},
		ServiceClass: "class",
	}
	_, err = es.Add(&instance)
	assert.NoError(t, err)

	client.On("CreateService", mock.Anything, &instance).Return(nil).Once()
	err = handler.Add(&instance)
	assert.NoError(t, err)
}

func TestServiceInstanceDelete(t *testing.T) {

}

func TestServiceInstanceSync(t *testing.T) {

}

func TestServiceBindingSync(t *testing.T) {

}

func TestServiceBindingAdd(t *testing.T) {

}

func TestServiceBindingDelete(t *testing.T) {

}
