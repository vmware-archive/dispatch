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
	if from.Name != "" {
		meta.Name = from.Name
	}
	if from.Org != "" {
		meta.Org = from.Org
	}
	if from.Project != "" {
		meta.Project = from.Project
	}
	if from.CreatedTime != 0 {
		meta.CreatedTime = from.CreatedTime
	}
}
