///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package backend

import (
	"context"

	knclientset "github.com/knative/build/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware/dispatch/pkg/api/v1"
	derrors "github.com/vmware/dispatch/pkg/errors"
	"github.com/vmware/dispatch/pkg/utils"
	"github.com/vmware/dispatch/pkg/utils/knaming"
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
	builds := h.knbuildClient.BuildV1alpha1().Builds(image.Org)

	build := FromImage(h.imageConfig, image)
	createdBuild, err := builds.Create(build)
	if err != nil {
		log.Errorf("error creating knative build %s: %v", image.Name, err)
		return nil, derrors.NewServerError(err)
	}
	return ToImage(createdBuild), nil
}

// GetImage gets image
func (h *knBuild) GetImage(ctx context.Context, meta *v1.Meta) (*v1.Image, error) {
	builds := h.knbuildClient.BuildV1alpha1().Builds(meta.Org)

	build, err := builds.Get(knaming.ImageName(*meta), metav1.GetOptions{})
	if kerrors.IsNotFound(err) {
		return nil, derrors.NewObjectNotFoundError(err)
	}
	if err != nil {
		log.Errorf("error getting knative build %s: %v", meta.Name, err)
		return nil, derrors.NewServerError(err)
	}

	img := ToImage(build)
	return img, nil
}

// DeleteImage deletes image
func (h *knBuild) DeleteImage(ctx context.Context, meta *v1.Meta) error {
	builds := h.knbuildClient.BuildV1alpha1().Builds(meta.Org)

	err := builds.Delete(knaming.ImageName(*meta), &metav1.DeleteOptions{})
	if kerrors.IsNotFound(err) {
		return derrors.NewObjectNotFoundError(err)
	}
	if err != nil {
		log.Errorf("error deleting knative build %s: %v", meta.Name, err)
		return derrors.NewServerError(err)
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
		log.Errorf("error listing knative builds: %v", err)
		return nil, derrors.NewServerError(err)
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

	updated, err := builds.Update(FromImage(h.imageConfig, image))
	if kerrors.IsNotFound(err) {
		return nil, derrors.NewObjectNotFoundError(err)
	}
	if err != nil {
		log.Errorf("error updating knative build %s: %v", image.Name, err)
		return nil, derrors.NewServerError(err)
	}
	return ToImage(updated), nil
}
