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

// ImageConfig contains images build configuration data
type ImageConfig struct {
	ImageTemplate   string
	ServciceAccount string
}

type knBuild struct {
	knbuildClient knclientset.Interface
	imageConfig   *ImageConfig
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

	// TODO: make ImageConfig come from configmap or other configurable
	imageConfig := &ImageConfig{
		ImageTemplate:   "image-template",
		ServciceAccount: "dispatch-build",
	}

	return &knBuild{
		knbuildClient: knClient(kubeconfPath),
		imageConfig:   imageConfig,
	}
}

// AddImage adds a image as Knative Build
func (h *knBuild) AddImage(ctx context.Context, image *v1.Image) (*v1.Image, error) {
	builds := h.knbuildClient.BuildV1alpha1().Builds(image.Meta.Org)

	build := FromImage(h.imageConfig, image)
	createdBuild, err := builds.Create(build)
	if err != nil {
		return nil, errors.Wrap(err, "creating knative build")
	}
	return ToImage(createdBuild), nil
}

// GetImage gets image
func (h *knBuild) GetImage(ctx context.Context, meta *v1.Meta) (*v1.Image, error) {
	builds := h.knbuildClient.BuildV1alpha1().Builds(meta.Org)

	build, err := builds.Get(knaming.ImageName(*meta), metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "get knative build")
	}

	img := ToImage(build)
	return img, nil
}

// DeleteImage deletes image
func (h *knBuild) DeleteImage(ctx context.Context, meta *v1.Meta) error {
	builds := h.knbuildClient.BuildV1alpha1().Builds(meta.Org)

	_, err := builds.Get(knaming.ImageName(*meta), metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "getting knative Build %s before deleting", meta.Name)
	}

	err = builds.Delete(knaming.ImageName(*meta), &metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrapf(err, "deleting knative Build '%s'", meta.Name)
	}

	return nil
}

// ListImage lists all image (Knative Build)
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

// UpdateImage updates Build
func (h *knBuild) UpdateImage(ctx context.Context, image *v1.Image) (*v1.Image, error) {
	builds := h.knbuildClient.BuildV1alpha1().Builds(image.Meta.Org)

	_, err := builds.Get(knaming.ImageName(image.Meta), metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "getting knative Build %s before updating", image.Meta.Name)
	}

	updated, err := builds.Update(FromImage(h.imageConfig, image))
	if err != nil {
		return nil, errors.Wrapf(err, "updating knative Build '%s'", image.Meta.Name)
	}
	return ToImage(updated), nil
}
