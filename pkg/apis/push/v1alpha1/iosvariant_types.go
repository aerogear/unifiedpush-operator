package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// IOSVariantSpec defines the desired state of IOSVariant
// +k8s:openapi-gen=true
type IOSVariantSpec struct {
	// Description is human friendly description for the variant.
	Description string `json:"description,omitempty"`
	// Certificate defines the base64 encoded APNs certificate that is needed to establish a
	// connection to Apple's APNs Push Servers.
	Certificate string `json:"certificate"`
	// Passphrase defines the APNs passphrase that is needed to establish a connection to any
	// of Apple's APNs Push Servers.
	Passphrase string `json:"passphrase"`
	// Production defines if a connection to production APNS server should be used. If false, a connection to
	// Apple's Sandbox/Development APNs server will be established for this iOS variant.
	Production bool `json:"production"`
}

// IOSVariantStatus defines the observed state of IOSVariant
// +k8s:openapi-gen=true
type IOSVariantStatus struct {
	Ready     bool   `json:"ready"`
	VariantId string `json:"variantId,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IOSVariant is the Schema for the iosvariants API
// +k8s:openapi-gen=true
type IOSVariant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IOSVariantSpec   `json:"spec,omitempty"`
	Status IOSVariantStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IOSVariantList contains a list of IOSVariant
type IOSVariantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IOSVariant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IOSVariant{}, &IOSVariantList{})
}
