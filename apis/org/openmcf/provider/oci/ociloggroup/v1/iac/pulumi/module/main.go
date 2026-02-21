package module

import (
	"github.com/pkg/errors"
	ociloggroupv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ociloggroup/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/oci/pulumiociprovider"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *ociloggroupv1.OciLogGroupStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	ociProvider, err := pulumiociprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup oci provider")
	}

	logGroup, err := logGroupResource(ctx, locals, ociProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create log group")
	}

	if err := logResources(ctx, locals, ociProvider, logGroup); err != nil {
		return errors.Wrap(err, "failed to create logs")
	}

	return nil
}

func pulumiOciOpt(provider *oci.Provider) pulumi.ResourceOption {
	return pulumi.Provider(provider)
}
