///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package entitystore

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/stretchr/testify/assert"
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

func TestGetDataType(t *testing.T) {
	assert.Equal(t, dataType("testEntity"), getDataType(&testEntity{}))
}

func TestAdd(t *testing.T) {
	path, kv := makeKVStore(t)
	defer cleanKVStore(t, path, kv)
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
	err = es.GetById("testOrg", id, &retreived)
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
	err = es.GetById("testOrg", id, &retreived)
	assert.NoError(t, err, "Error fetching entity")

	retreived.Value = "updatedValue"
	oldRev := retreived.Revision
	rev, err := es.Update(oldRev, &retreived)
	assert.NoError(t, err, "Error putting updated entity")
	assert.NotEqual(t, oldRev, rev)
}

type EntityConstructor func() Entity

func TestList(t *testing.T) {
	path, kv := makeKVStore(t)
	defer cleanKVStore(t, path, kv)
	es := New(kv)

	e1 := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testEntity",
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
			Name:           "testEntity",
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
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	items := new([]testEntity)
	err = es.List("testOrg", nil, items)
	assert.NoError(t, err, "Error listing entities")
	assert.Len(t, *items, 3)

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
	path, kv := makeKVStore(t)
	defer cleanKVStore(t, path, kv)
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
