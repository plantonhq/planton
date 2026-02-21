package module

import (
	"github.com/pkg/errors"
	ociredisclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocirediscluster/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/oci/pulumiociprovider"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *ociredisclusterv1.OciRedisClusterStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	ociProvider, err := pulumiociprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup oci provider")
	}

	if err := redisCluster(ctx, locals, ociProvider); err != nil {
		return errors.Wrap(err, "failed to create redis cluster")
	}

	return nil
}

func pulumiOciOpt(provider *oci.Provider) pulumi.ResourceOption {
	return pulumi.Provider(provider)
}
