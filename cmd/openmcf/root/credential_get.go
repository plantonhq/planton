package root

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
	credentialv1 "github.com/plantonhq/openmcf/apis/org/openmcf/app/credential/v1"
	credentialv1connect "github.com/plantonhq/openmcf/apis/org/openmcf/app/credential/v1/credentialv1connect"
	"github.com/spf13/cobra"
)

var CredentialGetCmd = &cobra.Command{
	Use:   "credential:get",
	Short: "get a credential by ID",
	Long:  "Retrieve detailed information about a credential by providing its unique ID.",
	Run:   credentialGetHandler,
}

func init() {
	CredentialGetCmd.Flags().StringP("id", "i", "", "unique identifier of the credential (required)")
	CredentialGetCmd.MarkFlagRequired("id")
}

func credentialGetHandler(cmd *cobra.Command, args []string) {
	// Get credential ID
	id, _ := cmd.Flags().GetString("id")
	if id == "" {
		fmt.Println("Error: --id flag is required. Provide the credential ID")
		fmt.Println("Usage: openmcf credential:get --id=<credential-id>")
		os.Exit(1)
	}

	// Get backend URL from configuration
	backendURL, err := GetBackendURL()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Create Connect-RPC client
	client := credentialv1connect.NewCredentialQueryControllerClient(
		http.DefaultClient,
		backendURL,
	)

	// Prepare request
	req := &credentialv1.GetCredentialRequest{
		Id: id,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Make the API call
	resp, err := client.Get(ctx, connect.NewRequest(req))
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
		fmt.Printf("Error getting credential: %v\n", err)
		os.Exit(1)
	}

	cred := resp.Msg.Credential

	// Display credential details
	fmt.Println("Credential Details:")
	fmt.Println("===================")
	fmt.Printf("ID:         %s\n", cred.Id)
	fmt.Printf("Name:       %s\n", cred.Name)

	providerName := "UNKNOWN"
	switch cred.Provider {
	case credentialv1.Credential_GCP:
		providerName = "GCP"
	case credentialv1.Credential_AWS:
		providerName = "AWS"
	case credentialv1.Credential_AZURE:
		providerName = "Azure"
	}
	fmt.Printf("Provider:   %s\n", providerName)

	if cred.CreatedAt != nil {
		fmt.Printf("Created At: %s\n", cred.CreatedAt.AsTime().Format("2006-01-02 15:04:05"))
	}
	if cred.UpdatedAt != nil {
		fmt.Printf("Updated At: %s\n", cred.UpdatedAt.AsTime().Format("2006-01-02 15:04:05"))
	}

	fmt.Println("\nCredential Data:")
	fmt.Println("----------------")
	if cred.ProviderConfig != nil {
		switch cred.Provider {
		case credentialv1.Credential_GCP:
			if gcp := cred.ProviderConfig.GetGcp(); gcp != nil {
				fmt.Printf("Service Account Key (JSON): %s\n", maskSensitive(gcp.ServiceAccountKey))
			}
		case credentialv1.Credential_AWS:
			if aws := cred.ProviderConfig.GetAws(); aws != nil {
				fmt.Printf("Account ID:       %s\n", aws.AccountId)
				fmt.Printf("Access Key ID:   %s\n", maskSensitive(aws.AccessKeyId))
				fmt.Printf("Secret Key:       %s\n", maskSensitive(aws.SecretAccessKey))
				if aws.Region != "" {
					fmt.Printf("Region:          %s\n", aws.Region)
				}
				if aws.SessionToken != "" {
					fmt.Printf("Session Token:   %s\n", maskSensitive(aws.SessionToken))
				}
			}
		case credentialv1.Credential_AZURE:
			if azure := cred.ProviderConfig.GetAzure(); azure != nil {
				fmt.Printf("Client ID:        %s\n", azure.ClientId)
				fmt.Printf("Client Secret:    %s\n", maskSensitive(azure.ClientSecret))
				fmt.Printf("Tenant ID:        %s\n", azure.TenantId)
				fmt.Printf("Subscription ID:  %s\n", azure.SubscriptionId)
			}
		}
	}
}

func maskSensitive(s string) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "..." + s[len(s)-4:]
}
