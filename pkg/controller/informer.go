///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package controller

import (
	"time"

	log "github.com/sirupsen/logrus"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/trace"
)

// Informer send notifications to downstream event handlers when an event occurs
type Informer interface {
	AddEventHandlers(handlers EventHandlers)
	Run(stopChan <-chan bool)
}

// DefaultInformer implements a basic Informer
type DefaultInformer struct {
	listerWatcher ListerWatcher
	handlers      []EventHandlers
	ResyncPeriod  time.Duration
}

// NewDefaultInformer creates a new DefaultInformer
func NewDefaultInformer(lw ListerWatcher, resyncPeriod time.Duration) Informer {
	return &DefaultInformer{
		listerWatcher: lw,
		handlers:      []EventHandlers{},
		ResyncPeriod:  resyncPeriod,
	}
}

// AddEventHandlers add event handlers
func (d *DefaultInformer) AddEventHandlers(handler EventHandlers) {
	d.handlers = append(d.handlers, handler)
}

func (d *DefaultInformer) processItem(e entitystore.Entity) error {

	var err error
	for _, h := range d.handlers {
		if e.GetStatus() == entitystore.StatusERROR {
			err = h.ErrorFunc(e)
		}
		if e.GetStatus() == entitystore.StatusCREATING {
			err = h.AddFunc(e)
		}
		if e.GetStatus() == entitystore.StatusUPDATING {
			err = h.UpdateFunc(nil, e)
		}
		if e.GetStatus() == entitystore.StatusDELETING {
			err = h.DeleteFunc(e)
		}
		if err != nil {
			log.Print(err)
		}
	}
	return err
}

func (d *DefaultInformer) sync() error {

	entities, err := d.listerWatcher.List(ListOptions{})
	if err != nil {
		// TODO
		log.Print(err)
	}
	for _, e := range entities {
		log.Printf("sync: processing entity %s", e.GetName())
		d.processItem(e)
	}
	return nil
}

// Run starts the a endless loop
func (d *DefaultInformer) Run(stopChan <-chan bool) {

	watcher, err := d.listerWatcher.Watch(ListOptions{})
	if err != nil {
		// TODO
		log.Error(err)
	}

	defer trace.Trace("")()
	for {
		var err error
		select {
		case entity := <-watcher.ResultChan():
			log.Printf("received event=%s entity=%s", entity.GetStatus(), entity.GetName())
			err = d.processItem(entity)
		case <-time.After(d.ResyncPeriod):
			log.Printf("periodic syncing with the underlying driver")
			err = d.sync()
		case <-stopChan:
			return
		}
		if err != nil {
			log.Print(err)
		}
	}
}
