///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package imagemanager

import (
	"io"

	"github.com/pkg/errors"
)

type System interface {
	GetPackageManager() string
	WriteDockerfile(io.Writer, *BaseImage, *Image) error
}

var SystemMap = make(map[Os]System)

func WriteSystemDockerfile(dir string, dockerfile io.Writer, baseImage *BaseImage, image *Image) (string, error) {
	// Hard-coded photon for now
	system, ok := SystemMap[OsPhoton]
	if !ok {
		return "", errors.Errorf("No system for OS %s", OsPhoton)
	}
	err := system.WriteDockerfile(dockerfile, baseImage, image)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to write Dockerfile content for %s", OsPhoton)
	}
	return system.GetPackageManager(), nil
}
