///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package vcenter

import (
	log "github.com/sirupsen/logrus"
	"github.com/vmware/govmomi/vim25/types"
)

// NO TESTS

func processEventMetadata(e types.BaseEvent) interface{} {
	// TODO: Implement automated way of generating payload based on vSphere API WSDL
	switch concreteEvent := e.(type) {
	case *types.VmBeingCreatedEvent:
		return handleVMBeingCreatedEvent(concreteEvent)
	case *types.VmBeingDeployedEvent:
		return handleVMBeingDeployedEvent(concreteEvent)
	case *types.VmDeployedEvent:
		return handleVMDeployedEvent(concreteEvent)
	case *types.VmEvent:
		return handleVMEvent(concreteEvent)
	default:
		log.Debug("No dedicated handler")
		return nil
	}
}

func handleVMBeingCreatedEvent(e *types.VmBeingCreatedEvent) interface{} {
	return struct {
		VMName  string `json:"vm_name"`
		VMID    string `json:"vm_id"`
		NumCPUs int32  `json:"num_cpus"`
		MemMB   int64  `json:"mem_mb"`
	}{
		VMName:  e.ConfigSpec.Name,
		VMID:    e.Vm.Vm.String(),
		NumCPUs: e.ConfigSpec.NumCPUs * e.ConfigSpec.NumCoresPerSocket,
		MemMB:   e.ConfigSpec.MemoryMB,
	}
}

func handleVMBeingDeployedEvent(e *types.VmBeingDeployedEvent) interface{} {
	return struct {
		VMName      string `json:"vm_name"`
		VMID        string `json:"vm_id"`
		SrcTemplate string `json:"src_template"`
	}{
		VMName:      e.Vm.Name,
		VMID:        e.Vm.Vm.String(),
		SrcTemplate: e.SrcTemplate.Name,
	}
}

func handleVMDeployedEvent(e *types.VmDeployedEvent) interface{} {
	return struct {
		VMName      string `json:"vm_name"`
		VMID        string `json:"vm_id"`
		SrcTemplate string `json:"src_template"`
	}{
		VMName:      e.Vm.Name,
		VMID:        e.Vm.Vm.String(),
		SrcTemplate: e.SrcTemplate.Name,
	}
}

func handleVMEvent(e *types.VmEvent) interface{} {
	return struct {
		VMName     string `json:"vm_name"`
		VMID       string `json:"vm_id"`
		IsTemplate bool   `json:"is_template"`
	}{
		VMName:     e.Vm.Name,
		VMID:       e.Vm.Vm.String(),
		IsTemplate: e.Template,
	}
}
