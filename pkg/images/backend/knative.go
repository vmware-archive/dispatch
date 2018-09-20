///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package backend

import (
	"context"
	"log"

	"github.com/vmware/dispatch/pkg/utils/knaming"

	knclientset "github.com/knative/build/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
// func (h *knBuild) AddBaseImage(ctx context.Context, baseimage *v1.BaseImage) (*v1.BaseImage, error) {
// 	buildTpl := FromBaseImage(baseimage)

// 	createdBuildTpl, err := h.knbuildClient.BuildV1alpha1().BuildTemplates(baseimage.Meta.Org).Create(buildTpl)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "creating knative build template")
// 	}
// 	return ToBaseImage(createdBuildTpl), nil
// }

// GetBaseImage adds a baseimage as Knative BuildTemplate
// func (h *knBuild) ListBaseImages(ctx context.Context, meta *v1.Meta) ([]*v1.BaseImage, error) {

// 	baseimageList, err := h.knbuildClient.BuildV1alpha1().BuildTemplates(meta.Org).List(metav1.ListOptions{
// 		LabelSelector: knaming.ToLabelSelector(map[string]string{
// 			knaming.ProjectLabel: meta.Project,
// 		}),
// 	})
// 	if err != nil {
// 		return nil, errors.Wrap(err, "listing knative BuildTemplates")
// 	}

// 	var baseimages []*v1.BaseImage

// 	for i := range baseimageList.Items {
// 		objectMeta := &baseimageList.Items[i].ObjectMeta
// 		if objectMeta.Labels[knaming.OrgLabel] != "" {
// 			baseimages = append(baseimages, ToBaseImage(&baseimageList.Items[i]))
// 		}
// 	}
// 	return baseimages, nil
// }

// AddImage adds a image as Knative Build
func (h *knBuild) AddImage(ctx context.Context, image *v1.Image) (*v1.Image, error) {
	build := FromImage(image)

	createdBuild, err := h.knbuildClient.BuildV1alpha1().Builds(image.Meta.Org).Create(build)
	if err != nil {
		return nil, errors.Wrap(err, "creating knative build")
	}
	return ToImage(createdBuild), nil
}

func (h *knBuild) ListImage(ctx context.Context, meta *v1.Meta) ([]*v1.Image, error) {
	builds := h.knbuildClient.BuildV1alpha1().Builds(meta.Org)

	buildList, err := builds.List(metav1.ListOptions{
		LabelSelector: knaming.ToLabelSelector(map[string]string{
			knaming.ProjectLabel: meta.Project,
		}),
	})
	if err != nil {
		return nil, errors.Wrap(err, "listing knative build")
	}

	var images []*v1.Image

	for i := range buildList.Items {
		objectMeta := &buildList.Items[i].ObjectMeta
		if objectMeta.Labels[knaming.OrgLabel] != "" {
			images = append(images, ToImage(&buildList.Items[i]))
		}
	}
	return images, nil
}
