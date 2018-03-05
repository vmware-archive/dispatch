///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"testing"

	"github.com/vmware/dispatch/pkg/entity-store"
	mocks2 "github.com/vmware/dispatch/pkg/event-manager/drivers/mocks"
	"github.com/vmware/dispatch/pkg/event-manager/subscriptions/entities"
	"github.com/vmware/dispatch/pkg/event-manager/subscriptions/mocks"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func TestControllerRun(t *testing.T) {
	manager := &mocks.Manager{}
	k8sBackend := &mocks2.Backend{}
	es := helpers.MakeEntityStore(t)

	controller := NewEventController(manager, k8sBackend, es, EventControllerConfig{})
	controller.Start()
	controller.Shutdown()
}

func TestControllerRunWithSubs(t *testing.T) {
	manager := &mocks.Manager{}
	k8sBackend := &mocks2.Backend{}
	es := helpers.MakeEntityStore(t)

	controller := NewEventController(manager, k8sBackend, es, EventControllerConfig{})
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

	es.Add(sub)
}
