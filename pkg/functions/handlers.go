///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functions

import (
	"log"

	"github.com/go-openapi/runtime/middleware"
	"github.com/vmware/dispatch/pkg/functions/gen/restapi/operations"
	fnrunner "github.com/vmware/dispatch/pkg/functions/gen/restapi/operations/runner"
	fnstore "github.com/vmware/dispatch/pkg/functions/gen/restapi/operations/store"
)

//Handlers interface declares methods needed to implement functions API
type Handlers interface {
	addFunction(params fnstore.AddFunctionParams) middleware.Responder
	getFunction(params fnstore.GetFunctionParams) middleware.Responder
	deleteFunction(params fnstore.DeleteFunctionParams) middleware.Responder
	getFunctions(params fnstore.GetFunctionsParams) middleware.Responder
	updateFunction(params fnstore.UpdateFunctionParams) middleware.Responder
	runFunction(params fnrunner.RunFunctionParams) middleware.Responder
	getRun(params fnrunner.GetRunParams) middleware.Responder
	getRuns(params fnrunner.GetRunsParams) middleware.Responder
}

// ConfigureHandlers registers the function manager knHandlers to the API
func ConfigureHandlers(api middleware.RoutableAPI, h Handlers) {
	a, ok := api.(*operations.FunctionsAPI)
	if !ok {
		panic("Cannot configure api")
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
