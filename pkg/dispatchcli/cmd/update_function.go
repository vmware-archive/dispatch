///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"github.com/vmware/dispatch/pkg/function-manager/gen/client/store"
	"github.com/vmware/dispatch/pkg/function-manager/gen/models"
)

// CallUpdateFunction makes the API call to update a function
func CallUpdateFunction(input interface{}) error {
	function := input.(*models.Function)
	params := store.NewUpdateFunctionParams()
	params.FunctionName = *function.Name
	params.Body = function
	_, err := functionManagerClient().Store.UpdateFunction(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	return nil
}
