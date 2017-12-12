///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package controller

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	entitystore "github.com/vmware/dispatch/pkg/entity-store"

	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

const (
	testOrgID         = "testAPIManagerOrg"
	testResyncPeriod  = 2 * time.Second
	testSleepDuration = 2 * testResyncPeriod
)

func TestController(t *testing.T) {

	store := helpers.MakeEntityStore(t)
	lw := NewDefaultListWatcher(store, ListOptions{
		OrganizationID: testOrgID,
		EntityType:     reflect.TypeOf(entitystore.BaseEntity{}),
	})
	watcher, err := lw.Watch(ListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	deleteCounter := make(chan string, 100)
	addCounter := make(chan string, 100)

	handlers := EventHandlers{
		AddFunc: func(obj entitystore.Entity) error {
			t.Logf("call AddFunc %s", obj.GetName())
			addCounter <- obj.GetName()
			return nil
		},
		UpdateFunc: func(_, obj entitystore.Entity) error {
			t.Logf("UpdateFunc called %s", obj.GetName())
			return nil
		},
		DeleteFunc: func(obj entitystore.Entity) error {
			t.Logf("call DeleteFunc %s", obj.GetName())
			deleteCounter <- obj.GetName()
			return nil
		},
		ErrorFunc: func(obj entitystore.Entity) error {
			t.Errorf("handleError func not implemented yet")
			return nil
		},
	}
	informer := NewDefaultInformer(lw, testResyncPeriod)
	informer.AddEventHandlers(handlers)
	controller, err := NewDefaultController(informer)
	go controller.Run()
	defer controller.Shutdown()

	testEntityNames := []string{"test-a", "test-b", "test-c"}
	for _, name := range testEntityNames {
		ent := &entitystore.BaseEntity{
			Name:   name,
			Status: entitystore.StatusCREATING,
		}
		store.Add(ent)
		watcher.OnAction(ent)
	}

	for i := 0; i < len(testEntityNames); i++ {
		name := <-addCounter
		assert.Equal(t, testEntityNames[i], name)
		t.Logf("added %s", name)
	}

	for _, name := range testEntityNames {
		ent := &entitystore.BaseEntity{
			Name:   name,
			Status: entitystore.StatusDELETING,
		}
		store.Delete(testOrgID, ent.GetName(), ent)
		watcher.OnAction(ent)
	}

	for i := 0; i < len(testEntityNames); i++ {
		name := <-deleteCounter
		assert.Equal(t, testEntityNames[i], name)
		t.Logf("deleted %s", name)
	}
}
