///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functions

import (
	"testing"
)

func TestFnRun_doneNoPanicByDefault(t *testing.T) {
	f := new(FnRun)
	f.Wait()
	f.Done()
}

func TestFnRun_waitDone(t *testing.T) {
	f := new(FnRun)
	f.WaitChan = make(chan struct{})

	go func() {
		f.Done()
	}()
	f.Wait()
}
