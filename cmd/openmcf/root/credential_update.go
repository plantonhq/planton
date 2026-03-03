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

var CredentialUpdateCmd = &cobra.Command{
	Use:   "credential:update",
	Short: "update an existing cloud provider credential",
	Long:  "Update an existing cloud provider credential. The provider type must match the existing credential.",
	Run:   credentialUpdateHandler,
}

func init() {
	CredentialUpdateCmd.Flags().StringP("id", "i", "", "unique identifier of the credential (required)")
	CredentialUpdateCmd.Flags().StringP("name", "n", "", "name of the credential (required)")
	CredentialUpdateCmd.Flags().StringP("provider", "p", "", "cloud provider: gcp, aws, or azure (required)")

	// GCP flags
	CredentialUpdateCmd.Flags().String("service-account-key", "", "path to GCP service account key JSON file (required for GCP)")

	// AWS flags
	CredentialUpdateCmd.Flags().String("account-id", "", "AWS account ID (required for AWS)")
	CredentialUpdateCmd.Flags().String("access-key-id", "", "AWS access key ID (required for AWS)")
	CredentialUpdateCmd.Flags().String("secret-access-key", "", "AWS secret access key (required for AWS)")
	CredentialUpdateCmd.Flags().String("region", "", "AWS region (optional for AWS)")
	CredentialUpdateCmd.Flags().String("session-token", "", "AWS session token (optional for AWS)")

	// Azure flags
	CredentialUpdateCmd.Flags().String("client-id", "", "Azure client ID (required for Azure)")
	CredentialUpdateCmd.Flags().String("client-secret", "", "Azure client secret (required for Azure)")
	CredentialUpdateCmd.Flags().String("tenant-id", "", "Azure tenant ID (required for Azure)")
	CredentialUpdateCmd.Flags().String("subscription-id", "", "Azure subscription ID (required for Azure)")

	CredentialUpdateCmd.MarkFlagRequired("id")
	CredentialUpdateCmd.MarkFlagRequired("name")
	CredentialUpdateCmd.MarkFlagRequired("provider")
}

func credentialUpdateHandler(cmd *cobra.Command, args []string) {
	// Get common flags
	id, _ := cmd.Flags().GetString("id")
	name, _ := cmd.Flags().GetString("name")
	providerStr, _ := cmd.Flags().GetString("provider")

	if id == "" || name == "" || providerStr == "" {
		fmt.Println("Error: --id, --name, and --provider flags are required")
		fmt.Println("Usage: openmcf credential:update --id=<id> --name=<name> --provider=<provider> [provider-specific-flags]")
		os.Exit(1)
	}

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

	// Convert provider string to enum
	var provider credentialv1.Credential_CredentialProvider
	switch strings.ToLower(providerStr) {
	case "gcp":
		provider = credentialv1.Credential_GCP
	case "aws":
		provider = credentialv1.Credential_AWS
	case "azure":
		provider = credentialv1.Credential_AZURE
	default:
		fmt.Printf("Error: Invalid provider '%s'. Valid values: gcp, aws, azure\n", providerStr)
		os.Exit(1)
	}

	// Prepare request based on provider
	req := &credentialv1.UpdateCredentialRequest{
		Id:       id,
		Name:     name,
		Provider: provider,
	}

	switch provider {
	case credentialv1.Credential_GCP:
		serviceAccountKeyPath, _ := cmd.Flags().GetString("service-account-key")
		if serviceAccountKeyPath == "" {
			fmt.Println("Error: --service-account-key is required for GCP provider")
			os.Exit(1)
		}

		keyBytes, err := os.ReadFile(serviceAccountKeyPath)
		if err != nil {
			fmt.Printf("Error: Failed to read service account key file '%s': %v\n", serviceAccountKeyPath, err)
			os.Exit(1)
		}

		req.ProviderConfig = &credentialv1.CredentialProviderConfig{
			Data: &credentialv1.CredentialProviderConfig_Gcp{
				Gcp: &gcpv1.GcpProviderConfig{
					ServiceAccountKey: string(keyBytes),
				},
			},
		}

	case credentialv1.Credential_AWS:
		accountID, _ := cmd.Flags().GetString("account-id")
		accessKeyID, _ := cmd.Flags().GetString("access-key-id")
		secretAccessKey, _ := cmd.Flags().GetString("secret-access-key")
		region, _ := cmd.Flags().GetString("region")
		sessionToken, _ := cmd.Flags().GetString("session-token")

		if accountID == "" || accessKeyID == "" || secretAccessKey == "" {
			fmt.Println("Error: --account-id, --access-key-id, and --secret-access-key are required for AWS provider")
			os.Exit(1)
		}

		awsSpec := &awsv1.AwsProviderConfig{
			AccountId:       accountID,
			AccessKeyId:     accessKeyID,
			SecretAccessKey: secretAccessKey,
		}

		if region != "" {
			awsSpec.Region = region
		}
		if sessionToken != "" {
			awsSpec.SessionToken = sessionToken
		}

		req.ProviderConfig = &credentialv1.CredentialProviderConfig{
			Data: &credentialv1.CredentialProviderConfig_Aws{
				Aws: awsSpec,
			},
		}

	case credentialv1.Credential_AZURE:
		clientID, _ := cmd.Flags().GetString("client-id")
		clientSecret, _ := cmd.Flags().GetString("client-secret")
		tenantID, _ := cmd.Flags().GetString("tenant-id")
		subscriptionID, _ := cmd.Flags().GetString("subscription-id")

		if clientID == "" || clientSecret == "" || tenantID == "" || subscriptionID == "" {
			fmt.Println("Error: --client-id, --client-secret, --tenant-id, and --subscription-id are required for Azure provider")
			os.Exit(1)
		}

		req.ProviderConfig = &credentialv1.CredentialProviderConfig{
			Data: &credentialv1.CredentialProviderConfig_Azure{
				Azure: &azurev1.AzureProviderConfig{
					ClientId:       clientID,
					ClientSecret:   clientSecret,
					TenantId:       tenantID,
					SubscriptionId: subscriptionID,
				},
			},
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Make the API call
	resp, err := client.Update(ctx, connect.NewRequest(req))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeUnavailable {
			fmt.Printf("Error: Cannot connect to backend service at %s. Please check:\n", backendURL)
			fmt.Printf("  1. The backend service is running\n")
			fmt.Printf("  2. The backend URL is correct\n")
			fmt.Printf("  3. Network connectivity\n")
			os.Exit(1)
		}
		if connect.CodeOf(err) == connect.CodeNotFound {
			fmt.Printf("Error: Credential with ID '%s' not found\n", id)
			os.Exit(1)
		}
		if connect.CodeOf(err) == connect.CodeInvalidArgument {
			fmt.Printf("Error: Invalid request - %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Error updating credential: %v\n", err)
		os.Exit(1)
	}

	cred := resp.Msg.Credential

	// Display success message
	fmt.Println("✅ Credential updated successfully!")
	fmt.Printf("\nID:       %s\n", cred.Id)
	fmt.Printf("Name:     %s\n", cred.Name)

	providerName := "UNKNOWN"
	switch cred.Provider {
	case credentialv1.Credential_GCP:
		providerName = "GCP"
	case credentialv1.Credential_AWS:
		providerName = "AWS"
	case credentialv1.Credential_AZURE:
		providerName = "Azure"
	}
	fmt.Printf("Provider: %s\n", providerName)

	if cred.UpdatedAt != nil {
		fmt.Printf("Updated At: %s\n", cred.UpdatedAt.AsTime().Format("2006-01-02 15:04:05"))
	}
}
