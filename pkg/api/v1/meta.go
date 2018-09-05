///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package v1

//Meta holds common metadata for API objects
type Meta struct {
	// Name
	// Required: true
	// Pattern: ^[\w\d][\w\d\-]*[\w\d]|[\w\d]+$
	Name string `json:"name"`

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

	// BackingObject
	// ReadOnly: true
	BackingObject interface{} `json:"backingObject,omitempty"`
}
