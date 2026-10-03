package module

import (
	"github.com/pkg/errors"
	cloudflarewaitingroomv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarewaitingroom/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflarewaitingroomv1alpha1.CloudflareWaitingRoomIacInput,
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

	// 3. Create the waiting room (and its bypass rules, when declared).
	if err := waitingRoom(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create waiting room")
	}

	return nil
}
