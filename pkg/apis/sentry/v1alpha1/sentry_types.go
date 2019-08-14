package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SentrySpec defines the desired state of Sentry
// +k8s:openapi-gen=true
type SentrySpec struct {
	//SentryImage is the image of sentry we are running (defaults: docker.io/sentry:latest)
	SentryImage string `json:"sentryImage,omitempty"`
	//SentryWebReplicas is the number of web workers to spawn (defaults: 2)
	SentryWebReplicas int `json:"sentryWebReplicas,omitempty"`
	//SentryWorkers is the number of async workers to spawn (defaults: 3)
	SentryWorkers int `json:"sentryWorkers,omitempty"`
	//SentryEnvironment is the environment this sentry cluster belongs to (defaults: production)
	SentryEnvironment string `json:"sentryEnvironment,omitempty"`
	//SentrySecret is the secret holding the sentry-specific secret config values
	SentrySecret string `json:"sentrySecret"`
	//SentrySecretKeyKey is the key inside the sentry secret holding the salt hash string
	//for cryptography (defaults: SENTRY_SECRET_KEY)
	SentrySecretKeyKey string `json:"sentrySecretKeyKey,omitempty"`
	//PostgresPasswordKey is the key inside the sentry secret holding the password
	//to connect to the database (defaults: SENTRY_DB_PASSWORD)
	PostgresPasswordKey string `json:"postgresPasswordKey,omitempty"`
	//SentrySuperUserEmailKey is the key inside the sentry secret holding the
	//superuser's email address (defaults: "SENTRY_SU_EMAIL")
	SentrySuperUserEmailKey string `json:"sentrySuperUserEmailKey,omitempty"`
	//SentrySuperUserPasswordKey is the key inside the sentry secret holding the
	//superuser's password (defaults: "SENTRY_SU_PASSWORD")
	SentrySuperUserPasswordKey string `json:"sentrySuperUserPasswordKey,omitempty"`

	//PostgresHost is the name of server running postgres
	PostgresHost string `json:"postgresHost"`
	//PostgresPort is the port on which the database server is listening (defaults: 5432)
	PostgresPort int `json:"postgresPort,omitempty"`
	//PostgresDB is the database within postgres we're using
	PostgresDB string `json:"postgresDB"`
	//PostgresUser is the name of the secret containing the database username
	PostgresUser string `json:"postgresUser"`

	//RedisHost is the name of the server running redis
	RedisHost string `json:"redisHost"`
	//RedisPort is the port on which the redis server is listening (defaults: 6379)
	RedisPort int `json:"redisPort,omitempty"`
	//RedisDB is the name of the redis instance we're using (defaults: "0")
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

	if sp.SentryEnvironment == "" {
		sp.SentryEnvironment = "production"
	}

	if sp.SentryWebReplicas == 0 {
		sp.SentryWebReplicas = 2
	}

	if sp.SentryWorkers == 0 {
		sp.SentryWorkers = 3
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

	if sp.PostgresPasswordKey == "" {
		sp.PostgresPasswordKey = "SENTRY_DB_PASSWORD"
	}

	if sp.SentrySecretKeyKey == "" {
		sp.SentrySecretKeyKey = "SENTRY_SECRET_KEY"
	}

	if sp.SentrySuperUserEmailKey == "" {
		sp.SentrySuperUserEmailKey = "SENTRY_SU_EMAIL"
	}

	if sp.SentrySuperUserPasswordKey == "" {
		sp.SentrySuperUserPasswordKey = "SENTRY_SU_PASSWORD"
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
