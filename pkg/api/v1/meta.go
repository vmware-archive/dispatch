///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package v1

type Meta struct {
	// Name
	// Required: true
	// Pattern: ^[\w\d][\w\d\-]*$
	Name string `json:"name"`

	// Project
	Project string
}
