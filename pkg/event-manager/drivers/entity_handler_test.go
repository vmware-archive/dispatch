///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package drivers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/event-manager/drivers/entities"
	"github.com/vmware/dispatch/pkg/event-manager/drivers/mocks"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func mockDriverHandler(backend Backend, es entitystore.EntityStore) *EntityHandler {
	return &EntityHandler{
		store:   es,
		backend: backend,
	}
}

func TestDriverAdd(t *testing.T) {
	backend := &mocks.Backend{}
	es := helpers.MakeEntityStore(t)
	handler := mockDriverHandler(backend, es)
	driver := &entities.Driver{
		BaseEntity: entitystore.BaseEntity{
			Name:   "driver1",
			Status: entitystore.StatusCREATING,
		},
		Type: "vcenter",
	}
	es.Add(context.Background(), driver)
	backend.On("Deploy", mock.Anything, mock.Anything).Return(nil)
	assert.NoError(t, handler.Add(context.Background(), driver))

	exposed := &entities.Driver{
		BaseEntity: entitystore.BaseEntity{
			Name:   "driver2",
			Status: entitystore.StatusCREATING,
		},
		Type:   "cloudevent",
		Expose: true,
	}
	es.Add(context.Background(), exposed)
	backend.On("Deploy", mock.Anything, mock.Anything).Return(nil)
	assert.NoError(t, handler.Add(context.Background(), exposed))
}

func TestDriverDelete(t *testing.T) {
	backend := &mocks.Backend{}
	es := helpers.MakeEntityStore(t)
	handler := mockDriverHandler(backend, es)
	driver := &entities.Driver{
		BaseEntity: entitystore.BaseEntity{
			Name:           "driver1",
			Status:         entitystore.StatusDELETING,
			OrganizationID: testOrgID,
		},
		Type: "vcenter",
	}
	es.Add(context.Background(), driver)
	backend.On("Delete", mock.Anything, mock.Anything).Return(nil)
	assert.NoError(t, handler.Delete(context.Background(), driver))
	var drivers []*entities.Driver
	es.List(context.Background(), driver.OrganizationID, entitystore.Options{}, drivers)
	assert.Len(t, drivers, 0)
}
