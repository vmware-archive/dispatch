///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package imagemanager

import (
	"testing"
	"time"

	"github.com/vmware/dispatch/pkg/entity-store"

	"github.com/stretchr/testify/assert"
)

func TestImageEntityHandlerCheckAndLock(t *testing.T) {
	ih := imageEntityHandler{timeout: time.Millisecond}
	e := &Image{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: "test",
			ID:             "deadbeef",
		},
	}
	assert.False(t, ih.CheckAndLock(e))
	assert.True(t, ih.CheckAndLock(e))
	time.Sleep(time.Millisecond)
	assert.False(t, ih.CheckAndLock(e))
}
