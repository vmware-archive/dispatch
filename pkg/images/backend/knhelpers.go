///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package backend

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	knbuild "github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	dapi "github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/utils"
	"github.com/vmware/dispatch/pkg/utils/knaming"
)

// FromImage produced Knative Build from Dispatch Image
func FromImage(imageConfig *ImageConfig, image *dapi.Image) *knbuild.Build {
	if image == nil {
		return nil
	}

	systemPackagesContent := ""
	if image.SystemDependencies != nil && image.SystemDependencies.Packages != nil {
		var packages []string
		for _, pkg := range image.SystemDependencies.Packages {
			if pkg.Name == nil {
				continue
			}
			if pkg.Version != "" {
				packages = append(packages, fmt.Sprintf("%s-%s", *pkg.Name, pkg.Version))
			} else {
				packages = append(packages, *pkg.Name)
			}
		}

		systemPackagesContent = base64.StdEncoding.EncodeToString([]byte(strings.Join(packages, "\n")))
	}

	runtimePackagesContent := ""
	if image.RuntimeDependencies != nil && image.RuntimeDependencies.Manifest != "" {
		runtimePackagesContent = base64.StdEncoding.EncodeToString([]byte(image.RuntimeDependencies.Manifest))
	}

	// TODO: Image -> Build make this right and configurable
	return &knbuild.Build{
		ObjectMeta: knaming.ToObjectMeta(image.Meta, *image),
		Spec: knbuild.BuildSpec{
			ServiceAccountName: imageConfig.ServiceAccount,
			Source: &knbuild.SourceSpec{
				Custom: &corev1.Container{
					Image:   image.BaseImageURL,
					Command: []string{"cp"},
					Args:    []string{"/image-template/Dockerfile", "/workspace"},
				},
			},
			Template: &knbuild.TemplateInstantiationSpec{
				Name: imageConfig.ImageTemplate,
				Arguments: []knbuild.ArgumentSpec{
					knbuild.ArgumentSpec{
						Name:  "DESTINATION",
						Value: image.ImageURL,
					},
					knbuild.ArgumentSpec{
						Name:  "BASE_IMAGE",
						Value: image.BaseImageURL,
					},
					knbuild.ArgumentSpec{
						Name:  "SYSTEM_PACKAGES_CONTENT",
						Value: systemPackagesContent,
					},
					knbuild.ArgumentSpec{
						Name:  "PACKAGES_CONTENT",
						Value: runtimePackagesContent,
					},
				},
			},
			Timeout: &metav1.Duration{Duration: time.Minute * 10},
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

	image.Kind = dapi.ImageKind
	image.ID = strfmt.UUID(objMeta.UID)

	image.Revision = objMeta.GetResourceVersion()
	image.Name = build.Labels[knaming.NameLabel]
	image.Org = build.Labels[knaming.OrgLabel]
	image.Project = build.Labels[knaming.ProjectLabel]
	image.Status = dapi.StatusINITIALIZED

	for _, cond := range build.Status.Conditions {
		if cond.Status == corev1.ConditionTrue {
			image.Status = dapi.StatusREADY
		}
		if cond.Message != "" {
			image.Reason = append(image.Reason, cond.Message)
		}
	}

	image.Meta.BackingObject = build
	return &image
}
