///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"time"

	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/trace"
)

const (
	DefaultResyncPeriod = 60 * time.Second
	DefaultWorkerNumber = 100
)

// EventControllerConfig defines configuration for controller
type EventControllerConfig struct {
	ResyncPeriod   time.Duration
	OrganizationID string
	WorkerNumber   int
}

// NewEventController creates a new controller to manage the reconciliation of event manager entities
func NewEventController(manager SubscriptionManager, backend DriverBackend, store entitystore.EntityStore, config EventControllerConfig) controller.Controller {
	defer trace.Trace("")()

	if config.WorkerNumber == 0 {
		config.WorkerNumber = DefaultWorkerNumber
	}

	if config.ResyncPeriod == 0 {
		config.ResyncPeriod = DefaultResyncPeriod
	}

	c := controller.NewController(store, controller.Options{
		OrganizationID: config.OrganizationID,
		ResyncPeriod:   config.ResyncPeriod,
		Workers:        config.WorkerNumber,
	})

	c.AddEntityHandler(&driverEntityHandler{store: store, driverBackend: backend})
	c.AddEntityHandler(&subscriptionEntityHandler{store: store, manager: manager})

	return c
}
