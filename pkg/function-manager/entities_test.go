///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////package functionmanager

import (
	"testing"
)

func TestFnRun_doneNoPanicByDefault(t *testing.T) {
	f := new(FnRun)
	f.wait()
	f.done()
}

func TestFnRun_waitDone(t *testing.T) {
	f := new(FnRun)
	f.waitChan = make(chan struct{})

	go func() {
		f.done()
	}()
	f.wait()
}
