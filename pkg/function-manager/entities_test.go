///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package functionmanager

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
