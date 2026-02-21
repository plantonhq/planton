package module

import (
	"github.com/pkg/errors"
	ocivaultsecretv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocivaultsecret/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/vault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func secret(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciVaultSecret.Spec

	secretArgs := &vault.SecretArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		SecretName:    pulumi.String(spec.SecretName),
		VaultId:       pulumi.String(spec.VaultId.GetValue()),
		KeyId:         pulumi.String(spec.KeyId.GetValue()),
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
	}

	if spec.Description != "" {
		secretArgs.Description = pulumi.StringPtr(spec.Description)
	}

	if spec.EnableAutoGeneration {
		secretArgs.EnableAutoGeneration = pulumi.BoolPtr(true)
	}

	if spec.SecretContent != nil {
		secretArgs.SecretContent = buildSecretContent(spec.SecretContent)
	}

	if spec.SecretGenerationContext != nil {
		secretArgs.SecretGenerationContext = buildSecretGenerationContext(spec.SecretGenerationContext)
	}

	if len(spec.SecretRules) > 0 {
		secretArgs.SecretRules = buildSecretRules(spec.SecretRules)
	}

	if spec.RotationConfig != nil {
		secretArgs.RotationConfig = buildRotationConfig(spec.RotationConfig)
	}

	if len(spec.SecretMetadata) > 0 {
		secretArgs.Metadata = pulumi.ToStringMap(spec.SecretMetadata)
	}

	createdSecret, err := vault.NewSecret(ctx, locals.DisplayName, secretArgs, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create vault secret")
	}

	ctx.Export(OpSecretId, createdSecret.ID())
	ctx.Export(OpCurrentVersionNumber, createdSecret.CurrentVersionNumber)

	return nil
}

func buildSecretContent(
	sc *ocivaultsecretv1.OciVaultSecretSpec_SecretContent,
) *vault.SecretSecretContentArgs {
	args := &vault.SecretSecretContentArgs{
		ContentType: pulumi.String("BASE64"),
	}

	if sc.Content != "" {
		args.Content = pulumi.StringPtr(sc.Content)
	}

	if sc.Name != "" {
		args.Name = pulumi.StringPtr(sc.Name)
	}

	if sc.Stage != "" {
		args.Stage = pulumi.StringPtr(sc.Stage)
	}

	return args
}

func buildSecretGenerationContext(
	sgc *ocivaultsecretv1.OciVaultSecretSpec_SecretGenerationContext,
) *vault.SecretSecretGenerationContextArgs {
	args := &vault.SecretSecretGenerationContextArgs{
		GenerationType:     pulumi.String(generationTypeMap[sgc.GenerationType]),
		GenerationTemplate: pulumi.String(sgc.GenerationTemplate),
	}

	if sgc.PassphraseLength > 0 {
		args.PassphraseLength = pulumi.IntPtr(int(sgc.PassphraseLength))
	}

	if sgc.SecretTemplate != "" {
		args.SecretTemplate = pulumi.StringPtr(sgc.SecretTemplate)
	}

	return args
}

func buildSecretRules(
	rules []*ocivaultsecretv1.OciVaultSecretSpec_SecretRule,
) vault.SecretSecretRuleArrayInput {
	var result vault.SecretSecretRuleArray
	for _, r := range rules {
		ruleArgs := vault.SecretSecretRuleArgs{
			RuleType: pulumi.String(ruleTypeMap[r.RuleType]),
		}

		if r.IsSecretContentRetrievalBlockedOnExpiry {
			ruleArgs.IsSecretContentRetrievalBlockedOnExpiry = pulumi.BoolPtr(true)
		}

		if r.SecretVersionExpiryInterval != "" {
			ruleArgs.SecretVersionExpiryInterval = pulumi.StringPtr(r.SecretVersionExpiryInterval)
		}

		if r.TimeOfAbsoluteExpiry != "" {
			ruleArgs.TimeOfAbsoluteExpiry = pulumi.StringPtr(r.TimeOfAbsoluteExpiry)
		}

		if r.IsEnforcedOnDeletedSecretVersions {
			ruleArgs.IsEnforcedOnDeletedSecretVersions = pulumi.BoolPtr(true)
		}

		result = append(result, ruleArgs)
	}
	return result
}

func buildRotationConfig(
	rc *ocivaultsecretv1.OciVaultSecretSpec_RotationConfig,
) *vault.SecretRotationConfigArgs {
	args := &vault.SecretRotationConfigArgs{
		TargetSystemDetails: buildTargetSystemDetails(rc.TargetSystemDetails),
	}

	if rc.IsScheduledRotationEnabled {
		args.IsScheduledRotationEnabled = pulumi.BoolPtr(true)
	}

	if rc.RotationInterval != "" {
		args.RotationInterval = pulumi.StringPtr(rc.RotationInterval)
	}

	return args
}

func buildTargetSystemDetails(
	tsd *ocivaultsecretv1.OciVaultSecretSpec_RotationConfig_TargetSystemDetails,
) vault.SecretRotationConfigTargetSystemDetailsArgs {
	args := vault.SecretRotationConfigTargetSystemDetailsArgs{
		TargetSystemType: pulumi.String(targetSystemTypeMap[tsd.TargetSystemType]),
	}

	if tsd.AdbId != nil && tsd.AdbId.GetValue() != "" {
		args.AdbId = pulumi.StringPtr(tsd.AdbId.GetValue())
	}

	if tsd.FunctionId != nil && tsd.FunctionId.GetValue() != "" {
		args.FunctionId = pulumi.StringPtr(tsd.FunctionId.GetValue())
	}

	return args
}
