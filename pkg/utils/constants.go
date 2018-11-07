///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package utils

// APIKind a constant representing the kind of the API model
const APIKind = "API"

// ApplicationKind a constant to represent the kind of the Application model
const ApplicationKind = "Application"

// DriverKind a constant representing the kind of the Driver API model
const DriverKind = "Driver"

// DriverTypeKind a constant representing the kind of the DriverType API model
const DriverTypeKind = "DriverType"

// SubscriptionKind a constant representing the kind of the Subscription API model
const SubscriptionKind = "Subscription"

// FunctionKind a constant representing the kind of the Function model
const FunctionKind = "Function"

// ImageKind a constant representing the kind of the Image model
const ImageKind = "Image"

// BaseImageKind a constant representing the kind of the Base Image model
const BaseImageKind = "BaseImage"

// SecretKind a constant representing the kind of the Secret model
const SecretKind = "Secret"

// PolicyKind a constant representing the kind of the Policy model
const PolicyKind = "Policy"

// ServiceClassKind a constant representing the kind of the Service Class model
const ServiceClassKind = "ServiceClass"

// ServicePlanKind a constant representing the kind of the Service Plan model
const ServicePlanKind = "ServicePlan"

// ServiceInstanceKind a constant representing the kind of the Service Instance model
const ServiceInstanceKind = "ServiceInstance"

// ServiceBindingKind a constant representing the kind of the Service Binding model
const ServiceBindingKind = "ServiceBinding"

// ServiceAccountKind a constant representing the kind of the ServiceAccount Model
const ServiceAccountKind = "ServiceAccount"

// OrganizationKind a constant representing the kind of the Organization Model
const OrganizationKind = "Organization"

// LetsEncryptStaging a constant representing Let's Encrypt staging server url
const LetsEncryptStaging = "https://acme-staging.api.letsencrypt.org/directory"

// LetsEncryptProduction a constant representing Let's Encrypt production server url
const LetsEncryptProduction = "https://acme-v01.api.letsencrypt.org/directory"

// KeyLength a constant representing the length of TLS Key
const KeyLength = 2048

// KeyType a constant representing the length of TLS Key Type
const KeyType = "RSA PRIVATE KEY"

// CertType a constant representing the type of certificate
const CertType = "CERTIFICATE"
