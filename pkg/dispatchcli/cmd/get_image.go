///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"io"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/dispatchcli/cmd/utils"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	image "github.com/vmware/dispatch/pkg/image-manager/gen/client/image"
)

var (
	getImagesLong = i18n.T(`Get images.`)

	// TODO: add examples
	getImagesExample = i18n.T(``)
)

// NewCmdGetImage creates command responsible for getting images.
func NewCmdGetImage(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "image [IMAGE]",
		Short:   i18n.T("Get images"),
		Long:    getImagesLong,
		Example: getImagesExample,
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"images"},
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if len(args) > 0 {
				err = getImage(out, errOut, cmd, args)
			} else {
				err = getImages(out, errOut, cmd)
			}
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	return cmd
}

func getImage(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	client := imageManagerClient()
	params := &image.GetImageByNameParams{
		Context:   context.Background(),
		ImageName: args[0],
		Tags:      []string{},
	}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	resp, err := client.Image.GetImageByName(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatImageOutput(out, false, []*v1.Image{resp.Payload})
}

func getImages(out, errOut io.Writer, cmd *cobra.Command) error {
	client := imageManagerClient()
	params := &image.GetImagesParams{
		Context: context.Background(),
		Tags:    []string{},
	}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	resp, err := client.Image.GetImages(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatImageOutput(out, true, resp.Payload)
}

func formatImageOutput(out io.Writer, list bool, images []*v1.Image) error {
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(images)
		}
		return encoder.Encode(images[0])
	}
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Name", "URL", "BaseImage", "Status", "Created Date"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	for _, image := range images {
		table.Append([]string{*image.Name, image.DockerURL, *image.BaseImageName, string(image.Status), time.Unix(image.CreatedTime, 0).Local().Format(time.UnixDate)})
	}
	table.Render()
	return nil
}
