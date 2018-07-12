///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import "os/exec"

// Runner interface for executing commands
type Runner interface {
	Run(name string, arg ...string) ([]byte, error)
}

type execCmdRunner struct{}

func (r execCmdRunner) Run(name string, arg ...string) ([]byte, error) {
	return exec.Command(name, arg...).CombinedOutput()
}
