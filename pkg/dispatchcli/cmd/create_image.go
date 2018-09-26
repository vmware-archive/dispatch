///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	createImageLong = i18n.T(`Create dispatch image.`)

	// TODO: add examples
	createImageExample      = i18n.T(``)
	systemDependenciesFile  = i18n.T(``)
	runtimeDependenciesFile = i18n.T(``)
)

// NewCmdCreateImage creates command responsible for image creation.
func NewCmdCreateImage(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "image IMAGE_NAME BASE_IMAGE_URL",
		Short:   i18n.T("Create image"),
		Long:    createImageLong,
		Example: createImageExample,
		Args:    cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			c := imageManagerClient()
			err := createImage(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVar(&systemDependenciesFile, "system-deps", "", "path to file with system dependencies")
	cmd.Flags().StringVar(&runtimeDependenciesFile, "runtime-deps", "", "path to file with runtime dependencies")
	return cmd
}

// CallCreateImage makes the API call to create an image
func CallCreateImage(c client.ImagesClient) ModelAction {
	return func(i interface{}) error {
		imageModel := i.(*v1.Image)

		created, err := c.CreateImage(context.TODO(), dispatchConfig.Organization, imageModel)
		if err != nil {
			return err
		}
		*imageModel = *created
		return nil
	}
}

func createImage(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.ImagesClient) error {
	imageModel := &v1.Image{
		Meta: v1.Meta{
			Name: args[0],
		},
		BaseImageName: &args[1],
	}

	var systemDependencies v1.SystemDependencies
	if systemDependenciesFile != "" {
		fullPath := path.Join(workDir, systemDependenciesFile)
		b, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read system dependencies file: %s", err)
		}
		err = json.Unmarshal(b, &systemDependencies)
		if err != nil {
			return fmt.Errorf("failed to unmarshal system dependencies file: %s", err)
		}
	}
	var runtimeDependencies v1.RuntimeDependencies
	if runtimeDependenciesFile != "" {
		fullPath := path.Join(workDir, runtimeDependenciesFile)
		b, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read runtime dependencies file: %s", err)
		}
		runtimeDependencies.Manifest = string(b)
	}

	imageModel.SystemDependencies = &systemDependencies
	imageModel.RuntimeDependencies = &runtimeDependencies

	err := CallCreateImage(c)(imageModel)
	if err != nil {
		return err
	}
	if w, err := formatOutput(out, false, imageModel); w {
		return err
	}
	fmt.Fprintf(out, "Created image: %s\n", imageModel.Name)
	return nil
}
