///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package drivers

import (
	"context"

	knFeedTypes "github.com/knative/eventing/pkg/apis/feeds/v1alpha1"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/event-manager/drivers/entities"
	"github.com/vmware/dispatch/pkg/utils/knaming"
)

// NO TEST

// Backend defines the event driver backend interface
type Backend interface {
	Deploy(context.Context, *entities.Driver) error
	Expose(context.Context, *entities.Driver) error
	Update(context.Context, *entities.Driver) error
	Delete(context.Context, *entities.Driver) error
}

// FromDriverType produces a Knative EventSouce from a Dispatch DriverType
func FromDriverType(driver *v1.EventDriverType) *knFeedTypes.EventSource {
	eventSource := &knFeedTypes.EventSource{
		ObjectMeta: *knaming.ToObjectMeta(driver),
		Spec: knFeedTypes.EventSourceSpec{
			CommonEventSourceSpec: knFeedTypes.CommonEventSourceSpec{
				Source:     *driver.Name,
				Image:      *driver.Image,
				Parameters: nil,
			},
		},
	}

	return eventSource
}

// ToDriverType produces a Dispatch DriverType from a Knative EventSource
func ToDriverType(eventSource *knFeedTypes.EventSource) *v1.EventDriverType {
	driverType := &v1.EventDriverType{}
	knaming.FromObjectMeta(&eventSource.ObjectMeta, driverType)
	return driverType
}
