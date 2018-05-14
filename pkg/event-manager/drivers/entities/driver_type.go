///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package entities

import (
	"github.com/go-openapi/swag"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/utils"
)

// NO TESTS

// DriverType represents a custom type of driver (e.g. timer-driver)
type DriverType struct {
	entitystore.BaseEntity
	Image   string            `json:"image"`
	BuiltIn bool              `json:"builtIn"`
	Config  map[string]string `json:"config,omitempty"`
}

// ToModel creates swagger model from driver type
func (dt *DriverType) ToModel() *v1.EventDriverType {
	var tags []*v1.Tag
	for k, v := range dt.Tags {
		tags = append(tags, &v1.Tag{Key: k, Value: v})
	}
	var mconfig []*v1.Config
	for k, v := range dt.Config {
		mconfig = append(mconfig, &v1.Config{Key: k, Value: v})
	}
	return &v1.EventDriverType{
		Name:         swag.String(dt.Name),
		Image:        swag.String(dt.Image),
		Kind:         utils.DriverTypeKind,
		BuiltIn:      swag.Bool(false),
		Config:       mconfig,
		CreatedTime:  dt.CreatedTime.Unix(),
		ModifiedTime: dt.ModifiedTime.Unix(),
		Tags:         tags,
	}
}

// FromModel builds driver type from swagger model
func (dt *DriverType) FromModel(m *v1.EventDriverType, orgID string) {
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	config := make(map[string]string)
	for _, c := range m.Config {
		config[c.Key] = c.Value
	}

	dt.BaseEntity.OrganizationID = orgID
	dt.BaseEntity.Name = *m.Name
	dt.BaseEntity.Tags = tags
	dt.BuiltIn = false
	dt.Image = *m.Image
	dt.Config = config
}
