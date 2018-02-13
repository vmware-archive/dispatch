///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package controller

import (
	"context"
	"reflect"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"

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
	// Sync returns a list of entities which to process.  This method should call out and determine the actual state
	// of entities.
	Sync(organizationID string, resyncPeriod time.Duration) ([]entitystore.Entity, error)
}

const defaultWorkers = 1

// Options defines controller configuration
type Options struct {
	OrganizationID string
	Namespace      string

	ResyncPeriod time.Duration
	Workers      int
}

// Watcher channel type
type Watcher chan<- entitystore.Entity

// OnAction pushes an entity onto the watcher channel
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
	options Options

	entityHandlers map[reflect.Type]EntityHandler
}

// NewController creates a new controller
func NewController(options Options) Controller {
	defer trace.Trace("")()

	if options.Workers == 0 {
		options.Workers = defaultWorkers
	}

	return &DefaultController{
		done:    make(chan bool),
		watcher: make(chan entitystore.Entity),
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

// Watcher returns a watcher channel for the controller
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
	if e.GetDelete() {
		return h.Delete(e)
	}
	switch e.GetStatus() {
	case entitystore.StatusERROR:
		err = h.Error(e)
	case entitystore.StatusINITIALIZED, entitystore.StatusCREATING, entitystore.StatusMISSING:
		err = h.Add(e)
	case entitystore.StatusUPDATING:
		err = h.Update(e)
	case entitystore.StatusDELETING:
		err = h.Delete(e)
	case entitystore.StatusREADY:
		err = h.Update(e)
	default:
		err = errors.Errorf("invalid status: '%v'", e.GetStatus())
	}
	return err
}

func defaultSyncFilter(resyncPeriod time.Duration) entitystore.Filter {
	defer trace.Trace("")()

	now := time.Now().Add(-resyncPeriod)
	return entitystore.FilterEverything().Add(
		entitystore.FilterStat{
			Scope:   entitystore.FilterScopeField,
			Subject: "ModifiedTime",
			Verb:    entitystore.FilterVerbBefore,
			Object:  now,
		},
		entitystore.FilterStat{
			Scope:   entitystore.FilterScopeField,
			Subject: "Status",
			Verb:    entitystore.FilterVerbIn,
			Object: []entitystore.Status{
				entitystore.StatusERROR, entitystore.StatusCREATING, entitystore.StatusUPDATING, entitystore.StatusDELETING,
			},
		})
}

// DefaultSync simply returns a list of entities in non-READY state which have been modified since the resync period.
func DefaultSync(store entitystore.EntityStore, entityType reflect.Type, organizationID string, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	valuesPtr := reflect.New(reflect.SliceOf(entityType))
	opts := entitystore.Options{
		Filter: defaultSyncFilter(resyncPeriod),
	}
	if err := store.List(organizationID, opts, valuesPtr.Interface()); err != nil {
		return nil, err
	}
	values := valuesPtr.Elem()
	var entities []entitystore.Entity
	for i := 0; i < values.Len(); i++ {
		e := values.Index(i).Interface().(entitystore.Entity)
		entities = append(entities, e)
	}
	return entities, nil
}

func (dc *DefaultController) sync() error {
	defer trace.Trace("")()
	sem := semaphore.NewWeighted(int64(dc.options.Workers))
	ctx := context.Background()

	for _, handler := range dc.entityHandlers {
		entities, err := handler.Sync(dc.options.OrganizationID, dc.options.ResyncPeriod)
		if err != nil {
			return err
		}
		for _, e := range entities {
			if err := sem.Acquire(ctx, 1); err != nil {
				log.Printf("Failed to acquire semaphore: %v", err)
				break
			}
			go func(e entitystore.Entity) {
				defer sem.Release(1)
				log.Printf("sync: processing entity %s", e.GetName())
				if err := dc.processItem(e); err != nil {
					log.Error(err)
				}
			}(e)
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

	// Start a worker pool.  The pool scales up to dc.options.Workers.
	go func() {
		defer trace.Trace("")()
		sem := semaphore.NewWeighted(int64(dc.options.Workers))
		ctx := context.Background()

		for entity := range dc.watcher {
			if err := sem.Acquire(ctx, 1); err != nil {
				log.Printf("Failed to acquire semaphore: %v", err)
				break
			}
			go func(e entitystore.Entity) {
				defer sem.Release(1)
				log.Printf("received event=%s entity=%s", e.GetStatus(), e.GetName())
				if err := dc.processItem(e); err != nil {
					log.Error(err)
				}
			}(entity)
		}
	}()

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
