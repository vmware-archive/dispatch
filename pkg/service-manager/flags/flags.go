///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package flags

// NO TEST

// ServiceManagerFlags are configuration flags for the service manager
var ServiceManagerFlags = struct {
	Config       string `long:"config" description:"Path to Config file" default:"./config.dev.json"`
	DbFile       string `long:"db-file" description:"Backend DB URL/Path" default:"./db.bolt"`
	DbBackend    string `long:"db-backend" description:"Backend DB Name" default:"boltdb"`
	DbUser       string `long:"db-username" description:"Backend DB Username" default:"dispatch"`
	DbPassword   string `long:"db-password" description:"Backend DB Password" default:"dispatch"`
	DbDatabase   string `long:"db-database" description:"Backend DB Name" default:"dispatch"`
	OrgID        string `long:"organization" description:"(temporary) Static organization id" default:"dispatch"`
	ResyncPeriod int    `long:"resync-period" description:"The time period (in seconds) to sync with image repository" default:"10"`
	K8sConfig    string `long:"kubeconfig" description:"Path to kubernetes config file" default:""`
	SecretStore  string `long:"secret-store" description:"Secret store endpoint" default:"localhost:8003"`
	Tracer       string `long:"tracer" description:"Open Tracing Tracer endpoint" default:""`
}{}
