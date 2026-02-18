package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/plantonhq/openmcf/app/backend/internal/database"
	"github.com/plantonhq/openmcf/app/backend/pkg/models"

	"connectrpc.com/connect"
	credentialv1 "github.com/plantonhq/openmcf/apis/org/openmcf/app/credential/v1"
	auth0v1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/auth0"
	awsv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws"
	azurev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure"
	gcpv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp"
	alicloudv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud"
	ociv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci"
	openstackv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack"
	scalewayv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CredentialService implements the CredentialService RPC.
type CredentialService struct {
	credentialRepo *database.CredentialRepository
}

// NewCredentialService creates a new service instance.
func NewCredentialService(credentialRepo *database.CredentialRepository) *CredentialService {
	return &CredentialService{
		credentialRepo: credentialRepo,
	}
}

// Create creates a new cloud provider credential.
func (s *CredentialService) Create(
	ctx context.Context,
	req *connect.Request[credentialv1.CreateCredentialRequest],
) (*connect.Response[credentialv1.CreateCredentialResponse], error) {
	// Validate common fields
	if req.Msg.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name is required"))
	}
	if req.Msg.Provider == credentialv1.Credential_CREDENTIAL_PROVIDER_UNSPECIFIED {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("provider is required"))
	}

	now := time.Now()

	// Handle based on provider type
	switch req.Msg.Provider {
	case credentialv1.Credential_GCP:
		return s.createGcpCredential(ctx, req.Msg.Name, req.Msg.ProviderConfig, now)
	case credentialv1.Credential_AWS:
		return s.createAwsCredential(ctx, req.Msg.Name, req.Msg.ProviderConfig, now)
	case credentialv1.Credential_AZURE:
		return s.createAzureCredential(ctx, req.Msg.Name, req.Msg.ProviderConfig, now)
	case credentialv1.Credential_AUTH0:
		return s.createAuth0Credential(ctx, req.Msg.Name, req.Msg.ProviderConfig, now)
	case credentialv1.Credential_OPENSTACK:
		return s.createOpenStackCredential(ctx, req.Msg.Name, req.Msg.ProviderConfig, now)
	case credentialv1.Credential_SCALEWAY:
		return s.createScalewayCredential(ctx, req.Msg.Name, req.Msg.ProviderConfig, now)
	case credentialv1.Credential_ALICLOUD:
		return s.createAlicloudCredential(ctx, req.Msg.Name, req.Msg.ProviderConfig, now)
	case credentialv1.Credential_OCI:
		return s.createOciCredential(ctx, req.Msg.Name, req.Msg.ProviderConfig, now)
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unsupported provider: %v", req.Msg.Provider))
	}
}

// createGcpCredential creates a GCP credential.
func (s *CredentialService) createGcpCredential(
	ctx context.Context,
	name string,
	providerConfig *credentialv1.CredentialProviderConfig,
	now time.Time,
) (*connect.Response[credentialv1.CreateCredentialResponse], error) {
	if providerConfig == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("provider_config is required"))
	}
	gcpConfig, ok := providerConfig.Data.(*credentialv1.CredentialProviderConfig_Gcp)
	if !ok || gcpConfig == nil || gcpConfig.Gcp == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("gcp provider_config is required"))
	}
	if gcpConfig.Gcp.ServiceAccountKeyBase64 == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("service_account_key_base64 is required"))
	}

	// Frontend sends base64 encoded string, backend validates and stores it
	// Validate that it's valid base64
	decodedBytes, err := base64.StdEncoding.DecodeString(gcpConfig.Gcp.ServiceAccountKeyBase64)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid base64 encoded service account key: %w", err))
	}

	// Validate that decoded content is valid JSON
	var keyJSON map[string]interface{}
	if err := json.Unmarshal(decodedBytes, &keyJSON); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("service account key is not valid JSON: %w", err))
	}

	// Store base64 in database
	createdCredential, err := s.credentialRepo.CreateGcp(ctx, name, gcpConfig.Gcp.ServiceAccountKeyBase64)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create GCP credential: %w", err))
	}

	// Decode base64 for response (return decoded JSON string to frontend)
	decodedKeyString := string(decodedBytes)

	protoCredential := &credentialv1.Credential{
		Id:       createdCredential.ID.Hex(),
		Name:     createdCredential.Name,
		Provider: credentialv1.Credential_GCP,
		ProviderConfig: &credentialv1.CredentialProviderConfig{
			Data: &credentialv1.CredentialProviderConfig_Gcp{
				Gcp: &gcpv1.GcpProviderConfig{
					ServiceAccountKeyBase64: decodedKeyString, // Actually contains decoded JSON string
				},
			},
		},
	}

	if !createdCredential.CreatedAt.IsZero() {
		protoCredential.CreatedAt = timestamppb.New(createdCredential.CreatedAt)
	}
	if !createdCredential.UpdatedAt.IsZero() {
		protoCredential.UpdatedAt = timestamppb.New(createdCredential.UpdatedAt)
	}

	return connect.NewResponse(&credentialv1.CreateCredentialResponse{
		Credential: protoCredential,
	}), nil
}

// createAwsCredential creates an AWS credential.
func (s *CredentialService) createAwsCredential(
	ctx context.Context,
	name string,
	providerConfig *credentialv1.CredentialProviderConfig,
	now time.Time,
) (*connect.Response[credentialv1.CreateCredentialResponse], error) {
	if providerConfig == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("provider_config is required"))
	}
	awsConfig, ok := providerConfig.Data.(*credentialv1.CredentialProviderConfig_Aws)
	if !ok || awsConfig == nil || awsConfig.Aws == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("aws provider_config is required"))
	}
	if awsConfig.Aws.AccountId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("account_id is required"))
	}
	if awsConfig.Aws.AccessKeyId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("access_key_id is required"))
	}
	if awsConfig.Aws.SecretAccessKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("secret_access_key is required"))
	}

	region := awsConfig.Aws.Region
	sessionToken := awsConfig.Aws.SessionToken

	createdCredential, err := s.credentialRepo.CreateAws(ctx, name, awsConfig.Aws.AccountId, awsConfig.Aws.AccessKeyId, awsConfig.Aws.SecretAccessKey, region, sessionToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create AWS credential: %w", err))
	}

	protoCredential := &credentialv1.Credential{
		Id:       createdCredential.ID.Hex(),
		Name:     createdCredential.Name,
		Provider: credentialv1.Credential_AWS,
		ProviderConfig: &credentialv1.CredentialProviderConfig{
			Data: &credentialv1.CredentialProviderConfig_Aws{
				Aws: &awsv1.AwsProviderConfig{
					AccountId:       createdCredential.AccountID,
					AccessKeyId:     createdCredential.AccessKeyID,
					SecretAccessKey: createdCredential.SecretAccessKey,
					Region:          createdCredential.Region,
					SessionToken:    createdCredential.SessionToken,
				},
			},
		},
	}

	if !createdCredential.CreatedAt.IsZero() {
		protoCredential.CreatedAt = timestamppb.New(createdCredential.CreatedAt)
	}
	if !createdCredential.UpdatedAt.IsZero() {
		protoCredential.UpdatedAt = timestamppb.New(createdCredential.UpdatedAt)
	}

	return connect.NewResponse(&credentialv1.CreateCredentialResponse{
		Credential: protoCredential,
	}), nil
}

// createAzureCredential creates an Azure credential.
func (s *CredentialService) createAzureCredential(
	ctx context.Context,
	name string,
	providerConfig *credentialv1.CredentialProviderConfig,
	now time.Time,
) (*connect.Response[credentialv1.CreateCredentialResponse], error) {
	if providerConfig == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("provider_config is required"))
	}
	azureConfig, ok := providerConfig.Data.(*credentialv1.CredentialProviderConfig_Azure)
	if !ok || azureConfig == nil || azureConfig.Azure == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("azure provider_config is required"))
	}
	if azureConfig.Azure.ClientId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("client_id is required"))
	}
	if azureConfig.Azure.ClientSecret == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("client_secret is required"))
	}
	if azureConfig.Azure.TenantId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("tenant_id is required"))
	}
	if azureConfig.Azure.SubscriptionId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("subscription_id is required"))
	}

	createdCredential, err := s.credentialRepo.CreateAzure(ctx, name, azureConfig.Azure.ClientId, azureConfig.Azure.ClientSecret, azureConfig.Azure.TenantId, azureConfig.Azure.SubscriptionId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create Azure credential: %w", err))
	}

	protoCredential := &credentialv1.Credential{
		Id:       createdCredential.ID.Hex(),
		Name:     createdCredential.Name,
		Provider: credentialv1.Credential_AZURE,
		ProviderConfig: &credentialv1.CredentialProviderConfig{
			Data: &credentialv1.CredentialProviderConfig_Azure{
				Azure: &azurev1.AzureProviderConfig{
					ClientId:       createdCredential.ClientID,
					ClientSecret:   createdCredential.ClientSecret,
					TenantId:       createdCredential.TenantID,
					SubscriptionId: createdCredential.SubscriptionID,
				},
			},
		},
	}

	if !createdCredential.CreatedAt.IsZero() {
		protoCredential.CreatedAt = timestamppb.New(createdCredential.CreatedAt)
	}
	if !createdCredential.UpdatedAt.IsZero() {
		protoCredential.UpdatedAt = timestamppb.New(createdCredential.UpdatedAt)
	}

	return connect.NewResponse(&credentialv1.CreateCredentialResponse{
		Credential: protoCredential,
	}), nil
}

// createAuth0Credential creates an Auth0 credential.
func (s *CredentialService) createAuth0Credential(
	ctx context.Context,
	name string,
	providerConfig *credentialv1.CredentialProviderConfig,
	now time.Time,
) (*connect.Response[credentialv1.CreateCredentialResponse], error) {
	if providerConfig == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("provider_config is required"))
	}
	auth0Config, ok := providerConfig.Data.(*credentialv1.CredentialProviderConfig_Auth0)
	if !ok || auth0Config == nil || auth0Config.Auth0 == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("auth0 provider_config is required"))
	}
	if auth0Config.Auth0.Domain == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("domain is required"))
	}
	if auth0Config.Auth0.ClientId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("client_id is required"))
	}
	if auth0Config.Auth0.ClientSecret == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("client_secret is required"))
	}

	createdCredential, err := s.credentialRepo.CreateAuth0(ctx, name, auth0Config.Auth0.Domain, auth0Config.Auth0.ClientId, auth0Config.Auth0.ClientSecret)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create Auth0 credential: %w", err))
	}

	protoCredential := &credentialv1.Credential{
		Id:       createdCredential.ID.Hex(),
		Name:     createdCredential.Name,
		Provider: credentialv1.Credential_AUTH0,
		ProviderConfig: &credentialv1.CredentialProviderConfig{
			Data: &credentialv1.CredentialProviderConfig_Auth0{
				Auth0: &auth0v1.Auth0ProviderConfig{
					Domain:       createdCredential.Domain,
					ClientId:     createdCredential.ClientID,
					ClientSecret: createdCredential.ClientSecret,
				},
			},
		},
	}

	if !createdCredential.CreatedAt.IsZero() {
		protoCredential.CreatedAt = timestamppb.New(createdCredential.CreatedAt)
	}
	if !createdCredential.UpdatedAt.IsZero() {
		protoCredential.UpdatedAt = timestamppb.New(createdCredential.UpdatedAt)
	}

	return connect.NewResponse(&credentialv1.CreateCredentialResponse{
		Credential: protoCredential,
	}), nil
}

// List lists all credentials with optional provider filter.
func (s *CredentialService) List(
	ctx context.Context,
	req *connect.Request[credentialv1.ListCredentialsRequest],
) (*connect.Response[credentialv1.ListCredentialsResponse], error) {
	// Convert provider enum to string for database query
	var providerFilter *string
	if req.Msg.Provider != credentialv1.Credential_CREDENTIAL_PROVIDER_UNSPECIFIED {
		// Convert CredentialProvider enum to string
		var provider string
		switch req.Msg.Provider {
		case credentialv1.Credential_GCP:
			provider = "gcp"
		case credentialv1.Credential_AWS:
			provider = "aws"
		case credentialv1.Credential_AZURE:
			provider = "azure"
		case credentialv1.Credential_AUTH0:
			provider = "auth0"
		case credentialv1.Credential_OPENSTACK:
			provider = "openstack"
		case credentialv1.Credential_SCALEWAY:
			provider = "scaleway"
		case credentialv1.Credential_ALICLOUD:
			provider = "alicloud"
		case credentialv1.Credential_OCI:
			provider = "oci"
		}
		if provider != "" {
			providerFilter = &provider
		}
	}

	// Query database
	credentials, err := s.credentialRepo.List(ctx, providerFilter)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list credentials: %w", err))
	}

	// Convert to proto summaries (without sensitive data)
	summaries := make([]*credentialv1.CredentialSummary, 0, len(credentials))
	for _, cred := range credentials {
		summary := &credentialv1.CredentialSummary{
			Id:   cred["_id"].(primitive.ObjectID).Hex(),
			Name: cred["name"].(string),
		}

		// Convert provider string to enum
		providerStr := cred["provider"].(string)
		switch providerStr {
		case "gcp":
			summary.Provider = credentialv1.Credential_GCP
		case "aws":
			summary.Provider = credentialv1.Credential_AWS
		case "azure":
			summary.Provider = credentialv1.Credential_AZURE
		case "auth0":
			summary.Provider = credentialv1.Credential_AUTH0
		case "openstack":
			summary.Provider = credentialv1.Credential_OPENSTACK
		case "scaleway":
			summary.Provider = credentialv1.Credential_SCALEWAY
		case "alicloud":
			summary.Provider = credentialv1.Credential_ALICLOUD
		case "oci":
			summary.Provider = credentialv1.Credential_OCI
		}

		// Add timestamps if present
		if createdAt, ok := cred["created_at"].(primitive.DateTime); ok {
			summary.CreatedAt = timestamppb.New(createdAt.Time())
		}
		if updatedAt, ok := cred["updated_at"].(primitive.DateTime); ok {
			summary.UpdatedAt = timestamppb.New(updatedAt.Time())
		}

		summaries = append(summaries, summary)
	}

	return connect.NewResponse(&credentialv1.ListCredentialsResponse{
		Credentials: summaries,
	}), nil
}

// Get retrieves a credential by ID.
func (s *CredentialService) Get(
	ctx context.Context,
	req *connect.Request[credentialv1.GetCredentialRequest],
) (*connect.Response[credentialv1.GetCredentialResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	doc, err := s.credentialRepo.FindByID(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get credential: %w", err))
	}
	if doc == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("credential with ID '%s' not found", req.Msg.Id))
	}

	// Convert to proto based on provider
	providerStr := doc["provider"].(string)
	var protoCredential *credentialv1.Credential

	switch providerStr {
	case "gcp":
		gcpCred, err := convertBsonToGcpCredential(doc)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to convert credential: %w", err))
		}

		// Decode base64 to get the original service account key JSON string
		decodedKey, err := base64.StdEncoding.DecodeString(gcpCred.ServiceAccountKeyBase64)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to decode service account key: %w", err))
		}

		// Validate that the decoded content is valid JSON
		var keyJSON map[string]interface{}
		if err := json.Unmarshal(decodedKey, &keyJSON); err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("service account key is not valid JSON: %w", err))
		}

		// Return decoded JSON string (not base64) to frontend
		decodedKeyString := string(decodedKey)

		protoCredential = &credentialv1.Credential{
			Id:       gcpCred.ID.Hex(),
			Name:     gcpCred.Name,
			Provider: credentialv1.Credential_GCP,
			ProviderConfig: &credentialv1.CredentialProviderConfig{
				Data: &credentialv1.CredentialProviderConfig_Gcp{
					Gcp: &gcpv1.GcpProviderConfig{
						ServiceAccountKeyBase64: decodedKeyString, // Actually contains decoded JSON string
					},
				},
			},
		}
		if !gcpCred.CreatedAt.IsZero() {
			protoCredential.CreatedAt = timestamppb.New(gcpCred.CreatedAt)
		}
		if !gcpCred.UpdatedAt.IsZero() {
			protoCredential.UpdatedAt = timestamppb.New(gcpCred.UpdatedAt)
		}
	case "aws":
		awsCred, err := convertBsonToAwsCredential(doc)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to convert credential: %w", err))
		}
		protoCredential = &credentialv1.Credential{
			Id:       awsCred.ID.Hex(),
			Name:     awsCred.Name,
			Provider: credentialv1.Credential_AWS,
			ProviderConfig: &credentialv1.CredentialProviderConfig{
				Data: &credentialv1.CredentialProviderConfig_Aws{
					Aws: &awsv1.AwsProviderConfig{
						AccountId:       awsCred.AccountID,
						AccessKeyId:     awsCred.AccessKeyID,
						SecretAccessKey: awsCred.SecretAccessKey,
						Region:          awsCred.Region,
						SessionToken:    awsCred.SessionToken,
					},
				},
			},
		}
		if !awsCred.CreatedAt.IsZero() {
			protoCredential.CreatedAt = timestamppb.New(awsCred.CreatedAt)
		}
		if !awsCred.UpdatedAt.IsZero() {
			protoCredential.UpdatedAt = timestamppb.New(awsCred.UpdatedAt)
		}
	case "azure":
		azureCred, err := convertBsonToAzureCredential(doc)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to convert credential: %w", err))
		}
		protoCredential = &credentialv1.Credential{
			Id:       azureCred.ID.Hex(),
			Name:     azureCred.Name,
			Provider: credentialv1.Credential_AZURE,
			ProviderConfig: &credentialv1.CredentialProviderConfig{
				Data: &credentialv1.CredentialProviderConfig_Azure{
					Azure: &azurev1.AzureProviderConfig{
						ClientId:       azureCred.ClientID,
						ClientSecret:   azureCred.ClientSecret,
						TenantId:       azureCred.TenantID,
						SubscriptionId: azureCred.SubscriptionID,
					},
				},
			},
		}
		if !azureCred.CreatedAt.IsZero() {
			protoCredential.CreatedAt = timestamppb.New(azureCred.CreatedAt)
		}
		if !azureCred.UpdatedAt.IsZero() {
			protoCredential.UpdatedAt = timestamppb.New(azureCred.UpdatedAt)
		}
	case "auth0":
		auth0Cred, err := convertBsonToAuth0Credential(doc)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to convert credential: %w", err))
		}
		protoCredential = &credentialv1.Credential{
			Id:       auth0Cred.ID.Hex(),
			Name:     auth0Cred.Name,
			Provider: credentialv1.Credential_AUTH0,
			ProviderConfig: &credentialv1.CredentialProviderConfig{
				Data: &credentialv1.CredentialProviderConfig_Auth0{
					Auth0: &auth0v1.Auth0ProviderConfig{
						Domain:       auth0Cred.Domain,
						ClientId:     auth0Cred.ClientID,
						ClientSecret: auth0Cred.ClientSecret,
					},
				},
			},
		}
		if !auth0Cred.CreatedAt.IsZero() {
			protoCredential.CreatedAt = timestamppb.New(auth0Cred.CreatedAt)
		}
		if !auth0Cred.UpdatedAt.IsZero() {
			protoCredential.UpdatedAt = timestamppb.New(auth0Cred.UpdatedAt)
		}
	case "openstack":
		osCred, err := convertBsonToOpenStackCredential(doc)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to convert credential: %w", err))
		}
		protoCredential = &credentialv1.Credential{
			Id:             osCred.ID.Hex(),
			Name:           osCred.Name,
			Provider:       credentialv1.Credential_OPENSTACK,
			ProviderConfig: openstackModelToProtoConfig(osCred),
		}
		if !osCred.CreatedAt.IsZero() {
			protoCredential.CreatedAt = timestamppb.New(osCred.CreatedAt)
		}
		if !osCred.UpdatedAt.IsZero() {
			protoCredential.UpdatedAt = timestamppb.New(osCred.UpdatedAt)
		}
	case "scaleway":
		scwCred, err := convertBsonToScalewayCredential(doc)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to convert credential: %w", err))
		}
		protoCredential = &credentialv1.Credential{
			Id:             scwCred.ID.Hex(),
			Name:           scwCred.Name,
			Provider:       credentialv1.Credential_SCALEWAY,
			ProviderConfig: scalewayModelToProtoConfig(scwCred),
		}
		if !scwCred.CreatedAt.IsZero() {
			protoCredential.CreatedAt = timestamppb.New(scwCred.CreatedAt)
		}
		if !scwCred.UpdatedAt.IsZero() {
			protoCredential.UpdatedAt = timestamppb.New(scwCred.UpdatedAt)
		}
	case "alicloud":
		aliCred, err := convertBsonToAlicloudCredential(doc)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to convert credential: %w", err))
		}
		protoCredential = &credentialv1.Credential{
			Id:             aliCred.ID.Hex(),
			Name:           aliCred.Name,
			Provider:       credentialv1.Credential_ALICLOUD,
			ProviderConfig: alicloudModelToProtoConfig(aliCred),
		}
		if !aliCred.CreatedAt.IsZero() {
			protoCredential.CreatedAt = timestamppb.New(aliCred.CreatedAt)
		}
		if !aliCred.UpdatedAt.IsZero() {
			protoCredential.UpdatedAt = timestamppb.New(aliCred.UpdatedAt)
		}
	case "oci":
		ociCred, err := convertBsonToOciCredential(doc)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to convert credential: %w", err))
		}
		protoCredential = &credentialv1.Credential{
			Id:             ociCred.ID.Hex(),
			Name:           ociCred.Name,
			Provider:       credentialv1.Credential_OCI,
			ProviderConfig: ociModelToProtoConfig(ociCred),
		}
		if !ociCred.CreatedAt.IsZero() {
			protoCredential.CreatedAt = timestamppb.New(ociCred.CreatedAt)
		}
		if !ociCred.UpdatedAt.IsZero() {
			protoCredential.UpdatedAt = timestamppb.New(ociCred.UpdatedAt)
		}
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unsupported provider: %s", providerStr))
	}

	return connect.NewResponse(&credentialv1.GetCredentialResponse{
		Credential: protoCredential,
	}), nil
}

// Update updates an existing credential.
func (s *CredentialService) Update(
	ctx context.Context,
	req *connect.Request[credentialv1.UpdateCredentialRequest],
) (*connect.Response[credentialv1.UpdateCredentialResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}
	if req.Msg.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name is required"))
	}
	if req.Msg.Provider == credentialv1.Credential_CREDENTIAL_PROVIDER_UNSPECIFIED {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("provider is required"))
	}

	// Handle based on provider type
	switch req.Msg.Provider {
	case credentialv1.Credential_GCP:
		return s.updateGcpCredential(ctx, req.Msg.Id, req.Msg.Name, req.Msg.ProviderConfig)
	case credentialv1.Credential_AWS:
		return s.updateAwsCredential(ctx, req.Msg.Id, req.Msg.Name, req.Msg.ProviderConfig)
	case credentialv1.Credential_AZURE:
		return s.updateAzureCredential(ctx, req.Msg.Id, req.Msg.Name, req.Msg.ProviderConfig)
	case credentialv1.Credential_AUTH0:
		return s.updateAuth0Credential(ctx, req.Msg.Id, req.Msg.Name, req.Msg.ProviderConfig)
	case credentialv1.Credential_OPENSTACK:
		return s.updateOpenStackCredential(ctx, req.Msg.Id, req.Msg.Name, req.Msg.ProviderConfig)
	case credentialv1.Credential_SCALEWAY:
		return s.updateScalewayCredential(ctx, req.Msg.Id, req.Msg.Name, req.Msg.ProviderConfig)
	case credentialv1.Credential_ALICLOUD:
		return s.updateAlicloudCredential(ctx, req.Msg.Id, req.Msg.Name, req.Msg.ProviderConfig)
	case credentialv1.Credential_OCI:
		return s.updateOciCredential(ctx, req.Msg.Id, req.Msg.Name, req.Msg.ProviderConfig)
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unsupported provider: %v", req.Msg.Provider))
	}
}

// updateGcpCredential updates a GCP credential.
func (s *CredentialService) updateGcpCredential(
	ctx context.Context,
	id, name string,
	providerConfig *credentialv1.CredentialProviderConfig,
) (*connect.Response[credentialv1.UpdateCredentialResponse], error) {
	if providerConfig == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("provider_config is required"))
	}
	gcpConfig, ok := providerConfig.Data.(*credentialv1.CredentialProviderConfig_Gcp)
	if !ok || gcpConfig == nil || gcpConfig.Gcp == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("gcp provider_config is required"))
	}
	if gcpConfig.Gcp.ServiceAccountKeyBase64 == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("service_account_key_base64 is required"))
	}

	// Frontend sends base64 encoded string, backend validates and stores it
	// Validate that it's valid base64
	decodedBytes, err := base64.StdEncoding.DecodeString(gcpConfig.Gcp.ServiceAccountKeyBase64)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid base64 encoded service account key: %w", err))
	}

	// Validate that decoded content is valid JSON
	var keyJSON map[string]interface{}
	if err := json.Unmarshal(decodedBytes, &keyJSON); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("service account key is not valid JSON: %w", err))
	}

	// Store base64 in database
	updatedCredential, err := s.credentialRepo.UpdateGcp(ctx, id, name, gcpConfig.Gcp.ServiceAccountKeyBase64)
	if err != nil {
		if err.Error() == fmt.Sprintf("GCP credential with ID '%s' not found", id) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update GCP credential: %w", err))
	}

	// Decode base64 for response (return decoded JSON string to frontend)
	decodedKeyString := string(decodedBytes)

	protoCredential := &credentialv1.Credential{
		Id:       updatedCredential.ID.Hex(),
		Name:     updatedCredential.Name,
		Provider: credentialv1.Credential_GCP,
		ProviderConfig: &credentialv1.CredentialProviderConfig{
			Data: &credentialv1.CredentialProviderConfig_Gcp{
				Gcp: &gcpv1.GcpProviderConfig{
					ServiceAccountKeyBase64: decodedKeyString, // Actually contains decoded JSON string
				},
			},
		},
	}

	if !updatedCredential.CreatedAt.IsZero() {
		protoCredential.CreatedAt = timestamppb.New(updatedCredential.CreatedAt)
	}
	if !updatedCredential.UpdatedAt.IsZero() {
		protoCredential.UpdatedAt = timestamppb.New(updatedCredential.UpdatedAt)
	}

	return connect.NewResponse(&credentialv1.UpdateCredentialResponse{
		Credential: protoCredential,
	}), nil
}

// updateAwsCredential updates an AWS credential.
func (s *CredentialService) updateAwsCredential(
	ctx context.Context,
	id, name string,
	providerConfig *credentialv1.CredentialProviderConfig,
) (*connect.Response[credentialv1.UpdateCredentialResponse], error) {
	if providerConfig == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("provider_config is required"))
	}
	awsConfig, ok := providerConfig.Data.(*credentialv1.CredentialProviderConfig_Aws)
	if !ok || awsConfig == nil || awsConfig.Aws == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("aws provider_config is required"))
	}
	if awsConfig.Aws.AccountId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("account_id is required"))
	}
	if awsConfig.Aws.AccessKeyId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("access_key_id is required"))
	}
	if awsConfig.Aws.SecretAccessKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("secret_access_key is required"))
	}

	updatedCredential, err := s.credentialRepo.UpdateAws(ctx, id, name, awsConfig.Aws.AccountId, awsConfig.Aws.AccessKeyId, awsConfig.Aws.SecretAccessKey, awsConfig.Aws.Region, awsConfig.Aws.SessionToken)
	if err != nil {
		if err.Error() == fmt.Sprintf("AWS credential with ID '%s' not found", id) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update AWS credential: %w", err))
	}

	protoCredential := &credentialv1.Credential{
		Id:       updatedCredential.ID.Hex(),
		Name:     updatedCredential.Name,
		Provider: credentialv1.Credential_AWS,
		ProviderConfig: &credentialv1.CredentialProviderConfig{
			Data: &credentialv1.CredentialProviderConfig_Aws{
				Aws: &awsv1.AwsProviderConfig{
					AccountId:       updatedCredential.AccountID,
					AccessKeyId:     updatedCredential.AccessKeyID,
					SecretAccessKey: updatedCredential.SecretAccessKey,
					Region:          updatedCredential.Region,
					SessionToken:    updatedCredential.SessionToken,
				},
			},
		},
	}

	if !updatedCredential.CreatedAt.IsZero() {
		protoCredential.CreatedAt = timestamppb.New(updatedCredential.CreatedAt)
	}
	if !updatedCredential.UpdatedAt.IsZero() {
		protoCredential.UpdatedAt = timestamppb.New(updatedCredential.UpdatedAt)
	}

	return connect.NewResponse(&credentialv1.UpdateCredentialResponse{
		Credential: protoCredential,
	}), nil
}

// updateAzureCredential updates an Azure credential.
func (s *CredentialService) updateAzureCredential(
	ctx context.Context,
	id, name string,
	providerConfig *credentialv1.CredentialProviderConfig,
) (*connect.Response[credentialv1.UpdateCredentialResponse], error) {
	if providerConfig == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("provider_config is required"))
	}
	azureConfig, ok := providerConfig.Data.(*credentialv1.CredentialProviderConfig_Azure)
	if !ok || azureConfig == nil || azureConfig.Azure == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("azure provider_config is required"))
	}
	if azureConfig.Azure.ClientId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("client_id is required"))
	}
	if azureConfig.Azure.ClientSecret == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("client_secret is required"))
	}
	if azureConfig.Azure.TenantId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("tenant_id is required"))
	}
	if azureConfig.Azure.SubscriptionId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("subscription_id is required"))
	}

	updatedCredential, err := s.credentialRepo.UpdateAzure(ctx, id, name, azureConfig.Azure.ClientId, azureConfig.Azure.ClientSecret, azureConfig.Azure.TenantId, azureConfig.Azure.SubscriptionId)
	if err != nil {
		if err.Error() == fmt.Sprintf("Azure credential with ID '%s' not found", id) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update Azure credential: %w", err))
	}

	protoCredential := &credentialv1.Credential{
		Id:       updatedCredential.ID.Hex(),
		Name:     updatedCredential.Name,
		Provider: credentialv1.Credential_AZURE,
		ProviderConfig: &credentialv1.CredentialProviderConfig{
			Data: &credentialv1.CredentialProviderConfig_Azure{
				Azure: &azurev1.AzureProviderConfig{
					ClientId:       updatedCredential.ClientID,
					ClientSecret:   updatedCredential.ClientSecret,
					TenantId:       updatedCredential.TenantID,
					SubscriptionId: updatedCredential.SubscriptionID,
				},
			},
		},
	}

	if !updatedCredential.CreatedAt.IsZero() {
		protoCredential.CreatedAt = timestamppb.New(updatedCredential.CreatedAt)
	}
	if !updatedCredential.UpdatedAt.IsZero() {
		protoCredential.UpdatedAt = timestamppb.New(updatedCredential.UpdatedAt)
	}

	return connect.NewResponse(&credentialv1.UpdateCredentialResponse{
		Credential: protoCredential,
	}), nil
}

// updateAuth0Credential updates an Auth0 credential.
func (s *CredentialService) updateAuth0Credential(
	ctx context.Context,
	id, name string,
	providerConfig *credentialv1.CredentialProviderConfig,
) (*connect.Response[credentialv1.UpdateCredentialResponse], error) {
	if providerConfig == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("provider_config is required"))
	}
	auth0Config, ok := providerConfig.Data.(*credentialv1.CredentialProviderConfig_Auth0)
	if !ok || auth0Config == nil || auth0Config.Auth0 == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("auth0 provider_config is required"))
	}
	if auth0Config.Auth0.Domain == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("domain is required"))
	}
	if auth0Config.Auth0.ClientId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("client_id is required"))
	}
	if auth0Config.Auth0.ClientSecret == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("client_secret is required"))
	}

	updatedCredential, err := s.credentialRepo.UpdateAuth0(ctx, id, name, auth0Config.Auth0.Domain, auth0Config.Auth0.ClientId, auth0Config.Auth0.ClientSecret)
	if err != nil {
		if err.Error() == fmt.Sprintf("Auth0 credential with ID '%s' not found", id) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update Auth0 credential: %w", err))
	}

	protoCredential := &credentialv1.Credential{
		Id:       updatedCredential.ID.Hex(),
		Name:     updatedCredential.Name,
		Provider: credentialv1.Credential_AUTH0,
		ProviderConfig: &credentialv1.CredentialProviderConfig{
			Data: &credentialv1.CredentialProviderConfig_Auth0{
				Auth0: &auth0v1.Auth0ProviderConfig{
					Domain:       updatedCredential.Domain,
					ClientId:     updatedCredential.ClientID,
					ClientSecret: updatedCredential.ClientSecret,
				},
			},
		},
	}

	if !updatedCredential.CreatedAt.IsZero() {
		protoCredential.CreatedAt = timestamppb.New(updatedCredential.CreatedAt)
	}
	if !updatedCredential.UpdatedAt.IsZero() {
		protoCredential.UpdatedAt = timestamppb.New(updatedCredential.UpdatedAt)
	}

	return connect.NewResponse(&credentialv1.UpdateCredentialResponse{
		Credential: protoCredential,
	}), nil
}

// Delete deletes a credential by ID.
func (s *CredentialService) Delete(
	ctx context.Context,
	req *connect.Request[credentialv1.DeleteCredentialRequest],
) (*connect.Response[credentialv1.DeleteCredentialResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	err := s.credentialRepo.Delete(ctx, req.Msg.Id)
	if err != nil {
		if err.Error() == fmt.Sprintf("credential with ID '%s' not found", req.Msg.Id) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete credential: %w", err))
	}

	return connect.NewResponse(&credentialv1.DeleteCredentialResponse{
		Message: fmt.Sprintf("Credential with ID '%s' deleted successfully", req.Msg.Id),
	}), nil
}

// Helper functions to convert bson.M to typed credentials
func convertBsonToGcpCredential(doc bson.M) (*models.GcpCredential, error) {
	id, ok := doc["_id"].(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid _id field")
	}

	var createdAt, updatedAt time.Time
	if dt, ok := doc["created_at"].(primitive.DateTime); ok {
		createdAt = dt.Time()
	} else if t, ok := doc["created_at"].(time.Time); ok {
		createdAt = t
	}
	if dt, ok := doc["updated_at"].(primitive.DateTime); ok {
		updatedAt = dt.Time()
	} else if t, ok := doc["updated_at"].(time.Time); ok {
		updatedAt = t
	}

	return &models.GcpCredential{
		ID:                      id,
		Name:                    doc["name"].(string),
		ServiceAccountKeyBase64: doc["service_account_key_base64"].(string),
		CreatedAt:               createdAt,
		UpdatedAt:               updatedAt,
	}, nil
}

func convertBsonToAwsCredential(doc bson.M) (*models.AwsCredential, error) {
	id, ok := doc["_id"].(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid _id field")
	}

	var createdAt, updatedAt time.Time
	if dt, ok := doc["created_at"].(primitive.DateTime); ok {
		createdAt = dt.Time()
	} else if t, ok := doc["created_at"].(time.Time); ok {
		createdAt = t
	}
	if dt, ok := doc["updated_at"].(primitive.DateTime); ok {
		updatedAt = dt.Time()
	} else if t, ok := doc["updated_at"].(time.Time); ok {
		updatedAt = t
	}

	cred := &models.AwsCredential{
		ID:              id,
		Name:            doc["name"].(string),
		AccountID:       doc["account_id"].(string),
		AccessKeyID:     doc["access_key_id"].(string),
		SecretAccessKey: doc["secret_access_key"].(string),
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}

	if region, ok := doc["region"].(string); ok {
		cred.Region = region
	}
	if sessionToken, ok := doc["session_token"].(string); ok {
		cred.SessionToken = sessionToken
	}

	return cred, nil
}

func convertBsonToAzureCredential(doc bson.M) (*models.AzureCredential, error) {
	id, ok := doc["_id"].(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid _id field")
	}

	var createdAt, updatedAt time.Time
	if dt, ok := doc["created_at"].(primitive.DateTime); ok {
		createdAt = dt.Time()
	} else if t, ok := doc["created_at"].(time.Time); ok {
		createdAt = t
	}
	if dt, ok := doc["updated_at"].(primitive.DateTime); ok {
		updatedAt = dt.Time()
	} else if t, ok := doc["updated_at"].(time.Time); ok {
		updatedAt = t
	}

	return &models.AzureCredential{
		ID:             id,
		Name:           doc["name"].(string),
		ClientID:       doc["client_id"].(string),
		ClientSecret:   doc["client_secret"].(string),
		TenantID:       doc["tenant_id"].(string),
		SubscriptionID: doc["subscription_id"].(string),
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}, nil
}

func convertBsonToAuth0Credential(doc bson.M) (*models.Auth0Credential, error) {
	id, ok := doc["_id"].(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid _id field")
	}

	var createdAt, updatedAt time.Time
	if dt, ok := doc["created_at"].(primitive.DateTime); ok {
		createdAt = dt.Time()
	} else if t, ok := doc["created_at"].(time.Time); ok {
		createdAt = t
	}
	if dt, ok := doc["updated_at"].(primitive.DateTime); ok {
		updatedAt = dt.Time()
	} else if t, ok := doc["updated_at"].(time.Time); ok {
		updatedAt = t
	}

	return &models.Auth0Credential{
		ID:           id,
		Name:         doc["name"].(string),
		Domain:       doc["domain"].(string),
		ClientID:     doc["client_id"].(string),
		ClientSecret: doc["client_secret"].(string),
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, nil
}

// createOpenStackCredential creates an OpenStack credential.
func (s *CredentialService) createOpenStackCredential(
	ctx context.Context,
	name string,
	providerConfig *credentialv1.CredentialProviderConfig,
	now time.Time,
) (*connect.Response[credentialv1.CreateCredentialResponse], error) {
	if providerConfig == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("provider_config is required"))
	}
	osConfig, ok := providerConfig.Data.(*credentialv1.CredentialProviderConfig_Openstack)
	if !ok || osConfig == nil || osConfig.Openstack == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("openstack provider_config is required"))
	}
	if osConfig.Openstack.AuthUrl == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("auth_url is required"))
	}

	// Build the credential model from proto, determining the auth method from the oneof
	credModel, authMethod, err := openstackProtoToModel(name, osConfig.Openstack)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if authMethod == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("credentials are required: provide one of password, application_credential, or token"))
	}

	createdCredential, err := s.credentialRepo.CreateOpenStack(ctx, credModel)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create OpenStack credential: %w", err))
	}

	protoCredential := &credentialv1.Credential{
		Id:             createdCredential.ID.Hex(),
		Name:           createdCredential.Name,
		Provider:       credentialv1.Credential_OPENSTACK,
		ProviderConfig: openstackModelToProtoConfig(createdCredential),
	}
	if !createdCredential.CreatedAt.IsZero() {
		protoCredential.CreatedAt = timestamppb.New(createdCredential.CreatedAt)
	}
	if !createdCredential.UpdatedAt.IsZero() {
		protoCredential.UpdatedAt = timestamppb.New(createdCredential.UpdatedAt)
	}

	return connect.NewResponse(&credentialv1.CreateCredentialResponse{
		Credential: protoCredential,
	}), nil
}

// updateOpenStackCredential updates an OpenStack credential.
func (s *CredentialService) updateOpenStackCredential(
	ctx context.Context,
	id, name string,
	providerConfig *credentialv1.CredentialProviderConfig,
) (*connect.Response[credentialv1.UpdateCredentialResponse], error) {
	if providerConfig == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("provider_config is required"))
	}
	osConfig, ok := providerConfig.Data.(*credentialv1.CredentialProviderConfig_Openstack)
	if !ok || osConfig == nil || osConfig.Openstack == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("openstack provider_config is required"))
	}
	if osConfig.Openstack.AuthUrl == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("auth_url is required"))
	}

	credModel, authMethod, err := openstackProtoToModel(name, osConfig.Openstack)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if authMethod == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("credentials are required: provide one of password, application_credential, or token"))
	}

	updatedCredential, err := s.credentialRepo.UpdateOpenStack(ctx, id, credModel)
	if err != nil {
		if err.Error() == fmt.Sprintf("OpenStack credential with ID '%s' not found", id) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update OpenStack credential: %w", err))
	}

	protoCredential := &credentialv1.Credential{
		Id:             updatedCredential.ID.Hex(),
		Name:           updatedCredential.Name,
		Provider:       credentialv1.Credential_OPENSTACK,
		ProviderConfig: openstackModelToProtoConfig(updatedCredential),
	}
	if !updatedCredential.CreatedAt.IsZero() {
		protoCredential.CreatedAt = timestamppb.New(updatedCredential.CreatedAt)
	}
	if !updatedCredential.UpdatedAt.IsZero() {
		protoCredential.UpdatedAt = timestamppb.New(updatedCredential.UpdatedAt)
	}

	return connect.NewResponse(&credentialv1.UpdateCredentialResponse{
		Credential: protoCredential,
	}), nil
}

// openstackProtoToModel converts an OpenStackProviderConfig proto to the database model.
// Returns the model, the detected auth method name, and any validation error.
func openstackProtoToModel(name string, cfg *openstackv1.OpenStackProviderConfig) (*models.OpenStackCredential, string, error) {
	cred := &models.OpenStackCredential{
		Name:              name,
		AuthURL:           cfg.AuthUrl,
		Region:            cfg.Region,
		TenantName:        cfg.TenantName,
		TenantID:          cfg.TenantId,
		UserDomainName:    cfg.UserDomainName,
		UserDomainID:      cfg.UserDomainId,
		ProjectDomainName: cfg.ProjectDomainName,
		ProjectDomainID:   cfg.ProjectDomainId,
		Insecure:          cfg.Insecure,
		CACertFile:        cfg.CacertFile,
		EndpointType:      cfg.EndpointType,
	}

	var authMethod string
	switch c := cfg.Credentials.(type) {
	case *openstackv1.OpenStackProviderConfig_Password:
		if c.Password == nil {
			return nil, "", fmt.Errorf("password credentials block is empty")
		}
		if c.Password.UserName == "" {
			return nil, "", fmt.Errorf("user_name is required for password authentication")
		}
		if c.Password.Password == "" {
			return nil, "", fmt.Errorf("password is required for password authentication")
		}
		authMethod = "password"
		cred.AuthMethod = authMethod
		cred.UserName = c.Password.UserName
		cred.Password = c.Password.Password
	case *openstackv1.OpenStackProviderConfig_ApplicationCredential:
		if c.ApplicationCredential == nil {
			return nil, "", fmt.Errorf("application_credential block is empty")
		}
		if c.ApplicationCredential.Secret == "" {
			return nil, "", fmt.Errorf("secret is required for application credential authentication")
		}
		if c.ApplicationCredential.Id == "" && c.ApplicationCredential.Name == "" {
			return nil, "", fmt.Errorf("either id or name is required for application credential authentication")
		}
		authMethod = "application_credential"
		cred.AuthMethod = authMethod
		cred.ApplicationCredentialID = c.ApplicationCredential.Id
		cred.ApplicationCredentialName = c.ApplicationCredential.Name
		cred.ApplicationCredentialSecret = c.ApplicationCredential.Secret
	case *openstackv1.OpenStackProviderConfig_Token:
		if c.Token == nil {
			return nil, "", fmt.Errorf("token credentials block is empty")
		}
		if c.Token.Token == "" {
			return nil, "", fmt.Errorf("token is required for token authentication")
		}
		authMethod = "token"
		cred.AuthMethod = authMethod
		cred.Token = c.Token.Token
	}

	return cred, authMethod, nil
}

// openstackModelToProtoConfig converts an OpenStackCredential model back to proto CredentialProviderConfig.
func openstackModelToProtoConfig(cred *models.OpenStackCredential) *credentialv1.CredentialProviderConfig {
	cfg := &openstackv1.OpenStackProviderConfig{
		AuthUrl:           cred.AuthURL,
		Region:            cred.Region,
		TenantName:        cred.TenantName,
		TenantId:          cred.TenantID,
		UserDomainName:    cred.UserDomainName,
		UserDomainId:      cred.UserDomainID,
		ProjectDomainName: cred.ProjectDomainName,
		ProjectDomainId:   cred.ProjectDomainID,
		Insecure:          cred.Insecure,
		CacertFile:        cred.CACertFile,
		EndpointType:      cred.EndpointType,
	}

	switch cred.AuthMethod {
	case "password":
		cfg.Credentials = &openstackv1.OpenStackProviderConfig_Password{
			Password: &openstackv1.OpenStackPasswordCredentials{
				UserName: cred.UserName,
				Password: cred.Password,
			},
		}
	case "application_credential":
		cfg.Credentials = &openstackv1.OpenStackProviderConfig_ApplicationCredential{
			ApplicationCredential: &openstackv1.OpenStackApplicationCredentials{
				Id:     cred.ApplicationCredentialID,
				Name:   cred.ApplicationCredentialName,
				Secret: cred.ApplicationCredentialSecret,
			},
		}
	case "token":
		cfg.Credentials = &openstackv1.OpenStackProviderConfig_Token{
			Token: &openstackv1.OpenStackTokenCredentials{
				Token: cred.Token,
			},
		}
	}

	return &credentialv1.CredentialProviderConfig{
		Data: &credentialv1.CredentialProviderConfig_Openstack{
			Openstack: cfg,
		},
	}
}

// createScalewayCredential creates a Scaleway credential.
func (s *CredentialService) createScalewayCredential(
	ctx context.Context,
	name string,
	providerConfig *credentialv1.CredentialProviderConfig,
	now time.Time,
) (*connect.Response[credentialv1.CreateCredentialResponse], error) {
	if providerConfig == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("provider_config is required"))
	}
	scwConfig, ok := providerConfig.Data.(*credentialv1.CredentialProviderConfig_Scaleway)
	if !ok || scwConfig == nil || scwConfig.Scaleway == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("scaleway provider_config is required"))
	}
	if scwConfig.Scaleway.AccessKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("access_key is required"))
	}
	if scwConfig.Scaleway.SecretKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("secret_key is required"))
	}

	credModel := scalewayProtoToModel(name, scwConfig.Scaleway)

	createdCredential, err := s.credentialRepo.CreateScaleway(ctx, credModel)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create Scaleway credential: %w", err))
	}

	protoCredential := &credentialv1.Credential{
		Id:             createdCredential.ID.Hex(),
		Name:           createdCredential.Name,
		Provider:       credentialv1.Credential_SCALEWAY,
		ProviderConfig: scalewayModelToProtoConfig(createdCredential),
	}
	if !createdCredential.CreatedAt.IsZero() {
		protoCredential.CreatedAt = timestamppb.New(createdCredential.CreatedAt)
	}
	if !createdCredential.UpdatedAt.IsZero() {
		protoCredential.UpdatedAt = timestamppb.New(createdCredential.UpdatedAt)
	}

	return connect.NewResponse(&credentialv1.CreateCredentialResponse{
		Credential: protoCredential,
	}), nil
}

// updateScalewayCredential updates a Scaleway credential.
func (s *CredentialService) updateScalewayCredential(
	ctx context.Context,
	id, name string,
	providerConfig *credentialv1.CredentialProviderConfig,
) (*connect.Response[credentialv1.UpdateCredentialResponse], error) {
	if providerConfig == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("provider_config is required"))
	}
	scwConfig, ok := providerConfig.Data.(*credentialv1.CredentialProviderConfig_Scaleway)
	if !ok || scwConfig == nil || scwConfig.Scaleway == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("scaleway provider_config is required"))
	}
	if scwConfig.Scaleway.AccessKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("access_key is required"))
	}
	if scwConfig.Scaleway.SecretKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("secret_key is required"))
	}

	credModel := scalewayProtoToModel(name, scwConfig.Scaleway)

	updatedCredential, err := s.credentialRepo.UpdateScaleway(ctx, id, credModel)
	if err != nil {
		if err.Error() == fmt.Sprintf("Scaleway credential with ID '%s' not found", id) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update Scaleway credential: %w", err))
	}

	protoCredential := &credentialv1.Credential{
		Id:             updatedCredential.ID.Hex(),
		Name:           updatedCredential.Name,
		Provider:       credentialv1.Credential_SCALEWAY,
		ProviderConfig: scalewayModelToProtoConfig(updatedCredential),
	}
	if !updatedCredential.CreatedAt.IsZero() {
		protoCredential.CreatedAt = timestamppb.New(updatedCredential.CreatedAt)
	}
	if !updatedCredential.UpdatedAt.IsZero() {
		protoCredential.UpdatedAt = timestamppb.New(updatedCredential.UpdatedAt)
	}

	return connect.NewResponse(&credentialv1.UpdateCredentialResponse{
		Credential: protoCredential,
	}), nil
}

// scalewayProtoToModel converts a ScalewayProviderConfig proto to the database model.
func scalewayProtoToModel(name string, cfg *scalewayv1.ScalewayProviderConfig) *models.ScalewayCredential {
	return &models.ScalewayCredential{
		Name:           name,
		AccessKey:      cfg.AccessKey,
		SecretKey:      cfg.SecretKey,
		ProjectID:      cfg.ProjectId,
		OrganizationID: cfg.OrganizationId,
		Region:         cfg.Region,
		Zone:           cfg.Zone,
	}
}

// scalewayModelToProtoConfig converts a ScalewayCredential model back to proto CredentialProviderConfig.
func scalewayModelToProtoConfig(cred *models.ScalewayCredential) *credentialv1.CredentialProviderConfig {
	return &credentialv1.CredentialProviderConfig{
		Data: &credentialv1.CredentialProviderConfig_Scaleway{
			Scaleway: &scalewayv1.ScalewayProviderConfig{
				AccessKey:      cred.AccessKey,
				SecretKey:      cred.SecretKey,
				ProjectId:      cred.ProjectID,
				OrganizationId: cred.OrganizationID,
				Region:         cred.Region,
				Zone:           cred.Zone,
			},
		},
	}
}

func convertBsonToScalewayCredential(doc bson.M) (*models.ScalewayCredential, error) {
	id, ok := doc["_id"].(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid _id field")
	}

	var createdAt, updatedAt time.Time
	if dt, ok := doc["created_at"].(primitive.DateTime); ok {
		createdAt = dt.Time()
	} else if t, ok := doc["created_at"].(time.Time); ok {
		createdAt = t
	}
	if dt, ok := doc["updated_at"].(primitive.DateTime); ok {
		updatedAt = dt.Time()
	} else if t, ok := doc["updated_at"].(time.Time); ok {
		updatedAt = t
	}

	cred := &models.ScalewayCredential{
		ID:        id,
		Name:      doc["name"].(string),
		AccessKey: doc["access_key"].(string),
		SecretKey: doc["secret_key"].(string),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if v, ok := doc["project_id"].(string); ok {
		cred.ProjectID = v
	}
	if v, ok := doc["organization_id"].(string); ok {
		cred.OrganizationID = v
	}
	if v, ok := doc["region"].(string); ok {
		cred.Region = v
	}
	if v, ok := doc["zone"].(string); ok {
		cred.Zone = v
	}

	return cred, nil
}

func convertBsonToOpenStackCredential(doc bson.M) (*models.OpenStackCredential, error) {
	id, ok := doc["_id"].(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid _id field")
	}

	var createdAt, updatedAt time.Time
	if dt, ok := doc["created_at"].(primitive.DateTime); ok {
		createdAt = dt.Time()
	} else if t, ok := doc["created_at"].(time.Time); ok {
		createdAt = t
	}
	if dt, ok := doc["updated_at"].(primitive.DateTime); ok {
		updatedAt = dt.Time()
	} else if t, ok := doc["updated_at"].(time.Time); ok {
		updatedAt = t
	}

	cred := &models.OpenStackCredential{
		ID:        id,
		Name:      doc["name"].(string),
		AuthURL:   doc["auth_url"].(string),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if v, ok := doc["auth_method"].(string); ok {
		cred.AuthMethod = v
	}
	if v, ok := doc["region"].(string); ok {
		cred.Region = v
	}
	if v, ok := doc["user_name"].(string); ok {
		cred.UserName = v
	}
	if v, ok := doc["password"].(string); ok {
		cred.Password = v
	}
	if v, ok := doc["application_credential_id"].(string); ok {
		cred.ApplicationCredentialID = v
	}
	if v, ok := doc["application_credential_name"].(string); ok {
		cred.ApplicationCredentialName = v
	}
	if v, ok := doc["application_credential_secret"].(string); ok {
		cred.ApplicationCredentialSecret = v
	}
	if v, ok := doc["token"].(string); ok {
		cred.Token = v
	}
	if v, ok := doc["tenant_name"].(string); ok {
		cred.TenantName = v
	}
	if v, ok := doc["tenant_id"].(string); ok {
		cred.TenantID = v
	}
	if v, ok := doc["user_domain_name"].(string); ok {
		cred.UserDomainName = v
	}
	if v, ok := doc["user_domain_id"].(string); ok {
		cred.UserDomainID = v
	}
	if v, ok := doc["project_domain_name"].(string); ok {
		cred.ProjectDomainName = v
	}
	if v, ok := doc["project_domain_id"].(string); ok {
		cred.ProjectDomainID = v
	}
	if v, ok := doc["insecure"].(bool); ok {
		cred.Insecure = v
	}
	if v, ok := doc["cacert_file"].(string); ok {
		cred.CACertFile = v
	}
	if v, ok := doc["endpoint_type"].(string); ok {
		cred.EndpointType = v
	}

	return cred, nil
}

// createAlicloudCredential creates an Alibaba Cloud credential.
func (s *CredentialService) createAlicloudCredential(
	ctx context.Context,
	name string,
	providerConfig *credentialv1.CredentialProviderConfig,
	now time.Time,
) (*connect.Response[credentialv1.CreateCredentialResponse], error) {
	if providerConfig == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("provider_config is required"))
	}
	aliConfig, ok := providerConfig.Data.(*credentialv1.CredentialProviderConfig_Alicloud)
	if !ok || aliConfig == nil || aliConfig.Alicloud == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("alicloud provider_config is required"))
	}

	credModel, err := alicloudProtoToModel(name, aliConfig.Alicloud)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	createdCredential, err := s.credentialRepo.CreateAlicloud(ctx, credModel)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create Alibaba Cloud credential: %w", err))
	}

	protoCredential := &credentialv1.Credential{
		Id:             createdCredential.ID.Hex(),
		Name:           createdCredential.Name,
		Provider:       credentialv1.Credential_ALICLOUD,
		ProviderConfig: alicloudModelToProtoConfig(createdCredential),
	}
	if !createdCredential.CreatedAt.IsZero() {
		protoCredential.CreatedAt = timestamppb.New(createdCredential.CreatedAt)
	}
	if !createdCredential.UpdatedAt.IsZero() {
		protoCredential.UpdatedAt = timestamppb.New(createdCredential.UpdatedAt)
	}

	return connect.NewResponse(&credentialv1.CreateCredentialResponse{
		Credential: protoCredential,
	}), nil
}

// updateAlicloudCredential updates an Alibaba Cloud credential.
func (s *CredentialService) updateAlicloudCredential(
	ctx context.Context,
	id, name string,
	providerConfig *credentialv1.CredentialProviderConfig,
) (*connect.Response[credentialv1.UpdateCredentialResponse], error) {
	if providerConfig == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("provider_config is required"))
	}
	aliConfig, ok := providerConfig.Data.(*credentialv1.CredentialProviderConfig_Alicloud)
	if !ok || aliConfig == nil || aliConfig.Alicloud == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("alicloud provider_config is required"))
	}

	credModel, err := alicloudProtoToModel(name, aliConfig.Alicloud)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	updatedCredential, err := s.credentialRepo.UpdateAlicloud(ctx, id, credModel)
	if err != nil {
		if err.Error() == fmt.Sprintf("Alibaba Cloud credential with ID '%s' not found", id) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update Alibaba Cloud credential: %w", err))
	}

	protoCredential := &credentialv1.Credential{
		Id:             updatedCredential.ID.Hex(),
		Name:           updatedCredential.Name,
		Provider:       credentialv1.Credential_ALICLOUD,
		ProviderConfig: alicloudModelToProtoConfig(updatedCredential),
	}
	if !updatedCredential.CreatedAt.IsZero() {
		protoCredential.CreatedAt = timestamppb.New(updatedCredential.CreatedAt)
	}
	if !updatedCredential.UpdatedAt.IsZero() {
		protoCredential.UpdatedAt = timestamppb.New(updatedCredential.UpdatedAt)
	}

	return connect.NewResponse(&credentialv1.UpdateCredentialResponse{
		Credential: protoCredential,
	}), nil
}

// alicloudProtoToModel converts an AlicloudProviderConfig proto to the database model.
// Switches on the AuthenticationType enum to extract fields from the correct sub-message.
func alicloudProtoToModel(name string, cfg *alicloudv1.AlicloudProviderConfig) (*models.AlicloudCredential, error) {
	cred := &models.AlicloudCredential{
		Name:        name,
		Region:      cfg.Region,
		AccountId:   cfg.AccountId,
		AccountType: cfg.AccountType,
	}

	switch cfg.AuthenticationType {
	case alicloudv1.AuthenticationType_static_credentials:
		if cfg.StaticCredentials == nil {
			return nil, fmt.Errorf("static_credentials message is required when authentication_type is static_credentials")
		}
		if cfg.StaticCredentials.AccessKey == "" {
			return nil, fmt.Errorf("access_key is required for static credentials")
		}
		if cfg.StaticCredentials.SecretKey == "" {
			return nil, fmt.Errorf("secret_key is required for static credentials")
		}
		cred.AuthMethod = "static_credentials"
		cred.AccessKey = cfg.StaticCredentials.AccessKey
		cred.SecretKey = cfg.StaticCredentials.SecretKey

	case alicloudv1.AuthenticationType_sts_token:
		if cfg.StsToken == nil {
			return nil, fmt.Errorf("sts_token message is required when authentication_type is sts_token")
		}
		if cfg.StsToken.AccessKey == "" {
			return nil, fmt.Errorf("access_key is required for STS token authentication")
		}
		if cfg.StsToken.SecretKey == "" {
			return nil, fmt.Errorf("secret_key is required for STS token authentication")
		}
		if cfg.StsToken.SecurityToken == "" {
			return nil, fmt.Errorf("security_token is required for STS token authentication")
		}
		cred.AuthMethod = "sts_token"
		cred.AccessKey = cfg.StsToken.AccessKey
		cred.SecretKey = cfg.StsToken.SecretKey
		cred.SecurityToken = cfg.StsToken.SecurityToken

	case alicloudv1.AuthenticationType_ecs_role:
		if cfg.EcsRole == nil {
			return nil, fmt.Errorf("ecs_role message is required when authentication_type is ecs_role")
		}
		if cfg.EcsRole.EcsRoleName == "" {
			return nil, fmt.Errorf("ecs_role_name is required for ECS role authentication")
		}
		cred.AuthMethod = "ecs_role"
		cred.EcsRoleName = cfg.EcsRole.EcsRoleName

	case alicloudv1.AuthenticationType_assume_role:
		if cfg.AssumeRole == nil {
			return nil, fmt.Errorf("assume_role message is required when authentication_type is assume_role")
		}
		if cfg.AssumeRole.AccessKey == "" {
			return nil, fmt.Errorf("access_key is required for assume role authentication")
		}
		if cfg.AssumeRole.SecretKey == "" {
			return nil, fmt.Errorf("secret_key is required for assume role authentication")
		}
		if cfg.AssumeRole.RoleArn == "" {
			return nil, fmt.Errorf("role_arn is required for assume role authentication")
		}
		cred.AuthMethod = "assume_role"
		cred.AccessKey = cfg.AssumeRole.AccessKey
		cred.SecretKey = cfg.AssumeRole.SecretKey
		cred.RoleArn = cfg.AssumeRole.RoleArn
		cred.SessionName = cfg.AssumeRole.SessionName
		cred.Policy = cfg.AssumeRole.Policy
		cred.SessionExpiration = cfg.AssumeRole.SessionExpiration
		cred.ExternalId = cfg.AssumeRole.ExternalId

	case alicloudv1.AuthenticationType_assume_role_with_oidc:
		if cfg.AssumeRoleWithOidc == nil {
			return nil, fmt.Errorf("assume_role_with_oidc message is required when authentication_type is assume_role_with_oidc")
		}
		if cfg.AssumeRoleWithOidc.OidcProviderArn == "" {
			return nil, fmt.Errorf("oidc_provider_arn is required for OIDC authentication")
		}
		if cfg.AssumeRoleWithOidc.RoleArn == "" {
			return nil, fmt.Errorf("role_arn is required for OIDC authentication")
		}
		if cfg.AssumeRoleWithOidc.OidcToken == "" && cfg.AssumeRoleWithOidc.OidcTokenFile == "" {
			return nil, fmt.Errorf("either oidc_token or oidc_token_file is required for OIDC authentication")
		}
		cred.AuthMethod = "assume_role_with_oidc"
		cred.OidcProviderArn = cfg.AssumeRoleWithOidc.OidcProviderArn
		cred.RoleArn = cfg.AssumeRoleWithOidc.RoleArn
		cred.OidcToken = cfg.AssumeRoleWithOidc.OidcToken
		cred.OidcTokenFile = cfg.AssumeRoleWithOidc.OidcTokenFile
		cred.SessionName = cfg.AssumeRoleWithOidc.SessionName
		cred.Policy = cfg.AssumeRoleWithOidc.Policy
		cred.SessionExpiration = cfg.AssumeRoleWithOidc.SessionExpiration

	case alicloudv1.AuthenticationType_shared_credentials:
		if cfg.SharedCredentials == nil {
			return nil, fmt.Errorf("shared_credentials message is required when authentication_type is shared_credentials")
		}
		cred.AuthMethod = "shared_credentials"
		cred.CredentialsFile = cfg.SharedCredentials.CredentialsFile
		cred.Profile = cfg.SharedCredentials.Profile

	case alicloudv1.AuthenticationType_sidecar_credentials:
		if cfg.SidecarCredentials == nil {
			return nil, fmt.Errorf("sidecar_credentials message is required when authentication_type is sidecar_credentials")
		}
		if cfg.SidecarCredentials.CredentialsUri == "" {
			return nil, fmt.Errorf("credentials_uri is required for sidecar credentials")
		}
		cred.AuthMethod = "sidecar_credentials"
		cred.CredentialsUri = cfg.SidecarCredentials.CredentialsUri

	default:
		return nil, fmt.Errorf("authentication_type is required: specify one of static_credentials, sts_token, ecs_role, assume_role, assume_role_with_oidc, shared_credentials, or sidecar_credentials")
	}

	return cred, nil
}

// alicloudModelToProtoConfig converts an AlicloudCredential model back to proto CredentialProviderConfig.
func alicloudModelToProtoConfig(cred *models.AlicloudCredential) *credentialv1.CredentialProviderConfig {
	cfg := &alicloudv1.AlicloudProviderConfig{
		Region:      cred.Region,
		AccountId:   cred.AccountId,
		AccountType: cred.AccountType,
	}

	switch cred.AuthMethod {
	case "static_credentials":
		cfg.AuthenticationType = alicloudv1.AuthenticationType_static_credentials
		cfg.StaticCredentials = &alicloudv1.AlicloudStaticCredentials{
			AccessKey: cred.AccessKey,
			SecretKey: cred.SecretKey,
		}
	case "sts_token":
		cfg.AuthenticationType = alicloudv1.AuthenticationType_sts_token
		cfg.StsToken = &alicloudv1.AlicloudStsTokenCredentials{
			AccessKey:     cred.AccessKey,
			SecretKey:     cred.SecretKey,
			SecurityToken: cred.SecurityToken,
		}
	case "ecs_role":
		cfg.AuthenticationType = alicloudv1.AuthenticationType_ecs_role
		cfg.EcsRole = &alicloudv1.AlicloudEcsRoleCredentials{
			EcsRoleName: cred.EcsRoleName,
		}
	case "assume_role":
		cfg.AuthenticationType = alicloudv1.AuthenticationType_assume_role
		cfg.AssumeRole = &alicloudv1.AlicloudAssumeRoleCredentials{
			AccessKey:         cred.AccessKey,
			SecretKey:         cred.SecretKey,
			RoleArn:           cred.RoleArn,
			SessionName:       cred.SessionName,
			Policy:            cred.Policy,
			SessionExpiration: cred.SessionExpiration,
			ExternalId:        cred.ExternalId,
		}
	case "assume_role_with_oidc":
		cfg.AuthenticationType = alicloudv1.AuthenticationType_assume_role_with_oidc
		cfg.AssumeRoleWithOidc = &alicloudv1.AlicloudAssumeRoleWithOidcCredentials{
			OidcProviderArn:   cred.OidcProviderArn,
			RoleArn:           cred.RoleArn,
			OidcToken:         cred.OidcToken,
			OidcTokenFile:     cred.OidcTokenFile,
			SessionName:       cred.SessionName,
			Policy:            cred.Policy,
			SessionExpiration: cred.SessionExpiration,
		}
	case "shared_credentials":
		cfg.AuthenticationType = alicloudv1.AuthenticationType_shared_credentials
		cfg.SharedCredentials = &alicloudv1.AlicloudSharedCredentials{
			CredentialsFile: cred.CredentialsFile,
			Profile:         cred.Profile,
		}
	case "sidecar_credentials":
		cfg.AuthenticationType = alicloudv1.AuthenticationType_sidecar_credentials
		cfg.SidecarCredentials = &alicloudv1.AlicloudSidecarCredentials{
			CredentialsUri: cred.CredentialsUri,
		}
	}

	return &credentialv1.CredentialProviderConfig{
		Data: &credentialv1.CredentialProviderConfig_Alicloud{
			Alicloud: cfg,
		},
	}
}

func convertBsonToAlicloudCredential(doc bson.M) (*models.AlicloudCredential, error) {
	id, ok := doc["_id"].(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid _id field")
	}

	var createdAt, updatedAt time.Time
	if dt, ok := doc["created_at"].(primitive.DateTime); ok {
		createdAt = dt.Time()
	} else if t, ok := doc["created_at"].(time.Time); ok {
		createdAt = t
	}
	if dt, ok := doc["updated_at"].(primitive.DateTime); ok {
		updatedAt = dt.Time()
	} else if t, ok := doc["updated_at"].(time.Time); ok {
		updatedAt = t
	}

	cred := &models.AlicloudCredential{
		ID:        id,
		Name:      doc["name"].(string),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if v, ok := doc["auth_method"].(string); ok {
		cred.AuthMethod = v
	}
	if v, ok := doc["region"].(string); ok {
		cred.Region = v
	}
	if v, ok := doc["account_id"].(string); ok {
		cred.AccountId = v
	}
	if v, ok := doc["account_type"].(string); ok {
		cred.AccountType = v
	}
	if v, ok := doc["access_key"].(string); ok {
		cred.AccessKey = v
	}
	if v, ok := doc["secret_key"].(string); ok {
		cred.SecretKey = v
	}
	if v, ok := doc["security_token"].(string); ok {
		cred.SecurityToken = v
	}
	if v, ok := doc["ecs_role_name"].(string); ok {
		cred.EcsRoleName = v
	}
	if v, ok := doc["role_arn"].(string); ok {
		cred.RoleArn = v
	}
	if v, ok := doc["session_name"].(string); ok {
		cred.SessionName = v
	}
	if v, ok := doc["policy"].(string); ok {
		cred.Policy = v
	}
	if v, ok := doc["session_expiration"].(int32); ok {
		cred.SessionExpiration = v
	}
	if v, ok := doc["external_id"].(string); ok {
		cred.ExternalId = v
	}
	if v, ok := doc["oidc_provider_arn"].(string); ok {
		cred.OidcProviderArn = v
	}
	if v, ok := doc["oidc_token"].(string); ok {
		cred.OidcToken = v
	}
	if v, ok := doc["oidc_token_file"].(string); ok {
		cred.OidcTokenFile = v
	}
	if v, ok := doc["credentials_file"].(string); ok {
		cred.CredentialsFile = v
	}
	if v, ok := doc["profile"].(string); ok {
		cred.Profile = v
	}
	if v, ok := doc["credentials_uri"].(string); ok {
		cred.CredentialsUri = v
	}

	return cred, nil
}

// createOciCredential creates an OCI credential.
func (s *CredentialService) createOciCredential(
	ctx context.Context,
	name string,
	providerConfig *credentialv1.CredentialProviderConfig,
	now time.Time,
) (*connect.Response[credentialv1.CreateCredentialResponse], error) {
	if providerConfig == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("provider_config is required"))
	}
	ociConfig, ok := providerConfig.Data.(*credentialv1.CredentialProviderConfig_Oci)
	if !ok || ociConfig == nil || ociConfig.Oci == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("oci provider_config is required"))
	}

	credModel, err := ociProtoToModel(name, ociConfig.Oci)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	createdCredential, err := s.credentialRepo.CreateOci(ctx, credModel)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create OCI credential: %w", err))
	}

	protoCredential := &credentialv1.Credential{
		Id:             createdCredential.ID.Hex(),
		Name:           createdCredential.Name,
		Provider:       credentialv1.Credential_OCI,
		ProviderConfig: ociModelToProtoConfig(createdCredential),
	}
	if !createdCredential.CreatedAt.IsZero() {
		protoCredential.CreatedAt = timestamppb.New(createdCredential.CreatedAt)
	}
	if !createdCredential.UpdatedAt.IsZero() {
		protoCredential.UpdatedAt = timestamppb.New(createdCredential.UpdatedAt)
	}

	return connect.NewResponse(&credentialv1.CreateCredentialResponse{
		Credential: protoCredential,
	}), nil
}

// updateOciCredential updates an OCI credential.
func (s *CredentialService) updateOciCredential(
	ctx context.Context,
	id, name string,
	providerConfig *credentialv1.CredentialProviderConfig,
) (*connect.Response[credentialv1.UpdateCredentialResponse], error) {
	if providerConfig == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("provider_config is required"))
	}
	ociConfig, ok := providerConfig.Data.(*credentialv1.CredentialProviderConfig_Oci)
	if !ok || ociConfig == nil || ociConfig.Oci == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("oci provider_config is required"))
	}

	credModel, err := ociProtoToModel(name, ociConfig.Oci)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	updatedCredential, err := s.credentialRepo.UpdateOci(ctx, id, credModel)
	if err != nil {
		if err.Error() == fmt.Sprintf("OCI credential with ID '%s' not found", id) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update OCI credential: %w", err))
	}

	protoCredential := &credentialv1.Credential{
		Id:             updatedCredential.ID.Hex(),
		Name:           updatedCredential.Name,
		Provider:       credentialv1.Credential_OCI,
		ProviderConfig: ociModelToProtoConfig(updatedCredential),
	}
	if !updatedCredential.CreatedAt.IsZero() {
		protoCredential.CreatedAt = timestamppb.New(updatedCredential.CreatedAt)
	}
	if !updatedCredential.UpdatedAt.IsZero() {
		protoCredential.UpdatedAt = timestamppb.New(updatedCredential.UpdatedAt)
	}

	return connect.NewResponse(&credentialv1.UpdateCredentialResponse{
		Credential: protoCredential,
	}), nil
}

// ociProtoToModel converts an OciProviderConfig proto to the database model.
func ociProtoToModel(name string, cfg *ociv1.OciProviderConfig) (*models.OciCredential, error) {
	cred := &models.OciCredential{
		Name:   name,
		Region: cfg.Region,
	}

	switch cfg.AuthenticationType {
	case ociv1.AuthenticationType_api_key:
		if cfg.ApiKey == nil {
			return nil, fmt.Errorf("api_key message is required when authentication_type is api_key")
		}
		if cfg.ApiKey.TenancyOcid == "" {
			return nil, fmt.Errorf("tenancy_ocid is required for API key authentication")
		}
		if cfg.ApiKey.UserOcid == "" {
			return nil, fmt.Errorf("user_ocid is required for API key authentication")
		}
		if cfg.ApiKey.Fingerprint == "" {
			return nil, fmt.Errorf("fingerprint is required for API key authentication")
		}
		if cfg.ApiKey.PrivateKey == "" && cfg.ApiKey.PrivateKeyPath == "" {
			return nil, fmt.Errorf("either private_key or private_key_path is required for API key authentication")
		}
		cred.AuthMethod = "api_key"
		cred.TenancyOcid = cfg.ApiKey.TenancyOcid
		cred.UserOcid = cfg.ApiKey.UserOcid
		cred.Fingerprint = cfg.ApiKey.Fingerprint
		cred.PrivateKey = cfg.ApiKey.PrivateKey
		cred.PrivateKeyPassword = cfg.ApiKey.PrivateKeyPassword

	case ociv1.AuthenticationType_security_token:
		if cfg.SecurityToken == nil {
			return nil, fmt.Errorf("security_token message is required when authentication_type is security_token")
		}
		if cfg.SecurityToken.ConfigFileProfile == "" {
			return nil, fmt.Errorf("config_file_profile is required for security token authentication")
		}
		cred.AuthMethod = "security_token"
		cred.ConfigFileProfile = cfg.SecurityToken.ConfigFileProfile
		cred.PrivateKeyPassword = cfg.SecurityToken.PrivateKeyPassword

	case ociv1.AuthenticationType_instance_principal:
		cred.AuthMethod = "instance_principal"

	case ociv1.AuthenticationType_resource_principal:
		cred.AuthMethod = "resource_principal"

	case ociv1.AuthenticationType_oke_workload_identity:
		cred.AuthMethod = "oke_workload_identity"

	default:
		return nil, fmt.Errorf("authentication_type is required: specify one of api_key, instance_principal, security_token, resource_principal, or oke_workload_identity")
	}

	return cred, nil
}

// ociModelToProtoConfig converts an OciCredential model back to proto CredentialProviderConfig.
func ociModelToProtoConfig(cred *models.OciCredential) *credentialv1.CredentialProviderConfig {
	cfg := &ociv1.OciProviderConfig{
		Region: cred.Region,
	}

	switch cred.AuthMethod {
	case "api_key":
		cfg.AuthenticationType = ociv1.AuthenticationType_api_key
		cfg.ApiKey = &ociv1.OciApiKeyAuth{
			TenancyOcid:        cred.TenancyOcid,
			UserOcid:           cred.UserOcid,
			Fingerprint:        cred.Fingerprint,
			PrivateKey:         cred.PrivateKey,
			PrivateKeyPassword: cred.PrivateKeyPassword,
		}
	case "security_token":
		cfg.AuthenticationType = ociv1.AuthenticationType_security_token
		cfg.SecurityToken = &ociv1.OciSecurityTokenAuth{
			ConfigFileProfile:  cred.ConfigFileProfile,
			PrivateKeyPassword: cred.PrivateKeyPassword,
		}
	case "instance_principal":
		cfg.AuthenticationType = ociv1.AuthenticationType_instance_principal
	case "resource_principal":
		cfg.AuthenticationType = ociv1.AuthenticationType_resource_principal
	case "oke_workload_identity":
		cfg.AuthenticationType = ociv1.AuthenticationType_oke_workload_identity
	}

	return &credentialv1.CredentialProviderConfig{
		Data: &credentialv1.CredentialProviderConfig_Oci{
			Oci: cfg,
		},
	}
}

func convertBsonToOciCredential(doc bson.M) (*models.OciCredential, error) {
	id, ok := doc["_id"].(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid _id field")
	}

	var createdAt, updatedAt time.Time
	if dt, ok := doc["created_at"].(primitive.DateTime); ok {
		createdAt = dt.Time()
	} else if t, ok := doc["created_at"].(time.Time); ok {
		createdAt = t
	}
	if dt, ok := doc["updated_at"].(primitive.DateTime); ok {
		updatedAt = dt.Time()
	} else if t, ok := doc["updated_at"].(time.Time); ok {
		updatedAt = t
	}

	cred := &models.OciCredential{
		ID:        id,
		Name:      doc["name"].(string),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if v, ok := doc["auth_method"].(string); ok {
		cred.AuthMethod = v
	}
	if v, ok := doc["region"].(string); ok {
		cred.Region = v
	}
	if v, ok := doc["tenancy_ocid"].(string); ok {
		cred.TenancyOcid = v
	}
	if v, ok := doc["user_ocid"].(string); ok {
		cred.UserOcid = v
	}
	if v, ok := doc["fingerprint"].(string); ok {
		cred.Fingerprint = v
	}
	if v, ok := doc["private_key"].(string); ok {
		cred.PrivateKey = v
	}
	if v, ok := doc["private_key_password"].(string); ok {
		cred.PrivateKeyPassword = v
	}
	if v, ok := doc["config_file_profile"].(string); ok {
		cred.ConfigFileProfile = v
	}

	return cred, nil
}
