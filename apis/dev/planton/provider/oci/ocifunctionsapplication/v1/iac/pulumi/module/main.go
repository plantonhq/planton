package module

import (
	"github.com/pkg/errors"
	ocifunctionsapplicationv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocifunctionsapplication/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/oci/pulumiociprovider"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var shapeMap = map[ocifunctionsapplicationv1.OciFunctionsApplicationSpec_Shape]string{
	ocifunctionsapplicationv1.OciFunctionsApplicationSpec_generic_x86:     "GENERIC_X86",
	ocifunctionsapplicationv1.OciFunctionsApplicationSpec_generic_arm:     "GENERIC_ARM",
	ocifunctionsapplicationv1.OciFunctionsApplicationSpec_generic_x86_arm: "GENERIC_X86_ARM",
}

func Resources(ctx *pulumi.Context, stackInput *ocifunctionsapplicationv1.OciFunctionsApplicationStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	ociProvider, err := pulumiociprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup oci provider")
	}

	if err := applicationResource(ctx, locals, ociProvider); err != nil {
		return errors.Wrap(err, "failed to create functions application")
	}

	return nil
}

func pulumiOciOpt(provider *oci.Provider) pulumi.ResourceOption {
	return pulumi.Provider(provider)
}
