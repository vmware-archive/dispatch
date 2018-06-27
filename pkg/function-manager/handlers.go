///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functionmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
	dispatcherrors "github.com/vmware/dispatch/pkg/errors"
	"github.com/vmware/dispatch/pkg/event-manager/helpers"
	"github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations"
	fnrunner "github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations/runner"
	fnstore "github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations/store"
	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

func functionEntityToModel(f *functions.Function) *v1.Function {
	var tags []*v1.Tag
	for k, v := range f.Tags {
		tags = append(tags, &v1.Tag{Key: k, Value: v})
	}
	return &v1.Function{
		CreatedTime:      f.CreatedTime.Unix(),
		Name:             swag.String(f.Name),
		Kind:             utils.FunctionKind,
		ID:               strfmt.UUID(f.ID),
		FaasID:           strfmt.UUID(f.FaasID),
		Image:            swag.String(f.ImageName),
		FunctionImageURL: f.FunctionImageURL,
		Handler:          f.Handler,
		Schema: &v1.Schema{
			In:  f.Schema.In,
			Out: f.Schema.Out,
		},
		Reason:   f.Reason,
		Secrets:  f.Secrets,
		Services: f.Services,
		Timeout:  f.Timeout,
		Tags:     tags,
		Status:   v1.Status(f.Status),
	}
}

func functionListToModel(funcs []*functions.Function) []*v1.Function {
	body := make([]*v1.Function, 0, len(funcs))
	for _, f := range funcs {
		body = append(body, functionEntityToModel(f))
	}
	return body
}

func schemaModelToEntity(mSchema *v1.Schema) (*functions.Schema, error) {
	schema := new(functions.Schema)
	if mSchema != nil && mSchema.In != nil {
		schema.In = new(spec.Schema)
		b, _ := json.Marshal(mSchema.In)
		if err := json.Unmarshal(b, schema.In); err != nil {
			return nil, errors.Wrap(err, "could not decode schema.in")
		}
	}
	if mSchema != nil && mSchema.Out != nil {
		schema.Out = new(spec.Schema)
		b, _ := json.Marshal(mSchema.Out)
		if err := json.Unmarshal(b, schema.Out); err != nil {
			return nil, errors.Wrap(err, "could not decode schema.out")
		}
	}
	return schema, nil
}

func functionModelOntoEntity(m *v1.Function, s *functions.Source, e *functions.Function) error {
	e.BaseEntity.Name = *m.Name
	schema, err := schemaModelToEntity(m.Schema)
	if err != nil {
		return err
	}
	e.SourceName = s.Name
	e.Handler = m.Handler
	e.ImageName = *m.Image
	e.FaasID = string(m.FaasID)
	e.Timeout = m.Timeout
	e.Tags = map[string]string{}
	for _, t := range m.Tags {
		e.Tags[t.Key] = t.Value
	}
	e.Reason = m.Reason
	e.Schema = schema
	e.Secrets = m.Secrets
	e.Services = m.Services
	return nil
}

func functionModelOntoSourceEntity(m *v1.Function) *functions.Source {
	return &functions.Source{
		BaseEntity: entitystore.BaseEntity{
			Name: uuid.NewV4().String(),
		},
		Code:     m.Source.Code,
		Function: *m.Name,
		URL:      m.Source.URL,
	}
}

func runModelToEntity(m *v1.Run, f *functions.Function) *functions.FnRun {
	secrets := f.Secrets
	if secrets == nil {
		secrets = m.Secrets
	} else {
		secrets = append(secrets, m.Secrets...)
	}
	services := f.Services
	if services == nil {
		services = m.Services
	} else {
		services = append(services, m.Services...)
	}
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	var waitChan chan struct{}
	if m.Blocking {
		waitChan = make(chan struct{})
	}
	return &functions.FnRun{
		BaseEntity: entitystore.BaseEntity{
			Name:   uuid.NewV4().String(),
			Status: f.Status,
			Reason: f.Reason,
			Tags:   tags,
		},
		Blocking:     m.Blocking,
		Input:        m.Input,
		HTTPContext:  m.HTTPContext,
		Secrets:      secrets,
		Services:     services,
		FunctionName: f.Name,
		FunctionID:   f.ID,
		FaasID:       f.FaasID,
		Event:        helpers.CloudEventFromAPI(m.Event),
		WaitChan:     waitChan,
	}
}

func runEntityToModel(f *functions.FnRun) *v1.Run {
	tags := []*v1.Tag{}
	for k, v := range f.Tags {
		tags = append(tags, &v1.Tag{Key: k, Value: v})
	}
	return &v1.Run{
		ExecutedTime: f.CreatedTime.Unix(),
		FinishedTime: f.FinishedTime.Unix(),
		Name:         strfmt.UUID(f.Name),
		Blocking:     f.Blocking,
		Input:        f.Input,
		Output:       f.Output,
		Logs:         f.Logs,
		Error:        f.Error,
		Secrets:      f.Secrets,
		HTTPContext:  f.HTTPContext,
		FunctionName: f.FunctionName,
		FunctionID:   f.FunctionID,
		FaasID:       strfmt.UUID(f.FaasID),
		Status:       v1.Status(f.Status),
		Event:        (*v1.CloudEvent)(helpers.CloudEventToAPI(f.Event)),
		Reason:       f.Reason,
		Tags:         tags,
	}
}

func runListToModel(runs []*functions.FnRun) []*v1.Run {
	body := make([]*v1.Run, 0, len(runs))
	for _, r := range runs {
		body = append(body, runEntityToModel(r))
	}
	return body
}

// Handlers is the API handler for function manager
type Handlers struct {
	Watcher controller.Watcher

	Store entitystore.EntityStore
}

// NewHandlers is the constructor for the function manager API handlers
func NewHandlers(watcher controller.Watcher, store entitystore.EntityStore) *Handlers {
	return &Handlers{
		Watcher: watcher,
		Store:   store,
	}
}

// ImageGetter retrieves image from Image Manager
type ImageGetter interface {
	GetImage(ctx context.Context, organizationID string, imageName string) (*v1.Image, error)
}

// FileImageManager is an ImageManager which is backed by a static map of images
type FileImageManager struct {
	Images map[string]map[string]*v1.Image
}

// GetImage returns an image based queried by name
func (m *FileImageManager) GetImage(ctx context.Context, organizationID string, imageName string) (*v1.Image, error) {
	if image, ok := m.Images[organizationID][imageName]; ok {
		return image, nil
	}
	return nil, fmt.Errorf("missing image %s", imageName)
}

// FileImageManagerClient returns a FileImageManager after populating the map with a JSON file
func FileImageManagerClient(imageFilePath string) *FileImageManager {
	b, err := ioutil.ReadFile(imageFilePath)
	if err != nil {
		panic(fmt.Sprintf("Failed to read image file %s", imageFilePath))
	}
	images := make(map[string]map[string]*v1.Image)
	json.Unmarshal(b, &images)
	return &FileImageManager{
		Images: images,
	}
}

// ConfigureHandlers registers the function manager handlers to the API
func (h *Handlers) ConfigureHandlers(api middleware.RoutableAPI) {
	a, ok := api.(*operations.FunctionManagerAPI)
	if !ok {
		panic("Cannot configure api")
	}

	a.CookieAuth = func(token string) (interface{}, error) {
		// TODO: be able to retrieve user information from the cookie
		// currently just return the cookie
		return token, nil
	}

	a.BearerAuth = func(token string) (interface{}, error) {
		// TODO: Once IAM issues signed tokens, validate them here.
		return token, nil
	}

	a.Logger = log.Printf
	a.StoreAddFunctionHandler = fnstore.AddFunctionHandlerFunc(h.addFunction)
	a.StoreGetFunctionHandler = fnstore.GetFunctionHandlerFunc(h.getFunction)
	a.StoreDeleteFunctionHandler = fnstore.DeleteFunctionHandlerFunc(h.deleteFunction)
	a.StoreGetFunctionsHandler = fnstore.GetFunctionsHandlerFunc(h.getFunctions)
	a.StoreUpdateFunctionHandler = fnstore.UpdateFunctionHandlerFunc(h.updateFunction)
	a.RunnerRunFunctionHandler = fnrunner.RunFunctionHandlerFunc(h.runFunction)
	a.RunnerGetRunHandler = fnrunner.GetRunHandlerFunc(h.getRun)
	a.RunnerGetRunsHandler = fnrunner.GetRunsHandlerFunc(h.getRuns)
}

func (h *Handlers) addFunction(params fnstore.AddFunctionParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	s := functionModelOntoSourceEntity(params.Body)
	s.OrganizationID = params.XDispatchOrg
	s.Status = entitystore.StatusREADY

	if len(s.Code) == 0 && s.URL == "" {
		return fnstore.NewAddFunctionBadRequest().WithPayload(&v1.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String("either Code or URL must be specified"),
		})
	}

	e := &functions.Function{}
	if err := functionModelOntoEntity(params.Body, s, e); err != nil {
		return fnstore.NewAddFunctionBadRequest().WithPayload(&v1.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(err.Error()),
		})
	}
	e.OrganizationID = params.XDispatchOrg
	e.Status = entitystore.StatusINITIALIZED
	e.FaasID = uuid.NewV4().String()

	if _, err := h.Store.Add(ctx, s); err != nil {
		if entitystore.IsUniqueViolation(err) {
			return fnstore.NewAddFunctionConflict().WithPayload(&v1.Error{
				Code:    http.StatusConflict,
				Message: utils.ErrorMsgAlreadyExists("source for function", s.Function),
			})
		}
		log.Errorf("Store error when adding new source %s for function %s: %+v", s.Name, s.Function, err)
		return fnstore.NewAddFunctionDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("source for function", s.Function),
		})
	}

	defer func() {
		if _, err := h.Store.Update(ctx, s.Revision, s); err != nil {
			log.Errorf("Store error when updating source %s for function %s: %+v", s.Name, s.Function, err)
		}
	}()

	log.Debugf("trying to add entity to store")
	log.Debugf("entity org=%s, name=%s, id=%s, status=%s", e.OrganizationID, e.Name, e.ID, e.Status)
	if _, err := h.Store.Add(ctx, e); err != nil {
		s.Status = entitystore.StatusDELETING
		if entitystore.IsUniqueViolation(err) {
			return fnstore.NewAddFunctionConflict().WithPayload(&v1.Error{
				Code:    http.StatusConflict,
				Message: utils.ErrorMsgAlreadyExists("function", e.Name),
			})
		}
		log.Errorf("Store error when adding a new function %s: %+v", e.Name, err)
		return fnstore.NewAddFunctionDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("function", e.Name),
		})
	}

	h.Watcher.OnAction(ctx, e)

	return fnstore.NewAddFunctionCreated().WithPayload(functionEntityToModel(e))
}

func (h *Handlers) getFunction(params fnstore.GetFunctionParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	e := new(functions.Function)

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return fnstore.NewGetFunctionBadRequest().WithPayload(
			&v1.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}

	if err := h.Store.Get(ctx, params.XDispatchOrg, params.FunctionName, opts, e); err != nil {
		log.Debugf("Error returned by h.Store.Get: ", err)
		log.Infof("Received GET for non-existent function %s", params.FunctionName)
		return fnstore.NewGetFunctionNotFound().WithPayload(&v1.Error{
			Code:    http.StatusNotFound,
			Message: utils.ErrorMsgNotFound("function", params.FunctionName),
		})
	}
	return fnstore.NewGetFunctionOK().WithPayload(functionEntityToModel(e))
}

func (h *Handlers) deleteFunction(params fnstore.DeleteFunctionParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	e := new(functions.Function)

	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	if err := h.Store.Get(ctx, params.XDispatchOrg, params.FunctionName, opts, e); err != nil {
		log.Debugf("Error returned by h.Store.Get: %+v", err)
		log.Infof("Received DELETE for non-existent function %s", params.FunctionName)
		return fnstore.NewDeleteFunctionNotFound().WithPayload(&v1.Error{
			Code:    http.StatusNotFound,
			Message: utils.ErrorMsgNotFound("function", params.FunctionName),
		})
	}

	e.Status = entitystore.StatusDELETING

	log.Debugf("trying to delete the entity from store")
	log.Debugf("entity org=%s, name=%s, id=%s, status=%s", e.OrganizationID, e.Name, e.ID, e.Status)
	if _, err := h.Store.Update(ctx, e.Revision, e); err != nil {
		log.Errorf("Store error when deleting a function %s: %+v", params.FunctionName, err)
		return fnstore.NewDeleteFunctionDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("function", e.Name),
		})
	}
	h.Watcher.OnAction(ctx, e)
	m := functionEntityToModel(e)
	return fnstore.NewDeleteFunctionOK().WithPayload(m)
}

func (h *Handlers) getFunctions(params fnstore.GetFunctionsParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return fnstore.NewGetFunctionsBadRequest().WithPayload(
			&v1.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}

	var funcs []*functions.Function
	err = h.Store.List(ctx, params.XDispatchOrg, opts, &funcs)
	if err != nil {
		log.Errorf("Store error when listing functions: %+v\n", err)
		return fnstore.NewGetFunctionsDefault(http.StatusInternalServerError).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("error when listing functions"),
		})
	}
	return fnstore.NewGetFunctionsOK().WithPayload(functionListToModel(funcs))
}

func (h *Handlers) updateFunction(params fnstore.UpdateFunctionParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return fnstore.NewUpdateFunctionBadRequest().WithPayload(
			&v1.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}

	e := new(functions.Function)
	err = h.Store.Get(ctx, params.XDispatchOrg, params.FunctionName, opts, e)
	if err != nil {
		log.Debugf("Error returned by h.Store.Get: %+v", err)
		log.Infof("Received update for non-existent function %s", params.FunctionName)
		return fnstore.NewDeleteFunctionNotFound().WithPayload(&v1.Error{
			Code:    http.StatusNotFound,
			Message: utils.ErrorMsgNotFound("function", params.FunctionName),
		})
	}

	s := new(functions.Source)
	err = h.Store.Get(ctx, params.XDispatchOrg, e.SourceName, entitystore.Options{}, s)
	if err != nil {
		log.Debugf("Error returned by h.Store.Get: %+v", err)
		log.Infof("Received update for non-existent source %s", e.SourceName)
		return fnstore.NewDeleteFunctionNotFound().WithPayload(&v1.Error{
			Code:    http.StatusNotFound,
			Message: utils.ErrorMsgNotFound("source for function", e.Name),
		})
	}

	updateEntity := functionModelOntoSourceEntity(params.Body)
	updateEntity.Name = e.SourceName
	updateEntity.CreatedTime = s.CreatedTime
	updateEntity.ID = s.ID
	updateEntity.OrganizationID = s.OrganizationID
	updateEntity.Status = entitystore.StatusREADY

	if len(updateEntity.Code) == 0 && updateEntity.URL == "" {
		return fnstore.NewUpdateFunctionBadRequest().WithPayload(&v1.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String("either Code or URL must be specified"),
		})
	}

	if err := functionModelOntoEntity(params.Body, updateEntity, e); err != nil {
		return fnstore.NewUpdateFunctionBadRequest().WithPayload(&v1.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(err.Error()),
		})
	}
	// generating a new UUID will force the creation of a new function in the underlying FaaS
	e.FaasID = uuid.NewV4().String()
	e.Status = entitystore.StatusUPDATING

	if _, err := h.Store.Update(ctx, s.Revision, updateEntity); err != nil {
		log.Errorf("Store error when updating source %s for function %s: %+v", s.Name, s.Function, err)
		return fnstore.NewUpdateFunctionDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("source for function", s.Function),
		})
	}

	if _, err := h.Store.Update(ctx, e.Revision, e); err != nil {
		log.Errorf("Store error when updating function %s: %+v", params.FunctionName, err)
		return fnstore.NewUpdateFunctionDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("function", e.Name),
		})
	}

	h.Watcher.OnAction(ctx, e)

	m := functionEntityToModel(e)
	return fnstore.NewUpdateFunctionOK().WithPayload(m)
}

func (h *Handlers) runFunction(params fnrunner.RunFunctionParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	var err error

	if params.FunctionName == nil {
		return fnrunner.NewRunFunctionBadRequest().WithPayload(&v1.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String("Bad Request: No function specified"),
		})
	}

	if params.Body == nil {
		return fnrunner.NewRunFunctionBadRequest().WithPayload(&v1.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String("Bad Request: Invalid Payload"),
		})
	}
	log.Debugf("Execute a function with payload: %#v", *params.Body)

	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return fnrunner.NewRunFunctionBadRequest().WithPayload(
			&v1.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	f := new(functions.Function)
	if err := h.Store.Get(ctx, params.XDispatchOrg, *params.FunctionName, opts, f); err != nil {
		log.Debugf("Error returned by h.Store.Get: %+v", err)
		log.Infof("Trying to create run for non-existent function %s", *params.FunctionName)
		return fnrunner.NewRunFunctionNotFound().WithPayload(&v1.Error{
			Code:    http.StatusNotFound,
			Message: utils.ErrorMsgNotFound("function", *params.FunctionName),
		})
	}

	if f.Status != entitystore.StatusREADY {
		return fnrunner.NewRunFunctionNotFound().WithPayload(&v1.Error{
			Code:    http.StatusNotFound,
			Message: swag.String(fmt.Sprintf("function %s is not READY", *params.FunctionName)),
		})
	}

	run := runModelToEntity(params.Body, f)
	run.OrganizationID = params.XDispatchOrg
	run.Status = entitystore.StatusINITIALIZED

	if _, err := h.Store.Add(ctx, run); err != nil {
		log.Errorf("Store error when adding new function run %s: %+v", run.Name, err)
		return fnrunner.NewRunFunctionDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("function run", run.Name),
		})
	}

	h.Watcher.OnAction(ctx, run)

	if run.Blocking {
		run.Wait()
		return fnrunner.NewRunFunctionOK().WithPayload(runEntityToModel(run))
	}

	return fnrunner.NewRunFunctionAccepted().WithPayload(runEntityToModel(run))
}

func (h *Handlers) getRun(params fnrunner.GetRunParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	run := functions.FnRun{}

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}

	if params.Since != nil {
		opts.Filter.Add(
			entitystore.FilterStat{
				Scope:   entitystore.FilterScopeField,
				Subject: "ModifiedTime",
				Verb:    entitystore.FilterVerbAfter,
				Object:  time.Unix(*params.Since, 0),
			})
	}

	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return fnrunner.NewGetRunBadRequest().WithPayload(
			&v1.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}

	err = h.Store.Get(ctx, params.XDispatchOrg, params.RunName.String(), opts, &run)
	if err != nil || (params.FunctionName != nil && run.FunctionName != *params.FunctionName) {
		log.Debugf("Error returned by h.Store.Get: %+v", err)
		if params.FunctionName != nil {
			log.Infof("Get run failed for function %s and run %s", *params.FunctionName, params.RunName.String())
		} else {
			log.Infof("Get run failed for function run %s", params.RunName.String())
		}
		return fnrunner.NewGetRunNotFound().WithPayload(&v1.Error{
			Code:    http.StatusNotFound,
			Message: utils.ErrorMsgNotFound("function run", run.Name),
		})
	}
	return fnrunner.NewGetRunOK().WithPayload(runEntityToModel(&run))
}

func getFilteredRuns(ctx context.Context, store entitystore.EntityStore, orgID string, functionName *string, since *int64, tags []string) ([]*functions.FnRun, error) {
	var runs []*functions.FnRun
	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}

	if functionName != nil {
		opts.Filter.Add(
			entitystore.FilterStat{
				Scope:   entitystore.FilterScopeExtra,
				Subject: "FunctionName",
				Verb:    entitystore.FilterVerbEqual,
				Object:  *functionName,
			})
	}

	if since != nil {
		opts.Filter.Add(
			entitystore.FilterStat{
				Scope:   entitystore.FilterScopeField,
				Subject: "ModifiedTime",
				Verb:    entitystore.FilterVerbAfter,
				Object:  time.Unix(*since, 0),
			})
	}

	opts.Filter, err = utils.ParseTags(opts.Filter, tags)
	if err != nil {
		return nil, dispatcherrors.NewRequestError(err)
	}

	if err = store.List(ctx, orgID, opts, &runs); err != nil {
		if functionName != nil {
			log.Errorf("Store error when listing runs for function %s: %+v", *functionName, err)
		} else {
			log.Errorf("Store error when listing runs: %+v", err)
		}
		return nil, dispatcherrors.NewServerError(err)
	}
	return runs, nil
}

func (h *Handlers) getRuns(params fnrunner.GetRunsParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	runs, err := getFilteredRuns(ctx, h.Store, params.XDispatchOrg, params.FunctionName, params.Since, params.Tags)

	switch err.(type) {
	case *dispatcherrors.RequestError:
		return fnrunner.NewGetRunsBadRequest().WithPayload(
			&v1.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	case *dispatcherrors.ServerError:
		return fnrunner.NewGetRunsDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("error when listing function runs"),
		})
	}
	return fnrunner.NewGetRunsOK().WithPayload(runListToModel(runs))
}
