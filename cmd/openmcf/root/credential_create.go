package root

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	credentialv1 "github.com/plantonhq/openmcf/apis/org/openmcf/app/credential/v1"
	credentialv1connect "github.com/plantonhq/openmcf/apis/org/openmcf/app/credential/v1/credentialv1connect"
	awsv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws"
	azurev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure"
	gcpv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp"
	"github.com/spf13/cobra"
)

var CredentialCreateCmd = &cobra.Command{
	Use:   "credential:create",
	Short: "create a new cloud provider credential",
	Long:  "Create a new cloud provider credential for GCP, AWS, or Azure. Credentials are stored in the database and automatically used when deploying resources to the corresponding provider.",
	Run:   credentialCreateHandler,
}

func init() {
	CredentialCreateCmd.Flags().StringP("name", "n", "", "name of the credential (required)")
	CredentialCreateCmd.Flags().StringP("provider", "p", "", "cloud provider: gcp, aws, or azure (required)")

	// GCP flags
	CredentialCreateCmd.Flags().String("service-account-key", "", "path to GCP service account key JSON file (required for GCP)")

	// AWS flags
	CredentialCreateCmd.Flags().String("account-id", "", "AWS account ID (required for AWS)")
	CredentialCreateCmd.Flags().String("access-key-id", "", "AWS access key ID (required for AWS)")
	CredentialCreateCmd.Flags().String("secret-access-key", "", "AWS secret access key (required for AWS)")
	CredentialCreateCmd.Flags().String("region", "", "AWS region (optional for AWS)")
	CredentialCreateCmd.Flags().String("session-token", "", "AWS session token (optional for AWS)")

	// Azure flags
	CredentialCreateCmd.Flags().String("client-id", "", "Azure client ID (required for Azure)")
	CredentialCreateCmd.Flags().String("client-secret", "", "Azure client secret (required for Azure)")
	CredentialCreateCmd.Flags().String("tenant-id", "", "Azure tenant ID (required for Azure)")
	CredentialCreateCmd.Flags().String("subscription-id", "", "Azure subscription ID (required for Azure)")

	CredentialCreateCmd.MarkFlagRequired("name")
	CredentialCreateCmd.MarkFlagRequired("provider")
}

func credentialCreateHandler(cmd *cobra.Command, args []string) {
	// Get common flags
	name, _ := cmd.Flags().GetString("name")
	provider, _ := cmd.Flags().GetString("provider")

	if name == "" {
		fmt.Println("Error: --name flag is required")
		fmt.Println("Usage: openmcf credential:create --name=<name> --provider=<gcp|aws|azure> [provider-specific-flags]")
		os.Exit(1)
	}

	if provider == "" {
		fmt.Println("Error: --provider flag is required")
		fmt.Println("Usage: openmcf credential:create --name=<name> --provider=<gcp|aws|azure> [provider-specific-flags]")
		os.Exit(1)
	}

	// Normalize provider to lowercase
	provider = strings.ToLower(provider)

	// Get backend URL from configuration
	backendURL, err := GetBackendURL()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Create Connect-RPC client
	client := credentialv1connect.NewCredentialCommandControllerClient(
		http.DefaultClient,
		backendURL,
	)

	// Prepare request based on provider
	var req *credentialv1.CreateCredentialRequest

	switch provider {
	case "gcp":
		req, err = buildGcpCredentialRequest(cmd, name)
	case "aws":
		req, err = buildAwsCredentialRequest(cmd, name)
	case "azure":
		req, err = buildAzureCredentialRequest(cmd, name)
	default:
		fmt.Printf("Error: Unsupported provider '%s'. Supported providers: gcp, aws, azure\n", provider)
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Make the API call
	resp, err := client.Create(ctx, connect.NewRequest(req))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeUnavailable {
			fmt.Printf("Error: Cannot connect to backend service at %s. Please check:\n", backendURL)
			fmt.Printf("  1. The backend service is running\n")
			fmt.Printf("  2. The backend URL is correct\n")
			fmt.Printf("  3. Network connectivity\n")
			os.Exit(1)
		}
		if connect.CodeOf(err) == connect.CodeInvalidArgument {
			fmt.Printf("Error: Invalid request - %v\n", err)
			os.Exit(1)
		}
		if connect.CodeOf(err) == connect.CodeAlreadyExists {
			fmt.Printf("Error: A credential with name '%s' already exists\n", name)
			os.Exit(1)
		}
		fmt.Printf("Error creating credential: %v\n", err)
		os.Exit(1)
	}

	credential := resp.Msg.Credential

	// Display success message
	fmt.Println("✅ Credential created successfully!")
	fmt.Printf("\nID: %s\n", credential.Id)
	fmt.Printf("Name: %s\n", credential.Name)
	fmt.Printf("Provider: %s\n", strings.ToUpper(provider))
	if credential.CreatedAt != nil {
		fmt.Printf("Created At: %s\n", credential.CreatedAt.AsTime().Format("2006-01-02 15:04:05"))
	}
	fmt.Printf("\n💡 This credential can now be automatically used when deploying %s resources.\n", strings.ToUpper(provider))
}

func buildGcpCredentialRequest(cmd *cobra.Command, name string) (*credentialv1.CreateCredentialRequest, error) {
	keyFile, _ := cmd.Flags().GetString("service-account-key")
	if keyFile == "" {
		return nil, fmt.Errorf("--service-account-key is required for GCP provider")
	}

	// Read service account key file
	keyContent, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read service account key file '%s': %w", keyFile, err)
	}

	return &credentialv1.CreateCredentialRequest{
		Name:     name,
		Provider: credentialv1.Credential_GCP,
		ProviderConfig: &credentialv1.CredentialProviderConfig{
			Data: &credentialv1.CredentialProviderConfig_Gcp{
				Gcp: &gcpv1.GcpProviderConfig{
					ServiceAccountKey: string(keyContent),
				},
			},
		},
	}, nil
}

func buildAwsCredentialRequest(cmd *cobra.Command, name string) (*credentialv1.CreateCredentialRequest, error) {
	accountID, _ := cmd.Flags().GetString("account-id")
	accessKeyID, _ := cmd.Flags().GetString("access-key-id")
	secretAccessKey, _ := cmd.Flags().GetString("secret-access-key")
	region, _ := cmd.Flags().GetString("region")
	sessionToken, _ := cmd.Flags().GetString("session-token")

	if accountID == "" {
		return nil, fmt.Errorf("--account-id is required for AWS provider")
	}
	if accessKeyID == "" {
		return nil, fmt.Errorf("--access-key-id is required for AWS provider")
	}
	if secretAccessKey == "" {
		return nil, fmt.Errorf("--secret-access-key is required for AWS provider")
	}

	spec := &awsv1.AwsProviderConfig{
		AccountId:       accountID,
		AccessKeyId:     accessKeyID,
		SecretAccessKey: secretAccessKey,
	}

	if region != "" {
		spec.Region = region
	}
	if sessionToken != "" {
		spec.SessionToken = sessionToken
	}

	return &credentialv1.CreateCredentialRequest{
		Name:     name,
		Provider: credentialv1.Credential_AWS,
		ProviderConfig: &credentialv1.CredentialProviderConfig{
			Data: &credentialv1.CredentialProviderConfig_Aws{
				Aws: spec,
			},
		},
	}, nil
}

func buildAzureCredentialRequest(cmd *cobra.Command, name string) (*credentialv1.CreateCredentialRequest, error) {
	clientID, _ := cmd.Flags().GetString("client-id")
	clientSecret, _ := cmd.Flags().GetString("client-secret")
	tenantID, _ := cmd.Flags().GetString("tenant-id")
	subscriptionID, _ := cmd.Flags().GetString("subscription-id")

	if clientID == "" {
		return nil, fmt.Errorf("--client-id is required for Azure provider")
	}
	if clientSecret == "" {
		return nil, fmt.Errorf("--client-secret is required for Azure provider")
	}
	if tenantID == "" {
		return nil, fmt.Errorf("--tenant-id is required for Azure provider")
	}
	if subscriptionID == "" {
		return nil, fmt.Errorf("--subscription-id is required for Azure provider")
	}

	return &credentialv1.CreateCredentialRequest{
		Name:     name,
		Provider: credentialv1.Credential_AZURE,
		ProviderConfig: &credentialv1.CredentialProviderConfig{
			Data: &credentialv1.CredentialProviderConfig_Azure{
				Azure: &azurev1.AzureProviderConfig{
					ClientId:       clientID,
					ClientSecret:   clientSecret,
					TenantId:       tenantID,
					SubscriptionId: subscriptionID,
				},
			},
		},
	}, nil
}
