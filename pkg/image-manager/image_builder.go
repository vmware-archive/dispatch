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
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/go-openapi/swag"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/images"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

// BaseImageBuilder manages base images, which are referenced docker images
type BaseImageBuilder struct {
	baseImageChannel chan BaseImage
	done             chan bool
	es               entitystore.EntityStore
	dockerClient     docker.ImageAPIClient
	pullPeriod       time.Duration
}

// ImageBuilder manages building images
type ImageBuilder struct {
	imageChannel chan Image
	done         chan bool
	es           entitystore.EntityStore
	dockerClient docker.CommonAPIClient
	registryHost string
	registryAuth string
	pullPeriod   time.Duration

	PushImages bool
}

type imageStatusResult struct {
	Result int `json:"result"`
}

// FilterLastPulledBefore creates a filter, which will filter images that were last pulled before a specified duration
func FilterLastPulledBefore(duration time.Duration) entitystore.Filter {
	f := entitystore.FilterEverything()
	f.Add(entitystore.FilterStat{
		Scope:   entitystore.FilterScopeExtra,
		Subject: "LastPullTime",
		Verb:    entitystore.FilterVerbBefore,
		Object:  time.Now().Add(-duration),
	})
	return f
}

// NewBaseImageBuilder is the constructor for the BaseImageBuilder
func NewBaseImageBuilder(es entitystore.EntityStore, pullPeriod time.Duration) (*BaseImageBuilder, error) {
	dockerClient, err := docker.NewEnvClient()
	if err != nil {
		return nil, errors.Wrap(err, "Error creating docker client")
	}

	return &BaseImageBuilder{
		baseImageChannel: make(chan BaseImage),
		done:             make(chan bool),
		es:               es,
		dockerClient:     dockerClient,
		pullPeriod:       pullPeriod,
	}, nil
}

func (b *BaseImageBuilder) baseImagePull(ctx context.Context, baseImage *BaseImage) error {
	// TODO (bjung): Need to use a lock of some sort in case we have multiple instances of image builder running
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	log.Printf("Pulling image %s/%s from %s", baseImage.OrganizationID, baseImage.Name, baseImage.DockerURL)
	rc, err := b.dockerClient.ImagePull(ctx, baseImage.DockerURL, dockerTypes.ImagePullOptions{All: false})
	if err != nil {
		return errors.Wrapf(err, "error when pulling image %s", baseImage.DockerURL)
	}
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
			return errors.Wrap(err, "error unmarshalling docker status")
		}
		log.Debugf("Docker status: %+v\n", status)
		if status.Error != "" {
			return errors.Wrap(fmt.Errorf(status.ErrorDetail.Message), "docker error when pulling image")
		}
	}
	if scanner.Err() != nil {
		return errors.Wrap(scanner.Err(), "error when reading data returned by docker")
	}

	log.Debugf("Successfully updated base-image %s/%s", baseImage.OrganizationID, baseImage.Name)
	return err
}

func (b *BaseImageBuilder) baseImageDelete(ctx context.Context, baseImage *BaseImage) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Even though we are explicitly removing the image, other base images which point to the same docker URL will
	// continue to work.  They remain in READY status, and the next "poll" loop should re-pull the image.  If the
	// image is pulled as part of an image create, the image will be pulled immediately and should continue to work.
	_, err := b.dockerClient.ImageRemove(ctx, baseImage.DockerURL, dockerTypes.ImageRemoveOptions{
		Force:         true,
		PruneChildren: true,
	})
	// If the image status is NOT ready, errors are expected, continue delete
	if err != nil && baseImage.Status == StatusREADY {
		return errors.Wrapf(err, "Error deleting image %s/%s", baseImage.OrganizationID, baseImage.Name)
	}
	log.Printf("Successfully deleted base image %s/%s", baseImage.OrganizationID, baseImage.Name)
	return nil
}

// DockerImageStatus gathers the status of multiple docker images
func DockerImageStatus(ctx context.Context, client docker.ImageAPIClient, images []DockerImage) ([]entitystore.Entity, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
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

func (b *BaseImageBuilder) baseImageStatus(ctx context.Context) ([]entitystore.Entity, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
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
	err = b.es.ListGlobal(ctx, entitystore.Options{Filter: FilterLastPulledBefore(b.pullPeriod)}, &all)
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
			// repo, move to MISSING state
			switch s := bi.Status; s {
			case entitystore.StatusREADY:
				bi.Status = entitystore.StatusMISSING
				entities = append(entities, bi)
				log.Debugf("base image %s is missing in docker daemon", bi.Name)
			}
		} else {
			// If the image is present, move to READY if in a
			// non-DELETING state
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
func NewImageBuilder(es entitystore.EntityStore, registryHost, registryAuth string, pullPeriod time.Duration) (*ImageBuilder, error) {
	dockerClient, err := docker.NewEnvClient()
	if err != nil {
		return nil, errors.Wrap(err, "Error creating docker client")
	}

	return &ImageBuilder{
		imageChannel: make(chan Image),
		done:         make(chan bool),
		es:           es,
		PushImages:   true,
		dockerClient: dockerClient,
		registryHost: registryHost,
		registryAuth: registryAuth,
		pullPeriod:   pullPeriod,
	}, nil
}

const (
	imageTemplateLabel      = "io.dispatchframework.imageTemplate"
	imageTemplateDirDefault = "/image-template"
)

func (b *ImageBuilder) copyImageTemplate(tmpDir string, image string) error {
	resp, err := b.dockerClient.ContainerCreate(context.Background(), &container.Config{
		Image: image,
	}, nil, nil, "")
	if err != nil {
		return errors.Wrap(err, "failed to create base-image container")
	}
	defer b.dockerClient.ContainerRemove(context.Background(), resp.ID, dockerTypes.ContainerRemoveOptions{})

	bic, err := b.dockerClient.ContainerInspect(context.Background(), resp.ID)
	if err != nil {
		return errors.Wrapf(err, "failed to inspect base-image container")
	}

	imageTemplateDir := bic.Config.Labels[imageTemplateLabel]
	if imageTemplateDir == "" {
		imageTemplateDir = imageTemplateDirDefault
	}

	readCloser, _, err := b.dockerClient.CopyFromContainer(context.Background(), resp.ID, imageTemplateDir)
	defer readCloser.Close()

	return utils.Untar(tmpDir, strings.TrimPrefix(imageTemplateDir, "/")+"/", readCloser)
}

const (
	packagesFile       = "packages.txt"
	systemPackagesFile = "system-packages.txt"
)

func (b *ImageBuilder) writeSystemPackagesFile(file string, image *Image) error {
	buffer := new(bytes.Buffer)

	for _, p := range image.SystemDependencies.Packages {
		if p.Name == "" || p.Version == "" {
			return errors.Errorf("invalid system package: empty name or version, name='%s', version='%s'", p.Name, p.Version)
		}
		fmt.Fprintf(buffer, "%s %s\n", p.Name, p.Version)
	}

	return ioutil.WriteFile(file, buffer.Bytes(), 0644)
}

func (b *ImageBuilder) writePackagesFile(file string, image *Image) error {
	manifestFileContent := []byte(image.RuntimeDependencies.Manifest)
	return ioutil.WriteFile(file, manifestFileContent, 0644)
}

func (b *ImageBuilder) imagePull(ctx context.Context, image *Image) error {
	log.Debug("Pulling image %s/%s", image.OrganizationID, image.Name)
	if err := images.DockerError(b.dockerClient.ImagePull(context.Background(), image.DockerURL, dockerTypes.ImagePullOptions{})); err != nil {
		return errors.Wrapf(err, "failed to pull image '%s'", image.DockerURL)
	}
	image.LastPullTime = time.Now()
	return nil
}

func (b *ImageBuilder) imageCreate(ctx context.Context, image *Image, baseImage *BaseImage) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	tmpDir, err := ioutil.TempDir("", "image-build")
	if err != nil {
		return errors.Wrap(err, "failed to create a temp dir")
	}
	defer os.RemoveAll(tmpDir)

	if err := images.DockerError(b.dockerClient.ImagePull(context.Background(), baseImage.DockerURL, dockerTypes.ImagePullOptions{})); err != nil {
		return errors.Wrapf(err, "failed to pull image '%s'", baseImage.DockerURL)
	}

	if err := b.copyImageTemplate(tmpDir, baseImage.DockerURL); err != nil {
		return err
	}

	spFile := filepath.Join(tmpDir, systemPackagesFile)
	if err := b.writeSystemPackagesFile(spFile, image); err != nil {
		return errors.Wrapf(err, "failed to write packages file %s", spFile)
	}

	pFile := filepath.Join(tmpDir, packagesFile)
	if err := b.writePackagesFile(pFile, image); err != nil {
		return errors.Wrapf(err, "failed to write %s", pFile)
	}

	dockerURL := strings.Join([]string{b.registryHost, image.GetID() + ":latest"}, "/")
	buildArgs := map[string]*string{
		"BASE_IMAGE":           swag.String(baseImage.DockerURL),
		"SYSTEM_PACKAGES_FILE": swag.String(systemPackagesFile),
		"PACKAGES_FILE":        swag.String(packagesFile),
	}
	err = images.BuildAndPushFromDir(ctx, b.dockerClient, tmpDir, dockerURL, b.registryAuth, b.PushImages, buildArgs)
	if err != nil {
		return err
	}
	image.DockerURL = dockerURL
	image.Status = entitystore.StatusREADY
	image.LastPullTime = time.Now()
	// TODO (bjung) run `tndf list and pip3 freeze` after image is built to get the list of installed
	// dependencies
	return nil
}

func (b *ImageBuilder) imageStatus(ctx context.Context) ([]entitystore.Entity, error) {
	// Currently the status simply mirrors the base image.  This will change as we actually
	// start building upon the image
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	var all []*Image
	err := b.es.ListGlobal(ctx, entitystore.Options{Filter: FilterLastPulledBefore(b.pullPeriod)}, &all)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting list of images")
	}
	var images []DockerImage
	for _, i := range all {
		images = append(images, i)
	}
	return DockerImageStatus(ctx, b.dockerClient, images)
}

func (b *ImageBuilder) imageDelete(ctx context.Context, image *Image) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := b.dockerClient.ImageRemove(ctx, image.DockerURL, dockerTypes.ImageRemoveOptions{
		Force:         true,
		PruneChildren: true,
	})
	// If the image status is NOT ready, errors are expected, continue delete
	if err != nil && image.Status == entitystore.StatusREADY {
		return errors.Wrapf(err, "Error deleting image %s/%s", image.OrganizationID, image.Name)
	}
	log.Printf("Successfully deleted image %s/%s", image.OrganizationID, image.Name)
	return nil
}

func (b *ImageBuilder) imageUpdate(ctx context.Context, image *Image) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	var bi BaseImage
	err := b.es.Get(ctx, image.OrganizationID, image.BaseImageName, entitystore.Options{}, &bi)
	if err != nil {
		return errors.Wrapf(err, "Error getting base image entity %s/%s", image.OrganizationID, image.Name)
	}
	if image.Status != bi.Status {
		image.Status = bi.Status
		rev, err := b.es.Update(ctx, image.Revision, image)
		if err != nil {
			return errors.Wrapf(err, "Error updating image entity %s/%s", image.OrganizationID, image.Name)
		}
		image.Revision = uint64(rev)
	}
	log.Printf("Successfully updated image %s/%s", image.OrganizationID, image.Name)
	return nil
}
