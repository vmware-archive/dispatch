///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package entitystore

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/stretchr/testify/assert"
)

const TypeTestEntity = "testEntity"
const TypeOtherEntity = "otherEntity"

type testEntity struct {
	Entity
	Value string `json:"value"`
}

func (e *testEntity) GetType() DataType {
	return TypeTestEntity
}

func (e *testEntity) getValue() string {
	return e.Value
}

type otherEntity struct {
	Entity
	Other string `json:"other"`
}

func (e *otherEntity) GetType() DataType {
	return TypeOtherEntity
}

func (e *otherEntity) getValue() string {
	return e.Other
}

var typeMap TypeMap = map[DataType]reflect.Type{
	TypeTestEntity:  reflect.TypeOf(testEntity{}),
	TypeOtherEntity: reflect.TypeOf(otherEntity{}),
}

func makeKVStore(t *testing.T) (path string, kv store.Store) {
	boltdb.Register()
	file, err := ioutil.TempFile(os.TempDir(), "test")
	assert.NoError(t, err, "Cannot create temp file")
	kv, err = libkv.NewStore(
		store.BOLTDB,
		[]string{file.Name()},
		&store.Config{
			Bucket:            "test",
			ConnectionTimeout: 1 * time.Second,
			PersistConnection: true,
		},
	)
	assert.NoError(t, err, "Cannot create store")
	return file.Name(), kv
}

func cleanKVStore(t *testing.T, path string, kv store.Store) {
	kv.Close()
	os.Remove(path)
}

func TestAdd(t *testing.T) {
	path, kv := makeKVStore(t)
	defer cleanKVStore(t, path, kv)
	es := NewEntityStore(kv, typeMap)

	e := &testEntity{
		Entity: Entity{
			OrganizationID: "testOrg",
			Name:           "testEntity",
			Tags: map[string]string{
				"role": "test",
			},
		},
		Value: "testValue",
	}

	id, err := es.AddEntity(e)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	var retreived testEntity
	err = es.GetEntityById("testOrg", TypeTestEntity, id, &retreived)
	assert.NoError(t, err, "Error fetching entity")

	assert.Equal(t, "testOrg", retreived.OrganizationID)
	assert.Equal(t, "testEntity", retreived.Name)
	assert.Equal(t, "testValue", retreived.Value)
	assert.NotNil(t, retreived.Tags)
	assert.Equal(t, "test", retreived.Tags["role"])
	assert.NotNil(t, retreived.CreatedTime)
	assert.NotNil(t, retreived.ModifiedTime)
}

func TestPut(t *testing.T) {
	path, kv := makeKVStore(t)
	defer cleanKVStore(t, path, kv)
	es := NewEntityStore(kv, typeMap)

	e := &testEntity{
		Entity: Entity{
			OrganizationID: "testOrg",
			Name:           "testEntity",
			Tags: map[string]string{
				"role": "test",
			},
		},
		Value: "testValue",
	}

	id, err := es.AddEntity(e)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	_, err = es.PutEntity(100, e)
	assert.Error(t, err)

	var retreived testEntity
	err = es.GetEntityById("testOrg", TypeTestEntity, id, &retreived)
	assert.NoError(t, err, "Error fetching entity")

	retreived.Value = "updatedValue"
	oldRev := retreived.Revision
	rev, err := es.PutEntity(oldRev, &retreived)
	assert.NoError(t, err, "Error putting updated entity")
	assert.NotEqual(t, oldRev, rev)
}

type EntityConstructor func() StoredEntity

func TestList(t *testing.T) {
	path, kv := makeKVStore(t)
	defer cleanKVStore(t, path, kv)
	es := NewEntityStore(kv, typeMap)

	e1 := &testEntity{
		Entity: Entity{
			OrganizationID: "testOrg",
			Name:           "testEntity",
			Tags: map[string]string{
				"filter": "one",
			},
		},
		Value: "testValue",
	}

	id, err := es.AddEntity(e1)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	e2 := &testEntity{
		Entity: Entity{
			OrganizationID: "testOrg",
			Name:           "testEntity",
			Tags: map[string]string{
				"filter": "two",
			},
		},
		Value: "testValue",
	}

	id, err = es.AddEntity(e2)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	items, err := es.ListEntities("testOrg", TypeTestEntity, nil)
	assert.NoError(t, err, "Error listing entities")
	all, ok := items.([]testEntity)
	assert.True(t, ok)
	assert.Len(t, all, 2)

	filter := func(e StoredEntity) bool {
		if e.GetTags()["filter"] == "one" {
			return true
		}
		return false
	}
	items, err = es.ListEntities("testOrg", TypeTestEntity, filter)
	assert.NoError(t, err, "Error listing entities")
	all, ok = items.([]testEntity)
	assert.True(t, ok)
	assert.Len(t, all, 1)
}

func TestMixedTypes(t *testing.T) {
	path, kv := makeKVStore(t)
	defer cleanKVStore(t, path, kv)
	es := NewEntityStore(kv, typeMap)

	te := &testEntity{
		Entity: Entity{
			OrganizationID: "testOrg",
			Name:           "testEntity",
		},
		Value: "testValue",
	}
	oe := &otherEntity{
		Entity: Entity{
			OrganizationID: "testOrg",
			Name:           "otherEntity",
		},
		Other: "otherValue",
	}
	id, err := es.AddEntity(te)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)
	id, err = es.AddEntity(oe)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	items, err := es.ListEntities("testOrg", TypeTestEntity, nil)
	assert.NoError(t, err, "Error listing entities")
	testEntities, ok := items.([]testEntity)
	assert.True(t, ok)
	assert.Len(t, testEntities, 1)

	items, err = es.ListEntities("testOrg", TypeOtherEntity, nil)
	assert.NoError(t, err, "Error listing entities")
	otherEntities, ok := items.([]otherEntity)
	assert.True(t, ok)
	assert.Len(t, otherEntities, 1)
}
