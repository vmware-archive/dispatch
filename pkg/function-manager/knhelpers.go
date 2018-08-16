///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functionmanager

import (
	kntypes "github.com/knative/serving/pkg/apis/serving/v1alpha1"

	"github.com/vmware/dispatch/pkg/api/v1"
)

func ToKnService(function *v1.Function) *kntypes.Service {
	panic("impl me")
}

func FromKnService(service *kntypes.Service) *v1.Function {
	panic("impl me")
}
