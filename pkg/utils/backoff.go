///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package utils

import (
	cRand "crypto/rand"
	"math/big"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
)

const minSleep = int64(time.Second) / 4

func init() {
	seed, _ := cRand.Int(cRand.Reader, big.NewInt(1<<63-1))
	rand.Seed(seed.Int64())
}

// Backoff runs a function with a random backoff timeout
func Backoff(timeout time.Duration, f func() error) error {
	maxTimer := time.NewTimer(timeout)
	var err error

	attempt := 0
	sleepTime := time.Duration(minSleep + rand.Int63n(minSleep))

	for ; ; sleepTime = sleepTime + time.Duration(rand.Int63n(int64(sleepTime+1))) {
		attempt++

		log.Debugf("backoff: attempt # %v, sleepTime: %v", attempt, sleepTime)

		err = f()
		if err == nil {
			return nil
		}

		log.Debugf("backoff: error on attempt # %v: %v", attempt, err)

		sleepTimer := time.NewTimer(sleepTime)

		select {
		case <-sleepTimer.C:
			log.Debugf("backoff: retrying")
			continue
		case <-maxTimer.C:
			log.Debugf("backoff: retries timed out")
			return err
		}
	}
}
