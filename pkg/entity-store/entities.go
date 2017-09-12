///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package entitystore

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/docker/libkv/store"
	uuid "github.com/satori/go.uuid"
)

// State represents the current state
type State string

// Status represents the desired state/status
type Status map[string]string

// DataType represents the stored struct type
type DataType string

// Tags are filterable metadata as key pairs
type Tags map[string]string

// StoredEntity is the base interface for all stored objects
type StoredEntity interface {
	setID(string)
	setCreatedTime(time.Time)
	setModifiedTime(time.Time)
	setRevision(uint64)
	GetType() DataType
	GetTags() Tags
	GetKey(DataType) string
}

// Entity is the base struct for all stored objects
type Entity struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	OrganizationID string    `json:"organizationId"`
	CreatedTime    time.Time `json:"createdTime,omitempty"`
	ModifiedTime   time.Time `json:"modifiedTime,omitempty"`
	Revision       uint64    `json:"revison"`
	Version        uint64    `json:"version"`
	State          State     `json:"state"`
	Status         Status    `json:"status"`
	Tags           Tags      `json:"tags"`
}

// BuildKey is a utility for building the object key (also works for directories)
func BuildKey(organizationID string, dataType DataType, id ...string) string {
	sub := strings.Join(id, "/")
	return fmt.Sprintf("%s/%s/%s", organizationID, dataType, sub)
}

func (e *Entity) setID(id string) {
	e.ID = id
}

func (e *Entity) setRevision(revision uint64) {
	e.Revision = revision
}

func (e *Entity) setCreatedTime(createdTime time.Time) {
	e.CreatedTime = createdTime
}

func (e *Entity) setModifiedTime(modifiedTime time.Time) {
	e.ModifiedTime = modifiedTime
}

// GetTags retreives the entity tags
func (e *Entity) GetTags() Tags {
	return e.Tags
}

// GetKey builds the key for a give entity
func (e *Entity) GetKey(dataType DataType) string {
	return BuildKey(e.OrganizationID, dataType, e.ID)
}

// TypeMap is a mapping between data type constants and the reflect Type
type TypeMap map[DataType]reflect.Type

// EntityStore is a wrapper around libkv and provides convenience methods to
// serializing and deserializing objects
type EntityStore struct {
	kv      store.Store
	typeMap TypeMap
}

// NewEntityStore is the EntityStore constructor
func NewEntityStore(kv store.Store, typeMap TypeMap) *EntityStore {
	return &EntityStore{
		kv:      kv,
		typeMap: typeMap,
	}
}

// AddEntity adds new entities to the store
func (es *EntityStore) AddEntity(entity StoredEntity) (id string, err error) {
	id = uuid.NewV4().String()
	entity.setID(id)

	key := entity.GetKey(entity.GetType())

	now := time.Now()
	entity.setCreatedTime(now)
	entity.setModifiedTime(now)

	data, err := json.Marshal(entity)
	if err != nil {
		return "", err
	}

	err = es.kv.Put(key, data, &store.WriteOptions{IsDir: false})
	if err != nil {
		return "", err
	}
	return id, nil
}

// PutEntity updates existing entities to the store
func (es *EntityStore) PutEntity(lastRevision uint64, entity StoredEntity) (revision int64, err error) {
	key := entity.GetKey(entity.GetType())

	entity.setModifiedTime(time.Now())
	data, err := json.Marshal(entity)
	if err != nil {
		return 0, err
	}

	previous := &store.KVPair{
		Key:       key,
		LastIndex: lastRevision,
	}
	_, kv, err := es.kv.AtomicPut(key, data, previous, &store.WriteOptions{IsDir: false})
	if err != nil {
		return 0, err
	}
	entity.setRevision(kv.LastIndex)
	return int64(kv.LastIndex), nil
}

// GetEntityById gets a single entity by id from the store
func (es *EntityStore) GetEntityById(organizationID string, dataType DataType, id string, entity StoredEntity) error {
	key := BuildKey(organizationID, dataType, id)
	kv, err := es.kv.Get(key)
	if err != nil {
		return err
	}
	err = json.Unmarshal(kv.Value, entity)
	if err != nil {
		return err
	}
	entity.setRevision(kv.LastIndex)
	return nil
}

// Filter is a function type that operates on returned list results
type Filter func(StoredEntity) bool

// ListEntities fetches a list of entities of a single data type.  The result may be safely asserted.
func (es *EntityStore) ListEntities(organizationID string, dataType DataType, filter Filter) (interface{}, error) {
	key := BuildKey(organizationID, dataType)
	kvs, err := es.kv.List(key)
	if err != nil {
		return nil, err
	}

	slice := reflect.MakeSlice(reflect.SliceOf(es.typeMap[dataType]), 0, 0)

	for _, kv := range kvs {
		obj := reflect.New(es.typeMap[dataType])
		err = json.Unmarshal(kv.Value, obj.Interface())

		if filter != nil {
			if filter(obj.Interface().(StoredEntity)) {
				continue
			}
		}

		slice = reflect.Append(slice, obj.Elem())
	}
	return slice.Interface(), nil
}
