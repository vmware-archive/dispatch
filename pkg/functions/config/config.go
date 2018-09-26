///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package config

// NO TESTS

// StorageType is a storage label
type StorageType string

// Minio is the label for the minio storage type (s3)
const Minio StorageType = "minio"

// File is the label for the file storage type (NFS)
const File StorageType = "file"

// StorageFileConfig contains all of the file storage settings
type StorageFileConfig struct {
	SourceRootPath string
}

// MinioLocation is the minio "region"
type MinioLocation string

// DefaultMinioLocation is the default minio "region"
const DefaultMinioLocation MinioLocation = "us-east-1"

// StorageMinioConfig contains all of the minio storage settings
type StorageMinioConfig struct {
	Username     string
	Password     string
	MinioAddress string
	Location     MinioLocation
}

// StorageConfig contains all of the functions storage settings
type StorageConfig struct {
	Storage StorageType
	File    *StorageFileConfig
	Minio   *StorageMinioConfig
}
