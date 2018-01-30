///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vmware/dispatch/pkg/entity-store"

	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func mockDriverHandler(backend DriverBackend, es entitystore.EntityStore) *driverEntityHandler {
	return &driverEntityHandler{
		store:         es,
		driverBackend: backend,
	}
}

func TestDriverAdd(t *testing.T) {
	backend := &DriverBackendMock{}
	es := helpers.MakeEntityStore(t)
	handler := mockDriverHandler(backend, es)
	driver := &Driver{
		BaseEntity: entitystore.BaseEntity{
			Name:   "driver1",
			Status: entitystore.StatusCREATING,
		},
		Type: "vcenter",
	}
	es.Add(driver)
	backend.On("Deploy", mock.Anything).Return(nil)
	assert.NoError(t, handler.Add(driver))

}

func TestDriverDelete(t *testing.T) {
	backend := &DriverBackendMock{}
	es := helpers.MakeEntityStore(t)
	handler := mockDriverHandler(backend, es)
	driver := &Driver{
		BaseEntity: entitystore.BaseEntity{
			Name:   "driver1",
			Status: entitystore.StatusDELETING,
		},
		Type: "vcenter",
	}
	es.Add(driver)
	backend.On("Delete", mock.Anything).Return(nil)
	assert.NoError(t, handler.Delete(driver))
	var drivers []*Driver
	es.List("", entitystore.Options{}, drivers)
	assert.Len(t, drivers, 0)
}
