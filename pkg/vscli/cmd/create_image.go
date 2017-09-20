///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"errors"
	"io"

	"github.com/spf13/cobra"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
)

var (
	createImageLong = i18n.T(`Create serverless image.`)

	// TODO: add examples
	createImageExample = i18n.T(``)
)

// NewCmdCreateImage creates command responsible for image creation.
func NewCmdCreateImage(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "image [--name IMAGE_NAME] [--status STATUS] BASE_IMAGE_NAME",
		Short:   i18n.T("Create image"),
		Long:    createImageLong,
		Example: createImageExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := createImage(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	return cmd
}

func createImage(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	return errors.New("Not implemented")
}
