package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarewaitingroomv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarewaitingroom/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareWaitingRoom    *cloudflarewaitingroomv1alpha1.CloudflareWaitingRoom
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarewaitingroomv1alpha1.CloudflareWaitingRoomIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareWaitingRoom = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
