///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"io"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
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
			c := imageManagerClient()
			if len(args) > 0 {
				err = getImage(out, errOut, cmd, args, c)
			} else {
				err = getImages(out, errOut, cmd, c)
			}
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	return cmd
}

func getImage(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.ImagesClient) error {
	imageName := args[0]

	resp, err := c.GetImage(context.TODO(), dispatchConfig.Organization, imageName)
	if err != nil {
		return err
	}
	return formatImageOutput(out, false, []v1.Image{*resp})
}

func getImages(out, errOut io.Writer, cmd *cobra.Command, c client.ImagesClient) error {
	resp, err := c.ListImages(context.TODO(), dispatchConfig.Organization)
	if err != nil {
		return err
	}
	return formatImageOutput(out, true, resp)
}

func formatImageOutput(out io.Writer, list bool, images []v1.Image) error {
	if w, err := formatOutput(out, list, images); w {
		return err
	}
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Name", "Destination", "BaseImage", "Status", "Created Date"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	for _, image := range images {
		table.Append([]string{*image.Name, image.ImageDestination, *image.BaseImageName, string(image.Status), time.Unix(image.CreatedTime, 0).Local().Format(time.UnixDate)})
	}
	table.Render()
	return nil
}
