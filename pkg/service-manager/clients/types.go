///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package clients

import (
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/service-manager/entities"
)

// BrokerClient defines the event driver backend interface
type BrokerClient interface {
	ListServiceClasses() ([]entitystore.Entity, error)
	ListServiceInstances() ([]entitystore.Entity, error)
	ListServiceBindings() ([]entitystore.Entity, error)
	CreateService(*entities.ServiceClass, *entities.ServiceInstance) error
	CreateBinding(*entities.ServiceInstance, *entities.ServiceBinding) error
	DeleteService(*entities.ServiceInstance) error
}
