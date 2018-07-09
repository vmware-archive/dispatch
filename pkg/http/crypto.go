///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package http

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path"
	"time"

	"github.com/pkg/errors"
)

const (
	certFilename = "dispatch_self_signed_cert.pem"
	keyFilename  = "dispatch_private_key.pem"
)

// GeneratePKI generates private key and a self-signed certificate, stores their encoded representation
// into files, and returns paths to those files, if successful. If privateKey or certificate are not empty strings,
// files for them have already been created, regardless of the err value.
// For development & testing purposes only, not suitable for production usage.
func GeneratePKI(hosts []string) (privateKey string, certificate string, err error) {
	if len(hosts) == 0 {
		return privateKey, certificate, errors.New("at least one host or IP must be provided")
	}

	workDir, err := os.Getwd()
	if err != nil {
		return privateKey, certificate, errors.Wrap(err, "error getting current working directory")
	}
	certPath := path.Join(workDir, certFilename)
	keyPath := path.Join(workDir, keyFilename)

	var certExists, keyExists bool
	if _, err = os.Stat(certPath); !os.IsNotExist(err) {
		certExists = true
	}
	if _, err = os.Stat(keyPath); !os.IsNotExist(err) {
		keyExists = true
	}

	if certExists && keyExists {
		// files already exist, let's return them and hope they are indeed certificates.
		return keyPath, certPath, nil
	}
	// if one of them exists and the other does not, play safe.
	if certExists {
		return keyPath, certPath, fmt.Errorf("certificate %s already exists, please clean up first", certPath)
	}
	if keyExists {
		return keyPath, certPath, fmt.Errorf("key %s already exists, please clean up first", keyPath)
	}

	startDate := time.Now()
	endDate := startDate.Add(365 * 24 * time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return privateKey, certificate, errors.Wrap(err, "unable to generate certificate serial number")
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Dispatch"},
		},
		NotBefore: startDate,
		NotAfter:  endDate,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return privateKey, certificate, errors.Wrap(err, "error creating certificate")
	}

	certOut, err := os.Create(certFilename)
	if err != nil {
		return "", "", errors.Wrapf(err, "error creating %s file", certFilename)
	}

	certificate = certPath

	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return privateKey, certificate, errors.Wrap(err, "error encoding certificate")
	}
	certOut.Close()

	keyOut, err := os.OpenFile(keyFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return privateKey, certificate, errors.Wrapf(err, "error creating %s file", keyFilename)
	}

	privateKey = keyPath

	b, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return privateKey, certificate, errors.Wrap(err, "error marshalling private key")
	}

	err = pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	if err != nil {
		return privateKey, certificate, errors.Wrap(err, "error encoding private key")
	}
	keyOut.Close()

	return privateKey, certificate, nil
}
