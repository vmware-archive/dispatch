///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package identitymanager

import (
	"fmt"

	"github.com/casbin/casbin/model"
	"github.com/casbin/casbin/persist"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/entity-store"
)

type CasbinEntityAdapter struct {
	store entitystore.EntityStore
}

func NewCasbinEntityAdapter(store entitystore.EntityStore) *CasbinEntityAdapter {
	return &CasbinEntityAdapter{store: store}
}

func (a *CasbinEntityAdapter) LoadPolicy(model model.Model) error {
	var policies []*Policy

	// Policy status like CREATING, DELETING doesn't have any effect on
	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}

	log.Debug("Reloading policies")
	err := a.store.List(IdentityManagerFlags.OrgID, opts, &policies)
	for _, policy := range policies {
		// Casbin authorization rules are of the form (subject, resource, action) and hence the need to iterate over all rule fields.
		for _, rule := range policy.Rules {
			for _, subject := range rule.Subjects {
				for _, resource := range rule.Resources {
					for _, action := range rule.Actions {
						log.Debugf("Loading policy %s: rule %s, %s, %s", policy.Name, subject, resource, action)
						lineText := fmt.Sprintf("p, %s, %s, %s", subject, resource, action)
						persist.LoadPolicyLine(lineText, model)
					}
				}
			}
		}
	}
	return err
}

// SavePolicy saves all policy rules to the storage.
func (a *CasbinEntityAdapter) SavePolicy(model model.Model) error {
	return errors.New("not implemented")
}

// AddPolicy adds a policy rule to the storage.
func (a *CasbinEntityAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	return errors.New("not implemented")
}

// RemovePolicy removes a policy rule from the storage.
func (a *CasbinEntityAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return errors.New("not implemented")
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
func (a *CasbinEntityAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return errors.New("not implemented")
}
