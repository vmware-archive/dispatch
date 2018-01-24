///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/image-manager/gen/client/base_image"
	"github.com/vmware/dispatch/pkg/image-manager/gen/models"
)

var (
	updateBaseImageLong    = "Updates a base image from a given json representation"
	updateBaseImageExample = `{
		"createdTime": -62167218539,
		"dockerUrl": "vmware/dispatch-openfaas-nodejs6-base:0.0.3-dev1",
		"groups": null,
		"id": "e4b248cf-acfc-40cb-88e3-6cfaa0b7e9d8",
		"language": "nodejs6",
		"name": "nodejs6-base",
		"public": true,
		"reason": null,
		"status": "READY",
		"tags": [
			{
				"key": "role",
				"value": "test"
			}
		]
	}`
)

func NewCmdUpdateBaseImage(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "base-image BASE_IMAGE_FILE",
		Short:   i18n.T("Update base image"),
		Long:    updateBaseImageLong,
		Example: updateBaseImageExample,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := updateBaseImage(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	return cmd
}

func updateBaseImage(out io.Writer, errOut io.Writer, cmd *cobra.Command, args []string) error {
	baseImagePath := args[0]
	var baseImage = models.BaseImage{}

	if baseImagePath != "" {
		baseImageContent, err := ioutil.ReadFile(baseImagePath)
		if err != nil {
			message := fmt.Sprintf("Error when reading content of %s", baseImagePath)
			return formatCliError(err, message)
		}
		if err := json.Unmarshal(baseImageContent, &baseImage); err != nil {
			message := fmt.Sprintf("Error when parsing JSON from %s with error %s", baseImageContent, err)
			return formatCliError(err, message)
		}
	}

	params := base_image.NewUpdateBaseImageByNameParams()
	params.BaseImageName = *baseImage.Name
	params.Body = &baseImage
	_, err := imageManagerClient().BaseImage.UpdateBaseImageByName(params, GetAuthInfoWriter())

	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Updated base image: %s\n", *baseImage.Name)
	return nil
}
