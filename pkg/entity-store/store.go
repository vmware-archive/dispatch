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
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
)

// Status represents the current state
type Status string

// Spec represents the desired state/status
type Spec map[string]string

// dataType represents the stored struct type
type dataType string

// Tags are filterable metadata as key pairs
type Tags map[string]string

// Entity is the base interface for all stored objects
type Entity interface {
	setID(string)
	setCreatedTime(time.Time)
	setModifiedTime(time.Time)
	setRevision(uint64)
	GetTags() Tags
	getKey(dataType) string
}

// BaseEntity is the base struct for all stored objects
type BaseEntity struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	OrganizationID string    `json:"organizationId"`
	CreatedTime    time.Time `json:"createdTime,omitempty"`
	ModifiedTime   time.Time `json:"modifiedTime,omitempty"`
	Revision       uint64    `json:"revision"`
	Version        uint64    `json:"version"`
	Spec           Spec      `json:"state"`
	Status         Status    `json:"status"`
	Reason         []string  `json:"reason"`
	Tags           Tags      `json:"tags"`
}

// buildKey is a utility for building the object key (also works for directories)
func buildKey(organizationID string, dt dataType, id ...string) string {
	sub := strings.Join(id, "/")
	return fmt.Sprintf("%s/%s/%s", organizationID, dt, sub)
}

func getKey(entity Entity) string {
	return entity.getKey(getDataType(entity))
}

func getDataType(entity Entity) dataType {
	return dataType(reflect.ValueOf(entity).Type().Elem().Name())
}

func (e *BaseEntity) setID(id string) {
	e.ID = id
}

func (e *BaseEntity) setRevision(revision uint64) {
	e.Revision = revision
}

func (e *BaseEntity) setCreatedTime(createdTime time.Time) {
	e.CreatedTime = createdTime
}

func (e *BaseEntity) setModifiedTime(modifiedTime time.Time) {
	e.ModifiedTime = modifiedTime
}

// GetTags retreives the entity tags
func (e *BaseEntity) GetTags() Tags {
	return e.Tags
}

// getKey builds the key for a give entity
func (e *BaseEntity) getKey(dt dataType) string {
	return buildKey(e.OrganizationID, dt, e.ID)
}

type entityStore struct {
	kv store.Store
}

// Filter is a function type that operates on returned list results
type Filter func(Entity) bool

// EntityStore is a wrapper around libkv and provides convenience methods to
// serializing and deserializing objects
type EntityStore interface {
	// Add adds new entities to the store
	Add(entity Entity) (id string, err error)
	// Update updates existing entities to the store
	Update(lastRevision uint64, entity Entity) (revision int64, err error)
	// GetById gets a single entity by id from the store
	GetById(organizationID string, id string, entity Entity) error
	// List fetches a list of entities of a single data type satisfying the filter.
	// entities is a placeholder for results and must be a pointer to an empty slice of the desired entity type.
	List(organizationID string, filter Filter, entities interface{}) error
	// Delete delets a single entity from the store.
	Delete(organizationID string, id string, entity Entity) error
}

// New is the EntityStore constructor
func New(kv store.Store) EntityStore {
	return &entityStore{
		kv: kv,
	}
}

// Add adds new entities to the store
func (es *entityStore) Add(entity Entity) (id string, err error) {
	id = uuid.NewV4().String()
	entity.setID(id)

	key := getKey(entity)

	now := time.Now()
	entity.setCreatedTime(now)
	entity.setModifiedTime(now)

	data, err := json.Marshal(entity)
	if err != nil {
		return "", errors.Wrap(err, "serialization error, before adding")
	}

	err = es.kv.Put(key, data, &store.WriteOptions{IsDir: false})
	if err != nil {
		return "", err
	}
	return id, nil
}

// Update updates existing entities to the store
func (es *entityStore) Update(lastRevision uint64, entity Entity) (revision int64, err error) {
	key := getKey(entity)

	entity.setModifiedTime(time.Now())
	data, err := json.Marshal(entity)
	if err != nil {
		return 0, errors.Wrap(err, "serialization error, before updating")
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

// Delete delets a single entity from the store
// entity should be a zero-value of entity to be deleted.
func (es *entityStore) Delete(organizationID string, id string, entity Entity) error {
	key := buildKey(organizationID, getDataType(entity), id)
	return es.kv.Delete(key)
}

// GetById gets a single entity by id from the store
func (es *entityStore) GetById(organizationID string, id string, entity Entity) error {
	key := buildKey(organizationID, getDataType(entity), id)
	kv, err := es.kv.Get(key)
	if err != nil {
		return err
	}
	err = json.Unmarshal(kv.Value, entity)
	if err != nil {
		return errors.Wrap(err, "deserialization error, while getting")
	}
	entity.setRevision(kv.LastIndex)
	return nil
}

// List fetches a list of entities of a single data type satisfying the filter.
// entities is a placeholder for results and must be a pointer to an empty slice of the desired entity type.
func (es *entityStore) List(organizationID string, filter Filter, entities interface{}) error {
	rv := reflect.ValueOf(entities)
	if entities == nil || rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Slice {
		return errors.New("need a non-nil slice pointer")
	}
	slice := reflect.MakeSlice(rv.Elem().Type(), 0, 0)

	elemType := rv.Elem().Type().Elem()

	key := buildKey(organizationID, dataType(elemType.Name()))
	kvs, err := es.kv.List(key)
	if err != nil {
		return err
	}
	for _, kv := range kvs {
		obj := reflect.New(elemType)
		err = json.Unmarshal(kv.Value, obj.Interface())
		if err != nil {
			return errors.Wrap(err, "deserialization error, while listing")
		}

		if filter != nil {
			if !filter(obj.Interface().(Entity)) {
				continue
			}
		}

		slice = reflect.Append(slice, obj.Elem())
	}
	rv.Elem().Set(slice)

	return nil
}
