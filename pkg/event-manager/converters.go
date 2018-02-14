///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/event-manager/gen/models"
	events "github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/trace"
)

func subscriptionModelToEntity(m *models.Subscription) *Subscription {
	defer trace.Tracef("topic: %s, function: %s", *m.EventType, *m.Function)()
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	e := Subscription{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: EventManagerFlags.OrgID,
			Name:           *m.Name,
			Status:         entitystore.Status(m.Status),
			Tags:           tags,
		},
		EventType:  *m.EventType,
		SourceName: *m.SourceName,
		SourceType: *m.SourceType,
		Function:   *m.Function,
		Secrets:    m.Secrets,
	}
	return &e
}

func subscriptionEntityToModel(sub *Subscription) *models.Subscription {
	defer trace.Tracef("topic: %s, function: %s", sub.EventType, sub.Function)()

	var tags []*models.Tag
	for k, v := range sub.Tags {
		tags = append(tags, &models.Tag{Key: k, Value: v})
	}
	m := models.Subscription{
		Name:         swag.String(sub.Name),
		EventType:    swag.String(sub.EventType),
		SourceName:   swag.String(sub.SourceName),
		SourceType:   swag.String(sub.SourceType),
		Function:     &sub.Function,
		Status:       models.Status(sub.Status),
		Secrets:      sub.Secrets,
		CreatedTime:  sub.CreatedTime.Unix(),
		ModifiedTime: sub.ModifiedTime.Unix(),
		Tags:         tags,
	}
	return &m
}

func driverModelToEntity(m *models.Driver) *Driver {
	defer trace.Tracef("driverModelToEntity")
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	config := make(map[string]string)
	for _, c := range m.Config {
		config[c.Key] = c.Value
	}
	return &Driver{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: EventManagerFlags.OrgID,
			Name:           *m.Name,
			Tags:           tags,
		},
		Type:    *m.Type,
		Config:  config,
		Secrets: m.Secrets,
	}
}

func driverEntityToModel(d *Driver) *models.Driver {
	defer trace.Tracef("type: %s, name: %s", d.Name, d.Type)

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

func driverTypeModelToEntity(m *models.DriverType) *DriverType {
	defer trace.Tracef("name: %s, image: %s", *m.Name, *m.Image)
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	config := make(map[string]string)
	for _, c := range m.Config {
		config[c.Key] = c.Value
	}

	return &DriverType{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: EventManagerFlags.OrgID,
			Name:           *m.Name,
			Tags:           tags,
		},
		Image:  *m.Image,
		Mode:   *m.Mode,
		Config: config,
	}
}

func driverTypeEntityToModel(d *DriverType) *models.DriverType {
	defer trace.Tracef("name: %s, image: %s", d.Name, d.Image)

	var tags []*models.Tag
	for k, v := range d.Tags {
		tags = append(tags, &models.Tag{Key: k, Value: v})
	}
	var mconfig []*models.Config
	for k, v := range d.Config {
		mconfig = append(mconfig, &models.Config{Key: k, Value: v})
	}
	return &models.DriverType{
		Name:         swag.String(d.Name),
		Image:        swag.String(d.Image),
		Mode:         swag.String(d.Mode),
		Config:       mconfig,
		Status:       models.Status(d.Status),
		CreatedTime:  d.CreatedTime.Unix(),
		ModifiedTime: d.ModifiedTime.Unix(),
		Tags:         tags,
	}
}

func cloudEventFromSwagger(e *models.CloudEvent) *events.CloudEvent {
	return &events.CloudEvent{
		Namespace:          *e.Namespace,
		EventType:          *e.EventType,
		EventTypeVersion:   e.EventTypeVersion,
		CloudEventsVersion: *e.CloudEventsVersion,
		SourceType:         *e.SourceType,
		SourceID:           *e.SourceID,
		EventID:            *e.EventID,
		EventTime:          time.Time(e.EventTime),
		SchemaURL:          e.SchemaURL,
		ContentType:        e.ContentType,
		Extensions:         events.CloudEventExtensions(e.Extensions),
		Data:               e.Data,
	}
}

func cloudEventToSwagger(e *events.CloudEvent) *models.CloudEvent {
	return &models.CloudEvent{
		CloudEventsVersion: swag.String(e.CloudEventsVersion),
		ContentType:        e.ContentType,
		Data:               e.Data,
		EventID:            swag.String(e.EventID),
		EventTime:          strfmt.DateTime(e.EventTime),
		EventType:          swag.String(e.EventType),
		EventTypeVersion:   e.EventTypeVersion,
		Extensions:         e.Extensions,
		Namespace:          swag.String(e.Namespace),
		SchemaURL:          e.SchemaURL,
		SourceID:           swag.String(e.SourceID),
		SourceType:         swag.String(e.SourceType),
	}
}
