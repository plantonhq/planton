package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Credential represents a base credential document in MongoDB.
// Each provider has its own collection (e.g., aws_credentials, gcp_credentials).
type Credential struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Provider  string             `bson:"provider" json:"provider"` // "aws", "gcp", "azure", etc.
	Spec      interface{}        `bson:"spec" json:"spec"`         // Provider-specific credential spec
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// AwsCredential represents AWS credentials.
type AwsCredential struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name            string             `bson:"name" json:"name"`
	AccountID       string             `bson:"account_id" json:"account_id"`
	AccessKeyID     string             `bson:"access_key_id" json:"access_key_id"`
	SecretAccessKey string             `bson:"secret_access_key" json:"secret_access_key"`
	Region          string             `bson:"region,omitempty" json:"region,omitempty"`
	SessionToken    string             `bson:"session_token,omitempty" json:"session_token,omitempty"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

// GcpCredential represents GCP credentials.
type GcpCredential struct {
	ID                      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name                    string             `bson:"name" json:"name"`
	ServiceAccountKeyBase64 string             `bson:"service_account_key_base64" json:"service_account_key_base64"`
	CreatedAt               time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt               time.Time          `bson:"updated_at" json:"updated_at"`
}

// AzureCredential represents Azure credentials.
type AzureCredential struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name           string             `bson:"name" json:"name"`
	ClientID       string             `bson:"client_id" json:"client_id"`
	ClientSecret   string             `bson:"client_secret" json:"client_secret"`
	TenantID       string             `bson:"tenant_id" json:"tenant_id"`
	SubscriptionID string             `bson:"subscription_id" json:"subscription_id"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

// AtlasCredential represents MongoDB Atlas credentials.
type AtlasCredential struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name       string             `bson:"name" json:"name"`
	PublicKey  string             `bson:"public_key" json:"public_key"`
	PrivateKey string             `bson:"private_key" json:"private_key"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

// Auth0Credential represents Auth0 credentials.
type Auth0Credential struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Domain       string             `bson:"domain" json:"domain"`
	ClientID     string             `bson:"client_id" json:"client_id"`
	ClientSecret string             `bson:"client_secret" json:"client_secret"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

// CloudflareCredential represents Cloudflare credentials.
type CloudflareCredential struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name       string             `bson:"name" json:"name"`
	AuthScheme int32              `bson:"auth_scheme" json:"auth_scheme"` // 1 = api_token, 2 = legacy_api_key
	APIToken   string             `bson:"api_token,omitempty" json:"api_token,omitempty"`
	APIKey     string             `bson:"api_key,omitempty" json:"api_key,omitempty"`
	Email      string             `bson:"email,omitempty" json:"email,omitempty"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

// ConfluentCredential represents Confluent Cloud credentials.
type ConfluentCredential struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	APIKey    string             `bson:"api_key" json:"api_key"`
	APISecret string             `bson:"api_secret" json:"api_secret"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// SnowflakeCredential represents Snowflake credentials.
type SnowflakeCredential struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Account   string             `bson:"account" json:"account"`
	Region    string             `bson:"region" json:"region"`
	Username  string             `bson:"username" json:"username"`
	Password  string             `bson:"password" json:"password"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// OpenStackCredential represents OpenStack credentials.
// Supports three authentication methods: password, application_credential, and token.
// The AuthMethod field acts as a discriminator to determine which credential fields are active.
type OpenStackCredential struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name     string             `bson:"name" json:"name"`
	AuthURL  string             `bson:"auth_url" json:"auth_url"`
	Region   string             `bson:"region,omitempty" json:"region,omitempty"`
	// AuthMethod discriminates the active credential set: "password", "application_credential", or "token"
	AuthMethod string `bson:"auth_method" json:"auth_method"`
	// Password authentication fields
	UserName string `bson:"user_name,omitempty" json:"user_name,omitempty"`
	Password string `bson:"password,omitempty" json:"password,omitempty"`
	// Application credential authentication fields
	ApplicationCredentialID     string `bson:"application_credential_id,omitempty" json:"application_credential_id,omitempty"`
	ApplicationCredentialName   string `bson:"application_credential_name,omitempty" json:"application_credential_name,omitempty"`
	ApplicationCredentialSecret string `bson:"application_credential_secret,omitempty" json:"application_credential_secret,omitempty"`
	// Token authentication fields
	Token string `bson:"token,omitempty" json:"token,omitempty"`
	// Project/tenant context
	TenantName string `bson:"tenant_name,omitempty" json:"tenant_name,omitempty"`
	TenantID   string `bson:"tenant_id,omitempty" json:"tenant_id,omitempty"`
	// Domain context (Identity v3)
	UserDomainName    string `bson:"user_domain_name,omitempty" json:"user_domain_name,omitempty"`
	UserDomainID      string `bson:"user_domain_id,omitempty" json:"user_domain_id,omitempty"`
	ProjectDomainName string `bson:"project_domain_name,omitempty" json:"project_domain_name,omitempty"`
	ProjectDomainID   string `bson:"project_domain_id,omitempty" json:"project_domain_id,omitempty"`
	// TLS
	Insecure   bool   `bson:"insecure,omitempty" json:"insecure,omitempty"`
	CACertFile string `bson:"cacert_file,omitempty" json:"cacert_file,omitempty"`
	// Advanced
	EndpointType string    `bson:"endpoint_type,omitempty" json:"endpoint_type,omitempty"`
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at" json:"updated_at"`
}

// ScalewayCredential represents Scaleway credentials.
// Uses a flat access key / secret key authentication model.
type ScalewayCredential struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name           string             `bson:"name" json:"name"`
	AccessKey      string             `bson:"access_key" json:"access_key"`
	SecretKey      string             `bson:"secret_key" json:"secret_key"`
	ProjectID      string             `bson:"project_id,omitempty" json:"project_id,omitempty"`
	OrganizationID string             `bson:"organization_id,omitempty" json:"organization_id,omitempty"`
	Region         string             `bson:"region,omitempty" json:"region,omitempty"`
	Zone           string             `bson:"zone,omitempty" json:"zone,omitempty"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

// AlicloudCredential represents Alibaba Cloud credentials.
// Supports seven authentication methods via an AuthMethod string discriminator.
// Only the fields relevant to the active auth method are populated.
type AlicloudCredential struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	AuthMethod  string             `bson:"auth_method" json:"auth_method"`
	Region      string             `bson:"region,omitempty" json:"region,omitempty"`
	AccountId   string             `bson:"account_id,omitempty" json:"account_id,omitempty"`
	AccountType string             `bson:"account_type,omitempty" json:"account_type,omitempty"`
	// Static / STS / AssumeRole credentials
	AccessKey     string `bson:"access_key,omitempty" json:"access_key,omitempty"`
	SecretKey     string `bson:"secret_key,omitempty" json:"secret_key,omitempty"`
	SecurityToken string `bson:"security_token,omitempty" json:"security_token,omitempty"`
	// ECS role
	EcsRoleName string `bson:"ecs_role_name,omitempty" json:"ecs_role_name,omitempty"`
	// Assume role / OIDC shared fields
	RoleArn          string `bson:"role_arn,omitempty" json:"role_arn,omitempty"`
	SessionName      string `bson:"session_name,omitempty" json:"session_name,omitempty"`
	Policy           string `bson:"policy,omitempty" json:"policy,omitempty"`
	SessionExpiration int32  `bson:"session_expiration,omitempty" json:"session_expiration,omitempty"`
	ExternalId       string `bson:"external_id,omitempty" json:"external_id,omitempty"`
	// OIDC-specific
	OidcProviderArn string `bson:"oidc_provider_arn,omitempty" json:"oidc_provider_arn,omitempty"`
	OidcToken       string `bson:"oidc_token,omitempty" json:"oidc_token,omitempty"`
	OidcTokenFile   string `bson:"oidc_token_file,omitempty" json:"oidc_token_file,omitempty"`
	// Shared credentials
	CredentialsFile string `bson:"credentials_file,omitempty" json:"credentials_file,omitempty"`
	Profile         string `bson:"profile,omitempty" json:"profile,omitempty"`
	// Sidecar credentials
	CredentialsUri string    `bson:"credentials_uri,omitempty" json:"credentials_uri,omitempty"`
	CreatedAt      time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at" json:"updated_at"`
}

// KubernetesCredential represents Kubernetes cluster credentials.
type KubernetesCredential struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Provider  int32              `bson:"provider" json:"provider"` // 1 = gcp_gke, 2 = aws_eks, 3 = azure_aks, 4 = digital_ocean_doks
	Spec      interface{}        `bson:"spec" json:"spec"`         // Provider-specific k8s config
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
