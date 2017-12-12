///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package controller

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

const (
	testOrgIDListWatch = "testOrgID"
)

func TestListWatcher(t *testing.T) {

	store := helpers.MakeEntityStore(t)
	store.Add(&entitystore.BaseEntity{
		Name:           "test-a",
		OrganizationID: testOrgIDListWatch,
	})

	lw := NewDefaultListWatcher(store, ListOptions{
		EntityType:     reflect.TypeOf(entitystore.BaseEntity{}),
		OrganizationID: testOrgIDListWatch,
	})

	entities, err := lw.List(ListOptions{})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entities))
	assert.Equal(t, "test-a", entities[0].GetName())
}
