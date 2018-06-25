///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	log "github.com/sirupsen/logrus"

	"github.com/docker/docker/client"
)

func dockerClient(config *serverConfig) client.CommonAPIClient {
	dc, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("error creating docker client: %+v", err)
	}
	return dc
}
