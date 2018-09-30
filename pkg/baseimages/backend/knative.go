///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package backend

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	dapi "github.com/vmware/dispatch/pkg/api/v1"
	derrors "github.com/vmware/dispatch/pkg/errors"
	baseimage "github.com/vmware/dispatch/pkg/resources/baseimage/v1"
	clientset "github.com/vmware/dispatch/pkg/resources/gen/clientset/versioned"
	"github.com/vmware/dispatch/pkg/utils"
	"github.com/vmware/dispatch/pkg/utils/knaming"
)

type knative struct {
	resourcesClient clientset.Interface
}

func resourcesClient(kubeconfPath string) clientset.Interface {
	config, err := utils.KubeClientConfig(kubeconfPath)
	if err != nil {
		log.Fatalf("%+v", errors.Wrap(err, "error configuring k8s API client"))
	}
	return clientset.NewForConfigOrDie(config)
}

// KnativeBuild create new knative build backend
func KnativeBuild(kubeconfPath string) Backend {
	return &knative{
		resourcesClient: resourcesClient(kubeconfPath),
	}
}

// AddBaseImage adds a baseimage resource
func (h *knative) AddBaseImage(ctx context.Context, model *dapi.BaseImage) (*dapi.BaseImage, error) {
	client := h.resourcesClient.ResourcesV1().BaseImages(model.Org)

	baseImage := FromModel(model)
	createdBaseImage, err := client.Create(baseImage)
	if err != nil {
		log.Errorf("error creating BaseImage resource %s: %v", model.Name, err)
		return nil, derrors.NewServerError(err)
	}
	return ToModel(createdBaseImage), nil
}

// GetBaseImage gets baseimage resource
func (h *knative) GetBaseImage(ctx context.Context, meta *dapi.Meta) (*dapi.BaseImage, error) {
	client := h.resourcesClient.ResourcesV1().BaseImages(meta.Org)
	baseImage, err := client.Get(knaming.BaseImageName(*meta), metav1.GetOptions{})
	if kerrors.IsNotFound(err) {
		return nil, derrors.NewObjectNotFoundError(err)
	}
	if err != nil {
		log.Errorf("error getting BaseImage resource %s: %v", meta.Name, err)
		return nil, derrors.NewServerError(err)
	}
	return ToModel(baseImage), nil
}

// DeleteImage deletes image
func (h *knative) DeleteBaseImage(ctx context.Context, meta *dapi.Meta) error {
	client := h.resourcesClient.ResourcesV1().BaseImages(meta.Org)
	err := client.Delete(knaming.BaseImageName(*meta), nil)
	if kerrors.IsNotFound(err) {
		return derrors.NewObjectNotFoundError(err)
	}
	if err != nil {
		log.Errorf("error getting BaseImage resource %s: %v", meta.Name, err)
		return derrors.NewServerError(err)
	}
	return nil
}

// ListBaseImage lists all baseimage resources
func (h *knative) ListBaseImage(ctx context.Context, meta *dapi.Meta) ([]*dapi.BaseImage, error) {
	client := h.resourcesClient.ResourcesV1().BaseImages(meta.Org)
	baseImages, err := client.List(metav1.ListOptions{})
	if kerrors.IsNotFound(err) {
		return nil, derrors.NewObjectNotFoundError(err)
	}
	if err != nil {
		log.Errorf("error getting BaseImage resource %s: %v", meta.Name, err)
		return nil, derrors.NewServerError(err)
	}

	var models []*dapi.BaseImage
	for _, baseImage := range baseImages.Items {
		model := ToModel(&baseImage)
		models = append(models, model)
	}
	return models, nil
}

// UpdateBaseImage updates baseimage resources
func (h *knative) UpdateBaseImage(ctx context.Context, model *dapi.BaseImage) (*dapi.BaseImage, error) {
	client := h.resourcesClient.ResourcesV1().BaseImages(model.Org)
	updatedBaseImage, err := client.Update(FromModel(model))
	if kerrors.IsNotFound(err) {
		return nil, derrors.NewObjectNotFoundError(err)
	}
	if err != nil {
		log.Errorf("error getting BaseImage resource %s: %v", model.Name, err)
		return nil, derrors.NewServerError(err)
	}
	return ToModel(updatedBaseImage), nil
}

// FromModel translates a Dispatch BaseImage model to a Dispatch k8s BaseImage (CRD)
func FromModel(model *dapi.BaseImage) *baseimage.BaseImage {
	if model == nil {
		return nil
	}
	return &baseimage.BaseImage{
		ObjectMeta: knaming.ToObjectMeta(model.Meta, *model),
		Spec: baseimage.BaseImageSpec{
			ImageURL: *model.ImageURL,
			Language: *model.Language,
		},
	}
}

// ToModel translates a Dispatch k8s BaseImage (CRD) to a Dispatch BaseImage model
func ToModel(baseImage *baseimage.BaseImage) *dapi.BaseImage {
	if baseImage == nil {
		return nil
	}
	objMeta := &baseImage.ObjectMeta
	var model dapi.BaseImage
	if err := knaming.FromJSONString(objMeta.Annotations[knaming.InitialObjectAnnotation], &model); err != nil {
		// TODO
		panic(errors.Wrap(err, "decoding to BaseImage resource"))
	}
	utils.AdjustMeta(&model.Meta, dapi.Meta{CreatedTime: baseImage.CreationTimestamp.Unix()})

	model.Kind = dapi.BaseImageKind
	model.ID = strfmt.UUID(objMeta.UID)

	model.Meta.Name = baseImage.Labels[knaming.NameLabel]
	model.Meta.Org = baseImage.Labels[knaming.OrgLabel]
	model.Meta.Project = baseImage.Labels[knaming.ProjectLabel]
	model.Status = dapi.StatusREADY

	// TODO: build a controller to update status (image not found, etc.)

	model.Meta.BackingObject = baseImage
	return &model
}
