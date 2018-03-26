---
---

# Dispatch Development Guide

This document outlines the rules and conventions which should be adhered to across all of the dispatch code base.
The guidelines below apply to both new and existing code. Any deviations in existing code should be raised as issues
and addressed as bugs.

This is a living document. If there is a question around convention or consistency which this document does not
address, please raise an issue and this document will be updated to reflect the agreed upon course of action.

## Object/Entity Model

Entities are the structures which represent state. They are the objects/documents/rows in the database (depending
on implementation). The **Entity Store** is an abstraction around different database backends and implements a fairly
basic key/value interface with some limited querying and filtering.

### Base Entity

All entities have a base set of fields which are common accross all entity types.

```go
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
```

#### Kind

> TODO: Should we incude and explicit `kind` field?

#### Key

Entities are referenced by key. The format of the key is the same for all entities and is as follows:

    <orgnaization>/<kind>/<name>

#### Name and IDs

* All objects have both a `name` and `ID` field
* `name` is used as the **key** (namespaced with organization and entity type)
* No two entities may have the same key at the same time
    - Names, and therefore keys, can be reused
* IDs are globally unique for all time
* Relalationships (i.e. foreign keys) use `name` unless explicitly indicated (i.e. `relatedID`)
    * Use the entity type as the field name (no need for "Name" as a suffix, it's redundant)
* External services that reference Dispatch entities may use `ID` or a combination of `name` and `ID` as they often
  have different standards with regards to acceptable names/IDs and may require that names/IDs are unique for all
  time.

#### OrganizationID

> TODO: should this be renamed simply to Orgnaization?

All objects belong to a single organization. This is used as the top-level token in the construction of an entity key.

#### Spec

> TODO: spec field is currently unused and the status field is overloaded to handle desired state

The desired state of an object should be captured in the spec field.

#### Status

The current state of an object. Status is represented by an enumeration of strings:

    // StatusINITIALIZED object is INITIALIZED
	// it is the initial state of the object
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

	// StatusMISSING temporary error state
	// Used when external resources cannot be found
	StatusMISSING Status = "MISSING"

#### Reason

The `reason` field is used in conjunction with status, most often the ERROR status. Its intent is to give context to the
status (i.e. the reasons behind the error).

### Additional Properties

Each entity type inherits from the base entity and includes all of the base entity fields. The derived entities may
include any additional fields needed for to fully describe the entity. The only rule is that the common conventions are
applied for consistency:

* Foreign keys fields should simply be the name of the referenced object (i.e. `secret` or `secrets` not `secretName` or `secretNames`)
* Foreign key values should reflect the `name` value of the entity (as opposed to the `ID` value). The database is indexed on `name`.
    * Should the value be the constructed key (i.e. `<orgnaization>/<kind>/<name>`)?
* Fields should be camel-case and follow the go naming conventions.

## APIs