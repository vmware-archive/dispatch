///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package utils

import (
	"fmt"
	"math/rand"

	"github.com/dustinkirkland/golang-petname"
)

// RandomResourceName produces a random name consisting of two words and an integer suffix.
// suffix will always be a 6 digits integer.
// Example: wiggly-yellowtail-123456
func RandomResourceName() string {
	suffix := rand.Intn(900000) + 100000
	return fmt.Sprintf("%s-%d", petname.Generate(2, "-"), suffix)
}
