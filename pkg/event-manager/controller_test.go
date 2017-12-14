///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"testing"

	"github.com/vmware/dispatch/pkg/entity-store"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func TestControllerRun(t *testing.T) {
	manager := &SubscriptionManagerMock{}
	k8sBackend := &DriverBackendMock{}
	es := helpers.MakeEntityStore(t)

	controller := NewEventController(manager, k8sBackend, es, EventControllerConfig{})
	controller.Start()
	controller.Shutdown()
}

func TestControllerRunWithSubs(t *testing.T) {
	manager := &SubscriptionManagerMock{}
	k8sBackend := &DriverBackendMock{}
	es := helpers.MakeEntityStore(t)

	controller := NewEventController(manager, k8sBackend, es, EventControllerConfig{})
	defer controller.Shutdown()
	controller.Start()

	sub := &Subscription{
		BaseEntity: entitystore.BaseEntity{
			Name:   "sub1",
			Status: entitystore.StatusCREATING,
		},
		Topic: "test.topic",
		Subscriber: Subscriber{
			Type: FunctionSubscriber,
			Name: "test.function",
		},
	}

	es.Add(sub)
}
