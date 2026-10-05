package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarewaitingroomeventv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarewaitingroomevent/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig   *cloudflareprovider.CloudflareProviderConfig
	CloudflareWaitingRoomEvent *cloudflarewaitingroomeventv1alpha1.CloudflareWaitingRoomEvent
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarewaitingroomeventv1alpha1.CloudflareWaitingRoomEventIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareWaitingRoomEvent = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
