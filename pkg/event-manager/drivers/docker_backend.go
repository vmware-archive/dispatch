///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package drivers

// NO TEST

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/vmware/dispatch/pkg/entity-store"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/event-manager/drivers/entities"
	"github.com/vmware/dispatch/pkg/utils"
)

const (
	labelEventDriverID   = "dispatch-eventdriver-id"
	defaultDeployTimeout = 10 // seconds
	dispatchEventHostEnv = "dispatch-event-host"
	dispatchEventPortEnv = "dispatch-event-port"
	dispatchEventOrgEnv  = "dispatch-event-org"
)

type dockerBackend struct {
	dockerClient  docker.CommonAPIClient
	secretsClient client.SecretsClient
	DeployTimeout int
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
		DeployTimeout: defaultDeployTimeout,
	}, nil
}

func getHostIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			// skip loopback ip
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ip = ip.To4(); ip != nil {
				return ip.String(), nil
			}
		}
	}
	return "", errors.New("cannot get local host ip")
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

func (d *dockerBackend) getDriverSecrets(ctx context.Context, driver *entities.Driver) (map[string]string, error) {
	secrets := make(map[string]string)
	for _, secretName := range driver.Secrets {
		secret, err := d.secretsClient.GetSecret(ctx, driver.OrganizationID, secretName)
		if err != nil {
			return nil, err
		}
		for key, val := range secret.Secrets {
			secrets[key] = val
		}
	}

	// Add host ip, port and org as secrets
	hostIP, err := getHostIP()
	if err != nil {
		log.Infoln("cannot get Dispatch solo host ip")
	}
	// Only set host and port when not specified in given secrets
	if _, ok := secrets[dispatchEventHostEnv]; !ok {
		secrets[dispatchEventHostEnv] = hostIP
	}
	if _, ok := secrets[dispatchEventPortEnv]; !ok {
		secrets[dispatchEventPortEnv] = "8080"
	}
	if _, ok := secrets[dispatchEventOrgEnv]; !ok {
		secrets[dispatchEventOrgEnv] = driver.OrganizationID
	}
	return secrets, nil
}

// Deploy deploys driver
func (d *dockerBackend) Deploy(ctx context.Context, driver *entities.Driver) error {
	log.Infof("Docker backend: deploying driver %v", driver.Name)

	// get driver secrets
	secrets, err := d.getDriverSecrets(ctx, driver)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("err getting secrets of driver %s", driver.Image))
	}

	rc, err := d.dockerClient.ImagePull(ctx, driver.Image, types.ImagePullOptions{})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("err pulling image %s", driver.Image))
	}
	defer rc.Close()
	io.Copy(os.Stdout, rc)

	return utils.Backoff(time.Duration(d.DeployTimeout)*time.Second, func() error {
		created, err := d.dockerClient.ContainerCreate(ctx, &container.Config{
			Image: driver.Image,
			Env:   bindEnv(secrets),
			Cmd:   buildArgs(driver.Config),
			Labels: map[string]string{
				labelEventDriverID: driver.ID,
			},
		}, nil, nil, "")
		if err != nil {
			return errors.Wrap(err, "error creating container")
		}

		if err := d.dockerClient.ContainerStart(ctx, created.ID, types.ContainerStartOptions{}); err != nil {
			return errors.Wrap(err, "error starting container")
		}
		return nil
	})
}

// Expose exposes driver
func (d *dockerBackend) Expose(ctx context.Context, driver *entities.Driver) error {
	log.Infof("Docker backend: exposing driver %v", driver.Name)
	// TODO: expose
	return nil
}

// Update updates driver
func (d *dockerBackend) Update(ctx context.Context, driver *entities.Driver) error {
	// Update the driver in UPDATING status
	// docker doesn't support change env/args of running container,
	// will first delete then create
	if driver.Status == entitystore.StatusUPDATING {
		if err := d.Delete(ctx, driver); err != nil {
			return errors.Wrap(err, "error deleting container before update")
		}
		if err := d.Deploy(ctx, driver); err != nil {
			return errors.Wrap(err, "error deploying container during update")
		}
		log.Infof("driver %s updated", driver.Name)
	}
	return nil
}

// Delete deletes driver
func (d *dockerBackend) Delete(ctx context.Context, driver *entities.Driver) error {
	log.Infof("Docker backend: deleting driver %v", driver.Name)

	opts := filters.NewArgs()
	opts.Add("label", labelEventDriverID+"="+driver.ID)
	containers, err := d.dockerClient.ContainerList(ctx, types.ContainerListOptions{
		Filters: opts,
		All:     true,
	})
	if len(containers) != 1 || err != nil {
		return errors.Wrap(err, "error getting container while deleting event driver")
	}

	err = d.dockerClient.ContainerRemove(ctx, containers[0].ID, types.ContainerRemoveOptions{
		Force: true,
	})

	return err
}

func buildArgs(input map[string]string) []string {
	var args []string
	for key, val := range input {
		if val == "" {
			args = append(args, fmt.Sprintf("--%s", key))
		} else {
			args = append(args, fmt.Sprintf("--%s=%s", key, val))
		}

	}
	return args
}
