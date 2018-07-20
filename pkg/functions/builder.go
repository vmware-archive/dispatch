///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functions

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/go-openapi/swag"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/images"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

// DockerImageBuilder builds function images
type DockerImageBuilder struct {
	ImageRegistry string
	RegistryAuth  string
	PushImages    bool
	PullImages    bool

	docker docker.CommonAPIClient
}

// NewDockerImageBuilder is the constructor for the DockerImageBuilder
func NewDockerImageBuilder(imageRegistry, registryAuth string, docker docker.CommonAPIClient) *DockerImageBuilder {
	return &DockerImageBuilder{
		ImageRegistry: imageRegistry,
		RegistryAuth:  registryAuth,
		docker:        docker,
		PushImages:    true,
		PullImages:    true,
	}
}

const (
	functionTemplateLabel      = "io.dispatchframework.functionTemplate"
	functionTemplateDirDefault = "/function-template"
)

func (ib *DockerImageBuilder) copyFunctionTemplate(tmpDir string, image string) error {
	log.Debugf("Creating a container for image: %s", image)
	resp, err := ib.docker.ContainerCreate(context.Background(), &container.Config{
		Image: image,
	}, nil, nil, "")
	if err != nil {
		return errors.Wrapf(err, "failed to create container for image '%s'", image)
	}
	defer ib.docker.ContainerRemove(context.Background(), resp.ID, types.ContainerRemoveOptions{})

	ic, err := ib.docker.ContainerInspect(context.Background(), resp.ID)
	if err != nil {
		return errors.Wrapf(err, "failed to inspect image container, id='%s', image='%s'", resp.ID, image)
	}

	functionTemplateDir := ic.Config.Labels[functionTemplateLabel]
	if functionTemplateDir == "" {
		functionTemplateDir = functionTemplateDirDefault
	}

	readCloser, _, err := ib.docker.CopyFromContainer(context.Background(), resp.ID, functionTemplateDir)
	defer readCloser.Close()

	return utils.Untar(tmpDir, strings.TrimPrefix(functionTemplateDir, "/")+"/", readCloser)
}

// BuildImage packages a function into a docker image.  It also adds any FaaS specfic image layers
func (ib *DockerImageBuilder) BuildImage(ctx context.Context, f *Function, code []byte) (string, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	name := imageName(ib.ImageRegistry, f.FaasID)
	log.Debugf("Building image '%s'", name)

	tmpDir, err := ioutil.TempDir("", "func-build")
	if err != nil {
		return "", errors.Wrap(err, "failed to create a temp dir")
	}
	defer os.RemoveAll(tmpDir)
	log.Debugf("Created tmpDir: %s", tmpDir)

	if ib.PullImages {
		opts := types.ImagePullOptions{RegistryAuth: ib.RegistryAuth}
		if err := images.DockerError(ib.docker.ImagePull(ctx, f.ImageURL, opts)); err != nil {
			return "", errors.Wrapf(err, "failed to pull image '%s'", f.ImageURL)
		}
	}

	if err := writeSourceDir(tmpDir, code); err != nil {
		return "", errors.Wrap(err, "failed to write dockerfile")
	}

	if err := ib.copyFunctionTemplate(tmpDir, f.ImageURL); err != nil {
		return "", err
	}

	buildArgs := map[string]*string{
		"IMAGE":   swag.String(f.ImageURL),
		"HANDLER": swag.String(f.Handler),
	}
	err = images.BuildAndPushFromDir(ctx, ib.docker, tmpDir, name, ib.RegistryAuth, ib.PushImages, buildArgs)
	return name, err
}

func writeSourceDir(destDir string, code []byte) error {
	r, err := tarStream(code)
	if err != nil {
		return errors.Wrapf(err, "failed to get the tar stream, writing source dir to '%s'", destDir)
	}
	err = utils.Untar(destDir, "/", r)
	return errors.Wrapf(err, "failed to untar, writing source dir to '%s'", destDir)
}

func tarStream(bs []byte) (io.Reader, error) {
	gr, err := gzip.NewReader(bytes.NewReader(bs))
	if err != nil {
		return nil, errors.Wrap(err, "failed to read gzip stream")
	}
	return gr, nil
}

func imageName(registry, fnID string) string {
	return registry + "/func-" + fnID + ":latest"
}

// RemoveImage removes a function image from the docker host
func (ib *DockerImageBuilder) RemoveImage(ctx context.Context, f *Function) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	if err := images.Remove(ctx, ib.docker, f.FunctionImageURL); err != nil {
		return errors.Wrapf(err, "failed to delete docker image for function %s", f.Name)
	}
	return nil
}
