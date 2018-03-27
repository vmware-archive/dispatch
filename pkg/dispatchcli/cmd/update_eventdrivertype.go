///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"github.com/vmware/dispatch/pkg/event-manager/gen/client/drivers"
	"github.com/vmware/dispatch/pkg/event-manager/gen/models"
)

// CallUpdateDriverType makes the API call to update a driver type
func CallUpdateDriverType(input interface{}) error {
	driverType := input.(*models.DriverType)
	params := drivers.NewUpdateDriverTypeParams()
	params.DriverTypeName = *driverType.Name
	params.Body = driverType
	_, err := eventManagerClient().Drivers.UpdateDriverType(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	return nil
}
