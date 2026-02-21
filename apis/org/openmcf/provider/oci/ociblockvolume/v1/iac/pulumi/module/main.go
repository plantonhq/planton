package module

import (
	"github.com/pkg/errors"
	ociblockvolumev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ociblockvolume/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/oci/pulumiociprovider"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var autotuneTypeMap = map[ociblockvolumev1.OciBlockVolumeSpec_AutotunePolicy_AutotuneType]string{
	ociblockvolumev1.OciBlockVolumeSpec_AutotunePolicy_detached_volume:   "DETACHED_VOLUME",
	ociblockvolumev1.OciBlockVolumeSpec_AutotunePolicy_performance_based: "PERFORMANCE_BASED",
}

func Resources(ctx *pulumi.Context, stackInput *ociblockvolumev1.OciBlockVolumeStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	ociProvider, err := pulumiociprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup oci provider")
	}

	createdVolume, err := volume(ctx, locals, ociProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create block volume")
	}

	if err := backupPolicyAssignment(ctx, locals, ociProvider, createdVolume); err != nil {
		return errors.Wrap(err, "failed to create backup policy assignment")
	}

	return nil
}

func pulumiOciOpt(provider *oci.Provider) pulumi.ResourceOption {
	return pulumi.Provider(provider)
}
