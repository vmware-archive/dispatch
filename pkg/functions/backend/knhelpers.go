///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package backend

import (
	"strconv"
	"strings"

	"github.com/go-openapi/strfmt"
	knbuild "github.com/knative/build/pkg/apis/build/v1alpha1"
	knserve "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/pkg/errors"
	dapi "github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/utils/knaming"
	corev1 "k8s.io/api/core/v1"
)

//FromFunction produces a Knative Service from a Dispatch Function
func FromFunction(buildCfg *BuildConfig, function *dapi.Function) *knserve.Service {
	if function == nil {
		return nil
	}
	probe := &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/healthz",
			},
		},
		InitialDelaySeconds: 0,
	}
	envVars := append(
		fromSecrets(function.Secrets, function.Meta),
		corev1.EnvVar{Name: "SERVERS", Value: "1"},
		corev1.EnvVar{Name: "SECRETS", Value: strings.Join(function.Secrets, ",")},
		corev1.EnvVar{Name: "TIMEOUT", Value: strconv.FormatInt(function.Timeout, 10)},
	)

	return &knserve.Service{
		ObjectMeta: knaming.ToObjectMeta(function.Meta, *function),
		Spec: knserve.ServiceSpec{
			RunLatest: &knserve.RunLatestType{
				Configuration: knserve.ConfigurationSpec{
					Build: &knbuild.BuildSpec{
						// This source spec assumes that function data is stored on an attached
						// persistent volume.  It would be good to allow other source stores such
						// as http accessible blobs.
						Source: &knbuild.SourceSpec{
							Custom: &corev1.Container{
								Image:           buildCfg.BuildImage,
								ImagePullPolicy: corev1.PullIfNotPresent,
								Command:         []string{buildCfg.BuildCommand},
								Args:            []string{function.SourceURL, "/workspace"},
								VolumeMounts: []corev1.VolumeMount{
									corev1.VolumeMount{
										Name:      "function-store",
										ReadOnly:  true,
										MountPath: "/store",
									},
								},
							},
						},
						Template: &knbuild.TemplateInstantiationSpec{
							Name: buildCfg.BuildTemplate,
							Arguments: []knbuild.ArgumentSpec{
								knbuild.ArgumentSpec{Name: "DESTINATION_IMAGE", Value: function.FunctionImageURL},
								knbuild.ArgumentSpec{Name: "SOURCE_IMAGE", Value: function.ImageURL},
								knbuild.ArgumentSpec{Name: "HANDLER", Value: function.Handler},
							},
						},
						ServiceAccountName: buildCfg.ServiceAccount,
						Volumes: []corev1.Volume{
							corev1.Volume{
								// TODO: use ID or something unique to avoid collisions
								Name: "function-store",
								VolumeSource: corev1.VolumeSource{
									PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
										ClaimName: "function-store-claim",
										ReadOnly:  true,
									},
								},
							},
						},
					},
					RevisionTemplate: knserve.RevisionTemplateSpec{
						Spec: knserve.RevisionSpec{
							Container: corev1.Container{
								Image:          function.FunctionImageURL,
								Env:            envVars,
								LivenessProbe:  probe,
								ReadinessProbe: probe,
							},
							ContainerConcurrency: 1,
							//ServiceAccountName: function.Meta.Project, // TODO define a service account per function
						},
					},
				},
			},
		},
	}
}

func fromSecrets(secrets []string, meta dapi.Meta) []corev1.EnvVar {
	var r []corev1.EnvVar
	for _, secret := range secrets {
		meta := meta
		meta.Name = secret
		r = append(r, corev1.EnvVar{
			Name: knaming.SecretEnvVarName(secret),
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
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
func ToFunction(service *knserve.Service) *dapi.Function {
	if service == nil {
		return nil
	}
	objMeta := &service.ObjectMeta
	var function dapi.Function
	if err := knaming.FromJSONString(objMeta.Annotations[knaming.InitialObjectAnnotation], &function); err != nil {
		// TODO the right thing
		panic(errors.Wrap(err, "decoding into function"))
	}
	function.CreatedTime = service.CreationTimestamp.Unix()

	function.Kind = dapi.FunctionKind
	function.ID = strfmt.UUID(objMeta.UID)
	function.ModifiedTime = objMeta.CreationTimestamp.Unix()
	function.Status = dapi.StatusINITIALIZED
	for _, cond := range service.Status.Conditions {
		if cond.Type == knserve.ServiceConditionRoutesReady && cond.Status == "True" {
			function.Status = dapi.StatusREADY
		}
		if cond.LastTransitionTime.Inner.Unix() > function.ModifiedTime {
			function.ModifiedTime = cond.LastTransitionTime.Inner.Unix()
		}
		if cond.Message != "" {
			function.Reason = append(function.Reason, cond.Reason)
		}
	}
	function.BackingObject = service
	return &function
}
