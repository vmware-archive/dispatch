///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package images

// NO TESTS

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

// DockerError scans for errors in docker commands
func DockerError(r io.ReadCloser, err error) error {
	if err != nil {
		return err
	}
	defer r.Close()
	s := bufio.NewScanner(r)
	for s.Scan() {
		log.Debug(s.Text())
		result := struct {
			Message *string `json:"message,omitempty"`
			Error   *string `json:"error,omitempty"`
		}{}
		if err := json.Unmarshal(s.Bytes(), &result); err != nil {
			return errors.Wrapf(err, "failed to parse docker response: %s", s.Text())
		}
		if result.Error != nil {
			return errors.New(*result.Error)
		}
	}
	return nil
}

// BuildAndPushFromDir will tar up a docker image, build it, and push it
func BuildAndPushFromDir(ctx context.Context, client docker.ImageAPIClient, dir, name, registryAuth string, buildArgs map[string]*string) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	if err := Build(ctx, client, dir, name, buildArgs); err != nil {
		return err
	}
	return Push(ctx, client, name, registryAuth)
}

// Build a docker image
func Build(ctx context.Context, client docker.ImageAPIClient, dir, name string, buildArgs map[string]*string) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	files, _ := ioutil.ReadDir(dir)
	for _, f := range files {
		log.Infof("Packing %s", f.Name())
		b, _ := ioutil.ReadFile(filepath.Join(dir, f.Name()))
		log.Debug(string(b))
	}

	tarBall := new(bytes.Buffer)
	if err := utils.Tar(dir, tarBall); err != nil {
		return errors.Wrap(err, "failed to create a tarball archive")
	}

	log.Infof("Building image %s from tarball", name)
	r, err := client.ImageBuild(context.Background(), tarBall, types.ImageBuildOptions{
		BuildArgs: buildArgs,
		Tags:      []string{name},
	})
	return errors.Wrapf(DockerError(r.Body, err), "failed to build image '%s'", name)
}

// Push a docker image
func Push(ctx context.Context, client docker.ImageAPIClient, name, registryAuth string) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	opts := types.ImagePushOptions{}
	if registryAuth != "" {
		opts.RegistryAuth = registryAuth
	}

	if err := DockerError(client.ImagePush(ctx, name, opts)); err != nil {
		return errors.Wrapf(err, "failed to push the image %s", name)
	}
	return nil
}
