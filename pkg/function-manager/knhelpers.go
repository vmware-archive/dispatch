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
)

//FromFunction produces a Knative Service from a Dispatch Function
func FromFunction(function *dapi.Function) *kntypes.Service {
	probe := &v1.Probe{
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path: "/healthz",
			},
		},
	}
	envVars := append(
		fromSecrets(function.Secrets, function.Meta),
		v1.EnvVar{Name: "SERVERS", Value: "1"},
		v1.EnvVar{Name: "SECRETS", Value: strings.Join(function.Secrets, ",")},
	)
	return &kntypes.Service{
		ObjectMeta: *knaming.ToObjectMeta(function.Meta, function),
		Spec: kntypes.ServiceSpec{
			RunLatest: &kntypes.RunLatestType{
				Configuration: kntypes.ConfigurationSpec{
					RevisionTemplate: kntypes.RevisionTemplateSpec{
						Spec: kntypes.RevisionSpec{
							Container: v1.Container{
								Image:          function.FunctionImageURL,
								Env:            envVars,
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

func fromSecrets(secrets []string, meta dapi.Meta) []v1.EnvVar {
	var r []v1.EnvVar
	for _, secret := range secrets {
		meta := meta
		meta.Name = secret
		r = append(r, v1.EnvVar{
			Name: knaming.SecretEnvVarName(secret),
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: knaming.SecretName(meta),
					},
					Key: knaming.TheSecretKey,
				},
			},
		})
	}
	return r
}

//ToFunction produces a Dispatch Function from a Knative Service
func ToFunction(service *kntypes.Service) *dapi.Function {
	objMeta := &service.ObjectMeta
	var function dapi.Function
	if err := knaming.FromJSONString(objMeta.Annotations[knaming.InitialObjectAnnotation], &function); err != nil {
		// TODO the right thing
		panic(errors.Wrap(err, "decoding into function"))
	}
	return &function
}
