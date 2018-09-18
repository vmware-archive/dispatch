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

var (
	defaultDockerfile         = "Dockerfile"
	defaultSystemPackagesFile = "system-packages.txt"
	defaultPackagesFile       = "packages.txt"
)

// FromBaseImage produced Knative BuildTemplate from Dispatch BaseImage
func FromBaseImage(baseimage *dapi.BaseImage) *knbuild.BuildTemplate {
	if baseimage == nil {
		return nil
	}

	// TODO: BaseImage -> BuildTemplate make this right and configurable
	return &knbuild.BuildTemplate{
		ObjectMeta: knaming.ToObjectMeta(baseimage.Meta, *baseimage),
		Spec: knbuild.BuildTemplateSpec{
			Parameters: []knbuild.ParameterSpec{
				knbuild.ParameterSpec{
					Name:        string("DESTINATION"),
					Description: string("The destination to push the image"),
				},
				knbuild.ParameterSpec{
					Name:        string("BASE_IMAGE"),
					Description: string("The base image which this image is built from"),
				},
				knbuild.ParameterSpec{
					Name:        string("DOCKERFILE"),
					Description: string("Path to the Dockerfile to build"),
					Default:     &defaultDockerfile,
				},
				knbuild.ParameterSpec{
					Name:        string("SYSTEM_PACKAGES_CONTENT"),
					Description: string("System packages content that will be written to SYSTEM_PACKAGES_FILE"),
				},
				knbuild.ParameterSpec{
					Name:        string("SYSTEM_PACKAGES_FILE"),
					Description: string("Path to file with system dependencies"),
					Default:     &defaultSystemPackagesFile,
				},
				knbuild.ParameterSpec{
					Name:        string("PACKAGES_CONTENT"),
					Description: string("Runtime packages content that will be written to PACKAGES_FILE"),
				},
				knbuild.ParameterSpec{
					Name:        string("PACKAGES_FILE"),
					Description: string("Path to file with runtime dependencies"),
					Default:     &defaultPackagesFile,
				},
			},
			Steps: []corev1.Container{
				corev1.Container{
					Name:    string("write-system-package-files"),
					Image:   string("vmware/photon2:latest"),
					Command: []string{"/bin/bash"},
					Args:    []string{"-c", "echo -n ${SYSTEM_PACKAGES_CONTENT} | base64 -d > /workspace/${SYSTEM_PACKAGES_FILE}"},
				},
				corev1.Container{
					Name:    string("write-package-files"),
					Image:   string("vmware/photon2:latest"),
					Command: []string{"/bin/bash"},
					Args:    []string{"-c", "echo -n ${PACKAGES_CONTENT} | base64 -d > /workspace/${PACKAGES_FILE}"},
				},
				corev1.Container{
					Name:  string("build-and-push"),
					Image: string("gcr.io/kaniko-project/executor:v0.3.0"),
					Args: []string{
						"--dockerfile=${DOCKERFILE}",
						"--destination=${DESTINATION}",
						"--build-arg=BASE_IMAGE=${BASE_IMAGE}",
						"--build-arg=SYSTEM_PACKAGES_FILE=${SYSTEM_PACKAGES_FILE}",
						"--build-arg=PACKAGES_FILE=${PACKAGES_FILE}",
					},
				},
			},
		},
	}
}

// ToBaseImage producedDispatch BaseImage from Knative BuildTemplate
func ToBaseImage(knbuildTpl *knbuild.BuildTemplate) *dapi.BaseImage {
	if knbuildTpl == nil {
		return nil
	}
	objMeta := &knbuildTpl.ObjectMeta
	var baseimage dapi.BaseImage
	if err := knaming.FromJSONString(objMeta.Annotations[knaming.InitialObjectAnnotation], &baseimage); err != nil {
		// TODO
		panic(errors.Wrap(err, "decoding to BaseImage"))
	}
	utils.AdjustMeta(&baseimage.Meta, dapi.Meta{CreatedTime: knbuildTpl.CreationTimestamp.Unix()})

	baseimage.Kind = utils.BaseImageKind
	baseimage.ID = strfmt.UUID(objMeta.UID)
	// TODO: BuildTemplate -> BaseImage
	// baseimage.Status = dapi.StatusINITIALIZED
	return &baseimage
}

// FromImage produced Knative Build from Dispatch Image
func FromImage(image *dapi.Image) *knbuild.Build {
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
			//ServiceAccountName:
			Source: &knbuild.SourceSpec{
				Custom: &corev1.Container{
					Image:   *image.BaseImageName, // TODO
					Command: []string{"cp"},
					Args:    []string{"/image-template/Dockerfile", "/workspace"},
				},
			},
			Template: &knbuild.TemplateInstantiationSpec{
				Name: "image-template" + *image.Name,
				Arguments: []knbuild.ArgumentSpec{
					knbuild.ArgumentSpec{
						Name:  "DESTINATION",
						Value: image.DockerURL, // TODO
					},
					knbuild.ArgumentSpec{
						Name:  "BASE_IMAGE",
						Value: *image.BaseImageName, // TODO
					},
					knbuild.ArgumentSpec{
						Name:  "SYSTEM_PACKAGES_CONTENT",
						Value: string(systemPackagesContent), // TODO
					},
					knbuild.ArgumentSpec{
						Name:  "PACKAGES_CONTENT",
						Value: string(runtimePackagesContent), // TODO
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
	// TODO: Build -> Image
	// image.Status = dapi.StatusINITIALIZED
	return &image
}
