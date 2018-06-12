///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package subscriptions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/event-manager/subscriptions/entities"
	mocks2 "github.com/vmware/dispatch/pkg/event-manager/subscriptions/mocks"

	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func mockSubscriptionHandler(manager Manager, es entitystore.EntityStore) *EntityHandler {
	return &EntityHandler{
		store:   es,
		manager: manager,
	}
}

func TestSubscriptionAdd(t *testing.T) {
	manager := &mocks2.Manager{}
	es := helpers.MakeEntityStore(t)
	handler := mockSubscriptionHandler(manager, es)
	sub := &entities.Subscription{
		BaseEntity: entitystore.BaseEntity{
			Name:           "sub1",
			Status:         entitystore.StatusCREATING,
			OrganizationID: testOrgID,
		},
		EventType: "test.topic",
		Function:  "test.function",
	}
	es.Add(context.Background(), sub)
	manager.On("Create", mock.Anything, mock.Anything).Return(nil)
	assert.NoError(t, handler.Add(context.Background(), sub))

}

func TestSubscriptionDelete(t *testing.T) {
	manager := &mocks2.Manager{}
	es := helpers.MakeEntityStore(t)
	handler := mockSubscriptionHandler(manager, es)
	sub := &entities.Subscription{
		BaseEntity: entitystore.BaseEntity{
			Name:           "sub1",
			Status:         entitystore.StatusDELETING,
			OrganizationID: testOrgID,
		},
		EventType: "test.topic",
		Function:  "test.function",
	}
	es.Add(context.Background(), sub)
	manager.On("Delete", mock.Anything, mock.Anything).Return(nil)
	assert.NoError(t, handler.Delete(context.Background(), sub))
	var subs []*entities.Subscription
	es.List(context.Background(), "", entitystore.Options{}, subs)
	assert.Len(t, subs, 0)
}
