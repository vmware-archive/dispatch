///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"github.com/vmware/dispatch/pkg/event-manager/gen/client/drivers"
	"github.com/vmware/dispatch/pkg/event-manager/gen/models"
)

// CallUpdateDriver makes the API call to update an event driver
func CallUpdateDriver(input interface{}) error {
	eventDriver := input.(*models.Driver)
	params := drivers.NewUpdateDriverParams()
	params.DriverName = *eventDriver.Name
	params.Body = eventDriver
	_, err := eventManagerClient().Drivers.UpdateDriver(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	return nil
}
