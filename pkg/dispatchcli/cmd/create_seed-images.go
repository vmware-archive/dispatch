///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

const createSeedImagesLong = `Create base-images and images to quick-start with the current version of Dispatch`

var (
	outputFile string
	imagesB64  string
)

// NewCmdCreateSeedImages creates command responsible for creation of seed images and base-images.
func NewCmdCreateSeedImages(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "seed-images",
		Short:   i18n.T("Create seed base-images and images"),
		Long:    createSeedImagesLong,
		Args:    cobra.ExactArgs(0),
		Aliases: []string{"seed"},
		Run: func(cmd *cobra.Command, args []string) {
			err := createSeedImages(out)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&outputFile, "output-file", "O", "", "seed images YAML gets written to this file (nothing gets created)")
	return cmd
}

func createSeedImages(out io.Writer) error {
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

	if outputFile != "" {
		err := ioutil.WriteFile(outputFile, bs.Bytes(), 0644)
		return errors.Wrapf(err, "error writing images YAML to '%s'", outputFile)
	}

	initCreateMap()
	return importBytes(out, bs.Bytes(), createMap, "Created")
}
