///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package identitymanager

// NO TESTS

type attributesRecord struct {
	subject           string
	resource          string
	path              string
	action            Action
	isResourceRequest bool
}
