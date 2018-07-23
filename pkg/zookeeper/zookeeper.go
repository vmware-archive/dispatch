///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package zookeeper

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/samuel/go-zookeeper/zk"
	log "github.com/sirupsen/logrus"
)

// NO TESTS

const (
	// NodeCreated represents the creation of a znode
	NodeCreated = zk.EventNodeCreated
)

// Driver is an interface that abstracts zookeeper calls
// This is a simple interface that allows for the creation/deletion of nodes
// Also we can watch for the creation of a node
type Driver interface {
	Connect(url string) error
	GetConnection() *zk.Conn
	CreateNode(path string, data []byte) error
	DeleteNode(path string) error
	WatchForNode(path string) (<-chan zk.Event, error)
	LockEntity(name string) (string, bool)
	ReleaseEntity(lock string)
	GetData(path string) ([]byte, error)
	Close()
}

// Zdriver is a simple implementation of the Driver interface
// This allows us to create and delete nodes, and create watches
type Zdriver struct {
	client *zk.Conn
	acl    []zk.ACL
}

// LockEntity Locks a given entity, preventing other drivers from modifying it
// Returns the path to the lock and whether we own it or not
// This uses the locking scheme given in the official zookeeper recipes & solutions
// Source: http://zookeeper.apache.org/doc/r3.1.2/recipes.html
// The only change here is that we don't want to spin waiting for the lock
func (d *Zdriver) LockEntity(entity string) (string, bool) {
	createOn := fmt.Sprintf("/entities/%v/lock-", entity)
	path, err := d.client.CreateProtectedEphemeralSequential(createOn, []byte("lock"), d.acl)
	if err != nil {
		log.Warnf("Unable to create lock node for %v: %v", entity, err)
	}
	children, _, err := d.client.Children(fmt.Sprintf("/entities/%v", entity))
	if err != nil {
		log.Warnf("Unable to get children of entity %v: %v", entity, err)
	}
	sfx := strings.Split(path, "lock-")[1]
	for _, child := range children {
		ch := strings.Split(child, "lock-")[1]
		if ch < sfx {
			d.DeleteNode(path)
			return "", false
		}
	}
	return path, true
}

// ReleaseEntity releases the entity by deleting the znode that represents the lock
func (d *Zdriver) ReleaseEntity(path string) {
	err := d.client.Delete(path, -1)
	if err != nil {
		log.Fatalf("Unable to delete lock %v: %v", path, err)
	}
	log.Infof("Released lock %v", path)
}

// NewDriver is just a constructor for the Zdriver class
func NewDriver(url string) (*Zdriver, error) {
	var driver Zdriver
	if err := driver.Connect(url); err != nil {
		return nil, errors.Errorf("Unable to connect to zookeeper client %v: %v", url, err)
	}
	driver.acl = zk.WorldACL(zk.PermAll)
	return &driver, nil
}

// Connect connects a driver to a zookeeper instance
func (d *Zdriver) Connect(url string) error {
	client, _, err := zk.Connect([]string{url}, time.Second)
	if err != nil {
		return err
	}
	d.client = client
	return nil
}

// GetConnection returns a connection to the same instance of zookeeper the driver is connected to
func (d *Zdriver) GetConnection() *zk.Conn {
	return d.client
}

// CreateNode create a znode along a path
func (d *Zdriver) CreateNode(path string, data []byte) error {
	if exists, _, err := d.client.Exists(path); exists {
		log.Debugf("Znode %v already exists", path)
		return nil
	} else if err != nil {
		return errors.Errorf("Unable to access znode %v: %v", path, err)
	}
	_, err := d.client.Create(path, data, int32(0), d.acl)
	if err != nil {
		return errors.Errorf("Unable to create znode %v: %v", path, err)
	}
	log.Debugf("Successfully Created Znode: %v", path)
	return nil
}

// DeleteNode deletes the znode at the given path
func (d *Zdriver) DeleteNode(path string) error {
	err := d.client.Delete(path, -1)
	if err != nil {
		return errors.Errorf("Unable to delete znode %v: %v", path, err)
	}
	log.Debugf("Successfully Deleted Znode: %V", path)
	return nil
}

// Close closes the connection to zookeeper
func (d *Zdriver) Close() {
	d.client.Close()
}

// WatchForNode watches for the creation of a specific znode
func (d *Zdriver) WatchForNode(path string) (<-chan zk.Event, error) {
	_, _, watch, err := d.client.ExistsW(path)
	if err != nil {
		return nil, errors.Errorf("Unable to watch for %v: %v", path, err)
	}
	return watch, nil
}

// GetData just grabs whatever data is at a node
func (d *Zdriver) GetData(path string) ([]byte, error) {
	data, _, err := d.client.Get(path)
	if err != nil {
		return nil, errors.Errorf("Can't get data from %v: %v", path, err)
	}
	return data, nil
}
