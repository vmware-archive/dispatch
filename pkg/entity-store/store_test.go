///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package entitystore

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	postgresConfig = BackendConfig{
		Backend:  "postgres",
		Address:  "192.168.99.104:5432",
		Username: "testuser",
		Password: "testpasswd",
		Bucket:   "testdb",
	}
)

type testEntity struct {
	BaseEntity
	Value string `json:"value" db:"value"`
}

func (e *testEntity) getValue() string {
	return e.Value
}

type testEntitySecond struct {
	BaseEntity
	Value string `json:"value" db:"value"`
}

func (e *testEntitySecond) getValue() string {
	return e.Value
}

type otherEntity struct {
	BaseEntity
	Other string `json:"other"`
}

func (e *otherEntity) getOther() string {
	return e.Other
}

func TestLibkvEntityStore(t *testing.T) {

	file, err := ioutil.TempFile(os.TempDir(), "test")
	assert.NoError(t, err, "Cannot create temp file")
	defer os.Remove(file.Name())

	libkvConfig := BackendConfig{
		Backend: "boltdb",
		Address: file.Name(),
		Bucket:  "test",
	}
	es, err := NewFromBackend(libkvConfig)
	assert.NoError(t, err, "Cannot create store")

	testGet(t, es)
	testAdd(t, es)
	testPut(t, es)
	testList(t, es)
	testListSamePrefix(t, es)
	testListWithFilter(t, es)
	testListWithFilterOnTags(t, es)
	testDelete(t, es)
	testInvalidNames(t, es)
	testMixedTypes(t, es)

	os.Remove(file.Name())
}

func testGet(t *testing.T, es EntityStore) {

	e := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testEntityGet",
			Tags: map[string]string{
				"role": "test",
			},
		},
		Value: "testValueGet",
	}

	id, err := es.Add(context.Background(), e)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	var retreived testEntity
	err = es.Get(context.Background(), "testOrg", "testEntityGet", Options{}, &retreived)

	assert.NoError(t, err)
	assert.Equal(t, "testOrg", retreived.OrganizationID)
	assert.Equal(t, "testEntityGet", retreived.Name)
	assert.Equal(t, "testValueGet", retreived.Value)
	assert.NotNil(t, retreived.Tags)
	assert.Equal(t, "test", retreived.Tags["role"])
	assert.NotNil(t, retreived.CreatedTime)
	assert.NotNil(t, retreived.ModifiedTime)

	var missing testEntity
	err = es.Get(context.Background(), "testOrg", "missing", Options{}, &missing)
	assert.Error(t, err, "No error returned for missing entity")

	// clean up
	err = es.Delete(context.Background(), "testOrg", "testEntityGet", e)
	assert.NoError(t, err, "Error clean up")
}

func testInvalidNames(t *testing.T, es EntityStore) {

	e := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
		},
		Value: "testInvalidNames",
	}

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
		_, err := es.Add(context.Background(), e)
		if tt.valid {
			assert.NoError(t, err, "Name is valid")
			// clean up
			err = es.Delete(context.Background(), "testOrg", tt.name, e)
			assert.NoError(t, err, "Error clean up")
		} else {
			assert.Error(t, err, fmt.Sprintf("Name %s should be flagged as invalid", tt.name))
		}
	}
}

func testAdd(t *testing.T, es EntityStore) {

	e := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testEntityAdd",
			Tags: map[string]string{
				"role": "test",
			},
		},
		Value: "testValueAdd",
	}

	id, err := es.Add(context.Background(), e)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	var retreived testEntity
	err = es.Get(context.Background(), "testOrg", e.Name, Options{}, &retreived)
	assert.NoError(t, err, "Error fetching entity")

	// clean up
	err = es.Delete(context.Background(), "testOrg", "testEntityAdd", e)
	assert.NoError(t, err, "Error clean up")
}

func testPut(t *testing.T, es EntityStore) {

	e := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testEntityPut",
			Tags: map[string]string{
				"role": "test",
			},
		},
		Value: "testValuePut",
	}

	id, err := es.Add(context.Background(), e)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	_, err = es.Update(context.Background(), 100, e)
	assert.Error(t, err)

	var retreived, updated testEntity
	err = es.Get(context.Background(), "testOrg", e.Name, Options{}, &retreived)
	assert.NoError(t, err, "Error fetching entity")

	retreived.Value = "updatedValue"
	oldRev := retreived.Revision
	rev, err := es.Update(context.Background(), oldRev, &retreived)
	assert.NoError(t, err, "Error putting updated entity")
	assert.NotEqual(t, oldRev, rev)
	err = es.Get(context.Background(), "testOrg", retreived.Name, Options{}, &updated)
	assert.Equal(t, updated.Revision, retreived.Revision, "Revision does not match")

	// cannot update an non-exist entity
	nonexistEntity := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "noSuchEntity",
		},
		Value: "noSuchValue",
	}
	_, err = es.Update(context.Background(), 0, nonexistEntity)
	assert.Error(t, err)

	// clean up
	err = es.Delete(context.Background(), "testOrg", "testEntityPut", e)
	assert.NoError(t, err, "Error clean up")
}

func testList(t *testing.T, es EntityStore) {

	e1 := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testEntityList1",
			Status:         StatusERROR,
			Tags: map[string]string{
				"filter": "one",
			},
		},
		Value: "testValue1",
	}

	id, err := es.Add(context.Background(), e1)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	e2 := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testEntityList2",
			Status:         StatusCREATING,
			Tags: map[string]string{
				"filter": "two",
			},
		},
		Value: "testValue2",
	}

	id, err = es.Add(context.Background(), e2)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	id, err = es.Add(context.Background(), e2)
	assert.Error(t, err, "Should not allow adding entities of same name")

	var items []*testEntity
	err = es.List(context.Background(), "testOrg", Options{}, &items)
	assert.NoError(t, err, "Error listing entities")
	assert.Len(t, items, 2)

	for _, item := range items {
		var i testEntity
		err = es.Get(context.Background(), "testOrg", item.GetName(), Options{}, &i)
		assert.NoError(t, err, "Error getting entity")
		assert.Equal(t, i.Revision, item.Revision, "Revision does not match")
	}

	filter := FilterEverything().Add(
		FilterStat{
			Scope:   FilterScopeField,
			Subject: "Status",
			Verb:    FilterVerbEqual,
			Object:  StatusERROR,
		})

	items = []*testEntity{}
	err = es.List(context.Background(), "testOrg", Options{Filter: filter}, &items)
	require.NoError(t, err, "Error listing entities")
	require.Len(t, items, 1)
	assert.Equal(t, string(StatusERROR), string(items[0].Status))

	// clean up
	err = es.Delete(context.Background(), "testOrg", "testEntityList1", e1)
	assert.NoError(t, err, "Error clean up")
	err = es.Delete(context.Background(), "testOrg", "testEntityList2", e2)
	assert.NoError(t, err, "Error clean up")
}

func testListSamePrefix(t *testing.T, es EntityStore) {

	e1 := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testEntityList1",
			Status:         StatusERROR,
		},
		Value: "testValue1",
	}

	id, err := es.Add(context.Background(), e1)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	e2 := &testEntitySecond{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testEntityList2",
			Status:         StatusCREATING,
		},
		Value: "testValue2",
	}

	id, err = es.Add(context.Background(), e2)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	var items []*testEntity
	err = es.List(context.Background(), "testOrg", Options{}, &items)
	assert.NoError(t, err, "Error listing entities")
	assert.Len(t, items, 1)

	var itemsSecond []*testEntitySecond
	err = es.List(context.Background(), "testOrg", Options{}, &itemsSecond)
	assert.NoError(t, err, "Error listing entities")
	assert.Len(t, items, 1)

	// clean up
	err = es.Delete(context.Background(), "testOrg", "testEntityList1", e1)
	assert.NoError(t, err, "Error clean up")
	err = es.Delete(context.Background(), "testOrg", "testEntityList2", e2)
	assert.NoError(t, err, "Error clean up")
}

func testListWithFilterOnTags(t *testing.T, es EntityStore) {

	testFoo := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testFoo",
			Status:         StatusREADY,
			Tags: map[string]string{
				"Application": "foo",
			},
		},
	}
	_, err := es.Add(context.Background(), testFoo)
	assert.NoError(t, err, "Error adding entity")

	testBar := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testBar",
			Status:         StatusREADY,
			Tags: map[string]string{
				"Application": "bar",
			},
		},
	}
	_, err = es.Add(context.Background(), testBar)
	assert.NoError(t, err, "Error adding entity")

	var result []*testEntity
	filterBar := FilterByApplication("bar")
	err = es.List(context.Background(), "testOrg", Options{Filter: filterBar}, &result)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "testBar", result[0].Name)

	// clean up
	es.Delete(context.Background(), "testOrg", testBar.Name, testBar)
	es.Delete(context.Background(), "testOrg", testFoo.Name, testFoo)
}

func testListWithFilter(t *testing.T, es EntityStore) {

	testTimeBeforeEntity := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testTimeBefore",
			Status:         StatusREADY,
		},
		Value: "testTimeBefore",
	}
	_, err := es.Add(context.Background(), testTimeBeforeEntity)
	assert.NoError(t, err, "Error adding entity")

	testTime := time.Now()
	time.Sleep(time.Second)

	testDeletedEntity := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testDeleted",
			Status:         StatusDELETED,
			Delete:         true,
		},
		Value: "testDeleted",
	}
	_, err = es.Add(context.Background(), testDeletedEntity)
	assert.NoError(t, err)

	testEqualValueEntity := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testEqualValue",
			Status:         StatusDELETING,
		},
		Value: "testEqualValue",
	}
	_, err = es.Add(context.Background(), testEqualValueEntity)
	assert.NoError(t, err)

	testInEntity := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testIn",
			Status:         StatusCREATING,
		},
		Value: "testIn",
	}
	_, err = es.Add(context.Background(), testInEntity)
	assert.NoError(t, err)

	filterTimeBefore := FilterStat{
		Scope:   FilterScopeField,
		Subject: "CreatedTime",
		Verb:    FilterVerbBefore,
		Object:  testTime,
	}
	filterEqualValue := FilterStat{Scope: FilterScopeExtra, Subject: "Value", Verb: FilterVerbEqual, Object: "testEqualValue"}
	filterDeleted := FilterStat{Scope: FilterScopeField, Subject: "Delete", Verb: FilterVerbEqual, Object: true}
	filterIn := FilterStat{
		Scope:   FilterScopeField,
		Subject: "Status", Verb: FilterVerbIn,
		Object: []Status{StatusCREATING, StatusDELETING, StatusERROR}}

	var result []*testEntity
	err = es.List(context.Background(), "testOrg", Options{Filter: FilterEverything().Add(filterTimeBefore)}, &result)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "testTimeBefore", result[0].Name)

	err = es.List(context.Background(), "testOrg", Options{Filter: FilterEverything().Add(filterEqualValue)}, &result)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "testEqualValue", result[0].Name)

	err = es.List(context.Background(), "testOrg", Options{Filter: FilterEverything().Add(filterEqualValue).Add(filterIn)}, &result)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "testEqualValue", result[0].Name)

	err = es.List(context.Background(), "testOrg", Options{Filter: FilterEverything().Add(filterDeleted)}, &result)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "testDeleted", result[0].Name)

	err = es.List(context.Background(), "testOrg", Options{Filter: FilterEverything().Add(filterIn)}, &result)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// clean up
	es.Delete(context.Background(), "testOrg", testInEntity.Name, testInEntity)
	es.Delete(context.Background(), "testOrg", testEqualValueEntity.Name, testEqualValueEntity)
	es.Delete(context.Background(), "testOrg", testTimeBeforeEntity.Name, testTimeBeforeEntity)
	es.Delete(context.Background(), "testOrg", testDeletedEntity.Name, testDeletedEntity)
}

func testMixedTypes(t *testing.T, es EntityStore) {

	te := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testEntityMixedTypes",
		},
		Value: "testValue",
	}
	oe := &otherEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "otherEntityMixedTypes",
		},
		Other: "otherValue",
	}
	id, err := es.Add(context.Background(), te)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)
	id, err = es.Add(context.Background(), oe)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	var testEntities []*testEntity
	err = es.List(context.Background(), "testOrg", Options{}, &testEntities)
	assert.NoError(t, err, "Error listing entities")
	assert.Len(t, testEntities, 1)

	var otherEntities []*otherEntity
	err = es.List(context.Background(), "testOrg", Options{}, &otherEntities)
	assert.NoError(t, err, "Error listing entities")
	assert.Len(t, otherEntities, 1)

	// clean up
	err = es.Delete(context.Background(), "testOrg", "testEntityMixedTypes", te)
	assert.NoError(t, err, "Error clean up")
	err = es.Delete(context.Background(), "testOrg", "otherEntityMixedTypes", oe)
	assert.NoError(t, err, "Error clean up")
}

func Test_getType(t *testing.T) {
	var something interface{} = &BaseEntity{}

	eType := reflect.TypeOf((*Entity)(nil)).Elem()

	assert.True(t, reflect.TypeOf(something).Implements(eType))
}

func testDelete(t *testing.T, es EntityStore) {

	e := &testEntity{
		BaseEntity: BaseEntity{
			OrganizationID: "testOrg",
			Name:           "testEntityDelete",
			Tags: map[string]string{
				"role": "test",
			},
		},
		Value: "testValue",
	}

	id, err := es.Add(context.Background(), e)
	assert.NoError(t, err, "Error adding entity")
	assert.NotNil(t, id)

	err = es.Delete(context.Background(), "testOrg", "testEntityDelete", e)
	assert.NoError(t, err, "Error deleting entity")
	var retreived testEntity
	err = es.Get(context.Background(), "testOrg", "testEntityDelete", Options{}, &retreived)
	assert.Error(t, err)
}
