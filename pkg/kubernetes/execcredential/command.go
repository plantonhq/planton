// The cobra command serving the ExecCredential protocol, mounted by every
// engine-spawning Planton binary (the OSS CLI via the engine command set; any host
// that embeds the engine inherits it the same way).
package execcredential

import (
	"context"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/kubernetes/kubetoken"
	"github.com/spf13/cobra"
)

// Command is the hidden protocol endpoint deploy engines re-invoke through the
// kubeconfig's exec entry. Hidden because it is machine-called plumbing, not a
// user-facing verb -- humans never run it by hand.
var Command = &cobra.Command{
	Use:    "kube-exec-credential",
	Hidden: true,
	Short:  "emit a Kubernetes ExecCredential token for the cluster named by the environment",
	Long: `Serve the Kubernetes client-go ExecCredential protocol
(client.authentication.k8s.io/v1).

Kubeconfigs rendered by Planton for managed clusters (EKS, GKE, AKS) name
this command as their credential source; the kubernetes and helm providers
invoke it whenever the current token expires. All inputs arrive via
environment variables templated into the kubeconfig -- never via arguments.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(cmd.Context(), cmd.OutOrStdout())
	},
}

// run mints a token for the provider selected by the environment and emits it in
// protocol shape. Kept separate from the cobra wiring so tests drive it in-process.
func run(ctx context.Context, out io.Writer) error {
	provider := os.Getenv(ProviderEnvVar)

	var token kubetoken.Token
	var err error
	switch provider {
	case ProviderAwsEks:
		// Static AWS credentials (when the connection carries them) arrive under the
		// SDK's standard env names, so the ambient chain inside the minter finds them.
		token, err = kubetoken.MintEksToken(ctx, kubetoken.EksTokenOptions{
			ClusterName: os.Getenv(EksClusterNameEnvVar),
			Region:      os.Getenv(EksRegionEnvVar),
		})
	case ProviderGcpGke:
		// A pre-minted token (GoogleAccessTokenEnvVar) needs no option: with no key,
		// the minter's ambient chain reads it before falling back to ADC.
		token, err = kubetoken.MintGkeToken(ctx, kubetoken.GkeTokenOptions{
			ServiceAccountKeyJSON: os.Getenv(GkeServiceAccountKeyEnvVar),
		})
	case ProviderAzureAks:
		// A pre-minted access token is used as is; with neither it nor a client
		// secret, the minter falls back to the ambient Azure credential chain
		// (environment, managed identity, Azure CLI login).
		token, err = kubetoken.MintAksToken(ctx, kubetoken.AksTokenOptions{
			TenantID:     os.Getenv(AksTenantIdEnvVar),
			ClientID:     os.Getenv(AksClientIdEnvVar),
			ClientSecret: os.Getenv(AksClientSecretEnvVar),
			AccessToken:  os.Getenv(AksAccessTokenEnvVar),
		})
	case "":
		return errors.Errorf("%s is not set; this command is invoked by deploy engines "+
			"through a Planton-rendered kubeconfig, never run directly", ProviderEnvVar)
	default:
		return errors.Errorf("unsupported provider %q in %s", provider, ProviderEnvVar)
	}
	if err != nil {
		return errors.Wrapf(err, "minting %s token", provider)
	}

	return Emit(out, token)
}
