///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functionmanager

import (
	"strings"

	kntypes "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/pkg/errors"
	dapi "github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/utils/knaming"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

//FromFunction produces a Knative Service from a Dispatch Function
func FromFunction(function *dapi.Function) *kntypes.Service {
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

//ToFunction produces a Dispatch Function from a Knative Service
func ToFunction(service *kntypes.Service) *dapi.Function {
	objMeta := &service.ObjectMeta
	var function dapi.Function
	if err := knaming.FromJSONString(objMeta.Labels[knaming.InitialObjectAnnotation], &function); err != nil {
		// TODO the right thing
		panic(errors.Wrap(err, "decoding into function"))
	}
	return &function
}
