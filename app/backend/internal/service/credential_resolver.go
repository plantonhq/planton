package service

import (
	"context"
	"fmt"
	"strings"

	credentialv1 "github.com/plantonhq/openmcf/apis/org/openmcf/app/credential/v1"
	alicloudv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud"
	auth0v1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/auth0"
	ociv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci"
	awsv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws"
	azurev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure"
	gcpv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp"
	openstackv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack"
	scalewayv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/app/backend/internal/database"
	"github.com/plantonhq/openmcf/app/backend/pkg/models"
	"github.com/plantonhq/openmcf/pkg/crkreflect"
)

// CredentialResolver resolves provider credentials from the database based on provider.
type CredentialResolver struct {
	credentialRepo *database.CredentialRepository
}

// NewCredentialResolver creates a new credential resolver instance.
func NewCredentialResolver(credentialRepo *database.CredentialRepository) *CredentialResolver {
	return &CredentialResolver{
		credentialRepo: credentialRepo,
	}
}

// ResolveProviderConfig resolves provider credentials from the database based on the provider from cloud resource kind.
// Returns a CredentialProviderConfig proto message that can be used for deployment.
func (r *CredentialResolver) ResolveProviderConfig(
	ctx context.Context,
	kindName string,
) (*credentialv1.CredentialProviderConfig, error) {
	// Step 1: Get the CloudResourceKind enum from kind name
	kindEnum, err := crkreflect.KindByKindName(kindName)
	if err != nil {
		return nil, fmt.Errorf("failed to get kind enum for '%s': %w", kindName, err)
	}

	// Step 2: Get the provider from the kind
	providerEnum := crkreflect.GetProvider(kindEnum)
	if providerEnum == cloudresourcekind.CloudResourceProvider_cloud_resource_provider_unspecified {
		return nil, fmt.Errorf("provider not configured for cloud resource kind '%s'", kindName)
	}

	// Step 3: Convert provider enum to string (e.g., "aws", "gcp", "azure")
	providerString := providerEnumToString(providerEnum)
	if providerString == "" {
		return nil, fmt.Errorf("unsupported provider: %v", providerEnum)
	}

	// Step 4: Fetch the first credential from the repository based on provider
	credInterface, err := r.credentialRepo.FindFirstByProvider(ctx, providerString)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s credential: %w", providerString, err)
	}
	if credInterface == nil {
		return nil, fmt.Errorf("no %s credential found. Please create a %s credential first", strings.ToUpper(providerString), strings.ToUpper(providerString))
	}

	// Convert to CredentialProviderConfig based on provider enum.
	// Switch on the enum directly for type safety -- the string is only needed for the DB query above.
	switch providerEnum {
	case cloudresourcekind.CloudResourceProvider_aws:
		awsCred := credInterface.(*models.AwsCredential)
		return &credentialv1.CredentialProviderConfig{
			Data: &credentialv1.CredentialProviderConfig_Aws{
				Aws: &awsv1.AwsProviderConfig{
					AccountId:       awsCred.AccountID,
					AccessKeyId:     awsCred.AccessKeyID,
					SecretAccessKey: awsCred.SecretAccessKey,
					Region:          awsCred.Region,
					SessionToken:    awsCred.SessionToken,
				},
			},
		}, nil

	case cloudresourcekind.CloudResourceProvider_gcp:
		gcpCred := credInterface.(*models.GcpCredential)
		return &credentialv1.CredentialProviderConfig{
			Data: &credentialv1.CredentialProviderConfig_Gcp{
				Gcp: &gcpv1.GcpProviderConfig{
					ServiceAccountKeyBase64: gcpCred.ServiceAccountKeyBase64,
				},
			},
		}, nil

	case cloudresourcekind.CloudResourceProvider_azure:
		azureCred := credInterface.(*models.AzureCredential)
		return &credentialv1.CredentialProviderConfig{
			Data: &credentialv1.CredentialProviderConfig_Azure{
				Azure: &azurev1.AzureProviderConfig{
					ClientId:       azureCred.ClientID,
					ClientSecret:   azureCred.ClientSecret,
					TenantId:       azureCred.TenantID,
					SubscriptionId: azureCred.SubscriptionID,
				},
			},
		}, nil

	case cloudresourcekind.CloudResourceProvider_auth0:
		auth0Cred := credInterface.(*models.Auth0Credential)
		return &credentialv1.CredentialProviderConfig{
			Data: &credentialv1.CredentialProviderConfig_Auth0{
				Auth0: &auth0v1.Auth0ProviderConfig{
					Domain:       auth0Cred.Domain,
					ClientId:     auth0Cred.ClientID,
					ClientSecret: auth0Cred.ClientSecret,
				},
			},
		}, nil

	case cloudresourcekind.CloudResourceProvider_openstack:
		osCred := credInterface.(*models.OpenStackCredential)
		cfg := &openstackv1.OpenStackProviderConfig{
			AuthUrl:           osCred.AuthURL,
			Region:            osCred.Region,
			TenantName:        osCred.TenantName,
			TenantId:          osCred.TenantID,
			UserDomainName:    osCred.UserDomainName,
			UserDomainId:      osCred.UserDomainID,
			ProjectDomainName: osCred.ProjectDomainName,
			ProjectDomainId:   osCred.ProjectDomainID,
			Insecure:          osCred.Insecure,
			CacertFile:        osCred.CACertFile,
			EndpointType:      osCred.EndpointType,
		}
		switch osCred.AuthMethod {
		case "password":
			cfg.Credentials = &openstackv1.OpenStackProviderConfig_Password{
				Password: &openstackv1.OpenStackPasswordCredentials{
					UserName: osCred.UserName,
					Password: osCred.Password,
				},
			}
		case "application_credential":
			cfg.Credentials = &openstackv1.OpenStackProviderConfig_ApplicationCredential{
				ApplicationCredential: &openstackv1.OpenStackApplicationCredentials{
					Id:     osCred.ApplicationCredentialID,
					Name:   osCred.ApplicationCredentialName,
					Secret: osCred.ApplicationCredentialSecret,
				},
			}
		case "token":
			cfg.Credentials = &openstackv1.OpenStackProviderConfig_Token{
				Token: &openstackv1.OpenStackTokenCredentials{
					Token: osCred.Token,
				},
			}
		}
		return &credentialv1.CredentialProviderConfig{
		Data: &credentialv1.CredentialProviderConfig_Openstack{
			Openstack: cfg,
			},
		}, nil

	case cloudresourcekind.CloudResourceProvider_scaleway:
		scwCred := credInterface.(*models.ScalewayCredential)
		return &credentialv1.CredentialProviderConfig{
			Data: &credentialv1.CredentialProviderConfig_Scaleway{
				Scaleway: &scalewayv1.ScalewayProviderConfig{
					AccessKey:      scwCred.AccessKey,
					SecretKey:      scwCred.SecretKey,
					ProjectId:      scwCred.ProjectID,
					OrganizationId: scwCred.OrganizationID,
					Region:         scwCred.Region,
					Zone:           scwCred.Zone,
				},
			},
		}, nil

	case cloudresourcekind.CloudResourceProvider_alicloud:
		aliCred := credInterface.(*models.AlicloudCredential)
		cfg := &alicloudv1.AlicloudProviderConfig{
			Region:      aliCred.Region,
			AccountId:   aliCred.AccountId,
			AccountType: aliCred.AccountType,
		}
		switch aliCred.AuthMethod {
		case "static_credentials":
			cfg.AuthenticationType = alicloudv1.AuthenticationType_static_credentials
			cfg.StaticCredentials = &alicloudv1.AlicloudStaticCredentials{
				AccessKey: aliCred.AccessKey,
				SecretKey: aliCred.SecretKey,
			}
		case "sts_token":
			cfg.AuthenticationType = alicloudv1.AuthenticationType_sts_token
			cfg.StsToken = &alicloudv1.AlicloudStsTokenCredentials{
				AccessKey:     aliCred.AccessKey,
				SecretKey:     aliCred.SecretKey,
				SecurityToken: aliCred.SecurityToken,
			}
		case "ecs_role":
			cfg.AuthenticationType = alicloudv1.AuthenticationType_ecs_role
			cfg.EcsRole = &alicloudv1.AlicloudEcsRoleCredentials{
				EcsRoleName: aliCred.EcsRoleName,
			}
		case "assume_role":
			cfg.AuthenticationType = alicloudv1.AuthenticationType_assume_role
			cfg.AssumeRole = &alicloudv1.AlicloudAssumeRoleCredentials{
				AccessKey:         aliCred.AccessKey,
				SecretKey:         aliCred.SecretKey,
				RoleArn:           aliCred.RoleArn,
				SessionName:       aliCred.SessionName,
				Policy:            aliCred.Policy,
				SessionExpiration: aliCred.SessionExpiration,
				ExternalId:        aliCred.ExternalId,
			}
		case "assume_role_with_oidc":
			cfg.AuthenticationType = alicloudv1.AuthenticationType_assume_role_with_oidc
			cfg.AssumeRoleWithOidc = &alicloudv1.AlicloudAssumeRoleWithOidcCredentials{
				OidcProviderArn:   aliCred.OidcProviderArn,
				RoleArn:           aliCred.RoleArn,
				OidcToken:         aliCred.OidcToken,
				OidcTokenFile:     aliCred.OidcTokenFile,
				SessionName:       aliCred.SessionName,
				Policy:            aliCred.Policy,
				SessionExpiration: aliCred.SessionExpiration,
			}
		case "shared_credentials":
			cfg.AuthenticationType = alicloudv1.AuthenticationType_shared_credentials
			cfg.SharedCredentials = &alicloudv1.AlicloudSharedCredentials{
				CredentialsFile: aliCred.CredentialsFile,
				Profile:         aliCred.Profile,
			}
		case "sidecar_credentials":
			cfg.AuthenticationType = alicloudv1.AuthenticationType_sidecar_credentials
			cfg.SidecarCredentials = &alicloudv1.AlicloudSidecarCredentials{
				CredentialsUri: aliCred.CredentialsUri,
			}
		}
		return &credentialv1.CredentialProviderConfig{
			Data: &credentialv1.CredentialProviderConfig_Alicloud{
				Alicloud: cfg,
			},
		}, nil

	case cloudresourcekind.CloudResourceProvider_oci:
		ociCred := credInterface.(*models.OciCredential)
		cfg := &ociv1.OciProviderConfig{
			Region: ociCred.Region,
		}
		switch ociCred.AuthMethod {
		case "api_key":
			cfg.AuthenticationType = ociv1.AuthenticationType_api_key
			cfg.ApiKey = &ociv1.OciApiKeyAuth{
				TenancyOcid:        ociCred.TenancyOcid,
				UserOcid:           ociCred.UserOcid,
				Fingerprint:        ociCred.Fingerprint,
				PrivateKey:         ociCred.PrivateKey,
				PrivateKeyPassword: ociCred.PrivateKeyPassword,
			}
		case "security_token":
			cfg.AuthenticationType = ociv1.AuthenticationType_security_token
			cfg.SecurityToken = &ociv1.OciSecurityTokenAuth{
				ConfigFileProfile:  ociCred.ConfigFileProfile,
				PrivateKeyPassword: ociCred.PrivateKeyPassword,
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
		}, nil

	default:
		return nil, fmt.Errorf("provider '%s' is not yet supported for automatic credential resolution", providerString)
	}
}

// providerEnumToString converts CloudResourceProvider enum to a lowercase string.
// The enum String() method returns values like "aws", "gcp", "azure", etc.
func providerEnumToString(provider cloudresourcekind.CloudResourceProvider) string {
	name := provider.String()
	// Handle special case for unspecified
	if name == "cloud_resource_provider_unspecified" {
		return ""
	}
	// For other values, String() returns the enum name directly (e.g., "aws", "gcp")
	return strings.ToLower(name)
}
