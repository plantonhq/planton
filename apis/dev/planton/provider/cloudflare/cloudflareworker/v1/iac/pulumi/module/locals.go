package module

import (
	cloudflareprovider "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare"
	cloudflareworkerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflareworker/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles quick references copied from the stack‑input.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareWorker         *cloudflareworkerv1.CloudflareWorker
}

// initializeLocals mirrors the pattern used in existing modules.
func initializeLocals(_ *pulumi.Context, stackInput *cloudflareworkerv1.CloudflareWorkerStackInput) *Locals {
	locals := &Locals{}
	locals.CloudflareWorker = stackInput.Target
	locals.CloudflareProviderConfig = stackInput.ProviderConfig
	return locals
}
