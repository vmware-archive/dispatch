///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package controller

import (
	"reflect"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
)

// ListOptions defines filters for lister and watcher
type ListOptions struct {
	EntityType     reflect.Type
	OrganizationID string
	Namespace      string
}

// Watcher defines an watcher interface
type Watcher interface {
	OnAction(entitystore.Entity)
	ResultChan() <-chan entitystore.Entity
}

// ListerWatcher defines an ListerWatcher interface
type ListerWatcher interface {
	List(options ListOptions) ([]entitystore.Entity, error)
	Watch(options ListOptions) (Watcher, error)
}

// DefaultWatcher defines a basic Watcher
type DefaultWatcher struct {
	filter entitystore.Filter
	result chan entitystore.Entity
}

// NewDefaultWatcher defines a default watcher
func NewDefaultWatcher(filter entitystore.Filter) Watcher {
	return &DefaultWatcher{
		filter: filter,
		result: make(chan entitystore.Entity),
	}
}

// OnAction put the event into chan
func (d *DefaultWatcher) OnAction(e entitystore.Entity) {
	if d.filter(e) {
		go func() {
			d.result <- e
		}()
	}
}

// ResultChan returns the pending entity
func (d *DefaultWatcher) ResultChan() <-chan entitystore.Entity {
	return d.result
}

// DefaultListerWatcher defines a basic ListerWatcher
type DefaultListerWatcher struct {
	options ListOptions
	filter  entitystore.Filter
	store   entitystore.EntityStore
	watcher Watcher
}

// NewDefaultListWatcher creates a new DefaultListWatcher
func NewDefaultListWatcher(es entitystore.EntityStore, options ListOptions) ListerWatcher {

	filter := func(e entitystore.Entity) bool {
		if options.EntityType != nil {
			return entitystore.GetDataType(e) == options.EntityType.Name()
		}
		return true
	}

	lw := &DefaultListerWatcher{
		options: options,
		store:   es,
		filter:  filter,
		watcher: NewDefaultWatcher(filter),
	}
	return lw
}

// Watch returns an Watcher
func (d *DefaultListerWatcher) Watch(_ ListOptions) (Watcher, error) {
	return d.watcher, nil
}

// List returns a list of entities
func (d *DefaultListerWatcher) List(_ ListOptions) ([]entitystore.Entity, error) {

	// create a pointer to the slice of the EntityType
	entitiesV := reflect.MakeSlice(reflect.SliceOf(d.options.EntityType), 0, 0)
	entitiesVPtr := reflect.New(entitiesV.Type())

	err := d.store.List(d.options.OrganizationID, d.filter, entitiesVPtr.Interface())
	if err != nil {
		err = errors.Wrap(err, "store error when listing entities:")
		log.Error(err)
		return nil, err
	}

	var result []entitystore.Entity
	entitiesV = entitiesVPtr.Elem()
	for i := 0; i < entitiesV.Len(); i++ {
		entV := entitiesV.Index(i)
		entVPtr := reflect.New(entV.Type())
		entVPtr.Elem().Set(entV)
		result = append(result, entVPtr.Interface().(entitystore.Entity))
	}
	return result, nil
}
