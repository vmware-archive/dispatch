///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/net/context"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	image "gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/client/image"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
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
		Args:    cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			err := getImages(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	return cmd
}

func getImages(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	client := imageManagerClient()
	params := &image.GetImagesParams{
		Context: context.Background(),
	}
	images, err := client.Image.GetImages(params)
	if err != nil {
		fmt.Println("list images returned an error")
		return err
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "URL", "BaseImage", "Status", "Created Date"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	for _, image := range images.Payload {
		table.Append([]string{*image.Name, image.DockerURL, *image.BaseImageName, string(image.Status), time.Unix(image.CreatedTime, 0).Local().Format(time.UnixDate)})
	}
	table.Render()
	return nil
}
