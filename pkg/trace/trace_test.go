///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package trace

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestTrace(t *testing.T) {
	var buf bytes.Buffer
	Logger.Out = &buf
	Logger.SetLevel(logrus.DebugLevel)
	end := Trace("Test")
	buf.Write([]byte("Content\n"))
	end()

	assert.Contains(t, buf.String(), "[BEGIN]")
	assert.Contains(t, buf.String(), "TestTrace")
	assert.Contains(t, buf.String(), "Test")
	assert.Contains(t, buf.String(), "Content")
	assert.Contains(t, buf.String(), "[END]")
}
