///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package utils

import (
	"testing"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func TestBackoff(t *testing.T) {
	n := 95

	assert.NoError(t, Backoff(5*time.Second, func() error {
		n = n / 2
		r := n % 2
		if r == 0 {
			return nil
		}
		return errors.Errorf("n = %v, r = %v", n, r)
	}))
}
