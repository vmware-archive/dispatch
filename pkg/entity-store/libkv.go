///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package entitystore

import (
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

func (es *libkvEntityStore) UpdateWithError(e Entity, err error) {
	if err != nil {
		e.setStatus(StatusERROR)
		e.setReason([]string{err.Error()})
	}
	if _, err2 := es.Update(e.GetRevision(), e); err2 != nil {
		log.Error(err2)
	}
}

// Add adds new entities to the store
func (es *libkvEntityStore) Add(entity Entity) (id string, err error) {
	err = precondition(entity)
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
func (es *libkvEntityStore) Update(lastRevision uint64, entity Entity) (revision int64, err error) {
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
func (es *libkvEntityStore) Delete(organizationID string, name string, entity Entity) error {
	key := buildKey(organizationID, getDataType(entity), name)
	return es.kv.Delete(key)
}

// Get gets a single entity by name from the store
func (es *libkvEntityStore) Get(organizationID string, name string, entity Entity) error {
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

func doFilterStat(fs FilterStat, entity Entity) (bool, error) {

	rv := reflect.ValueOf(entity).Elem()

	field := rv.FieldByName(fs.Subject)
	if !field.IsValid() {
		return false, errors.Errorf("error filtering: invalid field %s", fs.Subject)
	}

	switch fs.Verb {
	case FilterVerbEqual:
		return reflect.DeepEqual(field.Interface(), fs.Object), nil
	case FilterVerbIn:
		objects := reflect.ValueOf(fs.Object)
		if objects.Kind() != reflect.Slice {
			return false, errors.Errorf("error filtering: object of a 'in' operator must be a slice")
		}
		for i := 0; i < objects.Len(); i++ {
			if reflect.DeepEqual(field.Interface(), objects.Index(i).Interface()) {
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
		subject, ok := field.Interface().(time.Time)
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
	for _, fs := range filter {
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
func (es *libkvEntityStore) List(organizationID string, filter Filter, entities interface{}) error {
	rv := reflect.ValueOf(entities)
	if entities == nil || rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Slice {
		return errors.New("need a non-nil entity slice pointer")
	}
	slice := reflect.MakeSlice(rv.Elem().Type(), 0, 0)

	elemType := rv.Elem().Type().Elem()
	if !elemType.Implements(reflect.TypeOf((*Entity)(nil)).Elem()) {
		return errors.New("non-entity element type: maybe use pointers")
	}

	key := buildKey(organizationID, dataType(elemType.Elem().Name()))
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

		if filter != nil {
			ok, errFilter := doFilter(filter, entity)
			if errFilter != nil {
				log.Print(err)
				// log.Debug(err)
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
