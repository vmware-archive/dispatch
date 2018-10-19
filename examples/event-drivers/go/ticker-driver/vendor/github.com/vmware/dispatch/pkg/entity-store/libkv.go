///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package entitystore

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/docker/libkv/store"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type libkvEntityStore struct {
	kv store.Store
}

// newLibkv is the EntityStore constructor
func newLibkv(kv store.Store) EntityStore {
	return &libkvEntityStore{
		kv: kv,
	}
}

func (es *libkvEntityStore) UpdateWithError(ctx context.Context, e Entity, err error) {
	if err != nil {
		e.SetStatus(StatusERROR)
		e.SetReason([]string{err.Error()})
	}
	if _, err2 := es.Update(ctx, e.GetRevision(), e); err2 != nil {
		log.Error(err2)
	}
}

type kvUniqueViolation struct {
	key string
}

func (e *kvUniqueViolation) Error() string {
	return "non-unique key: " + e.key
}

func (*kvUniqueViolation) UniqueViolation() bool {
	return true
}

// Add adds new entities to the store
func (es *libkvEntityStore) Add(ctx context.Context, entity Entity) (id string, err error) {
	err = precondition(entity)
	if err != nil {
		return "", errors.Wrap(err, "Precondition failed")
	}

	key := getKey(entity)
	exists, err := es.kv.Exists(key)
	if err != nil && err != store.ErrKeyNotFound {
		return "", errors.Wrap(err, "error checking if the key exists")
	}
	if exists {
		return "", &kvUniqueViolation{key}
	}

	id = uuid.NewV4().String()
	entity.setID(id)

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
func (es *libkvEntityStore) Update(ctx context.Context, lastRevision uint64, entity Entity) (revision int64, err error) {
	if entity.GetOrganizationID() == "" {
		return 0, errors.Errorf("organizationID cannot be empty")
	}
	key := getKey(entity)

	exists, err := es.kv.Exists(key)
	if err != nil && err != store.ErrKeyNotFound {
		return 0, err
	}
	if !exists {
		return 0, errors.Errorf("Entity not found, cannot update")
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

// Delete deletes a single entity from the store
// entity should be a zero-value of entity to be deleted.
func (es *libkvEntityStore) Delete(ctx context.Context, organizationID string, name string, entity Entity) error {
	if organizationID == "" {
		return errors.Errorf("organizationID cannot be empty")
	}
	key := buildKey(getDataType(entity), organizationID, name)
	return es.kv.Delete(key)
}

// SoftDelete marks a single entity for deletion
func (es *libkvEntityStore) SoftDelete(ctx context.Context, entity Entity) error {
	entity.SetDelete(true)
	entity.SetStatus(StatusDELETING)
	_, err := es.Update(ctx, entity.GetRevision(), entity)
	return err
}

// Find gets a single entity by name from the store and returns a touple of found, error
func (es *libkvEntityStore) Find(ctx context.Context, organizationID string, name string, opts Options, entity Entity) (bool, error) {
	if organizationID == "" {
		return false, errors.Errorf("organizationID cannot be empty")
	}
	key := buildKey(getDataType(entity), organizationID, name)
	kv, err := es.kv.Get(key)
	if err != nil {
		if err == store.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	err = json.Unmarshal(kv.Value, entity)
	if err != nil {
		return false, errors.Wrap(err, "deserialization error, while getting")
	}

	if opts.Filter != nil {
		ok, err := doFilter(opts.Filter, entity)
		if err != nil {
			return false, errors.Wrap(err, "error filtering entity")
		}
		if !ok {
			return false, errors.Wrap(err, "error no such entity")
		}
	}
	entity.setRevision(kv.LastIndex)
	return true, nil
}

// Get gets a single entity by name from the store
func (es *libkvEntityStore) Get(ctx context.Context, organizationID string, name string, opts Options, entity Entity) error {
	found, err := es.Find(ctx, organizationID, name, opts, entity)
	if err != nil || !found {
		return errors.New("error getting: no such entity")
	}
	return err
}

func doFilterStat(fs FilterStat, entity Entity) (bool, error) {

	rv := reflect.ValueOf(entity).Elem()

	var subjectValue interface{}
	switch fs.Scope {
	case FilterScopeField, FilterScopeExtra:
		field := rv.FieldByName(fs.Subject)
		if !field.IsValid() {
			return false, errors.Errorf("error filtering: invalid field %s", fs.Subject)
		}
		subjectValue = field.Interface()
	case FilterScopeTag:
		tagField := rv.FieldByName("Tags")
		if !tagField.IsValid() {
			return false, errors.Errorf("unexpected error: Tags field not found in entity %s", entity)
		}
		tags, ok := tagField.Interface().(Tags)
		if !ok {
			return false, errors.Errorf("unexpected error: should be the an instance of type Tags")
		}
		subjectValue = tags[fs.Subject]
	}

	switch fs.Verb {
	case FilterVerbEqual:
		return reflect.DeepEqual(subjectValue, fs.Object), nil
	case FilterVerbIn:
		objects := reflect.ValueOf(fs.Object)
		if objects.Kind() != reflect.Slice {
			return false, errors.Errorf("error filtering: object of a 'in' operator must be a slice")
		}
		for i := 0; i < objects.Len(); i++ {
			if reflect.DeepEqual(subjectValue, objects.Index(i).Interface()) {
				return true, nil
			}
		}
		return false, nil
	case FilterVerbBefore, FilterVerbAfter:
		// must be time.Time
		object, ok := fs.Object.(time.Time)
		if !ok {
			return false, errors.Errorf("error filtering: object of a 'before' or 'after' verb must be an instance of time.Time")
		}
		subject, ok := subjectValue.(time.Time)
		if !ok {
			return false, errors.Errorf("error filtering: subject of a 'before' or 'after' verb must be an instance of time.Time")
		}
		if fs.Verb == FilterVerbBefore {
			return subject.Before(object), nil
		}
		return subject.After(object), nil
	default:
		return false, errors.Errorf("error filtering: invalid verb: %s", fs.Verb)
	}
}

func doFilter(filter Filter, entity Entity) (bool, error) {
	for _, fs := range filter.FilterStats() {
		ok, err := doFilterStat(fs, entity)
		if err != nil {
			log.Debugf("doFilter: error: %s", err)
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

// List fetches a list of entities of a single data type satisfying the filter.
// entities is a placeholder for results and must be a pointer to an empty slice of the desired entity type.
func (es *libkvEntityStore) List(ctx context.Context, organizationID string, opts Options, entities interface{}) error {
	if organizationID == "" {
		return errors.Errorf("organizationID cannot be empty")
	}
	return es.list(ctx, organizationID, opts, entities)
}

// ListGlobal fetches a list of entities of a single data type satisfying the filter across all organizations.
// entities is a placeholder for results and must be a pointer to an empty slice of the desired entity type.
func (es *libkvEntityStore) ListGlobal(ctx context.Context, opts Options, entities interface{}) error {
	return es.list(ctx, "", opts, entities)
}

func (es *libkvEntityStore) list(ctx context.Context, organizationID string, opts Options, entities interface{}) error {

	rv := reflect.ValueOf(entities)
	if entities == nil || rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Slice {
		return errors.New("need a non-nil entity slice pointer")
	}
	slice := reflect.MakeSlice(rv.Elem().Type(), 0, 0)

	elemType := rv.Elem().Type().Elem()
	if !elemType.Implements(reflect.TypeOf((*Entity)(nil)).Elem()) {
		return errors.New("non-entity element type: maybe use pointers")
	}

	key := buildKeyWithoutOrg(DataType(elemType.Elem().Name()))

	kvs, err := es.kv.List(key)
	if err != nil {
		if err == store.ErrKeyNotFound {
			rv.Elem().Set(slice)
			return nil
		}
		return err
	}
	for _, kv := range kvs {
		obj := reflect.New(elemType.Elem())
		entity := obj.Interface().(Entity)
		err = json.Unmarshal(kv.Value, entity)
		if err != nil {
			return errors.Wrap(err, "deserialization error, while listing")
		}

		if opts.Filter != nil {
			ok, errFilter := doFilter(opts.Filter, entity)
			if errFilter != nil {
				log.Print(err)
			}
			if !ok {
				continue
			}
		}
		entity.setRevision(kv.LastIndex)

		slice = reflect.Append(slice, obj)
	}
	rv.Elem().Set(slice)

	return nil
}
