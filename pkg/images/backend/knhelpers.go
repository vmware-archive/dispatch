///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package backend

import (
	"encoding/json"

	"github.com/go-openapi/strfmt"
	knbuild "github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/pkg/errors"
	dapi "github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/utils"
	"github.com/vmware/dispatch/pkg/utils/knaming"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

// FromImage produced Knative Build from Dispatch Image
func FromImage(imageConfig *ImageConfig, image *dapi.Image) *knbuild.Build {
	if image == nil {
		return nil
	}

	systemPackagesContent, err := json.Marshal(image.SystemDependencies)
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "parsing image with system-dependency"))
		return nil
	}
	runtimePackagesContent, err := json.Marshal(image.RuntimeDependencies)
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "parsing image with runtime-dependency"))
		return nil
	}

	// TODO: Image -> Build make this right and configurable
	return &knbuild.Build{
		ObjectMeta: knaming.ToObjectMeta(image.Meta, *image),
		Spec: knbuild.BuildSpec{
			ServiceAccountName: imageConfig.ServciceAccount,
			Source: &knbuild.SourceSpec{
				Custom: &corev1.Container{
					Image:   *image.BaseImageName,
					Command: []string{"cp"},
					Args:    []string{"/image-template/Dockerfile", "/workspace"},
				},
			},
			Template: &knbuild.TemplateInstantiationSpec{
				Name: imageConfig.ImageTemplate,
				Arguments: []knbuild.ArgumentSpec{
					knbuild.ArgumentSpec{
						Name:  "DESTINATION",
						Value: image.ImageDestination,
					},
					knbuild.ArgumentSpec{
						Name:  "BASE_IMAGE",
						Value: *image.BaseImageName,
					},
					knbuild.ArgumentSpec{
						Name:  "SYSTEM_PACKAGES_CONTENT",
						Value: string(systemPackagesContent),
					},
					knbuild.ArgumentSpec{
						Name:  "PACKAGES_CONTENT",
						Value: string(runtimePackagesContent),
					},
				},
			},
		},
	}
}

// ToImage producedDispatch Image from Knative Build
func ToImage(build *knbuild.Build) *dapi.Image {
	if build == nil {
		return nil
	}
	objMeta := &build.ObjectMeta
	var image dapi.Image
	if err := knaming.FromJSONString(objMeta.Annotations[knaming.InitialObjectAnnotation], &image); err != nil {
		// TODO
		panic(errors.Wrap(err, "decoding to Image"))
	}
	utils.AdjustMeta(&image.Meta, dapi.Meta{CreatedTime: build.CreationTimestamp.Unix()})

	image.Kind = utils.ImageKind
	image.ID = strfmt.UUID(objMeta.UID)

	image.Meta.Name = build.Labels[knaming.NameLabel]
	image.Meta.Org = build.Labels[knaming.OrgLabel]
	image.Meta.Project = build.Labels[knaming.ProjectLabel]
	image.Status = dapi.StatusINITIALIZED
	for _, cond := range build.Status.Conditions {
		if cond.Type == knbuild.BuildSucceeded && cond.Status == "True" {
			image.Status = dapi.StatusREADY
		}
		if cond.Message != "" {
			image.Reason = append(image.Reason, cond.Message)
		}
	}

	image.Meta.BackingObject = build
	return &image
}
