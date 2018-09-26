///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package backend

import (
	"strconv"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	knbuild "github.com/knative/build/pkg/apis/build/v1alpha1"
	knserve "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	dapi "github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/functions/config"
	"github.com/vmware/dispatch/pkg/utils/knaming"
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
	source := &knbuild.SourceSpec{
		Custom: &corev1.Container{
			Image:           buildCfg.BuildImage,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{buildCfg.BuildCommand},
			Args:            []string{function.SourceURL, "/workspace"},
		},
	}

	build := &knbuild.BuildSpec{
		Source: source,
		Template: &knbuild.TemplateInstantiationSpec{
			Name: buildCfg.BuildTemplate,
			Arguments: []knbuild.ArgumentSpec{
				knbuild.ArgumentSpec{Name: "DESTINATION_IMAGE", Value: function.FunctionImageURL},
				knbuild.ArgumentSpec{Name: "SOURCE_IMAGE", Value: function.ImageURL},
				knbuild.ArgumentSpec{Name: "HANDLER", Value: function.Handler},
			},
		},
		ServiceAccountName: buildCfg.ServiceAccount,
		Timeout:            metav1.Duration{Duration: time.Minute * 10},
	}

	// Add volume mount for copying files from
	if buildCfg.StorageConfig.Storage == config.File {
		source.Custom.VolumeMounts = []corev1.VolumeMount{
			corev1.VolumeMount{
				Name:      "function-store",
				ReadOnly:  true,
				MountPath: "/store",
			},
		}
		build.Volumes = []corev1.Volume{
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
		}
	}
	// Pass in minio/s3 config (should use secrets)
	if buildCfg.StorageConfig.Storage == config.Minio {
		source.Custom.Env = []corev1.EnvVar{
			corev1.EnvVar{Name: "MINIO_USER", Value: buildCfg.StorageConfig.Minio.Username},
			corev1.EnvVar{Name: "MINIO_PASSWORD", Value: buildCfg.StorageConfig.Minio.Password},
		}
	}

	// This is hackery!
	unstructuredBuild, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(&build)
	unstructuredBuild["kind"] = "Build"
	unstructuredBuild["apiVersion"] = "build.knative.dev/v1alpha1"

	return &knserve.Service{
		ObjectMeta: knaming.ToObjectMeta(function.Meta, *function),
		Spec: knserve.ServiceSpec{
			RunLatest: &knserve.RunLatestType{
				Configuration: knserve.ConfigurationSpec{
					Build: &unstructured.Unstructured{
						Object: unstructuredBuild,
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
							// TODO define a service account per function
							ServiceAccountName: buildCfg.ServiceAccount,
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
