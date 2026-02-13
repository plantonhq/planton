package module

import (
	"github.com/pkg/errors"
	scalewayloadbalancerv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway/scalewayloadbalancer/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/scaleway/pulumiscalewayprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point that provisions the complete
// ScalewayLoadBalancer composite:
//
//  1. Flexible IP (dedicated public IPv4)
//  2. Load Balancer appliance (with optional Private Network attachment)
//  3. TLS certificates (Let's Encrypt or custom PEM)
//  4. Backend server pools (with health checks)
//  5. Frontend listeners (with certificate references)
//
// Resources are created in dependency order. Backends must exist before
// frontends because frontends reference backend IDs.
func Resources(
	ctx *pulumi.Context,
	stackInput *scalewayloadbalancerv1.ScalewayLoadBalancerStackInput,
) error {
	// 1. Prepare locals (metadata, labels, resolved references).
	locals := initializeLocals(ctx, stackInput)

	// 2. Create a Scaleway provider from the supplied credential.
	scalewayProvider, err := pulumiscalewayprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup scaleway provider")
	}

	// 3. Create Flexible IP + Load Balancer + Private Network attachment.
	createdLb, err := loadBalancer(ctx, locals, scalewayProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create load balancer")
	}

	// 4. Create TLS certificates (before frontends, since frontends reference cert IDs).
	certMap, err := certificates(ctx, locals, scalewayProvider, createdLb.lb)
	if err != nil {
		return errors.Wrap(err, "failed to create certificates")
	}

	// 5. Create backend server pools (before frontends, since frontends reference backend IDs).
	backendMap, err := backends(ctx, locals, scalewayProvider, createdLb.lb)
	if err != nil {
		return errors.Wrap(err, "failed to create backends")
	}

	// 6. Create frontend listeners (references backends and certificates by name).
	if err := frontends(ctx, locals, scalewayProvider, createdLb.lb, backendMap, certMap); err != nil {
		return errors.Wrap(err, "failed to create frontends")
	}

	return nil
}
