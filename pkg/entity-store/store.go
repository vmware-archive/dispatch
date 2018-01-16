///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package entitystore

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/pkg/errors"
)

const (

	// StatusINITIALIZED objet is INITIALIZED
	//	this status is used by image manager
	StatusINITIALIZED Status = "INITIALIZED"

	// StatusCREATING object is CREATING
	// it is not a guarantee that the object is also created and ready in the underlying driver
	// should wait until the READY state to use the object
	StatusCREATING Status = "CREATING"

	// StatusREADY object is READY to be used
	StatusREADY Status = "READY"

	// StatusUPDATING object is UPDATING
	// it is not a guarantee that changes will be reflected by the underlying driver
	// when updated, will transfer to READY state
	StatusUPDATING Status = "UPDATING"

	// StatusDELETING object is DELETING
	// it is not a guarantee that it has been deleted from the underlying driver
	// user should not reuse the object name until it is transfered to DELETED state
	StatusDELETING Status = "DELETING"

	// StatusDELETED object is DELETED
	// Note for dispatch team:
	// leave this here, reserved for when we use UUID instead of entity name
	StatusDELETED Status = "DELETED"

	// StatusERROR unexpected error state
	// you should not use the object until the state is tranfered to READY
	// or the object is deleted
	StatusERROR Status = "ERROR"
)

// Status represents the current state
type Status string

// Reason represents the reason of current status
type Reason []string

// Spec represents the desired state/status
type Spec map[string]string

// dataType represents the stored struct type
type dataType string

// Tags are filterable metadata as key pairs
type Tags map[string]string

// Entity is the base interface for all stored objects
type Entity interface {
	setID(string)
	setName(string)
	setOrganizationID(string)
	setCreatedTime(time.Time)
	setModifiedTime(time.Time)
	setRevision(uint64)
	setVersion(uint64)
	setSpec(Spec)
	setStatus(Status)
	setReason(Reason)
	setTags(Tags)
	setDelete(bool)

	GetID() string
	GetName() string
	GetOrganizationID() string
	GetCreateTime() time.Time
	GetModifiedTime() time.Time
	GetRevision() uint64
	GetVersion() uint64
	GetSpec() Spec
	GetStatus() Status
	GetReason() Reason
	GetTags() Tags
	GetDelete() bool
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
	Reason         Reason    `json:"reason"`
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

func GetDataType(entity Entity) string {
	return string(getDataType(entity))
}

func (e *BaseEntity) setID(id string) {
	e.ID = id
}
func (e *BaseEntity) setName(name string) {
	e.Name = name
}
func (e *BaseEntity) setOrganizationID(o string) {
	e.OrganizationID = o
}
func (e *BaseEntity) setCreatedTime(createdTime time.Time) {
	e.CreatedTime = createdTime
}

func (e *BaseEntity) setModifiedTime(modifiedTime time.Time) {
	e.ModifiedTime = modifiedTime
}

func (e *BaseEntity) setRevision(revision uint64) {
	e.Revision = revision
}
func (e *BaseEntity) setVersion(version uint64) {
	e.Version = version
}
func (e *BaseEntity) setSpec(spec Spec) {
	e.Spec = spec
}

func (e *BaseEntity) setStatus(status Status) {
	e.Status = status
}

func (e *BaseEntity) setReason(reason Reason) {
	e.Reason = reason
}

func (e *BaseEntity) setTags(tags Tags) {
	e.Tags = tags
}
func (e *BaseEntity) setDelete(delete bool) {
	e.Delete = delete
}

// getKey builds the key for a give entity
func (e *BaseEntity) getKey(dt dataType) string {
	return buildKey(e.OrganizationID, dt, e.Name)
}

func (e *BaseEntity) GetID() string {
	return e.ID
}

// GetName retreives the entity name
func (e *BaseEntity) GetName() string {
	return e.Name
}

func (e *BaseEntity) GetOrganizationID() string {
	return e.OrganizationID
}

func (e *BaseEntity) GetCreateTime() time.Time {
	return e.CreatedTime
}

func (e *BaseEntity) GetModifiedTime() time.Time {
	return e.ModifiedTime
}

func (e *BaseEntity) GetRevision() uint64 {
	return e.Revision
}

func (e *BaseEntity) GetVersion() uint64 {
	return e.Version
}

func (e *BaseEntity) GetStatus() Status {
	return e.Status
}

func (e *BaseEntity) GetDelete() bool {
	return e.Delete
}

func (e *BaseEntity) GetReason() Reason {
	return e.Reason
}
func (e *BaseEntity) GetSpec() Spec {
	return e.Spec
}

// GetTags retreives the entity tags
func (e *BaseEntity) GetTags() Tags {
	return e.Tags
}

// Filter defines a set of criteria to filter entities when listing
type Filter []FilterStat

// FilterStat (Filter Statement) defines one filter criterion
type FilterStat struct {
	Subject string
	Verb    FilterVerb
	Object  interface{}
}

// FilterVerb describe the filter verb
type FilterVerb string

const (
	// FilterVerbIn tests containment
	FilterVerbIn FilterVerb = "in"

	// FilterVerbEqual tests equality
	FilterVerbEqual FilterVerb = "equal"

	// FilterVerbBefore tests two time.Time
	FilterVerbBefore FilterVerb = "before"

	// FilterVerbAfter tests two time.Time
	FilterVerbAfter FilterVerb = "after"
)

// type Filter func(Entity) bool

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
	// UpdateWithError is used by entity handlers to save changes and/or error status
	// e.g. `defer func() { h.store.UpdateWithError(e, err) }()`
	UpdateWithError(e Entity, err error)
}

// BackendConfig list a set of configuration values for backend DB
type BackendConfig struct {
	Backend  string
	Address  string
	Username string
	Password string
	Bucket   string
}

// NewFromBackend creates new entity store created from a backend DB
func NewFromBackend(config BackendConfig) (EntityStore, error) {

	switch config.Backend {
	case "postgres":
		es, err := newPostgres(config)
		if err != nil {
			return nil, errors.Wrapf(err, "error creating a(n) %s entity store", config.Backend)
		}
		return es, nil

	case string(store.BOLTDB):
		boltdb.Register()
		kv, err := libkv.NewStore(
			store.Backend(config.Backend),
			[]string{config.Address},
			&store.Config{
				Bucket:            config.Bucket,
				ConnectionTimeout: 1 * time.Second,
				PersistConnection: true,
				Username:          config.Username,
				Password:          config.Password,
			},
		)
		if err != nil {
			return nil, errors.Wrapf(err, "error creating a(n) %s entity store", config.Backend)
		}
		return newLibkv(kv), nil
	default:
		return nil, errors.Errorf("error creating an entity store %s: not supported", config.Backend)
	}
}

func precondition(entity Entity) error {
	var validName = regexp.MustCompile(`^[\w\d\-]+$`)
	if validName.MatchString(entity.GetName()) {
		return nil
	}
	return errors.Errorf("Invalid name %s, names may only contain letters, numbers, underscores and dashes", entity.GetName())
}
