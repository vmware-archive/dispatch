///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package controller

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/zookeeper"
)

// EntityHandler define an interface for entity operations of a generic controller
type EntityHandler interface {
	Type() reflect.Type
	Add(ctx context.Context, obj entitystore.Entity) error
	Update(ctx context.Context, obj entitystore.Entity) error
	Delete(ctx context.Context, obj entitystore.Entity) error
	Error(ctx context.Context, obj entitystore.Entity) error
	// Sync returns a list of entities which to process.  This method should call out and determine the actual state
	// of entities.
	Sync(ctx context.Context, resyncPeriod time.Duration) ([]entitystore.Entity, error)
}

const defaultWorkers = 1

// Options defines controller configuration
type Options struct {
	ServiceName string

	ResyncPeriod      time.Duration
	Workers           int
	Driver            zookeeper.Driver
	ZookeeperLocation string
}

// WatchEvent captures entity together with the associated context
type WatchEvent struct {
	Entity entitystore.Entity
	Ctx    context.Context
}

// Watcher channel type
type Watcher chan<- WatchEvent

// OnAction pushes an entity onto the watcher channel
func (w *Watcher) OnAction(ctx context.Context, e entitystore.Entity) {
	span, _ := trace.Trace(ctx, "")
	defer span.Finish()

	if w == nil || *w == nil {
		log.Warnf("nil watcher, skipping entity update: %s - %s", e.GetName(), e.GetStatus())
		return
	}
	// this event can outlive the context passed to OnAction, causing all sorts of troubles.
	// for example, HTTP request context is canceled when request is finished, which can result
	// in context being instantly canceled for any future WithTimeout or WithDeadline calls.
	// for this reason, we use fresh context with tracing span.
	*w <- WatchEvent{e, opentracing.ContextWithSpan(context.Background(), span)}
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
	watcher chan WatchEvent
	options Options
	driver  zookeeper.Driver

	entityHandlers map[reflect.Type]EntityHandler
}

// NewController creates a new controller
func NewController(options Options) Controller {
	if options.Workers == 0 {
		options.Workers = defaultWorkers
	}
	if options.ZookeeperLocation == "" {
		options.ZookeeperLocation = "127.0.0.1"
	}
	if options.Driver == nil {
		driver, err := zookeeper.NewDriver(options.ZookeeperLocation)
		if err != nil {
			log.Fatalf("Unable to get zookeeper driver for controller")
		}
		options.Driver = driver
		log.Infof("Connected to zookeeper at address %v", options.ZookeeperLocation)
	}
	return &DefaultController{
		done:    make(chan bool),
		watcher: make(chan WatchEvent),
		options: options,
		driver:  options.Driver,

		entityHandlers: map[reflect.Type]EntityHandler{},
	}
}

// Start starts the controller watch loop
func (dc *DefaultController) Start() {
	// Run sync once at the beginning to synchronize resources at service startup.
	// This should block until resources are synced to ensure proper handling of requests.
	dc.sync()

	go dc.run(dc.done)
}

// Shutdown stops the controller loop
func (dc *DefaultController) Shutdown() {
	dc.done <- true
}

// Watcher returns a watcher channel for the controller
func (dc *DefaultController) Watcher() Watcher {
	return dc.watcher
}

// AddEntityHandler adds entity handlers
func (dc *DefaultController) AddEntityHandler(h EntityHandler) {
	dc.entityHandlers[h.Type()] = h
}

func (dc *DefaultController) processItem(ctx context.Context, e entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	log.Debugf("Processing Item: %v (%v) with status %v", e.GetName(), e.GetID(), e.GetStatus())

	if err := dc.driver.CreateNode(fmt.Sprintf("/entities/%v", e.GetID()), []byte{}); err != nil {
		log.Fatalf("Unable to create znode for %v (%v): %v", e.GetName(), e.GetID(), err)
	} else {
		log.Infof("Created znode /entities/%v", e.GetID())
	}

	lock, canModify := dc.driver.LockEntity(e.GetID())
	if !canModify {
		return errors.Errorf("Failed to acquire lock for %v. Someone else is processing this entity", e.GetID())
	}
	log.Infof("Acquired lock for %v", e.GetID())
	defer dc.driver.ReleaseEntity(lock)

	var err error
	h, ok := dc.entityHandlers[reflect.TypeOf(e)]
	if !ok {
		return errors.Errorf("trying to process an entity with no entity handler: %v", reflect.TypeOf(e))
	}
	if e.GetDelete() {
		return h.Delete(ctx, e)
	}

	switch e.GetStatus() {
	case entitystore.StatusERROR:
		err = h.Error(ctx, e)
	case entitystore.StatusINITIALIZED, entitystore.StatusCREATING, entitystore.StatusMISSING:
		err = h.Add(ctx, e)
	case entitystore.StatusUPDATING:
		err = h.Update(ctx, e)
	case entitystore.StatusDELETING:
		err = h.Delete(ctx, e)
	case entitystore.StatusREADY:
		err = h.Update(ctx, e)
	default:
		err = errors.Errorf("invalid status: '%v'", e.GetStatus())
	}
	return err
}

func defaultSyncFilter(resyncPeriod time.Duration) entitystore.Filter {
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
				entitystore.StatusCREATING, entitystore.StatusUPDATING, entitystore.StatusDELETING,
			},
		})
}

// DefaultSync simply returns a list of entities in non-READY state which have been modified since the resync period.
func DefaultSync(ctx context.Context, store entitystore.EntityStore, entityType reflect.Type, resyncPeriod time.Duration, filter entitystore.Filter) ([]entitystore.Entity, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	valuesPtr := reflect.New(reflect.SliceOf(entityType))
	if filter == nil {
		filter = defaultSyncFilter(resyncPeriod)
	}
	opts := entitystore.Options{
		Filter: filter,
	}

	if err := store.ListGlobal(ctx, opts, valuesPtr.Interface()); err != nil {
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
	span, ctx := trace.Trace(context.Background(), "controller sync")
	defer span.Finish()
	if err := dc.driver.CreateNode("/entities", []byte{}); err != nil {
		log.Fatalf("Unable to create overarching znode %v", err)
	}
	sem := semaphore.NewWeighted(int64(dc.options.Workers))
	for _, handler := range dc.entityHandlers {
		entities, err := handler.Sync(ctx, dc.options.ResyncPeriod)
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
				log.Debugf("sync: processing entity %s (%v)", e.GetName(), e.GetStatus())
				if err := dc.processItem(ctx, e); err != nil {
					span.LogKV("error", err)
					log.Error(err)
				}
			}(e)
		}
	}
	return nil
}

// run runs the control loop
func (dc *DefaultController) run(stopChan <-chan bool) {
	resyncTicker := time.NewTicker(dc.options.ResyncPeriod)
	defer resyncTicker.Stop()

	defer close(dc.watcher)

	defer dc.driver.Close()

	if err := dc.driver.CreateNode("/entities", []byte{}); err != nil {
		log.Fatalf("Unable to create overarching znode %v", err)
	}
	// Start a worker pool.  The pool scales up to dc.options.Workers.
	go func() {
		sem := semaphore.NewWeighted(int64(dc.options.Workers))
		for watchEvent := range dc.watcher {
			if err := sem.Acquire(context.Background(), 1); err != nil {
				log.Warnf("Failed to acquire semaphore: %v", err)
				break
			}
			go func(event WatchEvent) {
				e := event.Entity
				defer sem.Release(1)
				log.Infof("received event=%s entity=%s", e.GetStatus(), e.GetName())
				if err := dc.processItem(event.Ctx, e); err != nil {
					log.Error(err)
				}
			}(watchEvent)
		}
	}()

	go func() {
		for range resyncTicker.C {
			func() {
				log.Debugf("%s periodic syncing with the underlying driver", dc.options.ServiceName)
				if err := dc.sync(); err != nil {
					log.Error(err)
				}
			}()
		}
	}()

	<-stopChan
}
