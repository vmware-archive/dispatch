///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package controller

import (
	"reflect"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/trace"
)

// EntityHandler define an interface for entity operations of a generic controller
type EntityHandler interface {
	Type() reflect.Type
	Add(obj entitystore.Entity) error
	Update(obj entitystore.Entity) error
	Delete(obj entitystore.Entity) error
	Error(obj entitystore.Entity) error
}

const defaultWorkers = 1

// Options defines controller configuration
type Options struct {
	OrganizationID string
	Namespace      string

	ResyncPeriod time.Duration
	Workers      int
}

type Watcher chan<- entitystore.Entity

func (w *Watcher) OnAction(e entitystore.Entity) {
	defer trace.Trace("")()

	if w == nil || *w == nil {
		log.Warnf("nil watcher, skipping entity update: %s - %s", e.GetName(), e.GetStatus())
		return
	}
	*w <- e
}

// Controller defines an interface for a generic controller
type Controller interface {
	Start()
	Shutdown()

	Watcher() Watcher

	AddEntityHandler(h EntityHandler)
}

// DefaultController defines a struct for a generic controller
type DefaultController struct {
	done    chan bool
	watcher chan entitystore.Entity
	store   entitystore.EntityStore
	options Options

	entityHandlers map[reflect.Type]EntityHandler
}

// NewController creates a new controller
func NewController(store entitystore.EntityStore, options Options) Controller {
	defer trace.Trace("")()

	if options.Workers == 0 {
		options.Workers = defaultWorkers
	}

	return &DefaultController{
		done:    make(chan bool),
		watcher: make(chan entitystore.Entity),
		store:   store,
		options: options,

		entityHandlers: map[reflect.Type]EntityHandler{},
	}
}

// Start starts the controller watch loop
func (dc *DefaultController) Start() {
	defer trace.Trace("")()

	go dc.run(dc.done)
}

// Shutdown stops the controller loop
func (dc *DefaultController) Shutdown() {
	defer trace.Trace("")()

	dc.done <- true
}

func (dc *DefaultController) Watcher() Watcher {
	defer trace.Trace("")()

	return dc.watcher
}

// AddEntityHandler adds entity handlers
func (dc *DefaultController) AddEntityHandler(h EntityHandler) {
	defer trace.Trace("")()

	dc.entityHandlers[h.Type()] = h
}

func (dc *DefaultController) processItem(e entitystore.Entity) error {
	defer trace.Trace("")()

	var err error
	h, ok := dc.entityHandlers[reflect.TypeOf(e)]
	if !ok {
		return errors.Errorf("trying to process an entity with no entity handler: %v", reflect.TypeOf(e))
	}
	switch e.GetStatus() {
	case entitystore.StatusERROR:
		err = h.Error(e)
	case entitystore.StatusCREATING:
		err = h.Add(e)
	case entitystore.StatusUPDATING:
		err = h.Update(e)
	case entitystore.StatusDELETING:
		err = h.Delete(e)
	default:
		err = errors.Errorf("invalid status: '%v'", e.GetStatus())
	}
	return err
}

func toSync(resyncPeriod time.Duration) entitystore.Filter {
	defer trace.Trace("")()

	now := time.Now().Add(-resyncPeriod)

	return []entitystore.FilterStat{
		entitystore.FilterStat{
			Subject: "ModifiedTime",
			Verb:    entitystore.FilterVerbBefore,
			Object:  now,
		},
		entitystore.FilterStat{
			Subject: "Status",
			Verb:    entitystore.FilterVerbIn,
			Object: []entitystore.Status{
				entitystore.StatusERROR, entitystore.StatusCREATING, entitystore.StatusUPDATING, entitystore.StatusDELETING,
			},
		},
	}
}

func (dc *DefaultController) sync() error {
	defer trace.Trace("")()

	for entityType := range dc.entityHandlers {
		entitiesPtr := reflect.New(reflect.SliceOf(entityType))
		if err := dc.store.List(dc.options.OrganizationID, toSync(dc.options.ResyncPeriod), entitiesPtr.Interface()); err != nil {
			return err
		}
		entities := entitiesPtr.Elem()
		for i := 0; i < entities.Len(); i++ {
			e := entities.Index(i).Interface().(entitystore.Entity)
			log.Printf("sync: processing entity %s", e.GetName())
			if err := dc.processItem(e); err != nil {
				log.Error(err)
			}
		}
	}
	return nil
}

// run runs the control loop
func (dc *DefaultController) run(stopChan <-chan bool) {
	defer trace.Trace("")()

	resyncTicker := time.NewTicker(dc.options.ResyncPeriod)
	defer resyncTicker.Stop()

	defer close(dc.watcher)

	// start workers
	for i := 0; i < dc.options.Workers; i++ {
		go func() {
			defer trace.Trace("")()

			for entity := range dc.watcher {
				func() {
					log.Printf("received event=%s entity=%s", entity.GetStatus(), entity.GetName())
					if err := dc.processItem(entity); err != nil {
						log.Error(err)
					}
				}()
			}
		}()
	}
	go func() {
		for range resyncTicker.C {
			func() {
				defer trace.Trace("")()

				log.Printf("periodic syncing with the underlying driver")
				if err := dc.sync(); err != nil {
					log.Error(err)
				}
			}()
		}
	}()

	<-stopChan
}
