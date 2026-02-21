package module

import (
	"github.com/pkg/errors"
	ocikmsvaultv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocikmsvault/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/kms"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func vault(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciKmsVault.Spec

	vaultArgs := &kms.VaultArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		DisplayName:   pulumi.String(locals.DisplayName),
		VaultType:     pulumi.String(vaultTypeMap[spec.VaultType]),
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
	}

	if spec.ExternalKeyManagerMetadata != nil {
		vaultArgs.ExternalKeyManagerMetadata = buildExternalKeyManagerMetadata(spec.ExternalKeyManagerMetadata)
	}

	createdVault, err := kms.NewVault(ctx, locals.DisplayName, vaultArgs, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create vault")
	}

	ctx.Export(OpVaultId, createdVault.ID())
	ctx.Export(OpCryptoEndpoint, createdVault.CryptoEndpoint)
	ctx.Export(OpManagementEndpoint, createdVault.ManagementEndpoint)

	return nil
}

func buildExternalKeyManagerMetadata(
	ekm *ocikmsvaultv1.OciKmsVaultSpec_ExternalKeyManagerMetadata,
) *kms.VaultExternalKeyManagerMetadataArgs {
	return &kms.VaultExternalKeyManagerMetadataArgs{
		ExternalVaultEndpointUrl: pulumi.String(ekm.ExternalVaultEndpointUrl),
		OauthMetadata: &kms.VaultExternalKeyManagerMetadataOauthMetadataArgs{
			ClientAppId:        pulumi.String(ekm.OauthMetadata.ClientAppId),
			ClientAppSecret:    pulumi.String(ekm.OauthMetadata.ClientAppSecret),
			IdcsAccountNameUrl: pulumi.String(ekm.OauthMetadata.IdcsAccountNameUrl),
		},
		PrivateEndpointId: pulumi.String(ekm.PrivateEndpointId),
	}
}
