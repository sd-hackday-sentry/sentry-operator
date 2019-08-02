package v1alpha1

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

	//SentryImage is the image of sentry we are running (defaults: docker.io/sentry:latest)
	SentryImage string `json:"sentryImage,omitempty"`

	//PostgresHost is the name of server running postgres
	PostgresHost string `json:"postgresHost"`
	//PostgresPort is the port on which the database server is listening (defaults: 5432)
	PostgresPort int `json:"postgresPort,omitempty"`
	//PostgresDB is the database within postgres we're using
	PostgresDB string `json:"postgresDB"`
	// PostgresUser is the name of the secret containing the database username
	PostgresUser string `json:"postgresUser"`
	// PostgresPassword is the name of the secret containing the database password
	PostgresPassword string `json:"postgresPassword"`

	// RedisHost is the name of the server running redis
	RedisHost string `json:"redisHost"`
	// RedisPort is the port on which the redis server is listening (defaults: 6379)
	RedisPort int `json:"redisPort,omitempty"`
	// RedisDB is the name of the redis instance we're using (defaults: "0")
	RedisDB string `json:"redisDB,omitempty"`
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

// SetDefaults set the default values for the sentry spec
func (s *Sentry) SetDefaults() {

	sp := &s.Spec

	if sp.SentryImage == "" {
		sp.SentryImage = "docker.io/sentry:latest"
	}

	if sp.PostgresPort == 0 {
		sp.PostgresPort = 5432
	}

	if sp.RedisPort == 0 {
		sp.RedisPort = 6379
	}

	if sp.RedisDB == "" {
		sp.RedisDB = "0"
	}
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
