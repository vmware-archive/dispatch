///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

const (
	jsonContentType       = "application/json"
	functionAPIPort       = "8080/tcp"
	healthcheckEndpoint   = "/healthz"
	labelFunctionID       = "dispatch-function-id"
	labelFunctionRevision = "dispatch-function-revision"
	defaultBackoff        = time.Second * 60
	defaultHost           = "127.0.0.1"
)

// dockerContainer represents a basic information about function container
type dockerContainer struct {
	ID               string
	ImageID          string
	ImageName        string
	FunctionID       string
	FunctionRevision string
	Port             string
	Host             string
}

// Client specifies the Docker client API interface required by docker driver
type Client interface {
	dockerclient.ContainerAPIClient
	dockerclient.ImageAPIClient
}

// Driver implements a FaaSDriver using Docker daemon. It's a simple driver without scaling or fault tolerance
// and is not recommended for production usage. It's goal is to provide a simple driver for demos, PoCs, and development
// use cases.
type Driver struct {
	// ExternalHost is a ip/hostname that function containers will be exposed with, and that is reachable to Dispatch.
	ExternalHost string
	// RetryTimeout specifies the maximum amount of time we should spend retrying calls to docker.
	RetryTimeout time.Duration

	docker         Client
	containerCache *sync.Map
}

// New creates a new Docker driver
func New(dockerClient Client) *Driver {

	d := &Driver{
		docker:         dockerClient,
		ExternalHost:   defaultHost,
		RetryTimeout:   defaultBackoff,
		containerCache: new(sync.Map),
	}

	return d
}

// Create creates new Docker container for a particular function. Currently, there is 1:1 mapping for a function.
func (d *Driver) Create(ctx context.Context, f *functions.Function) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	containerName := getID(f.Name, f.FaasID)

	resp, err := d.docker.ContainerCreate(ctx, &container.Config{
		Image:        f.FunctionImageURL,
		ExposedPorts: nat.PortSet{functionAPIPort: {}},
		Labels: map[string]string{
			labelFunctionID:       f.ID,
			labelFunctionRevision: f.FaasID,
		},
	}, &container.HostConfig{
		NetworkMode:  "bridge",
		PortBindings: nat.PortMap{functionAPIPort: []nat.PortBinding{{HostPort: "0"}}},
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
	}, nil, containerName)
	if err != nil {
		return errors.Wrapf(err, "error creating container %s", containerName)
	}

	containerID := resp.ID

	if err := d.docker.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		return errors.Wrapf(err, "error starting container %s with ID %s", containerName, containerID)
	}

	// We bind to port 0, we need to extract the actual port assigned to us.
	cDetails, err := d.docker.ContainerInspect(ctx, containerID)
	if err != nil {
		return errors.Wrapf(err, "error when inspecting container %s with ID %s", containerName, containerID)
	}
	binding, ok := cDetails.NetworkSettings.Ports[functionAPIPort]
	if !ok || len(binding) < 1 {
		return errors.Errorf("No port assigned to function container, docker error or no more ports available")
	}

	c := dockerContainer{
		ID:               cDetails.ID,
		ImageID:          cDetails.Image,
		ImageName:        f.FunctionImageURL,
		FunctionID:       f.ID,
		FunctionRevision: f.FaasID,
		Port:             binding[0].HostPort,
		Host:             binding[0].HostIP,
	}

	d.containerCache.Store(f.ID, c)

	// clear any containers that could have been created before (e.g. before update)
	go d.deleteContainers(ctx, f, false)

	// make sure the function has started
	return utils.Backoff(d.RetryTimeout, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		cDetails, err := d.docker.ContainerInspect(ctx, containerID)
		if err != nil {
			return errors.Wrapf(err, "error when inspecting container %s for function %s", containerID, f.Name)
		}
		if !cDetails.State.Running {
			return errors.Errorf("container %s for function %s not running", containerID, f.Name)
		}

		resp, err := http.Get("http://" + d.ExternalHost + ":" + c.Port + healthcheckEndpoint)
		if err != nil {
			return errors.Wrapf(err, "error when checking health for function %s container %s", f.Name, containerID)
		}

		if resp.StatusCode != http.StatusOK {
			return errors.Errorf("incorrect status code %d when checking health for function %s container %s", resp.StatusCode, f.Name, containerID)
		}

		return nil
	})
}

// Delete deletes the function container.
func (d *Driver) Delete(ctx context.Context, f *functions.Function) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()
	if err := d.deleteContainers(ctx, f, true); err != nil {
		// no error wrapping, delete already does that
		return err
	}
	return nil
}

func (d *Driver) deleteContainers(ctx context.Context, f *functions.Function, deleteActive bool) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	filter := filters.NewArgs()
	filter.Add("label", labelFunctionID+"="+f.ID)
	containers, err := d.docker.ContainerList(ctx, types.ContainerListOptions{
		Filters: filter,
	})
	if err != nil {
		return errors.Wrapf(err, "error when finding containers for function ID %s", f.ID)
	}

	for _, c := range containers {
		if c.Labels[labelFunctionRevision] == f.FaasID && !deleteActive {
			continue
		}
		log.Debugf("Deleting container %s", c.ID)
		err := d.docker.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			return errors.Wrapf(err, "error when deleting container %s for function %s", c.ID, f.ID)
		}
	}

	// Clear cache if active is also to be deleted
	if _, ok := d.containerCache.Load(f.ID); ok && deleteActive {
		d.containerCache.Delete(f.ID)
	}

	return nil
}

func (d *Driver) findActiveContainer(ctx context.Context, functionID, functionRevision string) (*dockerContainer, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	filter := filters.NewArgs()
	filter.Add("label", labelFunctionID+"="+functionID)
	filter.Add("label", labelFunctionRevision+"="+functionRevision)
	containers, err := d.docker.ContainerList(ctx, types.ContainerListOptions{
		Filters: filter,
	})

	if err != nil {
		return nil, errors.Wrapf(err, "error when finding container for function ID %s", functionID)
	}

	if len(containers) == 0 {
		return nil, nil
	}

	// We need to inspect container to get networking details
	cDetails, err := d.docker.ContainerInspect(ctx, containers[0].ID)
	if err != nil {
		return nil, errors.Wrapf(err, "error when inspecting container %s", containers[0].ID)
	}

	binding, ok := cDetails.NetworkSettings.Ports[functionAPIPort]
	if !ok || len(binding) < 1 {
		return nil, errors.Errorf("No port assigned to function container, docker error or no more ports available")
	}

	return &dockerContainer{
		ID:               cDetails.ID,
		ImageID:          cDetails.Image,
		ImageName:        cDetails.Config.Image,
		FunctionID:       functionID,
		FunctionRevision: functionRevision,
		Port:             binding[0].HostPort,
		Host:             binding[0].HostIP,
	}, nil
}

// GetRunnable creates runnable representation of the function
func (d *Driver) GetRunnable(e *functions.FunctionExecution) functions.Runnable {
	return func(ctx functions.Context, in interface{}) (interface{}, error) {
		bytesIn, _ := json.Marshal(functions.Message{Context: ctx, Payload: in})
		var c dockerContainer

		ci, ok := d.containerCache.Load(e.FunctionID)
		if ok {
			c = ci.(dockerContainer)
		} else {
			cRef, err := d.findActiveContainer(context.Background(), e.FunctionID, e.FaasID)
			if err != nil {
				return nil, &systemError{errors.Errorf("error retrieving container for function %s", e.FunctionID)}
			}
			if cRef == nil {
				return nil, &systemError{errors.Errorf("missing container for function %s", e.FunctionID)}
			}
			c = *cRef
			// TODO this is racy, need to synchronize it
			d.containerCache.Store(e.FunctionID, c)
		}

		postURL := "http://" + d.ExternalHost + ":" + c.Port + "/"
		res, err := http.Post(postURL, jsonContentType, bytes.NewReader(bytesIn))
		if err != nil {
			log.Errorf("Error when sending POST request to %s: %+v", postURL, err)
			return nil, &systemError{errors.Wrapf(err, "request to function container on %s failed", postURL)}
		}
		defer res.Body.Close()

		log.Debugf("docker.run.%s: status code: %v", e.FunctionID, res.StatusCode)
		switch res.StatusCode {
		case 200:
			resBytes, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return nil, &systemError{errors.Errorf("cannot read result from function container on URL: %s %s", postURL, err)}
			}
			var out functions.Message
			if err := json.Unmarshal(resBytes, &out); err != nil {
				return nil, &systemError{errors.Errorf("cannot JSON-parse result from function container: %s %s", err, string(resBytes))}
			}
			ctx.AddLogs(out.Context.Logs())
			ctx.SetError(out.Context.GetError())
			return out.Payload, nil

		default:
			bytesOut, err := ioutil.ReadAll(res.Body)
			if err == nil {
				return nil, &systemError{errors.Errorf("Server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))}
			}
			return nil, &systemError{errors.Wrapf(err, "Error performing request, status: %v", res.StatusCode)}
		}
	}
}

func getID(functionName string, id string) string {
	return fmt.Sprintf("dispatch-%s-%s", functionName, id)
}

type systemError struct {
	Err error `json:"err"`
}

func (err *systemError) Error() string {
	return err.Err.Error()
}

func (err *systemError) AsSystemErrorObject() interface{} {
	return err
}

func (err *systemError) StackTrace() errors.StackTrace {
	if e, ok := err.Err.(functions.StackTracer); ok {
		return e.StackTrace()
	}

	return nil
}
