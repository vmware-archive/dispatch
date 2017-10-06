///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
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
