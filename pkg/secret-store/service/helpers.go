///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package service

import (
	"github.com/pkg/errors"
	dapi "github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/utils"
	"github.com/vmware/dispatch/pkg/utils/knaming"
	k8sv1 "k8s.io/api/core/v1"
)

//FromSecret converts from Dispatch secret to k8s secret
// mutates the original secret: use with caution
func FromSecret(secret *dapi.Secret) *k8sv1.Secret {
	if secret == nil {
		return nil
	}

	return &k8sv1.Secret{
		ObjectMeta: knaming.ToObjectMeta(secret.Meta, *secret),
		Data:       map[string][]byte{knaming.TheSecretKey: []byte(knaming.ToJSONBytes(secret.Secrets))},
	}
}

//ToSecret converts from k8s secret to dispatch secret
func ToSecret(k8sSecret *k8sv1.Secret) *dapi.Secret {
	if k8sSecret == nil {
		return nil
	}

	objMeta := &k8sSecret.ObjectMeta

	secretValue := dapi.SecretValue{}
	if err := knaming.FromJSONBytes(k8sSecret.Data[knaming.TheSecretKey], &secretValue); err != nil {
		// TODO the right thing
		panic(errors.Wrap(err, "parsing JSON from k8s secret"))
	}

	var secret dapi.Secret
	if err := knaming.FromJSONString(objMeta.Annotations[knaming.InitialObjectAnnotation], &secret); err != nil {
		// TODO the right thing
		panic(errors.Wrap(err, "decoding into secret"))
	}
	utils.AdjustMeta(&secret.Meta, dapi.Meta{CreatedTime: k8sSecret.CreationTimestamp.Unix()})

	secret.Secrets = secretValue

	return &secret
}
