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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/trace"
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
		log.Debugf("Packing %s", f.Name())
		b, _ := ioutil.ReadFile(filepath.Join(dir, f.Name()))
		log.Debug(string(b))
	}

	tarBall := new(bytes.Buffer)
	if err := tarDir(dir, tarBall); err != nil {
		return errors.Wrap(err, "failed to create a tarball archive")
	}

	log.Debugf("Building image %s from tarball", name)
	r, err := client.ImageBuild(ctx, tarBall, types.ImageBuildOptions{
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

// Untar the tar stream r into the dst dir stripping prefix from file paths
func Untar(dst, prefix string, r io.Reader) error {
	tarReader := tar.NewReader(r)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(dst, strings.TrimPrefix(header.Name, prefix))
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}

func tarDir(source string, w io.Writer) error {
	tarball := tar.NewWriter(w)
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
