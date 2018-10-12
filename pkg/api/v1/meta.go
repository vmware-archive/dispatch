///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package v1

import strfmt "github.com/go-openapi/strfmt"

// NO TESTS

//Meta holds common metadata for API objects
type Meta struct {
	// Kind
	// Read Only: true
	// Pattern: ^[\w\d\-]+$
	Kind string `json:"kind,omitempty"`

	// Name
	// Required: true
	// Pattern: ^[\w\d][\w\d\-]*[\w\d]|[\w\d]+$
	Name string `json:"name"`

	// ID
	ID strfmt.UUID `json:"id,omitempty"`

	// Project
	// Pattern: ^[\w\d][\w\d\-]*[\w\d]|[\w\d]+$
	// Default: default
	Project string `json:"project,omitempty"`

	// Org
	// Pattern: ^[\w\d][\w\d\-]*[\w\d]|[\w\d]+$
	// Default: default
	Org string `json:"org,omitempty"`

	// CreatedTime
	// ReadOnly: true
	CreatedTime int64 `json:"createdTime,omitempty"`

	// ModifiedTime
	// ReadOnly: true
	ModifiedTime int64 `json:"modifiedTime,omitempty"`

	// Revision
	// ReadOnly: true
	Revision string `json:"revision,omitemtpy"`

	// Tags
	Tags []*Tag `json:"tags,omitempty"`

	// BackingObject
	// ReadOnly: true
	BackingObject interface{} `json:"backingObject,omitempty"`
}
