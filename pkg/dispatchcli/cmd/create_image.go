///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/image-manager/gen/client/image"
	models "github.com/vmware/dispatch/pkg/image-manager/gen/models"
)

var (
	createImageLong = i18n.T(`Create dispatch image.`)

	// TODO: add examples
	createImageExample = i18n.T(``)
)

// NewCmdCreateImage creates command responsible for image creation.
func NewCmdCreateImage(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "image IMAGE_NAME BASE_IMAGE_NAME",
		Short:   i18n.T("Create image"),
		Long:    createImageLong,
		Example: createImageExample,
		Args:    cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			err := createImage(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "associate with an application")
	return cmd
}

// CallCreateImage makes the API call to create an image
func CallCreateImage(i interface{}) error {
	client := imageManagerClient()
	body := i.(*models.Image)
	params := &image.AddImageParams{
		Body:    body,
		Context: context.Background(),
	}

	created, err := client.Image.AddImage(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	*body = *created.Payload
	return nil
}

func createImage(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	body := &models.Image{
		Name:          &args[0],
		BaseImageName: &args[1],
		Tags:          models.ImageTags{},
	}
	if cmdFlagApplication != "" {
		body.Tags = append(body.Tags, &models.Tag{
			Key:   "Application",
			Value: cmdFlagApplication,
		})
	}

	err := CallCreateImage(body)
	if err != nil {
		return err
	}
	if dispatchConfig.Json {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(*body)
	}
	fmt.Printf("created image: %s\n", *body.Name)
	return nil
}
