///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package openfaas

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/rand"
)

func Test_timeStampStr(t *testing.T) {
	ts := time.Now().UTC()
	expected := fmt.Sprintf("%04d%02d%02d-%02d%02d%02d", ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second())
	assert.Equal(t, expected, utcTimeStampStr(ts))
}

func Test_imageName(t *testing.T) {
	prefix := rand.String(9)
	name := rand.String(6)
	ts := rand.String(11)
	assert.Equal(t, prefix+"/func-"+name+":"+ts, imageName(prefix, name, ts))
}
