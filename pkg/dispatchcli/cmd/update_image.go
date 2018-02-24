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
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/image-manager/gen/client/image"
	models "github.com/vmware/dispatch/pkg/image-manager/gen/models"
)

var (
	updateImageLong    = ""
	updateImageExample = ""

	baseImageName string
)

// CallUpdateImage makes the service call to update an image.
func CallUpdateImage(input interface{}) error {
	img := input.(*models.Image)
	params := image.NewUpdateImageByNameParams()
	params.ImageName = *img.Name
	params.Body = img
	_, err := imageManagerClient().Image.UpdateImageByName(params, GetAuthInfoWriter())

	if err != nil {
		return formatAPIError(err, params)
	}

	return nil
}

// NewCmdUpdateImage creates command responsible for updating an image.
func NewCmdUpdateImage(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "image IMAGE_NAME",
		Short:   i18n.T("Update image"),
		Long:    updateImageLong,
		Example: updateImageExample,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := updateImage(out, errOut, cmd, args)
			CheckErr(err)
		},
	}

	cmd.Flags().StringVar(&baseImageName, "base-image-name", "", "base image to base this image on")
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "associate with an application")
	cmd.Flags().StringVar(&systemDependenciesFile, "system-deps", "", "path to file with system dependencies")
	cmd.Flags().StringVar(&runtimeDependenciesFile, "runtime-deps", "", "path to file with runtime dependencies")
	return cmd
}

func updateImage(out io.Writer, errOut io.Writer, cmd *cobra.Command, args []string) error {
	imageName := args[0]

	params := image.NewGetImageByNameParams()
	params.ImageName = imageName

	imageResp, err := imageManagerClient().Image.GetImageByName(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	img := *imageResp.Payload

	changed := false
	if cmd.Flags().Changed("base-image-name") {
		img.BaseImageName = &baseImageName
		changed = true
	}

	if cmd.Flags().Changed("system-deps") {
		var systemDependencies models.SystemDependencies
		if systemDependenciesFile != "" {
			fullPath := path.Join(workDir, systemDependenciesFile)
			b, err := ioutil.ReadFile(fullPath)
			if err != nil {
				return fmt.Errorf("Failed to read system dependencies file: %s", err)
			}
			err = json.Unmarshal(b, &systemDependencies)
			if err != nil {
				return fmt.Errorf("Failed to unmarshal system dependencies file: %s", err)
			}
		}

		img.SystemDependencies = &systemDependencies
		changed = true
	}

	if cmd.Flags().Changed("runtime-deps") {
		var runtimeDependencies models.RuntimeDependencies
		if runtimeDependenciesFile != "" {
			fullpath := path.Join(workDir, runtimeDependenciesFile)
			b, err := ioutil.ReadFile(fullpath)
			if err != nil {
				return fmt.Errorf("Failed to read runtime dependencies file: %s", err)
			}
			runtimeDependencies.Manifest = string(b)
		}

		img.RuntimeDependencies = &runtimeDependencies
		changed = true
	}

	if !changed {
		fmt.Fprintf(out, "No changes made")
		return nil
	}

	CallUpdateImage(&img)

	fmt.Fprintf(out, "Updated image: %s\n", imageName)
	return nil
}
