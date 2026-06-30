package module

import (
	"github.com/pkg/errors"
	ocivaultsecretv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocivaultsecret/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/oci/pulumiociprovider"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var generationTypeMap = map[ocivaultsecretv1.OciVaultSecretSpec_SecretGenerationContext_GenerationType]string{
	ocivaultsecretv1.OciVaultSecretSpec_SecretGenerationContext_bytes:      "BYTES",
	ocivaultsecretv1.OciVaultSecretSpec_SecretGenerationContext_passphrase: "PASSPHRASE",
	ocivaultsecretv1.OciVaultSecretSpec_SecretGenerationContext_ssh_key:    "SSH_KEY",
}

var ruleTypeMap = map[ocivaultsecretv1.OciVaultSecretSpec_SecretRule_RuleType]string{
	ocivaultsecretv1.OciVaultSecretSpec_SecretRule_secret_expiry_rule: "SECRET_EXPIRY_RULE",
	ocivaultsecretv1.OciVaultSecretSpec_SecretRule_secret_reuse_rule:  "SECRET_REUSE_RULE",
}

var targetSystemTypeMap = map[ocivaultsecretv1.OciVaultSecretSpec_RotationConfig_TargetSystemDetails_TargetSystemType]string{
	ocivaultsecretv1.OciVaultSecretSpec_RotationConfig_TargetSystemDetails_adb:      "ADB",
	ocivaultsecretv1.OciVaultSecretSpec_RotationConfig_TargetSystemDetails_function: "FUNCTION",
}

func Resources(ctx *pulumi.Context, stackInput *ocivaultsecretv1.OciVaultSecretStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	ociProvider, err := pulumiociprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup oci provider")
	}

	if err := secret(ctx, locals, ociProvider); err != nil {
		return errors.Wrap(err, "failed to create vault secret")
	}

	return nil
}

func pulumiOciOpt(provider *oci.Provider) pulumi.ResourceOption {
	return pulumi.Provider(provider)
}
