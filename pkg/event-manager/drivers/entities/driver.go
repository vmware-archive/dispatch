///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package entities

import (
	"github.com/go-openapi/swag"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/event-manager/gen/models"
)

// NO TESTS

// Driver represents an event driver instance(e.g. vcenter1.corp.local)
type Driver struct {
	entitystore.BaseEntity
	Type    string            `json:"type"`
	Config  map[string]string `json:"config,omitempty"`
	Secrets []string          `json:"secrets,omitempty"`
	Image   string            `json:"image"`
	Mode    string            `josn:"mode"`
}

// ToModel creates swagger model from the driver struct
func (d *Driver) ToModel() *models.Driver {
	var tags []*models.Tag
	for k, v := range d.Tags {
		tags = append(tags, &models.Tag{Key: k, Value: v})
	}
	var mconfig []*models.Config
	for k, v := range d.Config {
		mconfig = append(mconfig, &models.Config{Key: k, Value: v})
	}
	return &models.Driver{
		Name:         swag.String(d.Name),
		Type:         swag.String(d.Type),
		Config:       mconfig,
		Status:       models.Status(d.Status),
		CreatedTime:  d.CreatedTime.Unix(),
		ModifiedTime: d.ModifiedTime.Unix(),
		Secrets:      d.Secrets,
		Tags:         tags,
	}
}

// FromModel builds a driver struct from swagger model
func (d *Driver) FromModel(m *models.Driver, orgID string) {
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	config := make(map[string]string)
	for _, c := range m.Config {
		config[c.Key] = c.Value
	}
	d.BaseEntity = entitystore.BaseEntity{
		OrganizationID: orgID,
		Name:           *m.Name,
		Tags:           tags,
	}
	d.Type = *m.Type
	d.Config = config
	d.Secrets = m.Secrets

}
