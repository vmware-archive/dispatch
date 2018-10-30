///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package utils

import "io"

// Close provides a safe Close facility
func Close(i interface{}) {
	if c, ok := i.(io.Closer); ok {
		c.Close()
	}
}
