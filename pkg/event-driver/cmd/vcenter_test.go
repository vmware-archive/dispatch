///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewEventDriverCmd(t *testing.T) {

	os.Setenv("VCENTERURL", "127.0.0.1/testendpoint")
	command := NewEventDriverCmd(nil, nil, nil)
	command.SetArgs([]string{"vcenter"})
	command.Execute()

	assert.Equal(t, "127.0.0.1/testendpoint", viper.GetString("vcenterurl"))
}
