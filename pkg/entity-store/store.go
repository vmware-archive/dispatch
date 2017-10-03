///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package entitystore

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
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
	GetName() string
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
	Delete         bool      `json:"delete"`
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

// GetName retreives the entity name
func (e *BaseEntity) GetName() string {
	return e.Name
}

// GetTags retreives the entity tags
func (e *BaseEntity) GetTags() Tags {
	return e.Tags
}

// getKey builds the key for a give entity
func (e *BaseEntity) getKey(dt dataType) string {
	return buildKey(e.OrganizationID, dt, e.Name)
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
	// GetById gets a single entity by key from the store
	Get(organizationID string, key string, entity Entity) error
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

func (es *entityStore) precondition(entity Entity) error {
	var validName = regexp.MustCompile(`^[\w\d\-]+$`)
	if validName.MatchString(entity.GetName()) {
		return nil
	}
	return errors.Errorf("Invalid name %s, names may only contain letters, numbers, underscores and dashes", entity.GetName())
}

// Add adds new entities to the store
func (es *entityStore) Add(entity Entity) (id string, err error) {
	err = es.precondition(entity)
	if err != nil {
		return "", errors.Wrap(err, "Precondition failed")
	}

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

	_, resp, err := es.kv.AtomicPut(key, data, nil, &store.WriteOptions{IsDir: false})
	if err != nil {
		return "", err
	}
	entity.setRevision(resp.LastIndex)
	return id, nil
}

// Update updates existing entities to the store
func (es *entityStore) Update(lastRevision uint64, entity Entity) (revision int64, err error) {
	key := getKey(entity)

	exists, err := es.kv.Exists(key)
	if !exists {
		return 0, errors.Errorf("Entity not found, cannot update")
	}
	if err != nil {
		return 0, err
	}

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
func (es *entityStore) Delete(organizationID string, name string, entity Entity) error {
	key := buildKey(organizationID, getDataType(entity), name)
	return es.kv.Delete(key)
}

// Get gets a single entity by name from the store
func (es *entityStore) Get(organizationID string, name string, entity Entity) error {
	key := buildKey(organizationID, getDataType(entity), name)
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
		if err == store.ErrKeyNotFound {
			rv.Elem().Set(slice)
			return nil
		}
		return err
	}
	for _, kv := range kvs {
		obj := reflect.New(elemType)
		entity := obj.Interface().(Entity)
		err = json.Unmarshal(kv.Value, entity)
		if err != nil {
			return errors.Wrap(err, "deserialization error, while listing")
		}

		if filter != nil {
			if !filter(entity) {
				continue
			}
		}
		entity.setRevision(kv.LastIndex)

		slice = reflect.Append(slice, obj.Elem())
	}
	rv.Elem().Set(slice)

	return nil
}
