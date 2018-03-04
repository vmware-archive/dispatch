///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package imagemanager

import (
	"io"

	"github.com/pkg/errors"
)

// Runtime defines the Runtime interface
type Runtime interface {
	GetPackageManager() string
	PrepareManifest(string, *Image) error
	WriteDockerfile(io.Writer, *Image) error
}

// RuntimeMap tracks the mapping from language to runtime
var RuntimeMap = make(map[Language]Runtime)

// WriteRuntimeDockerfile creates the dockerfile for the given language
func WriteRuntimeDockerfile(dir string, dockerfile io.Writer, image *Image) (string, error) {
	runtime, ok := RuntimeMap[image.Language]
	if !ok {
		return "", errors.Errorf("No runtime for language %s", image.Language)
	}

	err := runtime.PrepareManifest(dir, image)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to write manifest for %s [%s]", image.Language, runtime.GetPackageManager())
	}
	err = runtime.WriteDockerfile(dockerfile, image)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to write Dockerfile content for %s", image.Language)
	}
	return runtime.GetPackageManager(), nil
}
