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

// IsDir determines if path is a directory
func IsDir(path string) (bool, error) {
	f, err := os.Stat(path)
	if err != nil {
		return false, errors.Wrapf(err, "failed to determine if path is dir: '%s'", path)
	}
	return f.IsDir(), nil
}

// Tar writes the tar stream of the source to w.
func Tar(source string, w io.Writer) error {
	source = filepath.Clean(source)
	prefix := source + "/"

	isDir, err := IsDir(source)
	if err != nil {
		return err
	}
	if !isDir {
		prefix = filepath.Dir(source)
	}

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

			header.Name = "./" + strings.TrimPrefix(path, prefix)
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

// TarGzBytes produces a tar.gz of source path and returns its byte content
func TarGzBytes(source string) ([]byte, error) {
	bs := &bytes.Buffer{}
	gw := gzip.NewWriter(bs)
	err := Tar(source, gw)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to tar source '%s'", source)
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
