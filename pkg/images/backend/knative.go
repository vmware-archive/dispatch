///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package backend

import (
	"context"
	"log"

	knclientset "github.com/knative/build/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/utils"
)

type knBuild struct {
	knbuildClient knclientset.Interface
}

func knClient(kubeconfPath string) knclientset.Interface {
	config, err := utils.KubeClientConfig(kubeconfPath)
	if err != nil {
		log.Fatalf("%+v", errors.Wrap(err, "error configuring k8s API client"))
	}
	return knclientset.NewForConfigOrDie(config)
}

// KnativeBuild create new knative build backend
func KnativeBuild(kubeconfPath string) Backend {
	return &knBuild{
		knbuildClient: knClient(kubeconfPath),
	}
}

// AddBaseImage adds a baseimage as Knative BuildTemplate
func (h *knBuild) AddBaseImage(ctx context.Context, baseimage *v1.BaseImage) (*v1.BaseImage, error) {
	buildTpl := FromBaseImage(baseimage)

	createdBuildTpl, err := h.knbuildClient.BuildV1alpha1().BuildTemplates(baseimage.Meta.Org).Create(buildTpl)
	if err != nil {
		return nil, errors.Wrap(err, "creating knative build template")
	}
	return ToBaseImage(createdBuildTpl), nil
}

// AddImage adds a image as Knative Build
func (h *knBuild) AddImage(ctx context.Context, image *v1.Image) (*v1.Image, error) {
	build := FromImage(image)

	createdBuild, err := h.knbuildClient.BuildV1alpha1().Builds(image.Meta.Org).Create(build)
	if err != nil {
		return nil, errors.Wrap(err, "creating knative build")
	}
	return ToImage(createdBuild), nil
}
