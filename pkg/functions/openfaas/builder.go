///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package openfaas

import (
	"archive/tar"
	"bytes"
	"context"
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

	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/trace"
)

func (d *ofDriver) Shutdown() {
	defer trace.Trace("")()
	defer func() { recover() }() // close can panic
	close(d.requests)
}

func cleanup(tmpDir string) {
	defer trace.Tracef("rm dir: %s", tmpDir)()
	os.RemoveAll(tmpDir)
}

func (d *ofDriver) processRequests() {
	for r := range d.requests {
		r.result <- d.buildAndPushImage(r)
		close(r.result)
	}
}

func (d *ofDriver) buildAndPushImage(request *imgRequest) *imgResult {
	defer trace.Tracef("function: '%s', base: '%s'", request.name, request.exec.Image)()
	name := imageName(d.imageRegistry, request.name, utcTimeStampStr(time.Now()))
	log.Debugf("Building image '%s'", name)

	tmpDir, err := ioutil.TempDir("", "func-build")
	if err != nil {
		return &imgResult{"", errors.Wrap(err, "failed to create a temp dir")}
	}
	defer cleanup(tmpDir)
	log.Debugf("Created tmpDir: %s", tmpDir)

	if r, err := d.docker.ImagePull(context.Background(), request.exec.Image, types.ImagePullOptions{}); err != nil {
		return &imgResult{"", errors.Wrap(err, "failed to create a temp dir")}
	} else {
		r.Close()
	}

	if err := writeDir(tmpDir, request.exec); err != nil {
		return &imgResult{"", err}
	}

	tarBall := &bytes.Buffer{}
	if err := tarDir(tmpDir, tar.NewWriter(tarBall)); err != nil {
		return &imgResult{"", err}
	}

	if r, err := d.docker.ImageBuild(context.Background(), tarBall, types.ImageBuildOptions{
		Tags:           []string{name},
		SuppressOutput: true,
	}); err != nil {
		return &imgResult{"", errors.Wrap(err, "failed to build an image")}
	} else {
		r.Body.Close()
	}

	if r, err := d.docker.ImagePush(context.Background(), name, types.ImagePushOptions{
		RegistryAuth: d.registryAuth,
	}); err != nil {
		return &imgResult{"", errors.Wrap(err, "failed to push the image")}
	} else {
		r.Close()
	}

	return &imgResult{image: name}
}

func writeDir(dir string, exec *functions.Exec) error {
	defer trace.Tracef("dir: %s", dir)()
	if err := os.MkdirAll(filepath.Join(dir, "function"), os.ModePerm); err != nil {
		return errors.Wrap(err, "error creating function dir")
	}
	dockerFileContent := []byte(fmt.Sprintf("FROM %s\nCOPY function function\n", exec.Image))
	if err := ioutil.WriteFile(filepath.Join(dir, "Dockerfile"), dockerFileContent, 0644); err != nil {
		return errors.Wrap(err, "failed to write Dockerfile")
	}
	if err := ioutil.WriteFile(filepath.Join(dir, "function", "func.js"), []byte(exec.Code), 0644); err != nil {
		return errors.Wrap(err, "failed to write function/func.js")
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

func imageName(registry, funcName, ts string) string {
	return registry + "/func-" + funcName + ":" + ts
}

func utcTimeStampStr(t time.Time) string {
	return t.UTC().Format("20060102-150405")
}
