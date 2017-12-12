///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package i18n

// NO TEST

import (
	"fmt"
)

// T is a placeholder for string translation. In future, the implementation
// should return the translated a string
func T(defaultValue string) string {
	return defaultValue
}

// Errorf produces an error with a translated error string.
// Substitution is performed via the `T` function above.
func Errorf(defaultValue string, args ...interface{}) error {
	return fmt.Errorf(T(defaultValue), args...)
}
