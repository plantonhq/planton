package module

import (
	"github.com/pkg/errors"
	scalewaydnsrecordv1 "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway/scalewaydnsrecord/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/scaleway/pulumiscalewayprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the entry point for provisioning Scaleway DNS records.
func Resources(
	ctx *pulumi.Context,
	stackInput *scalewaydnsrecordv1.ScalewayDnsRecordStackInput,
) error {
	// 1. Initialize locals.
	locals := initializeLocals(ctx, stackInput)

	// 2. Create Scaleway provider from credential.
	scalewayProvider, err := pulumiscalewayprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup scaleway provider")
	}

	// 3. Provision the DNS record.
	if err := dnsRecord(ctx, locals, scalewayProvider); err != nil {
		return errors.Wrap(err, "failed to create dns record")
	}

	return nil
}
