///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functionmanager

import (
	kntypes "github.com/knative/serving/pkg/apis/serving/v1alpha1"

	dapi "github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/utils/knaming"
)

func ToKnService(function *dapi.Function) *kntypes.Service {
	return &kntypes.Service{
		ObjectMeta: knaming.ToObjectMeta(function.Meta, function),
	}
}

func FromKnService(service *kntypes.Service) *dapi.Function {
	panic("impl me") // TODO impl
}
