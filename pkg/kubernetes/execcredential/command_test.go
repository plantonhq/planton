package execcredential

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRun_AwsEksEndToEnd drives the command exactly as a deploy engine would: provider
// selection and cluster identity from the env contract, static AWS credentials under
// the SDK's standard names, protocol JSON on stdout. Fully offline -- EKS token
// minting is pure local signing.
func TestRun_AwsEksEndToEnd(t *testing.T) {
	t.Setenv(ProviderEnvVar, ProviderAwsEks)
	t.Setenv(EksClusterNameEnvVar, "demo-cluster")
	t.Setenv(EksRegionEnvVar, "us-west-2")
	t.Setenv(AwsAccessKeyIDEnvVar, "AKIAIOSFODNN7EXAMPLE")
	t.Setenv(AwsSecretAccessKeyEnvVar, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")
	// Blank out ambient AWS state that could shadow the test credentials.
	t.Setenv(AwsSessionTokenEnvVar, "")
	t.Setenv("AWS_PROFILE", "")

	var out bytes.Buffer
	require.NoError(t, run(context.Background(), &out))

	var doc struct {
		APIVersion string `json:"apiVersion"`
		Status     struct {
			Token               string `json:"token"`
			ExpirationTimestamp string `json:"expirationTimestamp"`
		} `json:"status"`
	}
	require.NoError(t, json.Unmarshal(out.Bytes(), &doc))

	assert.Equal(t, "client.authentication.k8s.io/v1", doc.APIVersion)
	assert.True(t, strings.HasPrefix(doc.Status.Token, "k8s-aws-v1."))
	assert.NotEmpty(t, doc.Status.ExpirationTimestamp)
}

func TestRun_RejectsMissingProvider(t *testing.T) {
	t.Setenv(ProviderEnvVar, "")

	err := run(context.Background(), &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), ProviderEnvVar,
		"the error must name the missing contract variable")
}

func TestRun_RejectsUnknownProvider(t *testing.T) {
	t.Setenv(ProviderEnvVar, "digital_ocean_doks") // DOKS never uses exec credentials

	err := run(context.Background(), &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "digital_ocean_doks")
}

// TestRun_AzureAksRejectsIncompleteIdentity proves the AKS arm is wired: a client
// secret without its identity coordinates fails inside the minter, before any
// network activity -- and the error names the provider being minted for.
func TestRun_AzureAksRejectsIncompleteIdentity(t *testing.T) {
	t.Setenv(ProviderEnvVar, ProviderAzureAks)
	t.Setenv(AksClientSecretEnvVar, "test-secret")
	t.Setenv(AksTenantIdEnvVar, "")
	t.Setenv(AksClientIdEnvVar, "")

	err := run(context.Background(), &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "azure_aks")
}

// emittedToken runs the command and returns the bearer token it printed.
func emittedToken(t *testing.T) string {
	t.Helper()
	var out bytes.Buffer
	require.NoError(t, run(context.Background(), &out))
	var doc struct {
		Status struct {
			Token string `json:"token"`
		} `json:"status"`
	}
	require.NoError(t, json.Unmarshal(out.Bytes(), &doc))
	return doc.Status.Token
}

// TestRun_GcpGkeSuppliedToken: the token a GKE kubeconfig's exec entry carries under
// Google's standard name is the token the command hands the API server -- the
// keyless delegate's whole path through the helper, fully offline.
func TestRun_GcpGkeSuppliedToken(t *testing.T) {
	t.Setenv(ProviderEnvVar, ProviderGcpGke)
	t.Setenv(GkeServiceAccountKeyEnvVar, "")
	t.Setenv(GoogleAccessTokenEnvVar, "ya29.keyless-test-token")

	assert.Equal(t, "ya29.keyless-test-token", emittedToken(t))
}

// TestRun_AzureAksSuppliedToken: the AKS arm hands back the exec entry's pre-minted
// token as is, fully offline.
func TestRun_AzureAksSuppliedToken(t *testing.T) {
	t.Setenv(ProviderEnvVar, ProviderAzureAks)
	t.Setenv(AksTenantIdEnvVar, "11111111-2222-3333-4444-555555555555")
	t.Setenv(AksClientIdEnvVar, "")
	t.Setenv(AksClientSecretEnvVar, "")
	t.Setenv(AksAccessTokenEnvVar, "eyJ0eXAi.keyless-test-token")

	assert.Equal(t, "eyJ0eXAi.keyless-test-token", emittedToken(t))
}
