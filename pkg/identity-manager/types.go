///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package identitymanager

// NO TESTS

const (
	subjectUser          subjectKind = "user"
	subjectSvcAccount    subjectKind = "serviceAccount"
	subjectBootstrapUser subjectKind = "bootstrapUser"
)

type subjectKind string

type attributesRecord struct {
	subject           string
	resource          string
	path              string
	action            Action
	isResourceRequest bool
}

type authAccount struct {
	organizationID string
	subject        string
	kind           subjectKind
}
