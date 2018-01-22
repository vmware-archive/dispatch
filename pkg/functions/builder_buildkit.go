///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package functions

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	buildkit "github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/vmware/dispatch/pkg/trace"
)

var configTmpl = `{
	"auths": {
		"%s": {
			"auth": "%s"
		}
	}
}
`

// BuildKitImageBuilder implements ImageBuilder interface using BuildKit as a backend.
type BuildKitImageBuilder struct {
	imageRegistry string

	buildkitClient *buildkit.Client
}

// NewBuildKitImageBuilder creates new instance of ImageBuilder with BuildKit backend.
// Currently, registryAuth is base64-encoded string with JSON containing username and password.
func NewBuildKitImageBuilder(imageRegistry string, registryAuth string) (*BuildKitImageBuilder, error) {
	err := prepareRegistryCreds(imageRegistry, registryAuth)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare image registry credentials")
	}

	c, err := buildkit.New("tcp://127.0.0.1:1234")
	if err != nil {
		return nil, errors.Wrap(err, "error creating buildkit client")
	}

	return &BuildKitImageBuilder{
		imageRegistry:  imageRegistry,
		buildkitClient: c,
	}, nil
}

func (ib *BuildKitImageBuilder) BuildImage(fnName string, exec *Exec) (string, error) {
	defer trace.Tracef("function: '%s', base: '%s'", fnName, exec.Image)()
	targetName := imageName(ib.imageRegistry, fnName, utcTimeStampStr(time.Now()))
	log.Debugf("Building image '%s'", targetName)

	tmpDir, err := ioutil.TempDir("", "func-build")
	if err != nil {
		return "", errors.Wrap(err, "failed to create a temp dir")
	}
	defer cleanup(tmpDir)
	log.Debugf("Created tmpDir: %s", tmpDir)

	writeFuncFile(tmpDir, exec)
	if err != nil {
		return "", errors.Wrap(err, "failed to prepare function image")
	}

	err = ib.buildImage(fnName, targetName, tmpDir, exec)
	if err != nil {
		return "", errors.Wrap(err, "failed to build function image")
	}

	return targetName, nil
}

// buildImage constructs llb representation of function image and sends it to BuildKit daemon.
// Image after building is pushed to the image registry.
func (ib *BuildKitImageBuilder) buildImage(fnName string, targetImage string, srcDir string, exec *Exec) error {
	defer trace.Tracef("fnName: %s, targetImage: %s, srcDir: %s", fnName, targetImage, srcDir)()
	ctx := context.Background()
	eg, ctx := errgroup.WithContext(ctx)

	var buildStatusChan chan *buildkit.SolveStatus
	if log.GetLevel() == log.DebugLevel {
		buildStatusChan = make(chan *buildkit.SolveStatus)
	}

	baseImage := llb.Image(exec.Image)
	functionSource := llb.Local("function-source")
	funcImage := baseImage.With(
		copyFrom(exec.Image, functionSource, "/.", "/root/function/"),
	)

	def, err := funcImage.Marshal()
	if err != nil {
		return errors.Wrapf(err, "error when marshaling function image definition")
	}

	opts := buildkit.SolveOpt{
		Exporter:      buildkit.ExporterImage,
		ExporterAttrs: map[string]string{"name": targetImage, "push": "true"},
		LocalDirs:     map[string]string{"function-source": srcDir},
	}

	eg.Go(func() error {
		err := ib.buildkitClient.Solve(ctx, def, opts, buildStatusChan)
		return err
	})

	if log.GetLevel() == log.DebugLevel {
		eg.Go(func() error {
			for p := range buildStatusChan {
				for _, s := range p.Statuses {
					log.Debugf("status: %s %s %d\n", s.Vertex, s.ID, s.Current)
				}
			}
			return nil
		})
	}

	err = eg.Wait()
	if err != nil {
		return errors.Wrapf(err, "error executing buildkit solver")
	}

	return nil
}

// writeFuncFile creates necessary function file before pushing it to BuildKit
func writeFuncFile(dir string, exec *Exec) error {
	defer trace.Tracef("dir: %s", dir)()

	switch l := exec.Language; l {
	case "nodejs6":
		return ioutil.WriteFile(filepath.Join(dir, "func.js"), []byte(exec.Code), 0644)
	case "python3":
		return ioutil.WriteFile(filepath.Join(dir, "handler.py"), []byte(exec.Code), 0644)
	case "powershell":
		return ioutil.WriteFile(filepath.Join(dir, "handler.ps1"), []byte(exec.Code), 0644)
	}
	return fmt.Errorf("unknown or unavailable runtime language: %s", exec.Language)
}

// copyFrom has similar semantics as `COPY --from`
func copyFrom(img string, src llb.State, srcPath, destPath string) llb.StateOption {
	defer trace.Tracef("image: %s, src: %s, srcPath: %s, destPath: %s", img, src.GetDir(), srcPath, destPath)()
	return func(s llb.State) llb.State {
		return copy(img, src, srcPath, s, destPath)
	}
}

// copy mounts source and destination layers and copies content of srcPath into dest.
func copy(img string, src llb.State, srcPath string, dest llb.State, destPath string) llb.State {
	defer trace.Tracef("image: %s, src: %s, srcPath: %s, dest: %s, destPath: %s", img, src.GetDir(), srcPath, dest.GetDir(), destPath)()
	cpImage := llb.Image(img)
	cp := cpImage.Run(llb.Shlexf("cp -a /src%s /dest%s", srcPath, destPath))
	cp.AddMount("/src", src)
	return cp.AddMount("/dest", dest)
}

// prepareRegistryCreds creates a docker config file with basic auth credentials.
func prepareRegistryCreds(imageRegistry string, registryAuth string) error {
	configDir := os.Getenv("DOCKER_CONFIG")
	err := os.MkdirAll(configDir, 0644)
	if err != nil {
		return errors.Wrap(err, "error creating DOCKER_CONFIG directory")
	}

	// if imageRegistry does not contain . or :, it's username/org in Docker Hub.
	if !strings.ContainsAny(imageRegistry, ".:") {
		imageRegistry = "https://index.docker.io/v1/"
	}

	// We need to re-parse credentials from JSON expected by docker client into a basic auth
	// that can be stored in docker config.
	// TODO: Pass credentials as dictionary instead of base64-encoded JSON.
	type DockerCreds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var registryCreds DockerCreds

	credsJSON, err := base64.StdEncoding.DecodeString(registryAuth)
	if err != nil {
		return errors.Wrap(err, "error when decoding image registry credentials")
	}

	err = json.Unmarshal(credsJSON, &registryCreds)
	if err != nil {
		return errors.Wrap(err, "error when parsing JSON with image registry credentials")
	}

	basicAuthCreds := fmt.Sprintf("%s:%s", registryCreds.Username, registryCreds.Password)

	configContent := fmt.Sprintf(configTmpl, imageRegistry, base64.StdEncoding.EncodeToString([]byte(basicAuthCreds)))

	err = ioutil.WriteFile(filepath.Join(configDir, "config.json"), []byte(configContent), 0644)
	if err != nil {
		return errors.Wrap(err, "failed to create a docker config file")
	}

	return nil
}
