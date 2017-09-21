///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package entitystore

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	helpers "gitlab.eng.vmware.com/serverless/serverless/pkg/testing/store"
)

type testEntity struct {
	BaseEntity
	Value string `json:"value"`
}

func (e *testEntity) getValue() string {
	return e.Value
}

type otherEntity struct {
	BaseEntity
	Other string `json:"other"`
}

func (e *otherEntity) getOther() string {
	return e.Other
}

func TestGet(t *testing.T) {
	path, kv := helpers.MakeKVStore(t)
	defer helpers.CleanKVStore(t, path, kv)
	es := New(kv)

	e := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testEntity",
			Tags: map[string]string{
				"role": "test",
			},
		},
		Value: "testValue",
	}

	id, err := es.Add(e)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	var retreived testEntity
	err = es.Get("testOrg", "testEntity", &retreived)

	assert.Equal(t, "testOrg", retreived.OrganizationID)
	assert.Equal(t, "testEntity", retreived.Name)
	assert.Equal(t, "testValue", retreived.Value)
	assert.NotNil(t, retreived.Tags)
	assert.Equal(t, "test", retreived.Tags["role"])
	assert.NotNil(t, retreived.CreatedTime)
	assert.NotNil(t, retreived.ModifiedTime)

	var missing testEntity
	err = es.Get("testOrg", "missing", &missing)
	assert.Error(t, err, "No error returned for missing entity")
}

func TestAdd(t *testing.T) {
	path, kv := helpers.MakeKVStore(t)
	defer helpers.CleanKVStore(t, path, kv)
	es := New(kv)

	e := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testEntity",
			Tags: map[string]string{
				"role": "test",
			},
		},
		Value: "testValue",
	}

	id, err := es.Add(e)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	var retreived testEntity
	err = es.Get("testOrg", e.Name, &retreived)
	assert.NoError(t, err, "Error fetching entity")

	var nameTests = []struct {
		name  string
		valid bool
	}{
		{"invalid name", false},
		{"valid-name", true},
		{"valid_name", true},
		{"VALIDNAME", true},
		{"invalid!name", false},
	}
	for _, tt := range nameTests {
		e.Name = tt.name
		id, err = es.Add(e)
		if tt.valid {
			assert.NoError(t, err, "Name is valid")
		} else {
			assert.Error(t, err, fmt.Sprintf("Name %s should be flagged as invalid", tt.name))
		}
	}
}

func TestPut(t *testing.T) {
	path, kv := helpers.MakeKVStore(t)
	defer helpers.CleanKVStore(t, path, kv)
	es := New(kv)

	e := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testEntity",
			Tags: map[string]string{
				"role": "test",
			},
		},
		Value: "testValue",
	}

	id, err := es.Add(e)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	_, err = es.Update(100, e)
	assert.Error(t, err)

	var retreived testEntity
	err = es.Get("testOrg", e.Name, &retreived)
	assert.NoError(t, err, "Error fetching entity")

	retreived.Value = "updatedValue"
	oldRev := retreived.Revision
	rev, err := es.Update(oldRev, &retreived)
	assert.NoError(t, err, "Error putting updated entity")
	assert.NotEqual(t, oldRev, rev)
}

type EntityConstructor func() Entity

func TestList(t *testing.T) {
	path, kv := helpers.MakeKVStore(t)
	defer helpers.CleanKVStore(t, path, kv)
	es := New(kv)

	e1 := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testEntity1",
			Tags: map[string]string{
				"filter": "one",
			},
		},
		Value: "testValue",
	}

	id, err := es.Add(e1)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	e2 := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testEntity2",
			Tags: map[string]string{
				"filter": "two",
			},
		},
		Value: "testValue",
	}

	id, err = es.Add(e2)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	id, err = es.Add(e2)
	assert.Error(t, err, "Should not allow adding entities of same name")

	items := new([]testEntity)
	err = es.List("testOrg", nil, items)
	assert.NoError(t, err, "Error listing entities")
	assert.Len(t, *items, 2)

	filter := func(e Entity) bool {
		if e.GetTags()["filter"] == "one" {
			return true
		}
		return false
	}
	items = new([]testEntity)
	err = es.List("testOrg", filter, items)
	assert.NoError(t, err, "Error listing entities")
	assert.Len(t, *items, 1)
	assert.Equal(t, "one", (*items)[0].GetTags()["filter"])
}

func TestMixedTypes(t *testing.T) {
	path, kv := helpers.MakeKVStore(t)
	defer helpers.CleanKVStore(t, path, kv)
	es := New(kv)

	te := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testEntity",
		},
		Value: "testValue",
	}
	oe := &otherEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "otherEntity",
		},
		Other: "otherValue",
	}
	id, err := es.Add(te)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)
	id, err = es.Add(oe)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	testEntities := &[]testEntity{}
	err = es.List("testOrg", nil, testEntities)
	assert.NoError(t, err, "Error listing entities")
	assert.Len(t, *testEntities, 1)

	otherEntities := &[]otherEntity{}
	err = es.List("testOrg", nil, otherEntities)
	assert.NoError(t, err, "Error listing entities")
	assert.Len(t, *otherEntities, 1)
}

func TestDelete(t *testing.T) {
	path, kv := helpers.MakeKVStore(t)
	defer helpers.CleanKVStore(t, path, kv)
	es := New(kv)

	e := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testEntity",
			Tags: map[string]string{
				"role": "test",
			},
		},
		Value: "testValue",
	}

	id, err := es.Add(e)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	err = es.Delete("testOrg", "testEntity", e)
	assert.NoError(t, err, "Error deleting entity")
	var retreived testEntity
	err = es.Get("testOrg", "testEntity", &retreived)
	assert.Error(t, err)
}
