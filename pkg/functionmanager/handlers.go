///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package functionmanager

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	entitystore "gitlab.eng.vmware.com/serverless/serverless/pkg/entity-store"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functionmanager/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functionmanager/gen/restapi/operations"
	fnrunner "gitlab.eng.vmware.com/serverless/serverless/pkg/functionmanager/gen/restapi/operations/runner"
	fnstore "gitlab.eng.vmware.com/serverless/serverless/pkg/functionmanager/gen/restapi/operations/store"
)

// FunctionManagerFlags are configuration flags for the function manager
var FunctionManagerFlags = struct {
	DbFile string `long:"db-file" description:"Path to BoltDB file" default:"db.bolt"`
	OrgID  string `long:"organization" description:"(temporary) Static organization id" default:"serverless"`
}{}

func functionEntityToModel(f *Function) *models.Function {
	var tags []*models.Tag
	for k, v := range f.Tags {
		tags = append(tags, &models.Tag{Key: k, Value: v})
	}
	schema := models.Schema(f.Schema)
	m := models.Function{
		CreatedTime: f.CreatedTime.Unix(),
		Name:        swag.String(f.Name),
		ID:          strfmt.UUID(f.ID),
		Image:       swag.String(f.Image),
		Code:        swag.String(f.Code),
		Schema:      &schema,
		Language:    models.Language(f.Language),
		Active:      f.Active,

		Tags: tags,
	}
	return &m
}

func functionListToModel(funcs []Function) models.GetFunctionsOKBody {
	body := make(models.GetFunctionsOKBody, 0, len(funcs))
	for _, f := range funcs {
		body = append(body, functionEntityToModel(&f))
	}
	return body
}

func functionModelToEntity(m *models.Function) *Function {
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	e := Function{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: FunctionManagerFlags.OrgID,
			Name:           *m.Name,
			Tags:           tags,
		},
		Active:   m.Active,
		Code:     *m.Code,
		Image:    *m.Image,
		Schema:   Schema(*m.Schema),
		Language: string(m.Language),
	}
	e.ID = string(m.ID)
	return &e
}

func runModelToEntity(m *models.Run) *FnRun {
	e := FnRun{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: FunctionManagerFlags.OrgID,
		},
		Blocking:  m.Blocking,
		Arguments: m.Arguments,
	}
	e.ID = string(m.ID)
	return &e
}

func runEntityToModel(f *FnRun) *models.Run {
	m := models.Run{
		ExecutedTime: f.CreatedTime.Unix(),
		FinishedTime: f.ModifiedTime.Unix(),
		ID:           strfmt.UUID(f.ID),
		Blocking:     f.Blocking,
		Arguments:    f.Arguments,
	}
	return &m
}

func runListToModel(runs []FnRun) models.GetRunsOKBody {
	body := make(models.GetRunsOKBody, 0, len(runs))
	for _, r := range runs {
		body = append(body, runEntityToModel(&r))
	}
	return body
}

// ConfigureHandlers registers the function manager handlers to the API
func ConfigureHandlers(api middleware.RoutableAPI, store entitystore.EntityStore) {

	a, ok := api.(*operations.FunctionManagerAPI)
	if !ok {
		panic("Cannot configure api")
	}

	a.StoreAddFunctionHandler = fnstore.AddFunctionHandlerFunc(func(params fnstore.AddFunctionParams) middleware.Responder {
		functionRequest := params.Body
		e := functionModelToEntity(functionRequest)
		_, err := store.Add(e)
		if err != nil {
			return fnstore.NewAddFunctionMethodNotAllowed()
		}
		m := functionEntityToModel(e)
		return fnstore.NewAddFunctionAccepted().WithPayload(m)
	})

	a.StoreGetFunctionByIDHandler = fnstore.GetFunctionByIDHandlerFunc(func(params fnstore.GetFunctionByIDParams) middleware.Responder {
		e := Function{}
		err := store.GetById(FunctionManagerFlags.OrgID, params.FunctionID.String(), &e)
		if err != nil {
			return fnstore.NewGetFunctionByIDNotFound()
		}
		m := functionEntityToModel(&e)
		return fnstore.NewGetFunctionByIDOK().WithPayload(m)
	})

	a.StoreDeleteFunctionByIDHandler = fnstore.DeleteFunctionByIDHandlerFunc(func(params fnstore.DeleteFunctionByIDParams) middleware.Responder {
		e := Function{}
		err := store.GetById(FunctionManagerFlags.OrgID, params.FunctionID.String(), &e)
		if err != nil {
			return fnstore.NewDeleteFunctionByIDNotFound()
		}
		err = store.Delete(FunctionManagerFlags.OrgID, params.FunctionID.String(), &e)
		if err != nil {
			return fnstore.NewDeleteFunctionByIDBadRequest()
		}
		return fnstore.NewDeleteFunctionByIDNoContent()
	})

	a.StoreGetFunctionsHandler = fnstore.GetFunctionsHandlerFunc(func(params fnstore.GetFunctionsParams) middleware.Responder {
		funcs := []Function{}
		err := store.List(FunctionManagerFlags.OrgID, nil, funcs)
		if err != nil {
			return fnstore.NewGetFunctionsDefault(500)
		}
		return fnstore.NewGetFunctionsOK().WithPayload(functionListToModel(funcs))
	})

	a.StoreUpdateFunctionByIDHandler = fnstore.UpdateFunctionByIDHandlerFunc(func(params fnstore.UpdateFunctionByIDParams) middleware.Responder {
		e := Function{}
		err := store.GetById(FunctionManagerFlags.OrgID, params.FunctionID.String(), &e)
		if err != nil {
			return fnstore.NewDeleteFunctionByIDNotFound()
		}
		e.Active = params.Body.Active
		e.Code = *params.Body.Code
		tags := make(map[string]string)
		for _, t := range params.Body.Tags {
			tags[t.Key] = t.Value
		}
		e.Tags = tags
		e.Schema = Schema(*params.Body.Schema)
		_, err = store.Update(e.Revision, &e)
		if err != nil {
			return fnstore.NewUpdateFunctionByIDBadRequest()
		}
		m := functionEntityToModel(&e)
		return fnstore.NewGetFunctionByIDOK().WithPayload(m)
	})

	a.RunnerRunFunctionHandler = fnrunner.RunFunctionHandlerFunc(func(params fnrunner.RunFunctionParams) middleware.Responder {
		e := Function{}
		err := store.GetById(FunctionManagerFlags.OrgID, params.FunctionID.String(), &e)
		if err != nil {
			return fnrunner.NewRunFunctionNotFound()
		}
		run := runModelToEntity(params.Body)
		run.FunctionID = e.ID
		_, err = store.Add(run)
		if err != nil {
			return fnrunner.NewRunFunctionInternalServerError()
		}
		return fnrunner.NewRunFunctionAccepted().WithPayload(runEntityToModel(run))
	})

	a.RunnerGetRunByIDHandler = fnrunner.GetRunByIDHandlerFunc(func(params fnrunner.GetRunByIDParams) middleware.Responder {
		run := FnRun{}
		err := store.GetById(FunctionManagerFlags.OrgID, params.RunID.String(), &run)
		if err != nil || run.FunctionID != string(params.FunctionID) {
			return fnrunner.NewGetRunByIDNotFound()
		}
		return fnrunner.NewGetRunByIDOK().WithPayload(runEntityToModel(&run))
	})

	a.RunnerGetRunsHandler = fnrunner.GetRunsHandlerFunc(func(params fnrunner.GetRunsParams) middleware.Responder {
		e := Function{}
		err := store.GetById(FunctionManagerFlags.OrgID, params.FunctionID.String(), &e)
		if err != nil {
			return fnrunner.NewGetRunsNotFound()
		}
		filter := func(e entitystore.Entity) bool {
			if run, ok := e.(*FnRun); ok {
				return run.FunctionID == params.FunctionID.String()
			}
			return false
		}
		runs := []FnRun{}
		err = store.List(FunctionManagerFlags.OrgID, entitystore.Filter(filter), runs)
		if err != nil {
			return fnrunner.NewGetRunsNotFound()
		}
		return fnrunner.NewGetRunsOK().WithPayload(runListToModel(runs))
	})
}
