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

// Driver represents an event driver instance(e.g. vcenter1.corp.local)
type Driver struct {
	entitystore.BaseEntity
	Type    string            `json:"type"`
	Config  map[string]string `json:"config,omitempty"`
	Secrets []string          `json:"secrets,omitempty"`
	Image   string            `json:"image"`
	Expose  bool              `json:"expose"`
	URL     string            `json:"url"`
}

// ToModel creates swagger model from the driver struct
func (d *Driver) ToModel() *v1.EventDriver {
	var tags []*v1.Tag
	for k, v := range d.Tags {
		tags = append(tags, &v1.Tag{Key: k, Value: v})
	}
	var mconfig []*v1.Config
	for k, v := range d.Config {
		mconfig = append(mconfig, &v1.Config{Key: k, Value: v})
	}
	return &v1.EventDriver{
		Name:         swag.String(d.Name),
		Type:         swag.String(d.Type),
		Kind:         utils.DriverKind,
		Config:       mconfig,
		Status:       v1.Status(d.Status),
		CreatedTime:  d.CreatedTime.Unix(),
		ModifiedTime: d.ModifiedTime.Unix(),
		Secrets:      d.Secrets,
		URL:          d.URL,
		Expose:       d.Expose,
		Tags:         tags,
		Reason:       d.Reason,
	}
}

// FromModel builds a driver struct from swagger model
func (d *Driver) FromModel(m *v1.EventDriver, orgID string) {
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	config := make(map[string]string)
	for _, c := range m.Config {
		config[c.Key] = c.Value
	}
	d.BaseEntity.OrganizationID = orgID
	d.BaseEntity.Name = *m.Name
	d.BaseEntity.Tags = tags
	d.Type = *m.Type
	d.Config = config
	d.Secrets = m.Secrets
	d.URL = m.URL
	d.Expose = m.Expose
	d.Reason = m.Reason
}
