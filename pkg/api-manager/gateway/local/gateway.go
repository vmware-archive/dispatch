///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package local

import (
	"context"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"

	apimanager "github.com/vmware/dispatch/pkg/api-manager"
	"github.com/vmware/dispatch/pkg/api-manager/gateway"
	"github.com/vmware/dispatch/pkg/client"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/errors"
	"github.com/vmware/dispatch/pkg/http"
	"github.com/vmware/dispatch/pkg/trace"
)

// Gateway implements API Manager Gateway using local HTTP server.
type Gateway struct {
	Server *http.Server

	store    entitystore.EntityStore
	fnClient client.FunctionsClient

	sync.RWMutex
	pathLookup   map[string][]*gateway.API
	hostLookup   map[string][]*gateway.API
	methodLookup map[string][]*gateway.API
	apis         map[string]*gateway.API
}

// NewGateway creates a new local API gateway
func NewGateway(store entitystore.EntityStore, functionsClient client.FunctionsClient) (*Gateway, error) {
	c := &Gateway{
		fnClient:     functionsClient,
		Server:       http.NewServer(nil),
		store:        store,
		apis:         make(map[string]*gateway.API),
		pathLookup:   make(map[string][]*gateway.API),
		hostLookup:   make(map[string][]*gateway.API),
		methodLookup: make(map[string][]*gateway.API),
	}
	c.rebuildCache()
	return c, nil
}

// GetAPI gets the API
func (g *Gateway) GetAPI(ctx context.Context, name string) (*gateway.API, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	g.RLock()
	defer g.RUnlock()
	if api, ok := g.apis[name]; ok {
		apiCopy := *api
		return &apiCopy, nil
	}
	return nil, &errors.ObjectNotFoundError{Err: fmt.Errorf("api %s not found", name)}
}

// AddAPI adds an API to internal registry
func (g *Gateway) AddAPI(ctx context.Context, entity *gateway.API) (*gateway.API, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	g.Lock()
	defer g.Unlock()
	apiCopy := *entity
	g.apis[entity.Name] = &apiCopy
	g.rebuildCache()

	return entity, nil
}

// UpdateAPI updates the API
func (g *Gateway) UpdateAPI(ctx context.Context, name string, entity *gateway.API) (*gateway.API, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	g.Lock()
	defer g.Unlock()
	_, ok := g.apis[name]
	if !ok {
		return nil, &errors.ObjectNotFoundError{Err: fmt.Errorf("api %s not found", name)}
	}
	apiCopy := *entity
	g.apis[name] = &apiCopy
	g.rebuildCache()

	return entity, nil
}

// DeleteAPI deletes the API
func (g *Gateway) DeleteAPI(ctx context.Context, api *gateway.API) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	g.Lock()
	defer g.Unlock()
	delete(g.apis, api.Name)
	g.rebuildCache()
	return nil
}

// rebuildCache iterates over all configured APIs and populates lookup caches. Could optimized
// to only add changes.
func (g *Gateway) rebuildCache() {
	// Store should only be nil for tests
	if g.store != nil {
		var apis []*apimanager.API
		err := g.store.ListGlobal(context.TODO(), entitystore.Options{}, &apis)
		if err != nil {
			log.Errorf("error syncing APIs: %v", err)
		}
		g.apis = make(map[string]*gateway.API)
		for _, api := range apis {
			g.apis[api.Name] = &api.API
		}
	}
	g.hostLookup = make(map[string][]*gateway.API)
	g.methodLookup = make(map[string][]*gateway.API)
	g.pathLookup = make(map[string][]*gateway.API)
	for i, api := range g.apis {
		if !api.Enabled {
			continue
		}
		for _, method := range api.Methods {
			g.methodLookup[method] = append(g.methodLookup[method], g.apis[i])
		}
		for _, path := range api.URIs {
			g.pathLookup[path] = append(g.pathLookup[path], g.apis[i])
		}
		for _, host := range api.Hosts {
			g.hostLookup[host] = append(g.hostLookup[host], g.apis[i])
		}
	}
	log.Debugf("Cache rebuilt: methods: %#v, paths: %#v, hosts: %#v", g.methodLookup, g.pathLookup, g.hostLookup)
}
