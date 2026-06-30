package module

import (
	"github.com/pkg/errors"
	ocinetworkloadbalancerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocinetworkloadbalancer/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/oci/pulumiociprovider"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *ocinetworkloadbalancerv1.OciNetworkLoadBalancerStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	ociProvider, err := pulumiociprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup oci provider")
	}

	createdNlb, err := createNetworkLoadBalancer(ctx, locals, ociProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create network load balancer")
	}

	createdBackendSets, err := createBackendSets(ctx, locals, ociProvider, createdNlb)
	if err != nil {
		return errors.Wrap(err, "failed to create backend sets")
	}

	if err := createListeners(ctx, locals, ociProvider, createdNlb, createdBackendSets); err != nil {
		return errors.Wrap(err, "failed to create listeners")
	}

	return nil
}

func pulumiOciOpt(provider *oci.Provider) pulumi.ResourceOption {
	return pulumi.Provider(provider)
}
