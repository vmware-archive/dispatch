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
	createBaseImageLong = i18n.T(`Create base image.`)

	// TODO: add examples
	createBaseImageExample = i18n.T(``)
)

// NewCmdCreateBaseImage creates command responsible for base image creation.
func NewCmdCreateBaseImage(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "base-image [--name IMAGE_NAME] URL",
		Short:   i18n.T("Create base image"),
		Long:    createBaseImageLong,
		Example: createBaseImageExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := createBaseImage(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	return cmd
}

func createBaseImage(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	return errors.New("Not implemented")
}
