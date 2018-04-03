///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functions

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/go-openapi/swag"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/images"
	"github.com/vmware/dispatch/pkg/trace"
)

// DockerImageBuilder builds function images
type DockerImageBuilder struct {
	imageRegistry string
	registryAuth  string

	docker docker.CommonAPIClient
}

// NewDockerImageBuilder is the constructor for the DockerImageBuilder
func NewDockerImageBuilder(imageRegistry, registryAuth string, docker *docker.Client) *DockerImageBuilder {
	return &DockerImageBuilder{
		imageRegistry: imageRegistry,
		registryAuth:  registryAuth,
		docker:        docker,
	}
}

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

	functionTemplateDir := ic.Config.Labels["io.dispatchframework.functionTemplate"]

	readCloser, _, err := ib.docker.CopyFromContainer(context.Background(), resp.ID, functionTemplateDir)
	defer readCloser.Close()

	return images.Untar(tmpDir, strings.TrimPrefix(functionTemplateDir, "/")+"/", readCloser)
}

// BuildImage packages a function into a docker image.  It also adds any FaaS specfic image layers
func (ib *DockerImageBuilder) BuildImage(faas, fnID string, exec *Exec) (string, error) {
	defer trace.Tracef("function: '%s', base: '%s'", fnID, exec.Image)()
	name := imageName(ib.imageRegistry, faas, fnID)
	log.Debugf("Building image '%s'", name)

	tmpDir, err := ioutil.TempDir("", "func-build")
	if err != nil {
		return "", errors.Wrap(err, "failed to create a temp dir")
	}
	defer os.RemoveAll(tmpDir)

	log.Debugf("Created tmpDir: %s", tmpDir)

	if err := images.DockerError(ib.docker.ImagePull(context.Background(), exec.Image, types.ImagePullOptions{})); err != nil {
		return "", errors.Wrapf(err, "failed to pull image '%s'", exec.Image)
	}

	if err := ib.copyFunctionTemplate(tmpDir, exec.Image); err != nil {
		return "", err
	}

	if err := writeFunctionFile(tmpDir, exec); err != nil {
		return "", errors.Wrap(err, "failed to write dockerfile")
	}

	buildArgs := map[string]*string{
		"IMAGE":        swag.String(exec.Image),
		"FUNCTION_SRC": swag.String(functionFile),
	}
	err = images.BuildAndPushFromDir(ib.docker, tmpDir, name, ib.registryAuth, buildArgs)
	return name, err
}

const functionFile = "function.txt"

func writeFunctionFile(dir string, exec *Exec) error {
	functionPath := filepath.Join(dir, functionFile)
	err := ioutil.WriteFile(functionPath, []byte(exec.Code), 0644)
	return errors.Wrapf(err, "failed to write file, function '%s'", exec.Name)
}

func imageName(registry, faas, fnID string) string {
	return registry + "/func-" + faas + "-" + fnID + ":latest"
}
