///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package utils

// NO TESTS

import "fmt"

// AppendApplication append an application k
func AppendApplication(tags *[]string, application string) {
	if application != "" {
		*tags = append(*tags, fmt.Sprintf("Application=%s", application))
	}
}
