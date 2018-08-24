/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package servicecatalog

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterServiceBroker represents an entity that provides ClusterServiceClasses for use in the
// service catalog. ClusterServiceBroker is backed by an OSBAPI v2 broker supporting the
// latest minor version of the v2 major version.
type ClusterServiceBroker struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   ClusterServiceBrokerSpec
	Status ClusterServiceBrokerStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterServiceBrokerList is a list of Brokers.
type ClusterServiceBrokerList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []ClusterServiceBroker
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceBroker represents an entity that provides ServiceClasses for use in the
// service catalog. ServiceBroker is backed by an OSBAPI v2 broker supporting the
// latest minor version of the v2 major version.
type ServiceBroker struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   ServiceBrokerSpec
	Status ServiceBrokerStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceBrokerList is a list of Brokers.
type ServiceBrokerList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []ServiceBroker
}

// CommonServiceBrokerSpec represents a description of a Broker.
type CommonServiceBrokerSpec struct {
	// URL is the address used to communicate with the ServiceBroker.
	URL string

	// InsecureSkipTLSVerify disables TLS certificate verification when communicating with this Broker.
	// This is strongly discouraged.  You should use the CABundle instead.
	// +optional
	InsecureSkipTLSVerify bool

	// CABundle is a PEM encoded CA bundle which will be used to validate a Broker's serving certificate.
	// +optional
	CABundle []byte

	// RelistBehavior specifies the type of relist behavior the catalog should
	// exhibit when relisting ServiceClasses available from a broker.
	RelistBehavior ServiceBrokerRelistBehavior

	// RelistDuration is the frequency by which a controller will relist the
	// broker when the RelistBehavior is set to ServiceBrokerRelistBehaviorDuration.
	// Users are cautioned against configuring low values for the RelistDuration,
	// as this can easily overload the controller manager in an environment with
	// many brokers. The actual interval is intrinsically governed by the
	// configured resync interval of the controller, which acts as a minimum bound.
	// For example, with a resync interval of 5m and a RelistDuration of 2m, relists
	// will occur at the resync interval of 5m.
	RelistDuration *metav1.Duration

	// RelistRequests is a strictly increasing, non-negative integer counter that
	// can be manually incremented by a user to manually trigger a relist.
	RelistRequests int64

	// CatalogRestrictions is a set of restrictions on which of a broker's services
	// and plans have resources created for them.
	CatalogRestrictions *CatalogRestrictions
}

// CatalogRestrictions is a set of restrictions on which of a broker's services
// and plans have resources created for them.
//
// Some examples of this object are as follows:
//
// This is an example of a whitelist on service externalName.
// Goal: Only list Services with the externalName of FooService and BarService,
// Solution: restrictions := ServiceCatalogRestrictions{
// 		ServiceClass: ["externalName in (FooService, BarService)"]
// }
//
// This is an example of a blacklist on service externalName.
// Goal: Allow all services except the ones with the externalName of FooService and BarService,
// Solution: restrictions := ServiceCatalogRestrictions{
// 		ServiceClass: ["externalName notin (FooService, BarService)"]
// }
//
// This whitelists plans called "Demo", and blacklists (but only a single element in
// the list) a service and a plan.
// Goal: Allow all plans with the externalName demo, but not AABBCC, and not a specific service by name,
// Solution: restrictions := ServiceCatalogRestrictions{
// 		ServiceClass: ["name!=AABBB-CCDD-EEGG-HIJK"]
// 		ServicePlan: ["externalName in (Demo)", "name!=AABBCC"]
// }
//
// CatalogRestrictions strings have a special format similar to Label Selectors,
// except the catalog supports only a very specific property set.
//
// The predicate format is expected to be `<property><conditional><requirement>`
// Check the *Requirements type definition for which <property> strings will be allowed.
// <conditional> is allowed to be one of the following: ==, !=, in, notin
// <requirement> will be a string value if `==` or `!=` are used.
// <requirement> will be a set of string values if `in` or `notin` are used.
// Multiple predicates are allowed to be chained with a comma (,)
//
// ServiceClass allowed property names:
//   name - the value set to [Cluster]ServiceClass.Name
//   spec.externalName - the value set to [Cluster]ServiceClass.Spec.ExternalName
//   spec.externalID - the value set to [Cluster]ServiceClass.Spec.ExternalID
// ServicePlan allowed property names:
//   name - the value set to [Cluster]ServicePlan.Name
//   spec.externalName - the value set to [Cluster]ServicePlan.Spec.ExternalName
//   spec.externalID - the value set to [Cluster]ServicePlan.Spec.ExternalID
//   spec.free - the value set to [Cluster]ServicePlan.Spec.Free
//   spec.serviceClassName - the value set to ServicePlan.Spec.ServiceClassRef.Name
//   spec.clusterServiceClass.name - the value set to ClusterServicePlan.Spec.ClusterServiceClassRef.Name
type CatalogRestrictions struct {
	// ServiceClass represents a selector for plans, used to filter catalog re-lists.
	ServicePlan []string
	// ServicePlan represents a selector for classes, used to filter catalog re-lists.
	ServiceClass []string
}

// ClusterServiceBrokerSpec represents a description of a Broker.
type ClusterServiceBrokerSpec struct {
	CommonServiceBrokerSpec

	// AuthInfo contains the data that the service catalog should use to authenticate
	// with the Service Broker.
	AuthInfo *ClusterServiceBrokerAuthInfo
}

// ServiceBrokerSpec represents a description of a Broker.
type ServiceBrokerSpec struct {
	CommonServiceBrokerSpec

	// AuthInfo contains the data that the service catalog should use to authenticate
	// with the Service Broker.
	AuthInfo *ServiceBrokerAuthInfo
}

// ServiceBrokerRelistBehavior represents a type of broker relist behavior.
type ServiceBrokerRelistBehavior string

const (
	// ServiceBrokerRelistBehaviorDuration indicates that the broker will be
	// relisted automatically after the specified duration has passed.
	ServiceBrokerRelistBehaviorDuration ServiceBrokerRelistBehavior = "Duration"

	// ServiceBrokerRelistBehaviorManual indicates that the broker is only
	// relisted when the spec of the broker changes.
	ServiceBrokerRelistBehaviorManual ServiceBrokerRelistBehavior = "Manual"
)

// ClusterServiceBrokerAuthInfo is a union type that contains information on
// one of the authentication methods the the service catalog and brokers may
// support, according to the OpenServiceBroker API specification
// (https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md).
type ClusterServiceBrokerAuthInfo struct {
	// ClusterBasicAuthConfig provides configuration for basic authentication.
	Basic *ClusterBasicAuthConfig
	// ClusterBearerTokenAuthConfig provides configuration to send an opaque value as a bearer token.
	// The value is referenced from the 'token' field of the given secret.  This value should only
	// contain the token value and not the `Bearer` scheme.
	Bearer *ClusterBearerTokenAuthConfig
}

// ClusterBasicAuthConfig provides config for the basic authentication of
// cluster scoped brokers.
type ClusterBasicAuthConfig struct {
	// SecretRef is a reference to a Secret containing information the
	// catalog should use to authenticate to this ClusterServiceBroker.
	//
	// Required at least one of the fields:
	// - Secret.Data["username"] - username used for authentication
	// - Secret.Data["password"] - password or token needed for authentication
	SecretRef *ObjectReference
}

// ClusterBearerTokenAuthConfig provides config for the bearer token
// authentication of cluster scoped brokers.
type ClusterBearerTokenAuthConfig struct {
	// SecretRef is a reference to a Secret containing information the
	// catalog should use to authenticate to this ClusterServiceBroker.
	//
	// Required field:
	// - Secret.Data["token"] - bearer token for authentication
	SecretRef *ObjectReference
}

// ServiceBrokerAuthInfo is a union type that contains information on
// one of the authentication methods the the service catalog and brokers may
// support, according to the OpenServiceBroker API specification
// (https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md).
type ServiceBrokerAuthInfo struct {
	// BasicAuthConfig provides configuration for basic authentication.
	Basic *BasicAuthConfig
	// BearerTokenAuthConfig provides configuration to send an opaque value as a bearer token.
	// The value is referenced from the 'token' field of the given secret.  This value should only
	// contain the token value and not the `Bearer` scheme.
	Bearer *BearerTokenAuthConfig
}

// BasicAuthConfig provides config for the basic authentication of
// cluster scoped brokers.
type BasicAuthConfig struct {
	// SecretRef is a reference to a Secret containing information the
	// catalog should use to authenticate to this ServiceBroker.
	//
	// Required at least one of the fields:
	// - Secret.Data["username"] - username used for authentication
	// - Secret.Data["password"] - password or token needed for authentication
	SecretRef *LocalObjectReference
}

// BearerTokenAuthConfig provides config for the bearer token
// authentication of cluster scoped brokers.
type BearerTokenAuthConfig struct {
	// SecretRef is a reference to a Secret containing information the
	// catalog should use to authenticate to this ServiceBroker.
	//
	// Required field:
	// - Secret.Data["token"] - bearer token for authentication
	SecretRef *LocalObjectReference
}

const (
	// BasicAuthUsernameKey is the key of the username for SecretTypeBasicAuth secrets
	BasicAuthUsernameKey = "username"
	// BasicAuthPasswordKey is the key of the password or token for SecretTypeBasicAuth secrets
	BasicAuthPasswordKey = "password"

	// BearerTokenKey is the key of the bearer token for SecretTypeBearerTokenAuth secrets
	BearerTokenKey = "token"
)

// CommonServiceBrokerStatus represents the current status of a ServiceBroker.
type CommonServiceBrokerStatus struct {
	Conditions []ServiceBrokerCondition

	// ReconciledGeneration is the 'Generation' of the ServiceBrokerSpec that
	// was last processed by the controller. The reconciled generation is updated
	// even if the controller failed to process the spec.
	ReconciledGeneration int64

	// OperationStartTime is the time at which the current operation began.
	OperationStartTime *metav1.Time

	// LastCatalogRetrievalTime is the time the Catalog was last fetched from
	// the Service Broker
	LastCatalogRetrievalTime *metav1.Time
}

// ClusterServiceBrokerStatus represents the current status of a
// ClusterServiceBroker.
type ClusterServiceBrokerStatus struct {
	CommonServiceBrokerStatus
}

// ServiceBrokerStatus represents the current status of a ServiceBroker.
type ServiceBrokerStatus struct {
	CommonServiceBrokerStatus
}

// ServiceBrokerCondition contains condition information for a Service Broker.
type ServiceBrokerCondition struct {
	// Type of the condition, currently ('Ready').
	Type ServiceBrokerConditionType

	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus

	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	LastTransitionTime metav1.Time

	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	Reason string

	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	Message string
}

// ServiceBrokerConditionType represents a broker condition value.
type ServiceBrokerConditionType string

const (
	// ServiceBrokerConditionReady represents the fact that a given broker condition
	// is in ready state.
	ServiceBrokerConditionReady ServiceBrokerConditionType = "Ready"

	// ServiceBrokerConditionFailed represents information about a final failure
	// that should not be retried.
	ServiceBrokerConditionFailed ServiceBrokerConditionType = "Failed"
)

// ConditionStatus represents a condition's status.
type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in
// the condition; "ConditionFalse" means a resource is not in the condition;
// "ConditionUnknown" means kubernetes can't decide if a resource is in the
// condition or not. In the future, we could add other intermediate
// conditions, e.g. ConditionDegraded.
const (
	// ConditionTrue represents the fact that a given condition is true
	ConditionTrue ConditionStatus = "True"

	// ConditionFalse represents the fact that a given condition is false
	ConditionFalse ConditionStatus = "False"

	// ConditionUnknown represents the fact that a given condition is unknown
	ConditionUnknown ConditionStatus = "Unknown"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterServiceClassList is a list of ClusterServiceClasses.
type ClusterServiceClassList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []ClusterServiceClass
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterServiceClass represents an offering in the service catalog.
type ClusterServiceClass struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   ClusterServiceClassSpec
	Status ClusterServiceClassStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceClassList is a list of ServiceClasses.
type ServiceClassList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []ServiceClass
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceClass represents a namespaced offering in the service catalog.
type ServiceClass struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   ServiceClassSpec
	Status ServiceClassStatus
}

// ServiceClassStatus represents status information about a
// ServiceClass.
type ServiceClassStatus struct {
	CommonServiceClassStatus
}

// ClusterServiceClassStatus represents status information about a
// ClusterServiceClass.
type ClusterServiceClassStatus struct {
	CommonServiceClassStatus
}

// CommonServiceClassStatus represents common status information between
// cluster scoped and namespace scoped ServiceClasses.
type CommonServiceClassStatus struct {
	// RemovedFromBrokerCatalog indicates that the broker removed the service from its
	// catalog.
	RemovedFromBrokerCatalog bool
}

// CommonServiceClassSpec represents details about a ServiceClass
type CommonServiceClassSpec struct {
	// ExternalName is the name of this object that the Service Broker
	// exposed this Service Class as. Mutable.
	ExternalName string

	// ExternalID is the identity of this object for use with the OSB API.
	//
	// Immutable.
	ExternalID string

	// Description is a short description of this ServiceClass.
	Description string

	// Bindable indicates whether a user can create bindings to an ServiceInstance
	// provisioned from this service. ServicePlan has an optional field called
	// Bindable which overrides the value of this field.
	Bindable bool

	// Currently, this field is ALPHA: it may change or disappear at any time
	// and its data will not be migrated.
	//
	// BindingRetrievable indicates whether fetching a binding via a GET on
	// its endpoint is supported for all plans.
	BindingRetrievable bool

	// PlanUpdatable indicates whether instances provisioned from this
	// ServiceClass may change ServicePlans after being provisioned.
	PlanUpdatable bool

	// ExternalMetadata is a blob of information about the ServiceClass, meant
	// to be user-facing content and display instructions.  This field may
	// contain platform-specific conventional values.
	ExternalMetadata *runtime.RawExtension

	// Currently, this field is ALPHA: it may change or disappear at any time
	// and its data will not be migrated.
	//
	// Tags is a list of strings that represent different classification
	// attributes of the ServiceClass.  These are used in Cloud Foundry in a
	// way similar to Kubernetes labels, but they currently have no special
	// meaning in Kubernetes.
	Tags []string

	// Currently, this field is ALPHA: it may change or disappear at any time
	// and its data will not be migrated.
	//
	// Requires exposes a list of Cloud Foundry-specific 'permissions'
	// that must be granted to an instance of this service within Cloud
	// Foundry.  These 'permissions' have no meaning within Kubernetes and an
	// ServiceInstance provisioned from this ServiceClass will not work correctly.
	Requires []string
}

// ClusterServiceClassSpec represents the details about a ClusterServiceClass.
type ClusterServiceClassSpec struct {
	CommonServiceClassSpec

	// ClusterServiceBrokerName is the reference to the ClusterServiceBroker that
	// provides this ClusterServiceClass.
	//
	// Immutable.
	ClusterServiceBrokerName string
}

// ServiceClassSpec represents the details about a ServiceClass.
type ServiceClassSpec struct {
	CommonServiceClassSpec

	// ServiceBrokerName is the reference to the ServiceBroker that provides this
	// ServiceClass.
	//
	// Immutable.
	ServiceBrokerName string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterServicePlanList is a list of ClusterServicePlans.
type ClusterServicePlanList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []ClusterServicePlan
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterServicePlan represents a tier of a ClusterServiceClass.
type ClusterServicePlan struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   ClusterServicePlanSpec
	Status ClusterServicePlanStatus
}

// CommonServicePlanSpec represents details about the ServicePlan
type CommonServicePlanSpec struct {
	// ExternalName is the name of this object that the Service Broker
	// exposed this Service Plan as. Mutable.
	ExternalName string

	// ExternalID is the identity of this object for use with the OSB API.
	//
	// Immutable.
	ExternalID string

	// Description is a short description of this ServicePlan.
	Description string

	// Bindable indicates whether a user can create bindings to an ServiceInstance
	// using this ServicePlan.  If set, overrides the value of the
	// corresponding ServiceClassSpec Bindable field.
	Bindable *bool

	// Free indicates whether this ServicePlan is available at no cost.
	Free bool

	// ExternalMetadata is a blob of information about the plan, meant to be
	// user-facing content and display instructions.  This field may contain
	// platform-specific conventional values.
	ExternalMetadata *runtime.RawExtension

	// Currently, this field is ALPHA: it may change or disappear at any time
	// and its data will not be migrated.
	//
	// ServiceInstanceCreateParameterSchema is the schema for the parameters
	// that may be supplied when provisioning a new ServiceInstance on this plan.
	ServiceInstanceCreateParameterSchema *runtime.RawExtension

	// Currently, this field is ALPHA: it may change or disappear at any time
	// and its data will not be migrated.
	//
	// ServiceInstanceUpdateParameterSchema is the schema for the parameters
	// that may be updated once an ServiceInstance has been provisioned on this plan.
	// This field only has meaning if the corresponding ServiceClassSpec is PlanUpdatable.
	ServiceInstanceUpdateParameterSchema *runtime.RawExtension

	// Currently, this field is ALPHA: it may change or disappear at any time
	// and its data will not be migrated.
	//
	// ServiceBindingCreateParameterSchema is the schema for the parameters that
	// may be supplied binding to an ServiceInstance on this plan.
	ServiceBindingCreateParameterSchema *runtime.RawExtension

	// Currently, this field is ALPHA: it may change or disappear at any time
	// and its data will not be migrated.when a bind operation stored in the Secret when binding to a ServiceInstance on this plan.
	//
	// ServiceBindingCreateResponseSchema is the schema for the response that
	// will be returned by the broker when binding to a ServiceInstance on this plan.
	// The schema also contains the sub-schema for the credentials part of the
	// broker's response, which allows clients to see what the credentials
	// will look like even before the binding operation is performed.
	ServiceBindingCreateResponseSchema *runtime.RawExtension
}

// ClusterServicePlanSpec represents details about the ClusterServicePlan
type ClusterServicePlanSpec struct {
	CommonServicePlanSpec

	// ClusterServiceBrokerName is the name of the ClusterServiceBroker that offers this
	// ClusterServicePlan.
	ClusterServiceBrokerName string

	// ClusterServiceClassRef is a reference to the service class that
	// owns this plan.
	ClusterServiceClassRef ClusterObjectReference
}

// ClusterServicePlanStatus represents status information about a
// ClusterServicePlan.
type ClusterServicePlanStatus struct {
	CommonServicePlanStatus
}

// CommonServicePlanStatus represents status information about a
// ClusterServicePlan or a ServicePlan.
type CommonServicePlanStatus struct {
	// RemovedFromBrokerCatalog indicates that the broker removed the plan
	// from its catalog.
	RemovedFromBrokerCatalog bool
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServicePlanList is a list of ServicePlans.
type ServicePlanList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []ServicePlan
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServicePlan represents a tier of a ServiceClass.
type ServicePlan struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   ServicePlanSpec
	Status ServicePlanStatus
}

// ServicePlanSpec represents details about the ServicePlan
type ServicePlanSpec struct {
	CommonServicePlanSpec

	// ServiceBrokerName is the name of the ServiceBroker that offers this
	// ServicePlan.
	ServiceBrokerName string

	// ServiceClassRef is a reference to the service class that
	// owns this plan.
	ServiceClassRef LocalObjectReference
}

// ServicePlanStatus represents status information about a
// ServicePlan.
type ServicePlanStatus struct {
	CommonServicePlanStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceInstanceList is a list of instances.
type ServiceInstanceList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []ServiceInstance
}

// UserInfo holds information about the user that last changed a resource's spec.
type UserInfo struct {
	Username string
	UID      string
	Groups   []string
	Extra    map[string]ExtraValue
}

// ExtraValue contains additional information about a user that may be
// provided by the authenticator.
type ExtraValue []string

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceInstance represents a provisioned instance of a ClusterServiceClass.
type ServiceInstance struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   ServiceInstanceSpec
	Status ServiceInstanceStatus
}

// PlanReference defines the user specification for the desired
// (Cluster)ServicePlan and (Cluster)ServiceClass. Because there are
// multiple ways to specify the desired Class/Plan, this structure specifies the
// allowed ways to specify the intent. Note: a user may specify either cluster
// scoped OR namespace scoped identifiers, but NOT both, as they are mutually
// exclusive.
//
// Currently supported ways:
//  - ClusterServiceClassExternalName and ClusterServicePlanExternalName
//  - ClusterServiceClassExternalID and ClusterServicePlanExternalID
//  - ClusterServiceClassName and ClusterServicePlanName
//  - ServiceClassExternalName and ServicePlanExternalName
//  - ServiceClassExternalID and ServicePlanExternalID
//  - ServiceClassName and ServicePlanName
//
// For any of these ways, if a ClusterServiceClass only has one plan
// then the corresponding service plan field is optional.
type PlanReference struct {
	// ClusterServiceClassExternalName is the human-readable name of the
	// service as reported by the ClusterServiceBroker. Note that if the
	// ClusterServiceBroker changes the name of the ClusterServiceClass,
	// it will not be reflected here, and to see the current name of the
	// ClusterServiceClass, you should follow the ClusterServiceClassRef below.
	//
	// Immutable.
	ClusterServiceClassExternalName string
	// ClusterServicePlanExternalName is the human-readable name of the plan
	// as reported by the ClusterServiceBroker. Note that if the
	// ClusterServiceBroker changes the name of the ClusterServicePlan, it will
	// not be reflected here, and to see the current name of the
	// ClusterServicePlan, you should follow the ClusterServicePlanRef below.
	ClusterServicePlanExternalName string

	// ClusterServiceClassExternalID is the ClusterServiceBroker's external id
	// for the class.
	//
	// Immutable.
	ClusterServiceClassExternalID string

	// ClusterServicePlanExternalID is the ClusterServiceBroker's external id for
	// the plan.
	ClusterServicePlanExternalID string

	// ClusterServiceClassName is the kubernetes name of the ClusterServiceClass.
	//
	// Immutable.
	ClusterServiceClassName string
	// ClusterServicePlanName is kubernetes name of the ClusterServicePlan.
	ClusterServicePlanName string

	// ServiceClassExternalName is the human-readable name of the
	// service as reported by the ServiceBroker. Note that if the ServiceBroker
	// changes the name of the ServiceClass, it will not be reflected here,
	// and to see the current name of the ServiceClass, you should
	// follow the ServiceClassRef below.
	//
	// Immutable.
	ServiceClassExternalName string
	// ServicePlanExternalName is the human-readable name of the plan
	// as reported by the ServiceBroker. Note that if the ServiceBroker changes
	// the name of the ServicePlan, it will not be reflected here, and to see
	// the current name of the ServicePlan, you should follow the
	// ServicePlanRef below.
	ServicePlanExternalName string

	// ServiceClassExternalID is the ServiceBroker's external id for the class.
	//
	// Immutable.
	ServiceClassExternalID string

	// ServicePlanExternalID is the ServiceBroker's external id for the plan.
	ServicePlanExternalID string

	// ServiceClassName is the kubernetes name of the ServiceClass.
	//
	// Immutable.
	ServiceClassName string
	// ServicePlanName is kubernetes name of the ServicePlan.
	ServicePlanName string
}

// ServiceInstanceSpec represents the desired state of an Instance.
type ServiceInstanceSpec struct {
	PlanReference

	// ClusterServiceClassRef is a reference to the ClusterServiceClass
	// that the user selected. This is set by the controller based on the
	// cluster-scoped values specified in the PlanReference.
	ClusterServiceClassRef *ClusterObjectReference
	// ClusterServicePlanRef is a reference to the ClusterServicePlan
	// that the user selected. This is set by the controller based on the
	// cluster-scoped values specified in the PlanReference.
	ClusterServicePlanRef *ClusterObjectReference

	// ServiceClassRef is a reference to the ServiceClass that the user selected.
	// This is set by the controller based on the namespace-scoped values
	// specified in the PlanReference.
	ServiceClassRef *LocalObjectReference
	// ServicePlanRef is a reference to the ServicePlan that the user selected.
	// This is set by the controller based on the namespace-scoped values
	// specified in the PlanReference.
	ServicePlanRef *LocalObjectReference

	// Parameters is a set of the parameters to be passed to the underlying
	// broker. The inline YAML/JSON payload to be translated into equivalent
	// JSON object. If a top-level parameter name exists in multiples sources
	// among `Parameters` and `ParametersFrom` fields, it is considered to be
	// a user error in the specification
	//
	// The Parameters field is NOT secret or secured in any way and should
	// NEVER be used to hold sensitive information. To set parameters that
	// contain secret information, you should ALWAYS store that information
	// in a Secret and use the ParametersFrom field.
	//
	// +optional
	Parameters *runtime.RawExtension

	// List of sources to populate parameters.
	// If a top-level parameter name exists in multiples sources among
	// `Parameters` and `ParametersFrom` fields, it is
	// considered to be a user error in the specification
	// +optional
	ParametersFrom []ParametersFromSource

	// ExternalID is the identity of this object for use with the OSB API.
	//
	// Immutable.
	ExternalID string

	// Currently, this field is ALPHA: it may change or disappear at any time
	// and its data will not be migrated.
	//
	// UserInfo contains information about the user that last modified this
	// instance. This field is set by the API server and not settable by the
	// end-user. User-provided values for this field are not saved.
	// +optional
	UserInfo *UserInfo

	// UpdateRequests is a strictly increasing, non-negative integer counter that
	// can be manually incremented by a user to manually trigger an update. This
	// allows for parameters to be updated with any out-of-band changes that have
	// been made to the secrets from which the parameters are sourced.
	UpdateRequests int64
}

// ServiceInstanceStatus represents the current status of an Instance.
type ServiceInstanceStatus struct {
	// Conditions is an array of ServiceInstanceConditions capturing aspects of an
	// ServiceInstance's status.
	Conditions []ServiceInstanceCondition

	// AsyncOpInProgress is set to true if there is an ongoing async operation
	// against this ServiceInstance in progress.
	AsyncOpInProgress bool

	// OrphanMitigationInProgress is set to true if there is an ongoing orphan
	// mitigation operation against this ServiceInstance in progress.
	OrphanMitigationInProgress bool

	// LastOperation is the string that the broker may have returned when
	// an async operation started, it should be sent back to the broker
	// on poll requests as a query param.
	LastOperation *string

	// DashboardURL is the URL of a web-based management user interface for
	// the service instance.
	DashboardURL *string

	// CurrentOperation is the operation the Controller is currently performing
	// on the ServiceInstance.
	CurrentOperation ServiceInstanceOperation

	// ReconciledGeneration is the 'Generation' of the serviceInstanceSpec that
	// was last processed by the controller. The reconciled generation is updated
	// even if the controller failed to process the spec.
	// Deprecated: use ObservedGeneration with conditions set to true to find
	// whether generation was reconciled.
	ReconciledGeneration int64

	// ObservedGeneration is the 'Generation' of the serviceInstanceSpec that
	// was last processed by the controller. The observed generation is updated
	// whenever the status is updated regardless of operation result.
	ObservedGeneration int64

	// OperationStartTime is the time at which the current operation began.
	OperationStartTime *metav1.Time

	// InProgressProperties is the properties state of the ServiceInstance when
	// a Provision, Update or Deprovision is in progress.
	InProgressProperties *ServiceInstancePropertiesState

	// ExternalProperties is the properties state of the ServiceInstance which the
	// broker knows about.
	ExternalProperties *ServiceInstancePropertiesState

	// ProvisionStatus describes whether the instance is in the provisioned state.
	ProvisionStatus ServiceInstanceProvisionStatus

	// DeprovisionStatus describes what has been done to deprovision the
	// ServiceInstance.
	DeprovisionStatus ServiceInstanceDeprovisionStatus
}

// ServiceInstanceCondition contains condition information about an Instance.
type ServiceInstanceCondition struct {
	// Type of the condition, currently ('Ready').
	Type ServiceInstanceConditionType

	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus

	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	LastTransitionTime metav1.Time

	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	Reason string

	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	Message string
}

// ServiceInstanceConditionType represents a instance condition value.
type ServiceInstanceConditionType string

const (
	// ServiceInstanceConditionReady represents that a given InstanceCondition is in
	// ready state.
	ServiceInstanceConditionReady ServiceInstanceConditionType = "Ready"

	// ServiceInstanceConditionFailed represents information about a final failure
	// that should not be retried.
	ServiceInstanceConditionFailed ServiceInstanceConditionType = "Failed"

	// ServiceInstanceConditionOrphanMitigation represents information about an
	// orphan mitigation that is required after failed provisioning.
	ServiceInstanceConditionOrphanMitigation ServiceInstanceConditionType = "OrphanMitigation"
)

// ServiceInstanceOperation represents a type of operation the controller can
// be performing for a service instance in the OSB API.
type ServiceInstanceOperation string

const (
	// ServiceInstanceOperationProvision indicates that the ServiceInstance is
	// being Provisioned.
	ServiceInstanceOperationProvision ServiceInstanceOperation = "Provision"
	// ServiceInstanceOperationUpdate indicates that the ServiceInstance is
	// being Updated.
	ServiceInstanceOperationUpdate ServiceInstanceOperation = "Update"
	// ServiceInstanceOperationDeprovision indicates that the ServiceInstance is
	// being Deprovisioned.
	ServiceInstanceOperationDeprovision ServiceInstanceOperation = "Deprovision"
)

// ServiceInstancePropertiesState is the state of a ServiceInstance that
// the ServiceBroker knows about.
type ServiceInstancePropertiesState struct {
	// ClusterServicePlanExternalName is the name of the plan that the broker knows this
	// ServiceInstance to be on. This is the human readable plan name from the
	// OSB API.
	ClusterServicePlanExternalName string

	// ClusterServicePlanExternalID is the external ID of the plan that the
	// broker knows this ServiceInstance to be on.
	ClusterServicePlanExternalID string

	// ServicePlanExternalName is the name of the plan that the broker knows this
	// ServiceInstance to be on. This is the human readable plan name from the
	// OSB API.
	ServicePlanExternalName string

	// ServicePlanExternalID is the external ID of the plan that the
	// broker knows this ServiceInstance to be on.
	ServicePlanExternalID string

	// Parameters is a blob of the parameters and their values that the broker
	// knows about for this ServiceInstance.  If a parameter was sourced from
	// a secret, its value will be "<redacted>" in this blob.
	Parameters *runtime.RawExtension

	// ParametersChecksum is the checksum of the parameters that were sent.
	ParametersChecksum string

	// UserInfo is information about the user that made the request.
	UserInfo *UserInfo
}

// ServiceInstanceDeprovisionStatus is the status of deprovisioning a
// ServiceInstance
type ServiceInstanceDeprovisionStatus string

const (
	// ServiceInstanceDeprovisionStatusNotRequired indicates that a provision
	// request has not been sent for the ServiceInstance, so no deprovision
	// request needs to be made.
	ServiceInstanceDeprovisionStatusNotRequired ServiceInstanceDeprovisionStatus = "NotRequired"
	// ServiceInstanceDeprovisionStatusRequired indicates that a provision
	// request has been sent for the ServiceInstance. A deprovision request
	// must be made before deleting the ServiceInstance.
	ServiceInstanceDeprovisionStatusRequired ServiceInstanceDeprovisionStatus = "Required"
	// ServiceInstanceDeprovisionStatusSucceeded indicates that a deprovision
	// request has been sent for the ServiceInstance, and the request was
	// successful.
	ServiceInstanceDeprovisionStatusSucceeded ServiceInstanceDeprovisionStatus = "Succeeded"
	// ServiceInstanceDeprovisionStatusFailed indicates that deprovision
	// requests have been sent for the ServiceInstance but they failed. The
	// controller has given up on sending more deprovision requests.
	ServiceInstanceDeprovisionStatusFailed ServiceInstanceDeprovisionStatus = "Failed"
)

// ServiceInstanceProvisionStatus is the status of provisioning a
// ServiceInstance
type ServiceInstanceProvisionStatus string

const (
	// ServiceInstanceProvisionStatusProvisioned indicates that the instance
	// was provisioned.
	ServiceInstanceProvisionStatusProvisioned ServiceInstanceProvisionStatus = "Provisioned"
	// ServiceInstanceProvisionStatusNotProvisioned indicates that the instance
	// was not ever provisioned or was deprovisioned.
	ServiceInstanceProvisionStatusNotProvisioned ServiceInstanceProvisionStatus = "NotProvisioned"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceBindingList is a list of ServiceBindings.
type ServiceBindingList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []ServiceBinding
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceBinding represents a "used by" relationship between an application and an
// ServiceInstance.
type ServiceBinding struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   ServiceBindingSpec
	Status ServiceBindingStatus
}

// ServiceBindingSpec represents the desired state of a
// ServiceBinding.
//
// The spec field cannot be changed after a ServiceBinding is
// created.  Changes submitted to the spec field will be ignored.
type ServiceBindingSpec struct {
	// ServiceInstanceRef is the reference to the Instance this ServiceBinding is to.
	//
	// Immutable.
	ServiceInstanceRef LocalObjectReference

	// Parameters is a set of the parameters to be passed to the underlying
	// broker. The inline YAML/JSON payload to be translated into equivalent
	// JSON object. If a top-level parameter name exists in multiples sources
	// among `Parameters` and `ParametersFrom` fields, it is considered to be
	// a user error in the specification.
	//
	// The Parameters field is NOT secret or secured in any way and should
	// NEVER be used to hold sensitive information. To set parameters that
	// contain secret information, you should ALWAYS store that information
	// in a Secret and use the ParametersFrom field.
	//
	// +optional
	Parameters *runtime.RawExtension

	// List of sources to populate parameters.
	// If a top-level parameter name exists in multiples sources among
	// `Parameters` and `ParametersFrom` fields, it is
	// considered to be a user error in the specification
	// +optional
	ParametersFrom []ParametersFromSource

	// SecretName is the name of the secret to create in the ServiceBinding's
	// namespace that will hold the credentials associated with the ServiceBinding.
	SecretName string

	// List of transformations that should be applied to the credentials returned
	// by the broker before they are inserted into the Secret
	SecretTransforms []SecretTransform

	// ExternalID is the identity of this object for use with the OSB API.
	//
	// Immutable.
	ExternalID string

	// Currently, this field is ALPHA: it may change or disappear at any time
	// and its data will not be migrated.
	//
	// UserInfo contains information about the user that last modified this
	// ServiceBinding. This field is set by the API server and not
	// settable by the end-user. User-provided values for this field are not saved.
	// +optional
	UserInfo *UserInfo
}

// ServiceBindingStatus represents the current status of a ServiceBinding.
type ServiceBindingStatus struct {
	Conditions []ServiceBindingCondition

	// Currently, this field is ALPHA: it may change or disappear at any time
	// and its data will not be migrated.
	//
	// AsyncOpInProgress is set to true if there is an ongoing async operation
	// against this ServiceBinding in progress.
	AsyncOpInProgress bool

	// Currently, this field is ALPHA: it may change or disappear at any time
	// and its data will not be migrated.
	//
	// LastOperation is the string that the broker may have returned when
	// an async operation started, it should be sent back to the broker
	// on poll requests as a query param.
	LastOperation *string

	// CurrentOperation is the operation the Controller is currently performing
	// on the ServiceBinding.
	CurrentOperation ServiceBindingOperation

	// ReconciledGeneration is the 'Generation' of the
	// ServiceBindingSpec that was last processed by the controller.
	// The reconciled generation is updated even if the controller failed to
	// process the spec.
	ReconciledGeneration int64

	// OperationStartTime is the time at which the current operation began.
	OperationStartTime *metav1.Time

	// InProgressProperties is the properties state of the
	// ServiceBinding when a Bind is in progress. If the current
	// operation is an Unbind, this will be nil.
	InProgressProperties *ServiceBindingPropertiesState

	// ExternalProperties is the properties state of the
	// ServiceBinding which the broker knows about.
	ExternalProperties *ServiceBindingPropertiesState

	// OrphanMitigationInProgress is a flag that represents whether orphan
	// mitigation is in progress.
	OrphanMitigationInProgress bool

	// UnbindStatus describes what has been done to unbind a ServiceBinding
	UnbindStatus ServiceBindingUnbindStatus
}

// ServiceBindingCondition condition information for a ServiceBinding.
type ServiceBindingCondition struct {
	// Type of the condition, currently ('Ready').
	Type ServiceBindingConditionType

	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus

	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	LastTransitionTime metav1.Time

	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	Reason string

	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	Message string
}

// ServiceBindingConditionType represents a ServiceBindingCondition value.
type ServiceBindingConditionType string

const (
	// ServiceBindingConditionReady represents a ServiceBindingCondition is in ready state.
	ServiceBindingConditionReady ServiceBindingConditionType = "Ready"

	// ServiceBindingConditionFailed represents a ServiceBindingCondition that has failed
	// completely and should not be retried.
	ServiceBindingConditionFailed ServiceBindingConditionType = "Failed"
)

// ServiceBindingOperation represents a type of operation
// the controller can be performing for a binding in the OSB API.
type ServiceBindingOperation string

const (
	// ServiceBindingOperationBind indicates that the
	// ServiceBinding is being bound.
	ServiceBindingOperationBind ServiceBindingOperation = "Bind"
	// ServiceBindingOperationUnbind indicates that the
	// ServiceBinding is being unbound.
	ServiceBindingOperationUnbind ServiceBindingOperation = "Unbind"
)

// These are internal finalizer values to service catalog, must be qualified name.
const (
	FinalizerServiceCatalog string = "kubernetes-incubator/service-catalog"
)

// ServiceBindingPropertiesState is the state of a
// ServiceBinding that the ServiceBroker knows about.
type ServiceBindingPropertiesState struct {
	// Parameters is a blob of the parameters and their values that the broker
	// knows about for this ServiceBinding.  If a parameter was
	// sourced from a secret, its value will be "<redacted>" in this blob.
	Parameters *runtime.RawExtension

	// ParametersChecksum is the checksum of the parameters that were sent.
	ParametersChecksum string

	// UserInfo is information about the user that made the request.
	UserInfo *UserInfo
}

// ServiceBindingUnbindStatus is the status of unbinding a Binding
type ServiceBindingUnbindStatus string

const (
	// ServiceBindingUnbindStatusNotRequired indicates that a binding request
	// has not been sent for the ServiceBinding, so no unbinding request
	// needs to be made.
	ServiceBindingUnbindStatusNotRequired ServiceBindingUnbindStatus = "NotRequired"
	// ServiceBindingUnbindStatusRequired indicates that a binding request has
	// been sent for the ServiceBinding. An unbind request must be made before
	// deleting the ServiceBinding.
	ServiceBindingUnbindStatusRequired ServiceBindingUnbindStatus = "Required"
	// ServiceBindingUnbindStatusSucceeded indicates that a unbind request
	// has been sent for the ServiceBinding, and the request was successful.
	ServiceBindingUnbindStatusSucceeded ServiceBindingUnbindStatus = "Succeeded"
	// ServiceBindingUnbindStatusFailed indicates that unbind requests
	// have been sent for the ServiceBinding but they failed. The controller
	// has given up on sending more unbind requests.
	ServiceBindingUnbindStatusFailed ServiceBindingUnbindStatus = "Failed"
)

// ParametersFromSource represents the source of a set of Parameters
type ParametersFromSource struct {
	// The Secret key to select from.
	// The value must be a JSON object.
	// +optional
	SecretKeyRef *SecretKeyReference
}

// SecretKeyReference references a key of a Secret.
type SecretKeyReference struct {
	// The name of the secret in the pod's namespace to select from.
	Name string
	// The key of the secret to select from.  Must be a valid secret key.
	Key string
}

// ObjectReference contains enough information to let you locate the
// referenced object.
type ObjectReference struct {
	// Namespace of the referent.
	Namespace string
	// Name of the referent.
	Name string
}

// LocalObjectReference contains enough information to let you locate the
// referenced object inside the same namespace.
type LocalObjectReference struct {
	// Name of the referent.
	Name string
}

// ClusterObjectReference contains enough information to let you locate the
// cluster-scoped referenced object.
type ClusterObjectReference struct {
	// Name of the referent.
	Name string
}

// SecretTransform is a single transformation of the credentials returned
// from the broker
type SecretTransform struct {
	RenameKey   *RenameKeyTransform
	AddKey      *AddKeyTransform
	AddKeysFrom *AddKeysFromTransform
	RemoveKey   *RemoveKeyTransform
}

// RenameKeyTransform specifies that one of the credentials keys returned
// from the broker should be renamed
type RenameKeyTransform struct {
	From string
	To   string
}

// AddKeyTransform specifies that Service Catalog should add an
// additional entry to the Secret associated with the ServiceBinding.
type AddKeyTransform struct {
	Key                string
	Value              []byte
	StringValue        *string
	JSONPathExpression *string
}

// AddKeysFromTransform specifies that Service Catalog should merge
// an existing secret into the the Secret associated with the ServiceBinding.
type AddKeysFromTransform struct {
	SecretRef *ObjectReference
}

// RemoveKeyTransform specifies that one of the credentials keys returned
// from the broker should not be included in the credentials Secret.
type RemoveKeyTransform struct {
	Key string
}
