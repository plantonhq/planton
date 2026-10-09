// Package execcredential implements the Kubernetes client-go ExecCredential protocol
// (client.authentication.k8s.io/v1): the kubeconfigs Planton renders for managed
// clusters name an exec command that a deploy engine re-invokes whenever the current
// bearer token expires, which is what keeps long Helm installs alive on providers
// like EKS whose tokens die within minutes. The command is the engine-spawning
// Planton binary itself -- no separate helper artifact, no cloud CLI in any image.
//
// Everything the command needs arrives via environment variables (declared below and
// templated into the kubeconfig's exec env entries by pkg/kubernetes/kubeconfig), so
// no credential ever appears in process arguments.
package execcredential

import (
	"encoding/json"
	"io"
	"time"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/kubernetes/kubetoken"
)

// Environment contract between the kubeconfig builder and this command. The kubeconfig
// builder writes these names; the command reads them -- sharing the constants is what
// keeps the two ends from drifting.
const (
	// CommandPathEnvVar is how an engine-spawning host advertises its own executable
	// path to code that renders kubeconfigs in a DIFFERENT process (the pulumi module
	// binary). Hosts that render kubeconfigs in-process use os.Executable() directly.
	CommandPathEnvVar = "PLANTON_KUBE_CREDENTIAL_COMMAND"

	// ProviderEnvVar selects the token source; values are KubernetesProvider enum
	// value names (aws_eks, gcp_gke, azure_aks).
	ProviderEnvVar = "PLANTON_KUBE_CREDENTIAL_PROVIDER"

	EksClusterNameEnvVar = "PLANTON_EKS_CLUSTER_NAME"
	EksRegionEnvVar      = "PLANTON_EKS_REGION"

	GkeServiceAccountKeyEnvVar = "PLANTON_GKE_SERVICE_ACCOUNT_KEY"

	AksTenantIdEnvVar     = "PLANTON_AKS_TENANT_ID"
	AksClientIdEnvVar     = "PLANTON_AKS_CLIENT_ID"
	AksClientSecretEnvVar = "PLANTON_AKS_CLIENT_SECRET"
	AksAccessTokenEnvVar  = "PLANTON_AKS_ACCESS_TOKEN"
)

// Static AWS credentials ride the SDK's standard names so the ambient credential
// chain inside MintEksToken picks them up with zero plumbing; when the connection
// carries no static keys the kubeconfig simply omits them and the process's own
// ambient chain (profile, env, instance role) signs instead.
const (
	AwsAccessKeyIDEnvVar     = "AWS_ACCESS_KEY_ID"
	AwsSecretAccessKeyEnvVar = "AWS_SECRET_ACCESS_KEY"
	AwsSessionTokenEnvVar    = "AWS_SESSION_TOKEN"
)

// A pre-minted GKE token rides Google's standard name for the same reason: the GKE
// minter's ambient chain reads it first. Set on the kubeconfig's exec entry, it
// overrides any value the engine process inherited (client-go appends the entry's env
// after the parent's, and the last value wins) -- so the cluster identity can never be
// confused with another Google identity in the engine's environment, such as the state
// bucket's sign-in.
const GoogleAccessTokenEnvVar = kubetoken.GoogleOAuthAccessTokenEnvVar

// Provider values accepted in ProviderEnvVar (KubernetesProvider enum value names).
const (
	ProviderAwsEks   = "aws_eks"
	ProviderGcpGke   = "gcp_gke"
	ProviderAzureAks = "azure_aks"
)

// protocolAPIVersion is the only ExecCredential version emitted: v1 has been stable
// since Kubernetes 1.22 and both the tofu and pulumi kubernetes providers speak it.
const protocolAPIVersion = "client.authentication.k8s.io/v1"

// execCredential is the wire shape client-go expects on the command's stdout. Only
// status fields a bearer-token flow uses are modeled (no client certs).
type execCredential struct {
	APIVersion string               `json:"apiVersion"`
	Kind       string               `json:"kind"`
	Status     execCredentialStatus `json:"status"`
}

type execCredentialStatus struct {
	Token string `json:"token"`
	// ExpirationTimestamp tells client-go when to re-invoke the command; reporting the
	// token's honest expiry (rather than omitting it) is the entire point of the
	// protocol for short-lived EKS tokens.
	ExpirationTimestamp string `json:"expirationTimestamp"`
}

// Emit writes the token as an ExecCredential JSON document to w (the command's stdout).
func Emit(w io.Writer, token kubetoken.Token) error {
	if err := json.NewEncoder(w).Encode(execCredential{
		APIVersion: protocolAPIVersion,
		Kind:       "ExecCredential",
		Status: execCredentialStatus{
			Token:               token.Value,
			ExpirationTimestamp: token.ExpiresAt.UTC().Format(time.RFC3339),
		},
	}); err != nil {
		return errors.Wrap(err, "encoding ExecCredential JSON")
	}
	return nil
}
