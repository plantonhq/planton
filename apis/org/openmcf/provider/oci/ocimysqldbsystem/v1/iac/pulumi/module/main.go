package module

import (
	"github.com/pkg/errors"
	ocimysqldbsystemv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocimysqldbsystem/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/oci/pulumiociprovider"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *ocimysqldbsystemv1.OciMysqlDbSystemStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	ociProvider, err := pulumiociprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup oci provider")
	}

	if err := mysqlDbSystem(ctx, locals, ociProvider); err != nil {
		return errors.Wrap(err, "failed to create mysql db system")
	}

	return nil
}

func pulumiOciOpt(provider *oci.Provider) pulumi.ResourceOption {
	return pulumi.Provider(provider)
}
