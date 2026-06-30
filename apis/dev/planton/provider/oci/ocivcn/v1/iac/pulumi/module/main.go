package module

import (
	"github.com/pkg/errors"
	ocivcnv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocivcn/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/oci/pulumiociprovider"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *ocivcnv1.OciVcnStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	ociProvider, err := pulumiociprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup oci provider")
	}

	createdVcn, err := vcn(ctx, locals, ociProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create vcn")
	}

	spec := locals.OciVcn.Spec

	if spec.IsInternetGatewayEnabled {
		if err := internetGateway(ctx, locals, ociProvider, createdVcn); err != nil {
			return errors.Wrap(err, "failed to create internet gateway")
		}
	}

	if spec.IsNatGatewayEnabled {
		if err := natGateway(ctx, locals, ociProvider, createdVcn); err != nil {
			return errors.Wrap(err, "failed to create nat gateway")
		}
	}

	if spec.IsServiceGatewayEnabled {
		if err := serviceGateway(ctx, locals, ociProvider, createdVcn); err != nil {
			return errors.Wrap(err, "failed to create service gateway")
		}
	}

	return nil
}

// pulumiOciOpt is a convenience for passing the explicit OCI provider.
func pulumiOciOpt(provider *oci.Provider) pulumi.ResourceOption {
	return pulumi.Provider(provider)
}
