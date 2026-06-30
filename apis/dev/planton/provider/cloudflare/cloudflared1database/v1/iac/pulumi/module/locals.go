package module

import (
	cloudflareprovider "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare"
	cloudflared1databasev1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflared1database/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareD1Database     *cloudflared1databasev1.CloudflareD1Database
}

// initializeLocals copies stack‑input fields into the Locals struct.
// Mirrors the style used in other Planton modules.
func initializeLocals(_ *pulumi.Context, stackInput *cloudflared1databasev1.CloudflareD1DatabaseStackInput) *Locals {
	locals := &Locals{}
	locals.CloudflareD1Database = stackInput.Target
	locals.CloudflareProviderConfig = stackInput.ProviderConfig
	return locals
}
