///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functionmanager

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/trace"
)

// ControllerConfig is the function manager controller configuration
type ControllerConfig struct {
	ResyncPeriod      time.Duration
	ZookeeperLocation string
}

type funcEntityHandler struct {
	FaaS         functions.FaaSDriver
	Store        entitystore.EntityStore
	ImgClient    ImageGetter
	ImageBuilder functions.ImageBuilder
}

// Type returns the reflect.Type of a functions.Function
func (h *funcEntityHandler) Type() reflect.Type {
	return reflect.TypeOf(&functions.Function{})
}

// Add creates new functions (and function images) for the configured FaaS
func (h *funcEntityHandler) Add(ctx context.Context, obj entitystore.Entity) (err error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	e := obj.(*functions.Function)

	defer func() {
		log.Debugf("function org=%s, name=%s, id=%s, status=%s", e.OrganizationID, e.Name, e.ID, e.Status)
		h.Store.UpdateWithError(ctx, e, err)
	}()

	code, err := h.resolveSourceURL(ctx, e.OrganizationID, e.SourceURL)
	if err != nil {
		log.Debugf("failed to retrieve source because of %s", err)
		return errors.Wrap(err, "store error when retrieving source")
	}

	img, err := h.getImage(ctx, e.OrganizationID, e.ImageName)
	if err != nil {
		return errors.Wrapf(err, "Error when fetching image for function %s", e.Name)
	}

	if img.Status == v1.StatusERROR {
		return errors.Errorf("image in error status for function '%s', image name: '%s', reason: %v", e.ID, e.ImageName, img.Reason)
	}

	// If the image isn't ready yet, we cannot proceed.  The loop should pick up the work
	// next iteration.
	if img.Status != v1.StatusREADY {
		return
	}

	e.ImageURL = img.DockerURL
	e.Status = entitystore.StatusCREATING
	h.Store.UpdateWithError(ctx, e, nil)

	e.FunctionImageURL, err = h.ImageBuilder.BuildImage(ctx, e, code)
	if err != nil {
		return errors.Wrapf(err, "Error building image for function '%s'", e.ID)
	}

	if err := h.FaaS.Create(ctx, e); err != nil {
		return errors.Wrapf(err, "Driver error when creating a FaaS function")
	}

	e.Status = entitystore.StatusREADY

	return
}

func (h *funcEntityHandler) resolveSourceURL(ctx context.Context, organizationID string, sourceURL string) ([]byte, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	scheme, err := getScheme(sourceURL)
	if err != nil {
		return nil, errors.Wrapf(err, "Error when parsing scheme from source url %s", sourceURL)
	}
	switch scheme {
	case entityScheme:
		s := new(functions.Source)
		sourceName, err := getURLWithoutScheme(sourceURL)
		if err != nil {
			return nil, errors.Wrapf(err, "Error when parsing source url %s", sourceURL)
		}

		if err := h.Store.Get(ctx, organizationID, sourceName, entitystore.Options{}, s); err != nil {
			log.Debugf("Error returned by h.Store.Get: %+v", err)
			return nil, errors.Wrapf(err, "Failed to retreive source %s", sourceName)
		}

		return s.Code, nil
	default:
		return nil, errors.Errorf("Received non-supported source url %s", sourceURL)
	}
}

// Update updates functions (and function images) for the configured FaaS
func (h *funcEntityHandler) Update(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	e := obj.(*functions.Function)

	if err := h.FaaS.Delete(ctx, e); err != nil {
		log.Debugf("fail to delete function from faas driver because %s", err)
		return errors.Wrap(err, "error when deleting function from faas driver")
	}

	if err := h.ImageBuilder.RemoveImage(ctx, e); err != nil {
		log.Errorf("Failed to remove function image from docker client: %+v", err)
	}

	// generating a new UUID will force the creation of a new function in the underlying FaaS
	e.FaasID = uuid.NewV4().String()

	if err := h.Add(ctx, e); err != nil {
		return errors.Wrapf(err, "Error when updating function %s", e.Name)
	}

	return nil
}

// Delete deletes functions (and function images) for the configured FaaS
func (h *funcEntityHandler) Delete(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	e := obj.(*functions.Function)

	if err := h.FaaS.Delete(ctx, e); err != nil {
		log.Debugf("fail to delete from faas because %s", err)
		return errors.Wrapf(err, "Driver error when deleting a FaaS function")
	}

	runs, err := getFilteredRuns(ctx, h.Store, e.OrganizationID, &e.Name, nil, nil)
	if err != nil {
		return errors.Wrapf(err, "store error listing runs for function %s", e.Name)
	}
	for _, r := range runs {
		if err := h.Store.Delete(ctx, e.OrganizationID, r.Name, r); err != nil {
			log.Debugf("fail to delete entity because of %s", err)
			return errors.Wrap(err, "store error when deleting function run")
		}
	}

	if err := h.ImageBuilder.RemoveImage(ctx, e); err != nil {
		log.Errorf("Failed to remove function image from docker client: %+v", err)
	}

	scheme, err := getScheme(e.SourceURL)
	if err != nil {
		return errors.Wrapf(err, "Error when parsing scheme from source url %s", e.SourceURL)
	}
	if scheme == entityScheme {
		sourceName, err := getURLWithoutScheme(e.SourceURL)
		if err != nil {
			return errors.Wrapf(err, "Error when parsing source url %s", e.SourceURL)
		}
		var s functions.Source
		if err := h.Store.Delete(ctx, e.OrganizationID, sourceName, &s); err != nil {
			log.Debugf("fail to delete entity because of %s", err)
			return errors.Wrap(err, "store error when deleting source")
		}
	}

	log.Debugf("trying to delete entity=%s, org=%s, id=%s, status=%s\n", e.Name, e.OrganizationID, e.ID, e.Status)
	if err := h.Store.Delete(ctx, e.OrganizationID, e.Name, e); err != nil {
		log.Debugf("fail to delete entity because of %s", err)
		return errors.Wrap(err, "store error when deleting function")
	}
	log.Debugf("delete the entity successfully")

	return nil
}

// Error handles errors returned from the configured FaaS
func (h *funcEntityHandler) Error(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	// TODO implement me
	return nil
}

// Only return entities in INITIALIZED, UPDATING or DELETING status
// This is kind of a hack as smarter filtering is required.
func syncFilter(resyncPeriod time.Duration) entitystore.Filter {
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
func (h *funcEntityHandler) Sync(ctx context.Context, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	return controller.DefaultSync(ctx, h.Store, h.Type(), resyncPeriod, syncFilter(resyncPeriod))
}

func (h *funcEntityHandler) getImage(ctx context.Context, organizationID string, imageName string) (*v1.Image, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()
	resp, err := h.ImgClient.GetImage(ctx, organizationID, imageName)

	if err == nil {
		return resp, nil
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
	return reflect.TypeOf(&functions.FnRun{})
}

type invocationError struct {
	Err *v1.InvocationError `json:"err"`
}

func (err *invocationError) Error() string {
	if err.Err.Message != nil {
		return *(err.Err.Message)
	}
	return ""
}

// Add creates a function execution (run)
func (h *runEntityHandler) Add(ctx context.Context, obj entitystore.Entity) (err error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	run := obj.(*functions.FnRun)
	defer run.Done()

	defer func() { h.Store.UpdateWithError(ctx, run, err) }()

	run.Status = entitystore.StatusCREATING
	h.Store.UpdateWithError(ctx, run, nil)

	f := new(functions.Function)
	if err = h.Store.Get(ctx, run.OrganizationID, run.FunctionName, entitystore.Options{}, f); err != nil {
		return errors.Wrapf(err, "Error getting function from store: '%s'", run.FunctionName)
	}

	fctx := functions.Context{}

	if run.Event != nil {
		fctx[functions.EventKey] = run.Event
	}

	if len(run.HTTPContext) > 0 {
		fctx[functions.HTTPContextKey] = run.HTTPContext
	}

	fctx[functions.TimeoutKey] = f.Timeout

	output, err := h.Runner.Run(&functions.FunctionExecution{
		Context:        fctx,
		OrganizationID: run.OrganizationID,
		RunID:          run.ID,
		FunctionID:     run.FunctionID,
		FaasID:         run.FaasID,
		Schemas: &functions.Schemas{
			SchemaIn:  f.Schema.In,
			SchemaOut: f.Schema.Out,
		},
		Cookie:   "cookie",
		Secrets:  run.Secrets,
		Services: run.Services,
	}, run.Input)

	logs := fctx.Logs()
	run.Logs = &logs
	run.Output = output

	if err != nil {
		var stacktrace []string
		if e, ok := err.(functions.StackTracer); ok {
			if st := e.StackTrace(); st != nil {
				st := fmt.Sprintf("%+v", st)
				st = strings.Trim(st, "\n")
				stacktrace = strings.Split(st, "\n")
			}
		}
		message := err.Error()
		switch err.(type) {
		case functions.InputError:
			run.Error = &v1.InvocationError{Message: &message, Type: v1.ErrorTypeInputError, Stacktrace: stacktrace}
		case functions.FunctionError:
			run.Error = &v1.InvocationError{Message: &message, Type: v1.ErrorTypeFunctionError, Stacktrace: stacktrace}
		case functions.SystemError:
			run.Error = &v1.InvocationError{Message: &message, Type: v1.ErrorTypeSystemError, Stacktrace: stacktrace}
		default:
			log.Debugf("No invocation error type provided for error %s", err)
			run.Error = &v1.InvocationError{Message: &message, Stacktrace: stacktrace}
		}
		return errors.Wrapf(&invocationError{run.Error}, "error running function: %s", run.FunctionName)
	}

	run.Error = fctx.GetError()
	if run.Error != nil {
		return errors.Wrapf(&invocationError{run.Error}, "error running function: %s", run.FunctionName)
	}

	run.Status = entitystore.StatusREADY
	run.FinishedTime = time.Now()
	return
}

// Update updates a function execution (run)
func (h *runEntityHandler) Update(ctx context.Context, obj entitystore.Entity) (err error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	run := obj.(*functions.FnRun)
	defer func() { h.Store.UpdateWithError(ctx, run, err) }()
	return errors.Errorf("updating runs not supported, fn: '%s'", run.FunctionName)
}

// Delete deletes a function execution (run)
func (h *runEntityHandler) Delete(ctx context.Context, obj entitystore.Entity) (err error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	run := obj.(*functions.FnRun)
	defer func() { h.Store.UpdateWithError(ctx, run, err) }()
	return errors.Errorf("deleting runs not supported, fn: '%s'", run.FunctionName)
}

// Sync compares actual and desired state to return a list of function execution (run) entities which must be resolved
func (h *runEntityHandler) Sync(ctx context.Context, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	return controller.DefaultSync(ctx, h.Store, h.Type(), resyncPeriod, syncFilter(resyncPeriod))
}

// Error handles errors with regards to function execution entities (currently a no-op)
func (h *runEntityHandler) Error(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	// TODO implement me
	return nil
}

// NewController is the constructor for the function manager controller
func NewController(config *ControllerConfig, store entitystore.EntityStore, faas functions.FaaSDriver, runner functions.Runner, imgClient ImageGetter, imageBuilder functions.ImageBuilder) controller.Controller {

	c := controller.NewController(controller.Options{
		ResyncPeriod:      config.ResyncPeriod,
		Workers:           1000, // want more functions concurrently? add more workers // TODO configure workers
		ServiceName:       "functions",
		ZookeeperLocation: config.ZookeeperLocation,
	})
	c.AddEntityHandler(&funcEntityHandler{Store: store, FaaS: faas, ImgClient: imgClient, ImageBuilder: imageBuilder})
	c.AddEntityHandler(&runEntityHandler{Store: store, FaaS: faas, Runner: runner})

	return c
}
