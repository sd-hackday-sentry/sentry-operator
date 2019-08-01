package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SentrySpec defines the desired state of Sentry
// +k8s:openapi-gen=true
type SentrySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	//Name is the distinct name of the Sentry service we're running
	Name string `json:"name"`

	//Postgreshost is the name of server running postgres
	Postgreshost     string `json:"postgreshost"`
	//Postgresport is the port on which the database server is listening
	Postgresport     int    `json:"postgresport"`
	//Postgresname is the name of the database within postgres we're using
	Postgresname     string `json:"postgresname"`
	// Postgresuser is the name of the secret containing the database username
	Postgresuser     string `json:"postgresuser"`
	// Postgrespassword is the name of the secret containing the database password
	Postgrespassword string `json:"postgrespassword"`

	// Redishost is the name of the server running redis
	Redishost        string `json:"redishost"`
	// Redisport is the port on which the redis server is listening
	Redisport        int    `json:"redisport"`
	// Redisname is the name of the redis instance we're using
	Redisname        string `json:"redisname"`
}

// SentryStatus defines the observed state of Sentry
// +k8s:openapi-gen=true
type SentryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	Status string `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Sentry is the Schema for the sentries API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Sentry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SentrySpec   `json:"spec,omitempty"`
	Status SentryStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SentryList contains a list of Sentry
type SentryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Sentry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Sentry{}, &SentryList{})
}
