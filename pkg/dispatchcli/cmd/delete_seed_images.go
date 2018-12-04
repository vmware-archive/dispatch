///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	deleteSeedImagesLong = i18n.T(`Delete base images.`)

	// TODO: add examples
	deleteSeedImagesExample = i18n.T(``)
)

// NewCmdDeleteSeedImage creates command responsible for deleting base images.
func NewCmdDeleteSeedImage(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "seed-images",
		Short:   i18n.T("Delete base images"),
		Long:    deleteSeedImagesLong,
		Example: deleteSeedImagesExample,
		Args:    cobra.ExactArgs(0),
		Aliases: []string{"seed"},
		Run: func(cmd *cobra.Command, args []string) {
			c := imageManagerClient()
			err := deleteSeedImage(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}
	return cmd
}

func deleteSeedImage(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.ImagesClient) error {
	if imagesB64 == "" {
		return errors.New("embedded images YAML is empty")
	}

	sr := strings.NewReader(imagesB64)
	br := base64.NewDecoder(base64.StdEncoding, sr)
	gr, err := gzip.NewReader(br)
	if err != nil {
		return errors.Wrap(err, "error creating a gzip reader for embedded images YAML")
	}
	bs := &bytes.Buffer{}
	_, err = bs.ReadFrom(gr)
	if err != nil {
		return errors.Wrap(err, "error reading embedded images YAML: error reading from gzip reader")
	}

	initDeleteMap()
	return importBytes(out, bs.Bytes(), deleteMap, "Deleted")
}
