///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functionmanager

import (
	"strings"

	kntypes "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	dapi "github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/utils/knaming"
)

func ToKnService(function *dapi.Function) *kntypes.Service {
	probe := &v1.Probe{
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path: "/healthz",
				Port: intstr.FromInt(8080),
			},
		},
	}
	return &kntypes.Service{
		ObjectMeta: *knaming.ToObjectMeta(function.Meta, function),
		Spec: kntypes.ServiceSpec{
			RunLatest: &kntypes.RunLatestType{
				Configuration: kntypes.ConfigurationSpec{
					RevisionTemplate: kntypes.RevisionTemplateSpec{
						Spec: kntypes.RevisionSpec{
							Container: v1.Container{
								Image: function.FunctionImageURL,
								Env: []v1.EnvVar{
									{Name: "SERVERS", Value: "1"},
									{Name: "SECRETS", Value: strings.Join(function.Secrets, ",")},
								},
								LivenessProbe:  probe,
								ReadinessProbe: probe,
							},
							ConcurrencyModel:   kntypes.RevisionRequestConcurrencyModelSingle,
							ServiceAccountName: function.Meta.Project, // TODO now it's the default service-account for the project
						},
					},
				},
			},
		},
	}
}

func FromKnService(service *kntypes.Service) *dapi.Function {
	panic("impl me") // TODO impl
}
