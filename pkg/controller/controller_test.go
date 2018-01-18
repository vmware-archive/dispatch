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

	"github.com/vmware/dispatch/pkg/entity-store"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
	"github.com/vmware/dispatch/pkg/trace"
)

const (
	testOrgID         = "testAPIManagerOrg"
	testResyncPeriod  = 2 * time.Second
	testSleepDuration = 2 * testResyncPeriod
)

type testEntity struct {
	entitystore.BaseEntity
}

type testEntityHandler struct {
	t                         *testing.T
	store                     entitystore.EntityStore
	addCounter, deleteCounter chan string
}

func (h *testEntityHandler) Type() reflect.Type {
	return reflect.TypeOf(&testEntity{})
}

func (h *testEntityHandler) Add(obj entitystore.Entity) error {
	h.t.Logf("call Add %s", obj.GetName())
	h.addCounter <- obj.GetName()
	return nil
}

func (h *testEntityHandler) Update(obj entitystore.Entity) error {
	h.t.Logf("Update called %s", obj.GetName())
	return nil
}

func (h *testEntityHandler) Delete(obj entitystore.Entity) error {
	h.t.Logf("call Delete %s", obj.GetName())
	h.deleteCounter <- obj.GetName()
	return nil
}

func (h *testEntityHandler) Sync(organizationID string, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	defer trace.Trace("")()

	return DefaultSync(h.store, h.Type(), organizationID, resyncPeriod)
}

func (h *testEntityHandler) Error(obj entitystore.Entity) error {
	h.t.Errorf("handleError func not implemented yet")
	return nil
}

func TestController(t *testing.T) {

	store := helpers.MakeEntityStore(t)

	deleteCounter := make(chan string, 100)
	addCounter := make(chan string, 100)

	controller := NewController(Options{
		OrganizationID: testOrgID,
		ResyncPeriod:   testResyncPeriod,
	})
	controller.AddEntityHandler(&testEntityHandler{t: t, store: store, addCounter: addCounter, deleteCounter: deleteCounter})
	watcher := controller.Watcher()

	controller.Start()
	defer controller.Shutdown()

	testEntityNames := []string{"test-a", "test-b", "test-c"}
	for _, name := range testEntityNames {
		ent := &testEntity{entitystore.BaseEntity{
			Name:   name,
			Status: entitystore.StatusCREATING,
		}}
		store.Add(ent)
		watcher.OnAction(ent)
	}

	for i := 0; i < len(testEntityNames); i++ {
		name := <-addCounter
		assert.Equal(t, testEntityNames[i], name)
		t.Logf("added %s", name)
	}

	for _, name := range testEntityNames {
		ent := &testEntity{entitystore.BaseEntity{
			Name:   name,
			Status: entitystore.StatusDELETING,
		}}
		store.Delete(testOrgID, ent.GetName(), ent)
		watcher.OnAction(ent)
	}

	for i := 0; i < len(testEntityNames); i++ {
		name := <-deleteCounter
		assert.Equal(t, testEntityNames[i], name)
		t.Logf("deleted %s", name)
	}
}
