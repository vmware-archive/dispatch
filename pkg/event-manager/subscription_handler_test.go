///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vmware/dispatch/pkg/entity-store"

	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func mockSubscriptionHandler(manager SubscriptionManager, es entitystore.EntityStore) *subscriptionEntityHandler {
	return &subscriptionEntityHandler{
		store:   es,
		manager: manager,
	}
}

func TestSubscriptionAdd(t *testing.T) {
	manager := &SubscriptionManagerMock{}
	es := helpers.MakeEntityStore(t)
	handler := mockSubscriptionHandler(manager, es)
	sub := &Subscription{
		BaseEntity: entitystore.BaseEntity{
			Name:   "sub1",
			Status: entitystore.StatusCREATING,
		},
		Topic: "test.topic",
		Subscriber: Subscriber{
			Type: FunctionSubscriber,
			Name: "test.function",
		},
	}
	es.Add(sub)
	manager.On("Create", mock.Anything).Return(nil)
	assert.NoError(t, handler.Add(sub))

}

func TestSubscriptionDelete(t *testing.T) {
	manager := &SubscriptionManagerMock{}
	es := helpers.MakeEntityStore(t)
	handler := mockSubscriptionHandler(manager, es)
	sub := &Subscription{
		BaseEntity: entitystore.BaseEntity{
			Name:   "sub1",
			Status: entitystore.StatusDELETING,
		},
		Topic: "test.topic",
		Subscriber: Subscriber{
			Type: FunctionSubscriber,
			Name: "test.function",
		},
	}
	es.Add(sub)
	manager.On("Delete", mock.Anything).Return(nil)
	assert.NoError(t, handler.Delete(sub))
	var subs []*Subscription
	es.List("", entitystore.Options{}, subs)
	assert.Len(t, subs, 0)
}
