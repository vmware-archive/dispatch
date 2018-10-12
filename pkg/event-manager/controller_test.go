///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/vmware/dispatch/pkg/entity-store"
	mocks2 "github.com/vmware/dispatch/pkg/event-manager/drivers/mocks"
	"github.com/vmware/dispatch/pkg/event-manager/subscriptions/entities"
	"github.com/vmware/dispatch/pkg/event-manager/subscriptions/mocks"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
	zkmock "github.com/vmware/dispatch/pkg/zookeeper/mocks"
)

const (
	testZookeeperLocation = "zookeeper.zookeeper.svc.cluster.local"
)

func getTestDriver() *zkmock.Driver {
	driver := &zkmock.Driver{}
	driver.On("CreateNode", mock.Anything, mock.Anything).Return(nil)
	driver.On("GetConnection").Return(nil)
	driver.On("LockEntity", mock.Anything).Return("lock", true, nil)
	driver.On("ReleaseEntity", "lock").Return(nil, nil)
	driver.On("Close").Return(nil)
	return driver
}

func TestControllerRun(t *testing.T) {
	manager := &mocks.Manager{}
	k8sBackend := &mocks2.Backend{}
	es := helpers.MakeEntityStore(t)

	controller := NewEventController(manager, k8sBackend, es, EventControllerConfig{
		ZookeeperLocation: testZookeeperLocation,
		Driver:            getTestDriver(),
	})
	controller.Start()
	controller.Shutdown()
}

func TestControllerRunWithSubs(t *testing.T) {
	manager := &mocks.Manager{}
	k8sBackend := &mocks2.Backend{}
	es := helpers.MakeEntityStore(t)

	controller := NewEventController(manager, k8sBackend, es, EventControllerConfig{
		ZookeeperLocation: testZookeeperLocation,
		Driver:            getTestDriver(),
	})
	defer controller.Shutdown()
	controller.Start()

	sub := &entities.Subscription{
		BaseEntity: entitystore.BaseEntity{
			Name:   "sub1",
			Status: entitystore.StatusCREATING,
		},
		EventType: "test.topic",
		Function:  "test.function",
	}

	es.Add(context.Background(), sub)
}
