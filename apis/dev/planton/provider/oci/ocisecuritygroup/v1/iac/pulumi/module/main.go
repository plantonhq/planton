package module

import (
	"github.com/pkg/errors"
	ocisecuritygroupv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocisecuritygroup/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/oci/pulumiociprovider"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *ocisecuritygroupv1.OciSecurityGroupStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	ociProvider, err := pulumiociprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup oci provider")
	}

	createdNsg, err := nsg(ctx, locals, ociProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create network security group")
	}

	if err := securityRules(ctx, locals, ociProvider, createdNsg); err != nil {
		return errors.Wrap(err, "failed to create security rules")
	}

	return nil
}

func pulumiOciOpt(provider *oci.Provider) pulumi.ResourceOption {
	return pulumi.Provider(provider)
}
