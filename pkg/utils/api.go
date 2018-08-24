///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package utils

import (
	dapi "github.com/vmware/dispatch/pkg/api/v1"
)

//AdjustMeta replaces default values of Org and Project fields of Meta with specified values
func AdjustMeta(meta *dapi.Meta, from dapi.Meta) {
	meta.Org = from.Org
	meta.Project = from.Project
}
