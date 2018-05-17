///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"io"
	"time"

	"golang.org/x/net/context"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	getBaseImagesLong = i18n.T(`Get base images.`)

	// TODO: add examples
	getBaseImagesExample = i18n.T(``)
)

// NewCmdGetBaseImage creates command responsible for getting base images.
func NewCmdGetBaseImage(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "base-image [IMAGE_NAME]",
		Short:   i18n.T("Get base images"),
		Long:    getBaseImagesLong,
		Example: getBaseImagesExample,
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"base-images"},
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			c := imageManagerClient()
			if len(args) > 0 {
				err = getBaseImage(out, errOut, cmd, args, c)
			} else {
				err = getBaseImages(out, errOut, cmd, c)
			}
			CheckErr(err)
		},
	}
	return cmd
}

func getBaseImage(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.ImagesClient) error {
	baseImageName := args[0]
	resp, err := c.GetBaseImage(context.TODO(), dispatchConfig.Organization, baseImageName)
	if err != nil {
		return formatAPIError(err, baseImageName)
	}
	return formatBaseImageOutput(out, false, []v1.BaseImage{*resp})
}

func getBaseImages(out, errOut io.Writer, cmd *cobra.Command, c client.ImagesClient) error {
	resp, err := c.ListBaseImages(context.TODO(), dispatchConfig.Organization)
	if err != nil {
		return formatAPIError(err, nil)
	}
	return formatBaseImageOutput(out, true, resp)
}

func formatBaseImageOutput(out io.Writer, list bool, images []v1.BaseImage) error {
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(images)
		}
		return encoder.Encode(images[0])
	}
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Name", "URL", "Status", "Created Date"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	for _, image := range images {
		table.Append([]string{*image.Name, *image.DockerURL, string(image.Status), time.Unix(image.CreatedTime, 0).Local().Format(time.UnixDate)})
	}
	table.Render()
	return nil
}
