package module

import (
	"github.com/pkg/errors"
	cloudflarezerotrustmcpserverv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustmcpserver/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflarezerotrustmcpserverv1alpha1.CloudflareZeroTrustMcpServerIacInput,
) error {
	// 1. Prepare locals (metadata, credentials).
	locals := initializeLocals(ctx, iacInput)

	// 2. Create a Cloudflare provider from the supplied credential.
	cloudflareProvider, err := pulumicloudflareprovider.Get(
		ctx,
		iacInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup cloudflare provider")
	}

	// 3. Register the MCP server.
	if err := mcpServer(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create mcp server")
	}

	return nil
}
