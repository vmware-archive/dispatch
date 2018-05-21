///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package identitymanager

import (
	"context"
	"testing"

	"github.com/casbin/casbin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"time"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/identity-manager/mocks"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func TestPolicyAdd(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	model := casbin.NewModel(casbinPolicyModel)
	adapter := &mocks.AdapterMock{}
	adapter.On("LoadPolicy", mock.Anything).Return(nil)
	enforcer := casbin.NewSyncedEnforcer(model, adapter)
	adapter.AssertCalled(t, "LoadPolicy", model)
	adapter.AssertNumberOfCalls(t, "LoadPolicy", 1)
	handler := &policyEntityHandler{
		store:    es,
		enforcer: enforcer,
	}
	e := &Policy{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: IdentityManagerFlags.OrgID,
			Name:           "test-policy-1",
			Status:         entitystore.StatusREADY,
		},
		Rules: []Rule{},
	}
	es.Add(context.Background(), e)
	assert.NoError(t, handler.Add(context.Background(), e))
	// Ensures LoadPolicy is called after add
	adapter.AssertNumberOfCalls(t, "LoadPolicy", 2)
}

func TestPolicyDelete(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	model := casbin.NewModel(casbinPolicyModel)
	adapter := &mocks.AdapterMock{}
	adapter.On("LoadPolicy", mock.Anything).Return(nil)
	enforcer := casbin.NewSyncedEnforcer(model, adapter)
	adapter.AssertCalled(t, "LoadPolicy", model)
	adapter.AssertNumberOfCalls(t, "LoadPolicy", 1)
	handler := &policyEntityHandler{
		store:    es,
		enforcer: enforcer,
	}
	e := &Policy{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: IdentityManagerFlags.OrgID,
			Name:           "test-policy-1",
			Status:         entitystore.StatusREADY,
		},
		Rules: []Rule{},
	}
	es.Add(context.Background(), e)
	assert.NoError(t, handler.Delete(context.Background(), e))
	// Ensures LoadPolicy is called after delete
	adapter.AssertNumberOfCalls(t, "LoadPolicy", 2)
	err := es.Get(context.Background(), IdentityManagerFlags.OrgID, "test-policy-1", entitystore.Options{}, e)
	assert.Error(t, err)
}

func TestPolicyUpdate(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	model := casbin.NewModel(casbinPolicyModel)
	adapter := &mocks.AdapterMock{}
	adapter.On("LoadPolicy", mock.Anything).Return(nil)
	enforcer := casbin.NewSyncedEnforcer(model, adapter)
	adapter.AssertCalled(t, "LoadPolicy", model)
	adapter.AssertNumberOfCalls(t, "LoadPolicy", 1)
	handler := &policyEntityHandler{
		store:    es,
		enforcer: enforcer,
	}
	e := &Policy{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: IdentityManagerFlags.OrgID,
			Name:           "test-policy-1",
			Status:         entitystore.StatusREADY,
		},
		Rules: []Rule{},
	}
	es.Add(context.Background(), e)

	e.Rules = []Rule{
		Rule{
			Subjects:  []string{"user1@example.com"},
			Actions:   []string{"update,get"},
			Resources: []string{"image,function"},
		},
	}
	assert.NoError(t, handler.Update(context.Background(), e))

	// Ensures LoadPolicy is called after update
	adapter.AssertNumberOfCalls(t, "LoadPolicy", 2)
	err := es.Get(context.Background(), IdentityManagerFlags.OrgID, "test-policy-1", entitystore.Options{}, e)
	assert.NoError(t, err)
}

func TestPolicySync(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	model := casbin.NewModel(casbinPolicyModel)
	adapter := &mocks.AdapterMock{}
	adapter.On("LoadPolicy", mock.Anything).Return(nil)
	enforcer := casbin.NewSyncedEnforcer(model, adapter)
	adapter.AssertCalled(t, "LoadPolicy", model)
	adapter.AssertNumberOfCalls(t, "LoadPolicy", 1)
	handler := &policyEntityHandler{
		store:    es,
		enforcer: enforcer,
	}
	e := &Policy{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: IdentityManagerFlags.OrgID,
			Name:           "test-policy-1",
			Status:         entitystore.StatusCREATING,
		},
		Rules: []Rule{},
	}
	es.Add(context.Background(), e)
	entities, err := handler.Sync(context.Background(), IdentityManagerFlags.OrgID, time.Duration(5))
	assert.NoError(t, err)
	assert.Len(t, entities, 1)
	// Ensures LoadPolicy is called after add
	adapter.AssertNumberOfCalls(t, "LoadPolicy", 2)
}
