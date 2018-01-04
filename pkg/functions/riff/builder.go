///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package riff

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
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/trace"
)

func (d *riffDriver) Shutdown() {
	defer trace.Trace("")()
	defer func() { recover() }() // close can panic
	close(d.buildRequests)
}

func cleanup(tmpDir string) {
	defer trace.Tracef("rm dir: %s", tmpDir)()
	os.RemoveAll(tmpDir)
}

func (d *riffDriver) processRequests() {
	for r := range d.buildRequests {
		r.result <- d.buildAndPushImage(r)
		close(r.result)
	}
}

func dockerError(r io.ReadCloser, err error) error {
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

func (d *riffDriver) buildAndPushImage(request *imgRequest) *imgResult {
	defer trace.Tracef("function: '%s', base: '%s'", request.name, request.exec.Image)()
	name := imageName(d.imageRegistry, request.name, utcTimeStampStr(time.Now()))
	log.Debugf("Building image '%s'", name)

	tmpDir, err := ioutil.TempDir("", "func-build")
	if err != nil {
		return &imgResult{err: errors.Wrap(err, "failed to create a temp dir")}
	}
	defer cleanup(tmpDir)
	log.Debugf("Created tmpDir: %s", tmpDir)
	log.Printf("Pulling image: %s", request.exec.Image)

	if err := dockerError(d.docker.ImagePull(context.Background(), request.exec.Image, types.ImagePullOptions{})); err != nil {
		return &imgResult{err: errors.Wrap(err, "failed to pull image")}
	}

	if err := writeDir(tmpDir, request.exec); err != nil {
		return &imgResult{err: err}
	}

	tarBall := &bytes.Buffer{}
	if err := tarDir(tmpDir, tar.NewWriter(tarBall)); err != nil {
		return &imgResult{err: err}
	}

	if r, err := d.docker.ImageBuild(context.Background(), tarBall, types.ImageBuildOptions{
		Tags:           []string{name},
		SuppressOutput: true,
	}); err != nil {
		return &imgResult{err: errors.Wrap(err, "failed to build an image")}
	} else {
		r.Body.Close()
	}

	if err := dockerError(d.docker.ImagePush(context.Background(), name, types.ImagePushOptions{
		RegistryAuth: d.registryAuth,
	})); err != nil {
		return &imgResult{err: errors.Wrap(err, "failed to push the image")}
	}

	return &imgResult{image: name}
}

func writeDockerfile(dir string, exec *functions.Exec, functionFiles map[string]os.FileMode) error {
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

func writeDir(dir string, exec *functions.Exec) error {
	defer trace.Tracef("dir: %s", dir)()
	switch l := exec.Language; l {
	case "nodejs6":
		return writeDockerfile(dir, exec, map[string]os.FileMode{"func.js": 0644})
	case "python3":
		return writeDockerfile(dir, exec, map[string]os.FileMode{"handler.py": 0644, "requirements.txt": 0644})
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
