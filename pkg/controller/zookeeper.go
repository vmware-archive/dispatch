package controller

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/samuel/go-zookeeper/zk"
	log "github.com/sirupsen/logrus"
)

// ZKLock is a lock based on zookeeper
// Implements the sync.Locker interface
type ZKLock struct {
	client       *zk.Conn
	entity       string
	creationPath string
	LockPath     string
	Locked       bool
}

// NewZKLock is just a constructor method for a ZKLock
func NewZKLock(entity string, client *zk.Conn) *ZKLock {
	CreateZnode(client, fmt.Sprintf("/entities/%v", entity))
	return &ZKLock{
		entity:       entity,
		creationPath: fmt.Sprintf("/entities/%v/lock-", entity),
		client:       client,
		Locked:       false,
	}
}

// Lock locks the lock.
// This uses the locking scheme given in the official zookeeper recipes & solutions
// Source: http://zookeeper.apache.org/doc/r3.1.2/recipes.html
// The only change here is that we don't want to spin waiting for the lock
func (lck *ZKLock) Lock() {
	if lck.Locked {
		log.Fatalf("Lock is already held!")
	}
	acl := zk.WorldACL(zk.PermAll)
	path, err := lck.client.CreateProtectedEphemeralSequential(lck.creationPath, []byte("lock"), acl)
	lck.LockPath = path
	if err != nil {
		log.Warnf("Unable to create lock node for %v: %v", lck.entity, err)
	}
	children, _, err := lck.client.Children(fmt.Sprintf("/entities/%v", lck.entity))
	if err != nil {
		log.Warnf("Unable to get children of entity %v: %v", lck.entity, err)
	}
	sfx := strings.Split(path, "lock-")[1]
	for _, child := range children {
		ch := strings.Split(child, "lock-")[1]
		if ch < sfx {
			log.Debugf("%v: Couldn't Get Lock!\n", sfx)
			lck.Unlock()
		}
	}
	lck.Locked = true
}

// Unlock releases the lock by deleting the znode that represents the lock
func (lck *ZKLock) Unlock() {
	if !lck.Locked {
		log.Fatalf("Attempted unlock of unlocked lock!")
	}
	err := lck.client.Delete(lck.LockPath, -1)
	if err != nil {
		log.Fatalf("Unable to delete lock for %v: %v", lck.entity, err)
	}
	lck.LockPath = ""
	lck.Locked = false
	log.Infof("Released lock for %v", lck.creationPath)
}

// CreateZnode just creates Znode on path. If the node exists, it will return an error
func CreateZnode(zClient *zk.Conn, path string) error {
	acl := zk.WorldACL(zk.PermAll)

	if exists, _, err := zClient.Exists(path); exists {
		if err != nil {
			return err
		}
		return errors.Errorf("Node already exists")
	}
	_, err := zClient.Create(path, []byte("lock"), int32(0), acl)
	if err != nil {
		log.Warnf("Unable to create znode %v: %v", path, err)
		return err
	}
	log.Debugf("Successfully Created Znode: %v", path)
	return nil
}

// ZKAcquireLock checks whether a given client should be allowed access to an object
func ZKAcquireLock(zClient *zk.Conn, name string) (string, bool) {
	acl := zk.WorldACL(zk.PermAll)
	creationPath := fmt.Sprintf("/entities/%v/lock-", name)
	me, err := zClient.CreateProtectedEphemeralSequential(creationPath, []byte("lock"), acl)
	if err != nil {
		log.Warnf("Unable to create lock node for %v: %v", name, err)
	}
	children, _, err := zClient.Children(fmt.Sprintf("/entities/%v", name))
	if err != nil {
		log.Warnf("Unable to get children of %v: %v", name, err)
	}
	sfx := strings.Split(me, "lock-")[1]
	for _, child := range children {
		ch := strings.Split(child, "lock-")[1]
		if ch < sfx {
			log.Debugf("%v: Couldn't Get Lock!\n", sfx)
			return me, false
		}
	}
	return me, true
}

// ZKConnect opens a connection to zookeeper by creating a new client
func ZKConnect() *zk.Conn {
	client, _, err := zk.Connect([]string{"127.0.0.1"}, time.Second)
	if err != nil {
		log.Warnf("Unable to connect to zk: %v", err)
	}
	return client
}
