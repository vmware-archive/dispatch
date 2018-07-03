///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	log "github.com/sirupsen/logrus"

	entitystore "github.com/vmware/dispatch/pkg/entity-store"
)

func entityStore(config *serverConfig) entitystore.EntityStore {
	store, err := entitystore.NewFromBackend(
		entitystore.BackendConfig{
			Backend:  config.DatabaseBackend,
			Address:  config.DatabaseAddress,
			Bucket:   config.DatabaseBucket,
			Username: config.DatabaseUsername,
			Password: config.DatabasePassword,
		})
	if err != nil {
		log.Fatalln(err)
	}
	return store
}
