///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package utils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// TarDir writes the tar stream of the source dir to w
func TarDir(source string, w io.Writer) error {
	source = filepath.Clean(source) + "/"
	tarball := tar.NewWriter(w)
	return filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			header.Name = "./" + strings.TrimPrefix(path, source)
			log.Debugf("tar: writing header: %s", header.Name)

			if err := tarball.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarball, file)
			return err
		})
}

// TarGzBytes produces a tar.gz of sourceDir and returns its byte content
func TarGzBytes(sourceDir string) ([]byte, error) {
	bs := &bytes.Buffer{}
	gw := gzip.NewWriter(bs)
	err := TarDir(sourceDir, gw)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to tar dir '%s'", sourceDir)
	}
	if err := gw.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to close gzip writer")
	}
	return bs.Bytes(), nil
}

// Untar the tar stream r into the dst dir stripping prefix from file paths
func Untar(dst, prefix string, r io.Reader) error {
	tarReader := tar.NewReader(r)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(dst, strings.TrimPrefix(header.Name, prefix))
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}
