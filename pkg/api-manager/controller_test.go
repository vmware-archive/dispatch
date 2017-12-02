///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package apimanager

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vmware/dispatch/pkg/api-manager/gateway"
	"github.com/vmware/dispatch/pkg/api-manager/gateway/mocks"
	"github.com/vmware/dispatch/pkg/controller"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

const (
	testOrgID         = "testAPIManagerOrg"
	testResyncPeriod  = 2 * time.Second
	testSleepDuration = 2 * testResyncPeriod
)

func getTestController(t *testing.T, es entitystore.EntityStore, gw gateway.Gateway) (controller.Controller, controller.Watcher) {

	config := &ControllerConfig{
		ResyncPeriod:   testResyncPeriod,
		OrganizationID: testOrgID,
	}
	ctrl := NewController(config, es, gw)
	return ctrl, ctrl.Watcher()
}

func TestCtrlUpdateAPI(t *testing.T) {

	testAddAPI := &API{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: testOrgID,
			Name:           "testAddAPI",
			Status:         entitystore.StatusCREATING,
		},
		API: gateway.API{
			Name:      "testAddAPI",
			Function:  "testAddAPIFunc",
			URIs:      []string{"test.add.api", "test.add.api.com"},
			ID:        "123",
			CreatedAt: 123,
		},
	}

	mockedGateway := &mocks.Gateway{}
	mockedGateway.On("UpdateAPI", "testAddAPI", mock.Anything).Return(&testAddAPI.API, nil)
	es := helpers.MakeEntityStore(t)

	ctrl, watcher := getTestController(t, es, mockedGateway)
	ctrl.Start()
	defer ctrl.Shutdown()

	_, err := es.Add(testAddAPI)
	assert.Nil(t, err)
	watcher.OnAction(testAddAPI)

	time.Sleep(testSleepDuration)

	var actual API
	es.Get(testOrgID, testAddAPI.Name, &actual)
	assert.Equal(t, entitystore.StatusREADY, actual.Status)
}

func TestCtrlUpdateAPIError(t *testing.T) {

	mockedGateway := &mocks.Gateway{}
	mockedGateway.On("UpdateAPI", "testAddAPIReturnErr", mock.Anything).Return(nil, errors.New("mocked error"))
	es := helpers.MakeEntityStore(t)

	ctrl, watcher := getTestController(t, es, mockedGateway)
	ctrl.Start()
	defer ctrl.Shutdown()

	testAddAPIReturnErr := &API{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: testOrgID,
			Name:           "testAddAPIReturnErr",
			Status:         entitystore.StatusCREATING,
		},
		API: gateway.API{
			Name:     "testAddAPIReturnErr",
			Function: "testAddAPIFunc",
			URIs:     []string{"test.add.api", "test.add.api.com"},
		},
	}

	_, err := es.Add(testAddAPIReturnErr)
	assert.Nil(t, err)
	watcher.OnAction(testAddAPIReturnErr)

	time.Sleep(testSleepDuration)

	var actual API
	es.Get(testOrgID, testAddAPIReturnErr.Name, &actual)
	assert.Equal(t, entitystore.StatusERROR, actual.Status)
}

func TestCtrlDeleteAPI(t *testing.T) {

	mockedGateway := &mocks.Gateway{}
	// mockedGateway.On("DeleteAPI", "testDelAPI", mock.Anything).Return(nil)
	mockedGateway.On("DeleteAPI", mock.Anything).Return(nil)
	es := helpers.MakeEntityStore(t)

	ctrl, watcher := getTestController(t, es, mockedGateway)
	ctrl.Start()
	defer ctrl.Shutdown()

	testDelAPI := &API{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: testOrgID,
			Name:           "testDelAPI",
			Status:         entitystore.StatusDELETING,
		},
		API: gateway.API{
			Name:     "testDelAPI",
			Function: "testDelAPIFunc",
			URIs:     []string{"test.del.api", "test.del.api.com"},
		},
	}

	_, err := es.Add(testDelAPI)
	assert.Nil(t, err)
	watcher.OnAction(testDelAPI)

	time.Sleep(testSleepDuration)

	var entity API
	err = es.Get(testOrgID, testDelAPI.Name, &entity)
	assert.NotNil(t, err)
}

func TestCtrlDeleteAPIError(t *testing.T) {

	mockedGateway := &mocks.Gateway{}
	// mockedGateway.On("DeleteAPI", "testDelAPIReturnErr", mock.Anything).Return(errors.New("mocked error"))
	mockedGateway.On("DeleteAPI", mock.Anything).Return(errors.New("mocked error"))
	es := helpers.MakeEntityStore(t)

	ctrl, watcher := getTestController(t, es, mockedGateway)
	ctrl.Start()
	defer ctrl.Shutdown()

	testDelAPIReturnErr := &API{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: testOrgID,
			Name:           "testDelAPIReturnErr",
			Status:         entitystore.StatusDELETING,
		},
		API: gateway.API{
			Name:     "testDelAPIReturnErr",
			Function: "testDelAPIReturnErrFunc",
			URIs:     []string{"test.del.api", "test.del.api.com"},
		},
	}

	_, err := es.Add(testDelAPIReturnErr)
	assert.Nil(t, err)
	watcher.OnAction(testDelAPIReturnErr)

	time.Sleep(testSleepDuration)

	var entity API
	err = es.Get(testOrgID, testDelAPIReturnErr.Name, &entity)
	assert.Nil(t, err)
	assert.Equal(t, testDelAPIReturnErr.Status, entitystore.StatusDELETING)
	assert.Equal(t, testDelAPIReturnErr.Name, entity.Name)
}

func TestCtrlDeleteAPIAsync(t *testing.T) {

	mockedGateway := &mocks.Gateway{}
	// mockedGateway.On("DeleteAPI", "testDelAPIAsync", mock.Anything).Return(nil)
	mockedGateway.On("DeleteAPI", mock.Anything).Return(nil)
	mockedGateway.On("GetAPI", "testDelAPIAsync").Return(nil, nil)
	es := helpers.MakeEntityStore(t)

	ctrl, _ := getTestController(t, es, mockedGateway)
	ctrl.Start()
	defer ctrl.Shutdown()

	testDelAPIAsync := &API{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: testOrgID,
			Name:           "testDelAPIAsync",
			Status:         entitystore.StatusDELETING,
		},
		API: gateway.API{
			Name:     "testDelAPIAsync",
			Function: "testDelAPIAsyncFunc",
			URIs:     []string{"test.del.api", "test.del.api.com"},
		},
	}

	_, err := es.Add(testDelAPIAsync)
	assert.Nil(t, err)

	time.Sleep(testSleepDuration)

	// the api should be deleted by the controller now
	var entity API
	err = es.Get(testOrgID, testDelAPIAsync.Name, &entity)
	assert.NotNil(t, err)
}
