///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package controller

import (
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/trace"
)

// EventHandlers define an interface for handlers of a generic controller
type EventHandlers struct {
	AddFunc    func(obj entitystore.Entity) error
	UpdateFunc func(oldObj, newObj entitystore.Entity) error
	DeleteFunc func(obj entitystore.Entity) error
	ErrorFunc  func(obj entitystore.Entity) error
}

// Controller defines an interface for a generic controller
type Controller interface {
	Run()
	Shutdown()
}

// DefaultController defines a struct for a generic controller
type DefaultController struct {
	done     chan bool
	informer Informer
}

// NewDefaultController creates a new controller
func NewDefaultController(informer Informer) (Controller, error) {
	defer trace.Trace("")()
	return &DefaultController{
		done:     make(chan bool),
		informer: informer,
	}, nil
}

// Run starts the controller watch loop
func (d *DefaultController) Run() {
	defer trace.Trace("")()
	go d.informer.Run(d.done)
}

// Shutdown stops the controller loop
func (d *DefaultController) Shutdown() {
	defer trace.Trace("")()
	d.done <- true
}
