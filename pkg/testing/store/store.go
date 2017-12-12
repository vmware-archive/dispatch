///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package store

// NO TESTS

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/stretchr/testify/assert"
)

// MakeKVStore creates a temporary boltdb backed kv store
func MakeKVStore(t *testing.T) (path string, kv store.Store) {
	boltdb.Register()
	file, err := ioutil.TempFile(os.TempDir(), "test")
	assert.NoError(t, err, "Cannot create temp file")
	kv, err = libkv.NewStore(
		store.BOLTDB,
		[]string{file.Name()},
		&store.Config{
			Bucket:            "test",
			ConnectionTimeout: 1 * time.Second,
			PersistConnection: true,
		},
	)
	assert.NoError(t, err, "Cannot create store")
	return file.Name(), kv
}

// CleanKVStore deletes a boltdb backed kv store
func CleanKVStore(t *testing.T, path string, kv store.Store) {
	kv.Close()
	os.Remove(path)
}
