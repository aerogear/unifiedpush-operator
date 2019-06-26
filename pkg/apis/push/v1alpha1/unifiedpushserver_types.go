package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// UnifiedPushServerSpec defines the desired state of UnifiedPushServer
// +k8s:openapi-gen=true
type UnifiedPushServerSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Backups []UnifiedPushServerBackup `json:"backups,omitempty"`
}

// UnifiedPushServerStatus defines the observed state of UnifiedPushServer
// +k8s:openapi-gen=true
type UnifiedPushServerStatus struct {
	Phase StatusPhase `json:"phase"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// UnifiedPushServer is the Schema for the unifiedpushservers API
// +k8s:openapi-gen=true
type UnifiedPushServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UnifiedPushServerSpec   `json:"spec,omitempty"`
	Status UnifiedPushServerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// UnifiedPushServerList contains a list of UnifiedPushServer
type UnifiedPushServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UnifiedPushServer `json:"items"`
}

// Backup contains the info needed to configure a CronJob for backups
type UnifiedPushServerBackup struct {
	Name                         string            `json:"name"`
	Labels                       map[string]string `json:"labels"`
	Schedule                     string            `json:"schedule"`
	EncryptionKeySecretName      string            `json:"encryptionKeySecretName"`
	EncryptionKeySecretNamespace string            `json:"encryptionKeySecretNamespace"`
	BackendSecretName            string            `json:"backendSecretName"`
	BackendSecretNamespace       string            `json:"backendSecretNamespace"`
}

type StatusPhase string

var (
	PhaseEmpty     StatusPhase = ""
	PhaseComplete  StatusPhase = "Complete"
	PhaseProvision StatusPhase = "Provisioning"
)

func init() {
	SchemeBuilder.Register(&UnifiedPushServer{}, &UnifiedPushServerList{})
}
