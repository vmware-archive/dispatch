///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"context"
	"testing"

	"github.com/vmware/dispatch/pkg/entity-store"
	mocks2 "github.com/vmware/dispatch/pkg/event-manager/drivers/mocks"
	"github.com/vmware/dispatch/pkg/event-manager/subscriptions/entities"
	"github.com/vmware/dispatch/pkg/event-manager/subscriptions/mocks"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func TestControllerRun(t *testing.T) {
	manager := &mocks.Manager{}
	backend := &mocks2.Backend{}
	es := helpers.MakeEntityStore(t)

	controller := NewEventController(manager, backend, es, EventControllerConfig{})
	controller.Start()
	controller.Shutdown()
}

func TestControllerRunWithSubs(t *testing.T) {
	manager := &mocks.Manager{}
	backend := &mocks2.Backend{}
	es := helpers.MakeEntityStore(t)

	controller := NewEventController(manager, backend, es, EventControllerConfig{})
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
