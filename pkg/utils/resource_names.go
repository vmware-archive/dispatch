///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package utils

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/dustinkirkland/golang-petname"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomResourceName produces a random name consisting of two words and an integer suffix.
// suffix will always be a 6 digits integer.
// Example: wiggly-yellowtail-123456
func RandomResourceName() string {
	suffix := rand.Intn(900000) + 100000
	return fmt.Sprintf("%s-%d", petname.Generate(2, "-"), suffix)
}

// RandString generates a random string of length n
func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
