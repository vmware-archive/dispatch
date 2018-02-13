///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
// Package gen contains generated code
// this code needs to be considered by go, so it can't be in a file that starts with _ or .

package eventmanager

//go:generate mkdir -p gen
//go:generate go-bindata -o ./gen/bindata.go -pkg gen -prefix '../../swagger' ../../swagger
//go:generate swagger generate server -A EventManager -t ./gen -f ../../swagger/event-manager.yaml --exclude-main
