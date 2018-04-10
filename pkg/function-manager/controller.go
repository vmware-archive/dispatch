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

// ControllerConfig is the function manager controller configuration
type ControllerConfig struct {
	ResyncPeriod   time.Duration
	OrganizationID string
}

type funcEntityHandler struct {
	FaaS      functions.FaaSDriver
	Store     entitystore.EntityStore
	ImgClient ImageManager
}

// Type returns the reflect.Type of a functions.Function
func (h *funcEntityHandler) Type() reflect.Type {
	defer trace.Trace("")()

	return reflect.TypeOf(&functions.Function{})
}

// Add creates new functions (and function images) for the configured FaaS
func (h *funcEntityHandler) Add(obj entitystore.Entity) (err error) {
	defer trace.Trace("")()

	e := obj.(*functions.Function)

	defer func() {
		log.Debugf("function org=%s, name=%s, id=%s, status=%s", e.OrganizationID, e.Name, e.ID, e.Status)
		h.Store.UpdateWithError(e, err)
	}()

	img, err := h.getImage(e.ImageName)
	if err != nil {
		return errors.Wrapf(err, "Error when fetching image for function %s", e.Name)
	}

	if img.Status == imagemodels.StatusERROR {
		return errors.Errorf("image in error status for function '%s', image name: '%s', reason: %v", e.ID, e.ImageName, img.Reason)
	}

	// If the image isn't ready yet, we cannot proceed.  The loop should pick up the work
	// next iteration.
	if img.Status != imagemodels.StatusREADY {
		return
	}

	e.Status = entitystore.StatusCREATING
	h.Store.UpdateWithError(e, nil)

	if err := h.FaaS.Create(e, &functions.Exec{
		Code:  e.Code,
		Main:  e.Main,
		Image: img.DockerURL,
	}); err != nil {
		return errors.Wrapf(err, "Driver error when creating a FaaS function")
	}

	e.Status = entitystore.StatusREADY

	return
}

// Update updates functions (and function images) for the configured FaaS
// TODO: we are leaking images... the images should be deleted from the image
// repository
func (h *funcEntityHandler) Update(obj entitystore.Entity) error {
	defer trace.Trace("")()

	return h.Add(obj)
}

// Delete deletes functions (and function images) for the configured FaaS
// TODO: we are leaking images... the images should be deleted from the image
// repository
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

// Error handles errors returned from the configured FaaS
func (h *funcEntityHandler) Error(obj entitystore.Entity) error {
	defer trace.Trace("")()

	// TODO implement me
	return nil
}

// Only return entities in INITIALIZED, UPDATING or DELETING status
// This is kind of a hack as smarter filtering is required.
func syncFilter(resyncPeriod time.Duration) entitystore.Filter {
	defer trace.Trace("")()

	now := time.Now().Add(-resyncPeriod)
	return entitystore.FilterEverything().Add(
		entitystore.FilterStat{
			Scope:   entitystore.FilterScopeField,
			Subject: "ModifiedTime",
			Verb:    entitystore.FilterVerbBefore,
			Object:  now,
		},
		entitystore.FilterStat{
			Scope:   entitystore.FilterScopeField,
			Subject: "Status",
			Verb:    entitystore.FilterVerbIn,
			Object: []entitystore.Status{
				entitystore.StatusINITIALIZED, entitystore.StatusUPDATING, entitystore.StatusDELETING,
			},
		})
}

// Sync compares actual and desired state to return a list of function entities which must be resolved
func (h *funcEntityHandler) Sync(organizationID string, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	defer trace.Trace("")()

	return controller.DefaultSync(h.Store, h.Type(), organizationID, resyncPeriod, syncFilter(resyncPeriod))
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

// Type returns the reflect.Type of a functions.FnRun
func (h *runEntityHandler) Type() reflect.Type {
	defer trace.Trace("")()

	return reflect.TypeOf(&functions.FnRun{})
}

// Add creates a function execution (run)
func (h *runEntityHandler) Add(obj entitystore.Entity) (err error) {
	defer trace.Trace("")()

	run := obj.(*functions.FnRun)
	defer run.Done()

	defer func() { h.Store.UpdateWithError(run, err) }()

	run.Status = entitystore.StatusCREATING
	h.Store.UpdateWithError(run, nil)

	f := new(functions.Function)
	if err = h.Store.Get(FunctionManagerFlags.OrgID, run.FunctionName, entitystore.Options{}, f); err != nil {
		return errors.Wrapf(err, "Error getting function from store: '%s'", run.FunctionName)
	}

	ctx := functions.Context{}

	if run.Event != nil {
		ctx[functions.EventKey] = run.Event
	}

	if len(run.HTTPContext) > 0 {
		ctx[functions.HTTPContextKey] = run.HTTPContext
	}

	output, err := h.Runner.Run(&functions.FunctionExecution{
		Context:    ctx,
		RunID:      run.ID,
		FunctionID: run.FunctionID,
		FaasID:     run.FaasID,
		Schemas: &functions.Schemas{
			SchemaIn:  f.Schema.In,
			SchemaOut: f.Schema.Out,
		},
		Cookie:   "cookie",
		Secrets:  run.Secrets,
		Services: run.Services,
	}, run.Input)
	run.Logs = ctx.Logs()
	run.Output = output
	if err != nil {
		return errors.Wrapf(err, "error running function: %s", run.FunctionName)
	}

	run.Status = entitystore.StatusREADY
	run.FinishedTime = time.Now()

	return
}

// Update updates a function execution (run)
func (h *runEntityHandler) Update(obj entitystore.Entity) (err error) {
	defer trace.Trace("")()

	run := obj.(*functions.FnRun)
	defer func() { h.Store.UpdateWithError(run, err) }()
	return errors.Errorf("updating runs not supported, fn: '%s'", run.FunctionName)
}

// Delete deletes a function execution (run)
func (h *runEntityHandler) Delete(obj entitystore.Entity) (err error) {
	defer trace.Trace("")()

	run := obj.(*functions.FnRun)
	defer func() { h.Store.UpdateWithError(run, err) }()
	return errors.Errorf("deleting runs not supported, fn: '%s'", run.FunctionName)
}

// Sync compares actual and desired state to return a list of function execution (run) entities which must be resolved
func (h *runEntityHandler) Sync(organizationID string, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	defer trace.Trace("")()

	return controller.DefaultSync(h.Store, h.Type(), organizationID, resyncPeriod, syncFilter(resyncPeriod))
}

// Error handles errors with regards to function execution entities (currently a no-op)
func (h *runEntityHandler) Error(obj entitystore.Entity) error {
	defer trace.Trace("")()

	// TODO implement me
	return nil
}

// NewController is the contstructor for the function manager controller
func NewController(config *ControllerConfig, store entitystore.EntityStore, faas functions.FaaSDriver, runner functions.Runner, imgClient ImageManager) controller.Controller {

	defer trace.Trace("")()

	c := controller.NewController(controller.Options{
		OrganizationID: FunctionManagerFlags.OrgID,
		ResyncPeriod:   config.ResyncPeriod,
		Workers:        1000, // want more functions concurrently? add more workers // TODO configure workers
	})
	c.AddEntityHandler(&funcEntityHandler{Store: store, FaaS: faas, ImgClient: imgClient})
	c.AddEntityHandler(&runEntityHandler{Store: store, FaaS: faas, Runner: runner})

	return c
}
