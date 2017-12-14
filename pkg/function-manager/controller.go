///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functionmanager

import (
	"context"
	"reflect"
	"time"

	apiclient "github.com/go-openapi/runtime/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/image-manager/gen/client/image"
	imagemodels "github.com/vmware/dispatch/pkg/image-manager/gen/models"
	"github.com/vmware/dispatch/pkg/trace"
)

type ControllerConfig struct {
	ResyncPeriod   time.Duration
	OrganizationID string
}

type funcEntityHandler struct {
	FaaS      functions.FaaSDriver
	Store     entitystore.EntityStore
	ImgClient ImageManager
}

func (h *funcEntityHandler) Type() reflect.Type {
	defer trace.Trace("")()

	return reflect.TypeOf(&functions.Function{})
}

func (h *funcEntityHandler) Add(obj entitystore.Entity) (err error) {
	defer trace.Trace("")()

	e := obj.(*functions.Function)

	defer func() {
		log.Debugf("trying to update entity when the function is created in underying driver")
		log.Debugf("entity org=%s, name=%s, id=%s, status=%s", e.OrganizationID, e.Name, e.ID, e.Status)
		h.Store.UpdateWithError(e, err)
	}()

	img, err := h.getImage(e.ImageName)
	if err != nil {
		return errors.Wrapf(err, "Error when fetching image for function %s", e.Name)
	}
	if err := h.FaaS.Create(e, &functions.Exec{
		Code:     e.Code,
		Main:     e.Main,
		Image:    img.DockerURL,
		Language: string(img.Language),
	}); err != nil {
		return errors.Wrapf(err, "Driver error when creating a FaaS function")
	}

	e.Status = entitystore.StatusREADY

	return
}

func (h *funcEntityHandler) Update(obj entitystore.Entity) error {
	defer trace.Trace("")()

	return h.Add(obj)
}

func (h *funcEntityHandler) Delete(obj entitystore.Entity) error {
	defer trace.Trace("")()

	e := obj.(*functions.Function)

	if err := h.FaaS.Delete(e); err != nil {
		log.Debugf("fail to delete from faas because %s", err)
		return errors.Wrapf(err, "Driver error when deleting a FaaS function")
	}

	log.Debugf("trying to delete entity=%s, org=%s, id=%s, status=%s\n", e.Name, e.OrganizationID, e.ID, e.Status)
	if err := h.Store.Delete(e.OrganizationID, e.Name, e); err != nil {
		log.Debugf("fail to delete entity because of %s", err)
		return errors.Wrap(err, "store error when updating function")
	}
	log.Debugf("delete the entity successfully")

	return nil
}

func (h *funcEntityHandler) Error(obj entitystore.Entity) error {
	defer trace.Trace("")()

	// TODO implement me
	return nil
}

func (h *funcEntityHandler) getImage(imageName string) (*imagemodels.Image, error) {
	defer trace.Trace("")()

	apiKeyAuth := apiclient.APIKeyAuth("cookie", "header", "cookie") // TODO replace "cookie"
	resp, err := h.ImgClient.GetImageByName(
		&image.GetImageByNameParams{
			ImageName: imageName,
			Context:   context.Background(),
		}, apiKeyAuth)
	if err == nil {
		return resp.Payload, nil
	}
	return nil, errors.Wrapf(err, "failed to get image: '%s'", imageName)
}

type runEntityHandler struct {
	FaaS   functions.FaaSDriver
	Runner functions.Runner
	Store  entitystore.EntityStore
}

func (h *runEntityHandler) Type() reflect.Type {
	defer trace.Trace("")()

	return reflect.TypeOf(&functions.FnRun{})
}

func (h *runEntityHandler) Add(obj entitystore.Entity) (err error) {
	defer trace.Trace("")()

	run := obj.(*functions.FnRun)
	defer run.Done()

	defer func() { h.Store.UpdateWithError(run, err) }()

	f := new(functions.Function)
	if err := h.Store.Get(FunctionManagerFlags.OrgID, run.FunctionName, f); err != nil {
		return errors.Wrapf(err, "Error getting function from store: '%s'", run.FunctionName)
	}

	ctx := functions.Context{}

	output, err := h.Runner.Run(&functions.FunctionExecution{
		Context: ctx,
		Name:    run.FunctionName,
		ID:      run.FunctionID,
		Schemas: &functions.Schemas{
			SchemaIn:  f.Schema.In,
			SchemaOut: f.Schema.Out,
		},
		Cookie:  "cookie",
		Secrets: run.Secrets,
	}, run.Input)
	run.Logs = ctx.Logs()
	run.Output = output
	if err != nil {
		return errors.Wrapf(err, "error running function: %s", run.FunctionName)
	}

	run.Status = entitystore.StatusREADY

	return
}

func (h *runEntityHandler) Update(obj entitystore.Entity) (err error) {
	defer trace.Trace("")()

	run := obj.(*functions.FnRun)
	defer func() { h.Store.UpdateWithError(run, err) }()
	return errors.Errorf("updating runs not supported, fn: '%s'", run.FunctionName)
}

func (h *runEntityHandler) Delete(obj entitystore.Entity) (err error) {
	defer trace.Trace("")()

	run := obj.(*functions.FnRun)
	defer func() { h.Store.UpdateWithError(run, err) }()
	return errors.Errorf("deleting runs not supported, fn: '%s'", run.FunctionName)
}

func (h *runEntityHandler) Error(obj entitystore.Entity) error {
	defer trace.Trace("")()

	// TODO implement me
	return nil
}

func NewController(config *ControllerConfig, store entitystore.EntityStore, faas functions.FaaSDriver, runner functions.Runner, imgClient ImageManager) controller.Controller {

	defer trace.Trace("")()

	c := controller.NewController(store, controller.Options{
		OrganizationID: FunctionManagerFlags.OrgID,
		ResyncPeriod:   config.ResyncPeriod,
		Workers:        1000, // want more functions concurrently? add more workers // TODO configure workers
	})
	c.AddEntityHandler(&funcEntityHandler{Store: store, FaaS: faas, ImgClient: imgClient})
	c.AddEntityHandler(&runEntityHandler{Store: store, FaaS: faas, Runner: runner})

	return c
}
