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

	// StatusINITIALIZED object is INITIALIZED
	//	this status is used by image manager
	StatusINITIALIZED Status = "INITIALIZED"

	// StatusCREATING object is CREATING
	// it is not a guarantee that the object is also created and ready in the underlying driver
	// should wait until the READY state to use the object
	StatusCREATING Status = "CREATING"

	// StatusINTRANSIT object is currently being processed
	// and should NOT be picked up by periodic sync
	// TODO(imikushin) GC stale inflight objects - in case responsible process terminates (e.g. using process UUIDs)
	StatusINTRANSIT Status = "INTRANSIT"

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

	// StatusMISSING temporary error state
	// Used when external resources cannot be found
	StatusMISSING Status = "MISSING"
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
	SetSpec(Spec)
	SetStatus(Status)
	SetReason(Reason)
	SetTags(Tags)
	SetDelete(bool)

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

// GetDataType returns the data type of the given entity
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

// SetSpec sets the entity spec
func (e *BaseEntity) SetSpec(spec Spec) {
	e.Spec = spec
}

// SetStatus sets the entity status
func (e *BaseEntity) SetStatus(status Status) {
	e.Status = status
}

// SetReason sets the entity reason
func (e *BaseEntity) SetReason(reason Reason) {
	e.Reason = reason
}

// SetTags sets the entity tags
func (e *BaseEntity) SetTags(tags Tags) {
	e.Tags = tags
}

// SetDelete sets the entity delete status
func (e *BaseEntity) SetDelete(delete bool) {
	e.Delete = delete
}

// getKey builds the key for a give entity
func (e *BaseEntity) getKey(dt dataType) string {
	return buildKey(e.OrganizationID, dt, e.Name)
}

// GetID gets the entity ID
func (e *BaseEntity) GetID() string {
	return e.ID
}

// GetName retreives the entity name
func (e *BaseEntity) GetName() string {
	return e.Name
}

// GetOrganizationID gets the entity organizationID
func (e *BaseEntity) GetOrganizationID() string {
	return e.OrganizationID
}

// GetCreateTime gets the entity creation time
func (e *BaseEntity) GetCreateTime() time.Time {
	return e.CreatedTime
}

// GetModifiedTime gets the entity modification time
func (e *BaseEntity) GetModifiedTime() time.Time {
	return e.ModifiedTime
}

// GetRevision gets the entity revision
func (e *BaseEntity) GetRevision() uint64 {
	return e.Revision
}

// GetVersion gets the entity version
func (e *BaseEntity) GetVersion() uint64 {
	return e.Version
}

// GetStatus gets the entity status
func (e *BaseEntity) GetStatus() Status {
	return e.Status
}

// GetDelete gets the entity delete status
func (e *BaseEntity) GetDelete() bool {
	return e.Delete
}

// GetReason gets the entity reason
func (e *BaseEntity) GetReason() Reason {
	return e.Reason
}

// GetSpec gets the entity spec
func (e *BaseEntity) GetSpec() Spec {
	return e.Spec
}

// GetTags retreives the entity tags
func (e *BaseEntity) GetTags() Tags {
	return e.Tags
}

// EntityStore is a wrapper around libkv and provides convenience methods to
// serializing and deserializing objects
type EntityStore interface {
	// Add adds new entities to the store
	Add(entity Entity) (id string, err error)
	// Update updates existing entities to the store
	Update(lastRevision uint64, entity Entity) (revision int64, err error)
	// GetById gets a single entity by key from the store
	Get(organizationID string, key string, opts Options, entity Entity) error
	// List fetches a list of entities of a single data type satisfying the filter.
	// entities is a placeholder for results and must be a pointer to an empty slice of the desired entity type.
	List(organizationID string, opts Options, entities interface{}) error
	// Delete delets a single entity from the store.
	Delete(organizationID string, id string, entity Entity) error
	// UpdateWithError is used by entity handlers to save changes and/or error status
	// e.g. `defer func() { h.store.UpdateWithError(e, err) }()`
	UpdateWithError(e Entity, err error)
}

type uniqueViolation interface {
	UniqueViolation() bool
}

// IsUniqueViolation is a helper function to safely return UniqueViolation if available
func IsUniqueViolation(err error) bool {
	e, ok := errors.Cause(err).(uniqueViolation)
	return ok && e.UniqueViolation()
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
