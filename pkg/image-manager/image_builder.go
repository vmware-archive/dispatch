///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package imagemanager

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	dockerTypes "github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/pkg/errors"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/entity-store"
)

type BaseImageBuilder struct {
	baseImageChannel chan BaseImage
	done             chan bool
	es               entitystore.EntityStore
	dockerClient     *docker.Client
	namespace        string
	orgID            string
}

type imageStatusResult struct {
	Result int `json:"result"`
}

func NewBaseImageBuilder(es entitystore.EntityStore) (*BaseImageBuilder, error) {
	dockerClient, err := docker.NewEnvClient()
	if err != nil {
		return nil, errors.Wrap(err, "Error creating docker client")
	}

	return &BaseImageBuilder{
		baseImageChannel: make(chan BaseImage),
		done:             make(chan bool),
		es:               es,
		dockerClient:     dockerClient,
		namespace:        ImageManagerFlags.K8sNamespace,
		orgID:            ImageManagerFlags.OrgID,
	}, nil
}

func (b *BaseImageBuilder) dockerPull(baseImage *BaseImage) error {
	// TODO (bjung): Need to use a lock of some sort in case we have multiple instanances of image builder running
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	var bytes []byte
	log.Printf("Pulling image %s/%s from %s", baseImage.OrganizationID, baseImage.Name, baseImage.DockerURL)
	rc, err := b.dockerClient.ImagePull(ctx, baseImage.DockerURL, dockerTypes.ImagePullOptions{All: false})
	if err == nil {
		defer rc.Close()
		_, err = rc.Read(bytes)
		fmt.Printf("docker return: %s\n", string(bytes))
	}
	if err != nil {
		baseImage.Status = StatusERROR
		baseImage.Reason = []string{err.Error()}
	} else {
		baseImage.Status = StatusREADY
		baseImage.Reason = nil
	}
	_, err = b.es.Update(baseImage.Revision, baseImage)
	if err != nil {
		err = errors.Wrap(err, "Error pulling docker image")
	}
	log.Printf("Successfully updated image image %s/%s", baseImage.OrganizationID, baseImage.Name)
	return err
}

func (b *BaseImageBuilder) dockerDelete(baseImage *BaseImage) error {
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
	var deleted BaseImage
	err = b.es.Delete(b.orgID, baseImage.Name, &deleted)
	if err != nil {
		return errors.Wrapf(err, "Error deleting base image entity %s/%s", baseImage.OrganizationID, baseImage.Name)
	}
	log.Printf("Successfully deleted base image %s/%s", baseImage.OrganizationID, baseImage.Name)
	return nil
}

func (b *BaseImageBuilder) dockerStatus() ([]BaseImage, error) {
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
	var all []BaseImage
	err = b.es.List(b.orgID, nil, &all)
	if err != nil {
		return nil, errors.Wrap(err, "Error listing docker images")
	}
	for i, bi := range all {
		url := bi.DockerURL
		parts := strings.SplitN(url, ":", 2)
		if len(parts) == 1 {
			url = fmt.Sprintf("%s:latest", url)
		}

		status := StatusREADY
		if _, ok := imageMap[url]; !ok {
			status = StatusERROR
		}
		if bi.Status != status {
			bi.Status = status
			rev, err := b.es.Update(bi.Revision, &bi)
			if err != nil {
				log.Printf("Error updating %s/%s, continue", bi.OrganizationID, bi.Name)
			}
			bi.Revision = uint64(rev)
			all[i] = bi
		}
	}
	return all, err
}

func (b *BaseImageBuilder) poll() error {
	baseImages, err := b.dockerStatus()
	if err != nil {
		return err
	}
	for _, bi := range baseImages {
		log.Printf("Polling base image %s/%s, delete: %v", bi.OrganizationID, bi.Name, bi.Delete)
		if bi.Delete {
			err = b.dockerDelete(&bi)
		}
		if bi.Status == StatusERROR {
			err = b.dockerPull(&bi)
		}
		if err != nil {
			log.Print(err)
		}
	}
	return nil
}

func (b *BaseImageBuilder) watch() error {
	for {
		var err error
		select {
		case bi := <-b.baseImageChannel:
			log.Printf("Received base image update %s/%s, delete: %v", bi.OrganizationID, bi.Name, bi.Delete)
			if bi.Delete {
				err = b.dockerDelete(&bi)
			} else {
				err = b.dockerPull(&bi)
			}
		case <-time.After(60 * time.Second):
			log.Printf("Polling docker daemon")
			err = b.poll()
		case <-b.done:
			return nil
		}
		if err != nil {
			log.Print(err)
		}
	}
}

func (b *BaseImageBuilder) Run() {
	b.watch()
}

func (b *BaseImageBuilder) Shutdown() {
	b.done <- true
}
