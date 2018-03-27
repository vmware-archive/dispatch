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

// Subscription struct represents a single subscription of subscriber to publisher
type Subscription struct {
	entitystore.BaseEntity
	EventType  string   `json:"eventType"`
	SourceType string   `json:"sourceType"`
	Function   string   `json:"function"`
	Secrets    []string `json:"secrets,omitempty"`
}

// ToModel converts subscription to swagger model
func (s *Subscription) ToModel() *models.Subscription {
	var tags []*models.Tag
	for k, v := range s.Tags {
		tags = append(tags, &models.Tag{Key: k, Value: v})
	}
	m := models.Subscription{
		Name:         swag.String(s.Name),
		EventType:    swag.String(s.EventType),
		SourceType:   swag.String(s.SourceType),
		Function:     &s.Function,
		Status:       models.Status(s.Status),
		Secrets:      s.Secrets,
		CreatedTime:  s.CreatedTime.Unix(),
		ModifiedTime: s.ModifiedTime.Unix(),
		Tags:         tags,
	}
	return &m
}

// FromModel builds subscription based on swagger model
func (s *Subscription) FromModel(m *models.Subscription, orgID string) {
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	s.BaseEntity.OrganizationID = orgID
	s.BaseEntity.Name = *m.Name
	s.BaseEntity.Status = entitystore.Status(m.Status)
	s.BaseEntity.Tags = tags
	s.EventType = *m.EventType
	s.SourceType = *m.SourceType
	s.Function = *m.Function
	s.Secrets = m.Secrets
}
