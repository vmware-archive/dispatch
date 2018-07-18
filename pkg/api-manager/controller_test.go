///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package apimanager

import (
	"context"
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
	zkmock "github.com/vmware/dispatch/pkg/zookeeper/mocks"
)

const (
	testOrgID             = "dispatch"
	testResyncPeriod      = 500 * time.Millisecond
	testSleepDuration     = 2 * testResyncPeriod
	testZookeeperLocation = "zookeeper.zookeeper.svc.cluster.local"
)

func getTestDriver() *zkmock.Driver {
	driver := &zkmock.Driver{}
	driver.On("CreateNode", mock.Anything, mock.Anything).Return(nil)
	driver.On("GetConnection").Return(nil)
	driver.On("LockEntity", mock.Anything).Return("lock", true)
	driver.On("ReleaseEntity", "lock").Return(nil)
	driver.On("Close").Return(nil)
	return driver
}

func getTestController(t *testing.T, es entitystore.EntityStore, gw gateway.Gateway) (controller.Controller, controller.Watcher) {

	config := &ControllerConfig{
		ResyncPeriod: testResyncPeriod,
		Driver:       getTestDriver(),
	}
	ctrl := NewController(config, es, gw)
	return ctrl, ctrl.Watcher()
}

func TestCtrlAddAPI(t *testing.T) {

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
	mockedGateway.On("AddAPI", mock.Anything, mock.Anything).Return(&testAddAPI.API, nil)
	es := helpers.MakeEntityStore(t)

	ctrl, watcher := getTestController(t, es, mockedGateway)
	ctrl.Start()
	defer ctrl.Shutdown()

	_, err := es.Add(context.Background(), testAddAPI)
	assert.Nil(t, err)
	watcher.OnAction(context.Background(), testAddAPI)

	time.Sleep(testSleepDuration)

	var actual API
	es.Get(context.Background(), testOrgID, testAddAPI.Name, entitystore.Options{}, &actual)
	assert.Equal(t, entitystore.StatusREADY, actual.Status)
}

func TestCtrlUpdateAPI(t *testing.T) {

	testUpdateAPI := &API{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: testOrgID,
			Name:           "testUpdateAPI",
			Status:         entitystore.StatusUPDATING,
		},
		API: gateway.API{
			Name:      "testUpdateAPI",
			Function:  "testAddAPIFunc",
			URIs:      []string{"test.add.api", "test.add.api.com"},
			ID:        "123",
			CreatedAt: 123,
		},
	}

	mockedGateway := &mocks.Gateway{}
	mockedGateway.On("UpdateAPI", mock.Anything, "testUpdateAPI", mock.Anything).Return(&testUpdateAPI.API, nil)
	es := helpers.MakeEntityStore(t)

	ctrl, watcher := getTestController(t, es, mockedGateway)
	ctrl.Start()
	defer ctrl.Shutdown()

	_, err := es.Add(context.Background(), testUpdateAPI)
	assert.Nil(t, err)
	watcher.OnAction(context.Background(), testUpdateAPI)

	time.Sleep(testSleepDuration)

	var actual API
	es.Get(context.Background(), testOrgID, testUpdateAPI.Name, entitystore.Options{}, &actual)
	assert.Equal(t, entitystore.StatusREADY, actual.Status)
}

func TestCtrlAddAPIError(t *testing.T) {

	mockedGateway := &mocks.Gateway{}
	mockedGateway.On("AddAPI", mock.Anything, mock.Anything).Return(nil, errors.New("mocked error"))
	mockedGateway.On("DeleteAPI", mock.Anything, mock.Anything).Return(nil)
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

	_, err := es.Add(context.Background(), testAddAPIReturnErr)
	assert.Nil(t, err)
	watcher.OnAction(context.Background(), testAddAPIReturnErr)

	time.Sleep(testSleepDuration)

	var actual API
	es.Get(context.Background(), testOrgID, testAddAPIReturnErr.Name, entitystore.Options{}, &actual)
	assert.Equal(t, entitystore.StatusERROR, actual.Status)
}

func TestCtrlDeleteAPI(t *testing.T) {

	mockedGateway := &mocks.Gateway{}
	mockedGateway.On("DeleteAPI", mock.Anything, mock.Anything).Return(nil)
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

	_, err := es.Add(context.Background(), testDelAPI)
	assert.Nil(t, err)
	watcher.OnAction(context.Background(), testDelAPI)

	time.Sleep(testSleepDuration)

	var entity API
	err = es.Get(context.Background(), testOrgID, testDelAPI.Name, entitystore.Options{}, &entity)
	assert.NotNil(t, err)
}

func TestCtrlDeleteAPIError(t *testing.T) {

	mockedGateway := &mocks.Gateway{}
	mockedGateway.On("DeleteAPI", mock.Anything, mock.Anything).Return(errors.New("mocked error"))
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

	_, err := es.Add(context.Background(), testDelAPIReturnErr)
	assert.Nil(t, err)
	watcher.OnAction(context.Background(), testDelAPIReturnErr)

	time.Sleep(testSleepDuration)

	var entity API
	err = es.Get(context.Background(), testOrgID, testDelAPIReturnErr.Name, entitystore.Options{}, &entity)
	assert.Nil(t, err)
	assert.Equal(t, testDelAPIReturnErr.Status, entitystore.StatusDELETING)
	assert.Equal(t, testDelAPIReturnErr.Name, entity.Name)
}

func TestCtrlDeleteAPIAsync(t *testing.T) {

	mockedGateway := &mocks.Gateway{}
	mockedGateway.On("DeleteAPI", mock.Anything, mock.Anything).Return(nil)
	mockedGateway.On("GetAPI", mock.Anything, "testDelAPIAsync").Return(nil, nil)
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

	_, err := es.Add(context.Background(), testDelAPIAsync)
	assert.Nil(t, err)

	time.Sleep(testSleepDuration)

	// the api should be deleted by the controller now
	var entity API
	err = es.Get(context.Background(), testOrgID, testDelAPIAsync.Name, entitystore.Options{}, &entity)
	assert.NotNil(t, err)
}
