///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package functions

import (
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/images"
	"github.com/vmware/dispatch/pkg/trace"
)

// DockerImageBuilder builds function images
type DockerImageBuilder struct {
	imageRegistry       string
	registryAuth        string
	functionTemplateDir string

	docker *docker.Client
}

// NewDockerImageBuilder is the constructor for the DockerImageBuilder
func NewDockerImageBuilder(imageRegistry, registryAuth, functionTemplateDir string, docker *docker.Client) *DockerImageBuilder {
	return &DockerImageBuilder{
		imageRegistry:       imageRegistry,
		registryAuth:        registryAuth,
		functionTemplateDir: functionTemplateDir,
		docker:              docker,
	}
}

// BuildImage packages a function into a docker image.  It also adds any FaaS specfic image layers
func (ib *DockerImageBuilder) BuildImage(faas, fnID string, exec *Exec) (string, error) {
	defer trace.Tracef("function: '%s', base: '%s'", fnID, exec.Image)()
	name := imageName(ib.imageRegistry, faas, fnID)
	log.Debugf("Building image '%s'", name)

	tmpDir, err := ioutil.TempDir("", "func-build")
	if err != nil {
		return "", errors.Wrap(err, "failed to create a temp dir")
	}
	defer cleanup(tmpDir)

	log.Debugf("Created tmpDir: %s", tmpDir)
	log.Printf("Pulling image: %s", exec.Image)

	if err := images.DockerError(ib.docker.ImagePull(context.Background(), exec.Image, types.ImagePullOptions{})); err != nil {
		return "", errors.Wrap(err, "failed to pull image")
	}

	if err := writeDir(tmpDir, ib.functionTemplateDir, faas, exec); err != nil {
		return "", errors.Wrap(err, "failed to write dockerfile")
	}

	err = images.BuildAndPushFromDir(ib.docker, tmpDir, name, ib.registryAuth)
	return name, err
}

func cleanup(tmpDir string) {
	defer trace.Tracef("rm dir: %s", tmpDir)()
	os.RemoveAll(tmpDir)
}

func writeFunctionDockerfile(dir, functionTemplateDir, faas string, exec *Exec) error {
	functionFile := "function.txt"
	functionPath := filepath.Join(dir, functionFile)

	if err := ioutil.WriteFile(functionPath, []byte(exec.Code), 0644); err != nil {
		return errors.Wrapf(err, "failed to write function/%s", exec.Name)
	}

	templateArgs := struct {
		FaaS         string
		Language     string
		DockerURL    string
		FunctionFile string
	}{
		FaaS:         faas,
		Language:     exec.Language,
		DockerURL:    exec.Image,
		FunctionFile: functionFile,
	}

	srcDir := filepath.Join(functionTemplateDir, faas, exec.Language)
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("faas driver %s does not support language %s", faas, exec.Language)
	}
	templateFiles, err := ioutil.ReadDir(srcDir)
	if err != nil {
		return errors.Wrapf(err, "failed to read function template directory %s", srcDir)
	}

	for _, src := range templateFiles {
		dest, err := os.Create(filepath.Join(dir, src.Name()))
		defer dest.Close()
		if err != nil {
			return errors.Wrapf(err, "failed to create dest file %s/%s", dir, src.Name())
		}
		templatePath := filepath.Join(srcDir, src.Name())
		templateBytes, err := ioutil.ReadFile(templatePath)
		if err != nil {
			return errors.Wrapf(err, "failed to read template file %s/%s", srcDir, src.Name())
		}
		tmpl, err := template.New(dest.Name()).Parse(string(templateBytes))
		if err != nil {
			return errors.Wrapf(err, "failed to parse template %s", dest.Name())
		}
		err = tmpl.Execute(dest, &templateArgs)
		if err != nil {
			return errors.Wrapf(err, "failed to render template with args %s: %+v", dest.Name(), templateArgs)
		}
	}
	return nil
}

func writeDir(dir, functionTemplateDir, faas string, exec *Exec) error {
	defer trace.Tracef("dir: %s", dir)()
	return writeFunctionDockerfile(dir, functionTemplateDir, faas, exec)
}

func imageName(registry, faas, fnID string) string {
	return registry + "/func-" + faas + "-" + fnID + ":latest"
}
