///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package drivers

// NO TEST

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/event-manager/drivers/entities"
)

type dockerBackend struct {
	dockerClient  docker.CommonAPIClient
	secretsClient client.SecretsClient
}

// NewDockerBackend creates a new docker backend driver
func NewDockerBackend(secretsClient client.SecretsClient) (Backend, error) {
	dockerClient, err := docker.NewEnvClient()
	if err != nil {
		return nil, err
	}

	return &dockerBackend{
		dockerClient:  dockerClient,
		secretsClient: secretsClient,
	}, nil
}

func bindEnv(secrets map[string]string) []string {
	var vars []string
	for key, val := range secrets {
		// ENV=value
		envVar := strings.Replace(strings.ToUpper(key), "-", "_", -1) + "=" + val
		vars = append(vars, envVar)
	}
	return vars
}

// Deploy deploys driver
func (d *dockerBackend) Deploy(ctx context.Context, driver *entities.Driver) error {
	log.Infof("Docker backend: deploying driver %v", driver.Name)

	// get driver secrets
	secrets := make(map[string]string)
	for _, secretName := range driver.Secrets {
		secret, err := d.secretsClient.GetSecret(ctx, driver.OrganizationID, secretName)
		if err != nil {
			return err
		}
		for key, val := range secret.Secrets {
			secrets[key] = val
		}
	}

	// TODO: very first image pull won't finish before starting the container
	_, err := d.dockerClient.ImagePull(ctx, driver.Image, types.ImagePullOptions{})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("err pulling image %s", driver.Image))
	}

	created, err := d.dockerClient.ContainerCreate(ctx, &container.Config{
		Image: driver.Image,
		Env:   bindEnv(secrets),
		Cmd:   buildArgs(driver.Config),
	}, nil, nil, "")
	if err != nil {
		return errors.Wrap(err, "error creating container")
	}

	if err := d.dockerClient.ContainerStart(ctx, created.ID, types.ContainerStartOptions{}); err != nil {
		return errors.Wrap(err, "error starting container")
	}
	return nil
}

// Expose exposes driver
func (d *dockerBackend) Expose(ctx context.Context, driver *entities.Driver) error {
	log.Infof("Docker backend: exposing driver %v", driver.Name)
	return nil
}

// Update updates driver
func (d *dockerBackend) Update(ctx context.Context, driver *entities.Driver) error {
	log.Infof("Docker backend: updating driver %v", driver.Name)
	return nil
}

// Delete deletes driver
func (d *dockerBackend) Delete(ctx context.Context, driver *entities.Driver) error {
	log.Infof("Docker backend: deleting driver %v", driver.Name)
	return nil
}
