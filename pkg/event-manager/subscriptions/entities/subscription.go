///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package entities

import (
	"github.com/go-openapi/swag"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/entity-store"
)

// NO TESTS

// Subscription struct represents a single subscription of subscriber to publisher
type Subscription struct {
	entitystore.BaseEntity
	EventType string   `json:"eventType"`
	Function  string   `json:"function"`
	Secrets   []string `json:"secrets,omitempty"`
}

// ToModel converts subscription to swagger model
func (s *Subscription) ToModel() *v1.Subscription {
	var tags []*v1.Tag
	for k, v := range s.Tags {
		tags = append(tags, &v1.Tag{Key: k, Value: v})
	}
	m := v1.Subscription{
		Name:         swag.String(s.Name),
		Kind:         v1.SubscriptionKind,
		EventType:    swag.String(s.EventType),
		Function:     &s.Function,
		Status:       v1.Status(s.Status),
		Secrets:      s.Secrets,
		CreatedTime:  s.CreatedTime.Unix(),
		ModifiedTime: s.ModifiedTime.Unix(),
		Tags:         tags,
	}
	return &m
}

// FromModel builds subscription based on swagger model
func (s *Subscription) FromModel(m *v1.Subscription, orgID string) {
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	s.BaseEntity.OrganizationID = orgID
	s.BaseEntity.Name = *m.Name
	s.BaseEntity.Status = entitystore.Status(m.Status)
	s.BaseEntity.Tags = tags
	s.EventType = *m.EventType
	s.Function = *m.Function
	s.Secrets = m.Secrets
}
