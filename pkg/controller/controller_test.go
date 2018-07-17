///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package controller

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/vmware/dispatch/pkg/entity-store"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

const (
	testOrgID             = "testAPIManagerOrg"
	testResyncPeriod      = 2 * time.Second
	testSleepDuration     = 2 * testResyncPeriod
	testZookeeperLocation = "zookeeper.zookeeper.svc.cluster.local"
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

func (h *testEntityHandler) Add(ctx context.Context, obj entitystore.Entity) error {
	h.t.Logf("call Add %s", obj.GetName())
	h.addCounter <- obj.GetName()
	return nil
}

func (h *testEntityHandler) Update(ctx context.Context, obj entitystore.Entity) error {
	h.t.Logf("Update called %s", obj.GetName())
	return nil
}

func (h *testEntityHandler) Delete(ctx context.Context, obj entitystore.Entity) error {
	h.t.Logf("call Delete %s", obj.GetName())
	h.deleteCounter <- obj.GetName()
	return nil
}

func (h *testEntityHandler) Sync(ctx context.Context, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	return DefaultSync(context.Background(), h.store, h.Type(), resyncPeriod, nil)
}

func (h *testEntityHandler) Error(ctx context.Context, obj entitystore.Entity) error {
	h.t.Errorf("handleError func not implemented yet")
	return nil
}

func TestController(t *testing.T) {
	ctx := context.Background()
	store := helpers.MakeEntityStore(t)

	deleteCounter := make(chan string, 100)
	addCounter := make(chan string, 100)

	controller := NewController(Options{
		ResyncPeriod:      testResyncPeriod,
		ZookeeperLocation: testZookeeperLocation,
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
		store.Add(ctx, ent)
		watcher.OnAction(ctx, ent)
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
		store.Delete(ctx, testOrgID, ent.GetName(), ent)
		watcher.OnAction(ctx, ent)
	}

	for i := 0; i < len(testEntityNames); i++ {
		name := <-deleteCounter
		assert.Equal(t, testEntityNames[i], name)
		t.Logf("deleted %s", name)
	}
}
