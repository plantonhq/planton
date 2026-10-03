package module

import (
	"github.com/pkg/errors"
	awsclientvpnv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsclientvpn/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the Client VPN endpoint and its three folded,
// endpoint-scoped satellites: target network associations (one per subnet),
// authorization rules, and routes. The satellites have no identity outside
// their endpoint, which is why they fold here instead of being kinds of
// their own; each is still its own provider resource so membership edits
// apply in place.
func Resources(ctx *pulumi.Context, iacInput *awsclientvpnv1alpha1.AwsClientVpnIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static keys, keyless
	// web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsClientVpn.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdEndpoint, err := endpoint(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create Client VPN endpoint")
	}

	createdAssociations, err := networkAssociations(ctx, locals, provider, createdEndpoint)
	if err != nil {
		return errors.Wrap(err, "failed to associate target networks")
	}

	if err := authorizationRules(ctx, locals, provider, createdEndpoint); err != nil {
		return errors.Wrap(err, "failed to create authorization rules")
	}

	if err := routes(ctx, locals, provider, createdEndpoint, createdAssociations); err != nil {
		return errors.Wrap(err, "failed to create routes")
	}

	return nil
}
