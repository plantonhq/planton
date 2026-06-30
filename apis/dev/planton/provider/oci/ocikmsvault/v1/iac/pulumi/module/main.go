package module

import (
	"github.com/pkg/errors"
	ocikmsvaultv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocikmsvault/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/oci/pulumiociprovider"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var vaultTypeMap = map[ocikmsvaultv1.OciKmsVaultSpec_VaultType]string{
	ocikmsvaultv1.OciKmsVaultSpec_default_vault:   "DEFAULT",
	ocikmsvaultv1.OciKmsVaultSpec_virtual_private: "VIRTUAL_PRIVATE",
	ocikmsvaultv1.OciKmsVaultSpec_external:        "EXTERNAL",
}

func Resources(ctx *pulumi.Context, stackInput *ocikmsvaultv1.OciKmsVaultStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	ociProvider, err := pulumiociprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup oci provider")
	}

	if err := vault(ctx, locals, ociProvider); err != nil {
		return errors.Wrap(err, "failed to create kms vault")
	}

	return nil
}

func pulumiOciOpt(provider *oci.Provider) pulumi.ResourceOption {
	return pulumi.Provider(provider)
}
