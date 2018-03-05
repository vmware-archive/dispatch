///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package imagemanager

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	dockerTypes "github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/images"
	"github.com/vmware/dispatch/pkg/trace"
)

// BaseImageBuilder manages base images, which are referenced docker images
type BaseImageBuilder struct {
	baseImageChannel chan BaseImage
	done             chan bool
	es               entitystore.EntityStore
	dockerClient     docker.ImageAPIClient
	orgID            string
}

// ImageBuilder manages building images
type ImageBuilder struct {
	imageChannel chan Image
	done         chan bool
	es           entitystore.EntityStore
	dockerClient docker.ImageAPIClient
	orgID        string
	registryHost string
	registryAuth string
}

type imageStatusResult struct {
	Result int `json:"result"`
}

// NewBaseImageBuilder is the constructor for the BaseImageBuilder
func NewBaseImageBuilder(es entitystore.EntityStore) (*BaseImageBuilder, error) {
	defer trace.Trace("NewBaseImageBuilder")()
	dockerClient, err := docker.NewEnvClient()
	if err != nil {
		return nil, errors.Wrap(err, "Error creating docker client")
	}

	return &BaseImageBuilder{
		baseImageChannel: make(chan BaseImage),
		done:             make(chan bool),
		es:               es,
		dockerClient:     dockerClient,
		orgID:            ImageManagerFlags.OrgID,
	}, nil
}

func (b *BaseImageBuilder) baseImagePull(baseImage *BaseImage) error {
	defer trace.Trace("")()
	// TODO (bjung): Need to use a lock of some sort in case we have multiple instanances of image builder running
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	log.Printf("Pulling image %s/%s from %s", baseImage.OrganizationID, baseImage.Name, baseImage.DockerURL)
	rc, err := b.dockerClient.ImagePull(ctx, baseImage.DockerURL, dockerTypes.ImagePullOptions{All: false})
	if err == nil {
		defer rc.Close()
		scanner := bufio.NewScanner(rc)
		for scanner.Scan() {
			bytes := scanner.Bytes()
			status := struct {
				ErrorDetail struct {
					Message string `json:"message"`
				} `json:"errorDetail"`
				Error   string `json:"error"`
				Message string `json:"message"`
			}{}
			err = json.Unmarshal(bytes, &status)
			if err != nil {
				// Return immediately on unmarshal error (do not update status)
				// Assume this is a transient error.
				return errors.Wrap(err, "Error unmarshalling docker status")
			}
			log.Printf("Docker status: %+v\n", status)
			if status.Error != "" {
				err = fmt.Errorf(status.ErrorDetail.Message)
				break
			}
		}
		if scanner.Err() != nil {
			err = scanner.Err()
		}
	}
	log.Printf("Successfully updated base-image %s/%s", baseImage.OrganizationID, baseImage.Name)
	return err
}

func (b *BaseImageBuilder) baseImageDelete(baseImage *BaseImage) error {
	defer trace.Trace("")()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Even though we are explicitly removing the image, other base images which point to the same docker URL will
	// continue to work.  They remain in READY status, and the next "poll" loop should re-pull the image.  If the
	// image is pulled as part of an image create, the image will be pulled immediately and should continue to work.
	_, err := b.dockerClient.ImageRemove(ctx, baseImage.DockerURL, dockerTypes.ImageRemoveOptions{Force: true})
	// If the image status is NOT ready, errors are expected, continue delete
	if err != nil && baseImage.Status == StatusREADY {
		return errors.Wrapf(err, "Error deleting image %s/%s", baseImage.OrganizationID, baseImage.Name)
	}
	log.Printf("Successfully deleted base image %s/%s", baseImage.OrganizationID, baseImage.Name)
	return nil
}

// DockerImageStatus gathers the status of multiple docker images
func DockerImageStatus(client docker.ImageAPIClient, images []DockerImage) ([]entitystore.Entity, error) {
	defer trace.Trace("")()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	summary, err := client.ImageList(ctx, dockerTypes.ImageListOptions{All: false})
	if err != nil {
		return nil, errors.Wrap(err, "Error listing docker images")
	}
	imageMap := make(map[string]bool)
	for _, is := range summary {
		for _, t := range is.RepoTags {
			imageMap[t] = true
		}
	}
	var entities []entitystore.Entity

	if err != nil {
		return nil, errors.Wrap(err, "Error listing docker images")
	}
	for _, i := range images {
		url := i.GetDockerURL()
		parts := strings.SplitN(url, ":", 2)
		if len(parts) == 1 {
			url = fmt.Sprintf("%s:latest", url)
		}

		if _, ok := imageMap[url]; !ok {
			// If we are READY, but the image is missing from the
			// repo, move to ERROR state
			switch s := i.GetStatus(); s {
			case entitystore.StatusREADY:
				i.SetStatus(entitystore.StatusMISSING)
				entities = append(entities, i)
			}
		} else {
			// If the image is present, move to READY if in a
			// non-DELETING statue
			switch s := i.GetStatus(); s {
			case entitystore.StatusINITIALIZED,
				entitystore.StatusCREATING,
				entitystore.StatusUPDATING,
				entitystore.StatusERROR:
				i.SetStatus(entitystore.StatusREADY)
				entities = append(entities, i)
			}
		}
	}
	return entities, err
}

func (b *BaseImageBuilder) baseImageStatus() ([]entitystore.Entity, error) {
	defer trace.Trace("")()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	summary, err := b.dockerClient.ImageList(ctx, dockerTypes.ImageListOptions{All: false})
	if err != nil {
		return nil, errors.Wrap(err, "Error listing docker images")
	}
	imageMap := make(map[string]bool)
	for _, is := range summary {
		for _, t := range is.RepoTags {
			imageMap[t] = true
		}
	}
	var entities []entitystore.Entity
	var all []*BaseImage
	err = b.es.List(b.orgID, entitystore.Options{}, &all)
	if err != nil {
		return nil, errors.Wrap(err, "Error listing docker images")
	}
	for _, bi := range all {
		url := bi.DockerURL
		parts := strings.SplitN(url, ":", 2)
		if len(parts) == 1 {
			url = fmt.Sprintf("%s:latest", url)
		}

		if _, ok := imageMap[url]; !ok {
			// If we are READY, but the image is missing from the
			// repo, move to ERROR state
			switch s := bi.Status; s {
			case entitystore.StatusREADY:
				bi.Status = entitystore.StatusMISSING
				entities = append(entities, bi)
			}
		} else {
			// If the image is present, move to READY if in a
			// non-DELETING statue
			switch s := bi.Status; s {
			case entitystore.StatusINITIALIZED,
				entitystore.StatusCREATING,
				entitystore.StatusUPDATING,
				entitystore.StatusERROR:
				bi.Status = entitystore.StatusREADY
				entities = append(entities, bi)
			}
		}
	}
	return entities, err
}

// NewImageBuilder is the constructor for the ImageBuilder
func NewImageBuilder(es entitystore.EntityStore, registryHost, registryAuth string) (*ImageBuilder, error) {
	defer trace.Trace("NewBaseImageBuilder")()
	dockerClient, err := docker.NewEnvClient()
	if err != nil {
		return nil, errors.Wrap(err, "Error creating docker client")
	}

	return &ImageBuilder{
		imageChannel: make(chan Image),
		done:         make(chan bool),
		es:           es,
		dockerClient: dockerClient,
		orgID:        ImageManagerFlags.OrgID,
		registryHost: registryHost,
		registryAuth: registryAuth,
	}, nil
}

func cleanup(tmpDir string) {
	defer trace.Tracef("rm dir: %s", tmpDir)()
	os.RemoveAll(tmpDir)
}

func writeDockerFile(dir string, baseImage *BaseImage, image *Image) (string, error) {
	dockerfile := new(bytes.Buffer)
	_, err := WriteSystemDockerfile(dir, dockerfile, baseImage, image)
	if err != nil {
		return "", errors.Wrap(err, "failed to write Dockerfile")
	}
	format, err := WriteRuntimeDockerfile(dir, dockerfile, image)
	if err != nil {
		return "", errors.Wrap(err, "failed to write Dockerfile")
	}
	log.Printf("Dockerfile:\n%s\n", dockerfile.String())
	if err := ioutil.WriteFile(filepath.Join(dir, "Dockerfile"), dockerfile.Bytes(), 0644); err != nil {
		return "", errors.Wrap(err, "failed to write Dockerfile")
	}
	return format, nil
}

func (b *ImageBuilder) imageCreate(image *Image, baseImage *BaseImage) error {
	tmpDir, err := ioutil.TempDir("", "func-build")
	if err != nil {
		return errors.Wrap(err, "failed to create a temp dir")
	}
	defer cleanup(tmpDir)

	if err := images.DockerError(b.dockerClient.ImagePull(context.Background(), baseImage.DockerURL, dockerTypes.ImagePullOptions{})); err != nil {
		return errors.Wrap(err, "failed to pull image")
	}

	format, err := writeDockerFile(tmpDir, baseImage, image)
	if err != nil {
		return err
	}

	dockerURL := strings.Join([]string{b.registryHost, image.GetID() + ":latest"}, "/")
	err = images.BuildAndPushFromDir(b.dockerClient, tmpDir, dockerURL, b.registryAuth)
	if err != nil {
		return err
	}
	image.DockerURL = dockerURL
	image.Status = entitystore.StatusREADY
	image.RuntimeDependencies.Format = format
	// TODO (bjung) run `tndf list and pip3 freeze` after image is built to get the list of installed
	// dependencies
	return nil
}

func (b *ImageBuilder) imageStatus() ([]entitystore.Entity, error) {
	// Currently the status simply mirrors the base image.  This will change as we actually
	// start building upon the image
	var all []*Image
	err := b.es.List(b.orgID, entitystore.Options{}, &all)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting list of images")
	}
	var images []DockerImage
	for _, i := range all {
		images = append(images, i)
	}
	return DockerImageStatus(b.dockerClient, images)
}

func (b *ImageBuilder) imageDelete(image *Image) error {
	defer trace.Trace("")()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := b.dockerClient.ImageRemove(ctx, image.DockerURL, dockerTypes.ImageRemoveOptions{Force: true})
	// If the image status is NOT ready, errors are expected, continue delete
	if err != nil && image.Status == entitystore.StatusREADY {
		return errors.Wrapf(err, "Error deleting image %s/%s", image.OrganizationID, image.Name)
	}
	log.Printf("Successfully deleted image %s/%s", image.OrganizationID, image.Name)
	return nil
}

func (b *ImageBuilder) imageUpdate(image *Image) error {
	var bi BaseImage
	err := b.es.Get(b.orgID, image.BaseImageName, entitystore.Options{}, &bi)
	if err != nil {
		return errors.Wrapf(err, "Error getting base image entity %s/%s", image.OrganizationID, image.Name)
	}
	if image.Status != bi.Status {
		image.Status = bi.Status
		rev, err := b.es.Update(image.Revision, image)
		if err != nil {
			return errors.Wrapf(err, "Error updating image entity %s/%s", image.OrganizationID, image.Name)
		}
		image.Revision = uint64(rev)
	}
	log.Printf("Successfully updated image %s/%s", image.OrganizationID, image.Name)
	return nil
}
