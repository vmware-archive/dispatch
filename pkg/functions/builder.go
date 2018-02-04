///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package functions

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/trace"
)

type DockerImageBuilder struct {
	imageRegistry string
	registryAuth  string

	docker *docker.Client
}

func NewDockerImageBuilder(imageRegistry string, registryAuth string, docker *docker.Client) *DockerImageBuilder {
	return &DockerImageBuilder{
		imageRegistry: imageRegistry,
		registryAuth:  registryAuth,
		docker:        docker,
	}
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
	opts := types.ImagePushOptions{}
	if registryAuth != "" {
		opts.RegistryAuth = registryAuth
	}

	if err := DockerError(client.ImagePush(context.Background(), name, opts)); err != nil {
		return errors.Wrapf(err, "failed to push the image %s", name)
	}
	return nil
}

func (ib *DockerImageBuilder) BuildImage(fnName string, exec *Exec) (string, error) {
	defer trace.Tracef("function: '%s', base: '%s'", fnName, exec.Image)()
	name := imageName(ib.imageRegistry, fnName, utcTimeStampStr(time.Now()))
	log.Debugf("Building image '%s'", name)

	tmpDir, err := ioutil.TempDir("", "func-build")
	if err != nil {
		return "", errors.Wrap(err, "failed to create a temp dir")
	}
	defer cleanup(tmpDir)

	log.Debugf("Created tmpDir: %s", tmpDir)
	log.Printf("Pulling image: %s", exec.Image)

	if err := DockerError(ib.docker.ImagePull(context.Background(), exec.Image, types.ImagePullOptions{})); err != nil {
		return "", errors.Wrap(err, "failed to pull image")
	}

	if err := writeDir(tmpDir, exec); err != nil {
		return "", errors.Wrap(err, "failed to write dockerfile")
	}

	tarBall := &bytes.Buffer{}
	if err := tarDir(tmpDir, tar.NewWriter(tarBall)); err != nil {
		return "", errors.Wrap(err, "failed to create a tarball archive")
	}

	if r, err := ib.docker.ImageBuild(context.Background(), tarBall, types.ImageBuildOptions{
		Tags:           []string{name},
		SuppressOutput: true,
	}); err != nil {
		return "", errors.Wrap(err, "failed to build an image")
	} else {
		r.Body.Close()
	}

	if err := DockerError(ib.docker.ImagePush(context.Background(), name, types.ImagePushOptions{
		RegistryAuth: ib.registryAuth,
	})); err != nil {
		return "", errors.Wrapf(err, "failed to push the image to registry %s", ib.imageRegistry)
	}

	return name, nil
}

func cleanup(tmpDir string) {
	defer trace.Tracef("rm dir: %s", tmpDir)()
	os.RemoveAll(tmpDir)
}

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

func writeDockerfile(dir string, exec *Exec, functionFiles map[string]os.FileMode) error {
	defer trace.Tracef("writeDockerfile")()
	if err := os.MkdirAll(filepath.Join(dir, "function"), os.ModePerm); err != nil {
		return errors.Wrap(err, "error creating function dir")
	}
	dockerFileContent := []byte(fmt.Sprintf("FROM %s\nCOPY function function\n", exec.Image))
	if err := ioutil.WriteFile(filepath.Join(dir, "Dockerfile"), dockerFileContent, 0644); err != nil {
		return errors.Wrap(err, "failed to write Dockerfile")
	}
	for f, perm := range functionFiles {
		if err := ioutil.WriteFile(filepath.Join(dir, "function", f), []byte(exec.Code), perm); err != nil {
			return errors.Wrapf(err, "failed to write function/%s", f)
		}
	}
	return nil
}

func writeDir(dir string, exec *Exec) error {
	defer trace.Tracef("dir: %s", dir)()
	switch l := exec.Language; l {
	case "nodejs6":
		return writeDockerfile(dir, exec, map[string]os.FileMode{"func.js": 0644})
	case "python3":
		return writeDockerfile(dir, exec, map[string]os.FileMode{"handler.py": 0644, "requirements.txt": 0644})
	case "powershell":
		return writeDockerfile(dir, exec, map[string]os.FileMode{"handler.ps1": 0644})
	}
	return fmt.Errorf("Unknown or unavailable runtime language: %s", exec.Language)
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

func imageName(registry, funcName, ts string) string {
	return registry + "/func-" + funcName + ":" + ts
}

func utcTimeStampStr(t time.Time) string {
	return t.UTC().Format("20060102-150405")
}
