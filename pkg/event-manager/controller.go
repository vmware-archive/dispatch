///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"time"

	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/event-manager/drivers"
	"github.com/vmware/dispatch/pkg/event-manager/subscriptions"
	"github.com/vmware/dispatch/pkg/trace"
)

// Event manager constants
const (
	defaultResyncPeriod = 10 * time.Second
	defaultWorkerNumber = 100
)

// EventControllerConfig defines configuration for controller
type EventControllerConfig struct {
	ResyncPeriod   time.Duration
	OrganizationID string
	WorkerNumber   int
}

// NewEventController creates a new controller to manage the reconciliation of event manager entities
func NewEventController(manager subscriptions.Manager, backend drivers.Backend, store entitystore.EntityStore, config EventControllerConfig) controller.Controller {
	defer trace.Trace("")()

	if config.WorkerNumber == 0 {
		config.WorkerNumber = defaultWorkerNumber
	}

	if config.ResyncPeriod == 0 {
		config.ResyncPeriod = defaultResyncPeriod
	}

	c := controller.NewController(controller.Options{
		OrganizationID: config.OrganizationID,
		ResyncPeriod:   config.ResyncPeriod,
		Workers:        config.WorkerNumber,
	})

	c.AddEntityHandler(drivers.NewEntityHandler(store, backend))
	c.AddEntityHandler(subscriptions.NewEntityHandler(store, manager))

	return c
}
