///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package v1

import (
	strfmt "github.com/go-openapi/strfmt"
)

// NO TESTS

// SecretValue secret value
// swagger:model SecretValue
type SecretValue map[string]string

// Validate validates this secret value
func (m SecretValue) Validate(formats strfmt.Registry) error {
	return nil
}
