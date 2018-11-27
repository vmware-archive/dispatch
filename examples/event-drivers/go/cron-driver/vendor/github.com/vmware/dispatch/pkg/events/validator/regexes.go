///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package validator

import "regexp"

var (
	eventTypeRegexString = "^[\\w\\d\\.\\-]+$"
)

var (
	eventTypeRegex = regexp.MustCompile(eventTypeRegexString)
)
