///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package images

// NO TESTS

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/dispatch/pkg/trace"
)

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
			return errors.Wrapf(err, "failed to parse ImagePull response: %s", s.Text())
		}
		if result.Error != nil {
			return errors.New(*result.Error)
		}
	}
	return nil
}

func BuildAndPushFromDir(client docker.ImageAPIClient, dir, name, registryAuth string) error {
	tarBall := &bytes.Buffer{}
	if err := tarDir(dir, tar.NewWriter(tarBall)); err != nil {
		return errors.Wrap(err, "failed to create a tarball archive")
	}

	if r, err := client.ImageBuild(context.Background(), tarBall, types.ImageBuildOptions{
		Tags:           []string{name},
		SuppressOutput: true,
	}); err != nil {
		return errors.Wrap(err, "failed to build an image")
	} else {
		r.Body.Close()
	}
	opts := types.ImagePushOptions{
		RegistryAuth: registryAuth,
	}
	if err := DockerError(client.ImagePush(context.Background(), name, opts)); err != nil {
		return errors.Wrapf(err, "failed to push the image %s", name)
	}
	return nil
}

func tarDir(source string, tarball *tar.Writer) error {
	defer trace.Tracef("tar dir: %s", source)()
	return filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			header.Name = "." + strings.TrimPrefix(path, source)
			log.Debugf("tar: writing header: %s", header.Name)

			if err := tarball.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarball, file)
			return err
		})
}
