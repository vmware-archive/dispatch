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

	baseimage "gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/client/base_image"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
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
		Args:    cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			err := getBaseImages(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	return cmd
}

func getBaseImages(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	client := imageManagerClient()
	params := &baseimage.GetBaseImagesParams{
		Context: context.Background(),
	}
	images, err := client.BaseImage.GetBaseImages(params)
	if err != nil {
		fmt.Println("list base images returned an error")
		return err
	}
	data := [][]string{}
	for _, image := range images.Payload {
		data = append(data, []string{*image.Name, *image.DockerURL, string(image.Status), time.Unix(image.CreatedTime, 0).Local().Format(time.UnixDate)})
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "URL", "Status", "Created Date"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	table.AppendBulk(data)
	table.Render()
	return nil
}
