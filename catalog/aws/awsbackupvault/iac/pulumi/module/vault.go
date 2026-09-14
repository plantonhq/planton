package module

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/backup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"google.golang.org/protobuf/types/known/structpb"
)

// vault creates the backup vault (exactly one of the two arms - the
// spec's exactly-one union mirrors AWS's own VaultType discriminator)
// plus the standard arm's satellites, and exports outputs.
//
// Lifecycle facts the render below depends on:
//   - the three satellites (lock, policy, notifications) attach by
//     VAULT NAME and only to STANDARD vaults - the provider's readers
//     reject other vault types, so the union gates them structurally;
//   - force_destroy is deploy-side delete behavior, never reported
//     back by AWS - invisible to imports, asserted only at destroy;
//   - lock MODE is decided by changeable_for_days alone: unset =
//     governance (removable), set = compliance (immutable once the
//     cooling-off window passes) - and AWS never reports the window
//     back;
//   - an air-gapped vault is immutable apart from tags: every argument
//     forces replacement, and its recovery points cannot be manually
//     deleted (they age out by retention).
func vault(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec

	if spec.GetAirGapped() != nil {
		args := &backup.LogicallyAirGappedVaultArgs{
			// metadata.name is the vault name on both engines.
			Name: pulumi.String(locals.Target.Metadata.Name),
			// Both retention bounds are REQUIRED by AWS on this vault
			// type (min floor 7 days) and both force replacement.
			MinRetentionDays: pulumi.Int(int(spec.GetAirGapped().MinRetentionDays)),
			MaxRetentionDays: pulumi.Int(int(spec.GetAirGapped().MaxRetentionDays)),
			Tags:             pulumi.ToStringMap(locals.AwsTags),
		}
		// Rendered only on an explicit choice so the module never
		// fights the provider default (the AWS-owned key).
		if spec.GetAirGapped().EncryptionKeyArn.GetValue() != "" {
			args.EncryptionKeyArn = pulumi.String(spec.GetAirGapped().EncryptionKeyArn.GetValue())
		}

		createdVault, err := backup.NewLogicallyAirGappedVault(ctx, "air-gapped-vault", args, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrap(err, "create logically air-gapped vault")
		}

		ctx.Export(OpVaultArn, createdVault.Arn)
		ctx.Export(OpVaultName, createdVault.Name)
		return nil
	}

	standard := spec.GetStandard()

	args := &backup.VaultArgs{
		// metadata.name is the vault name on both engines.
		Name: pulumi.String(locals.Target.Metadata.Name),
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}
	// Rendered only on an explicit choice so the module never fights
	// the provider default (the AWS Backup service key).
	if standard.KmsKeyArn.GetValue() != "" {
		args.KmsKeyArn = pulumi.String(standard.KmsKeyArn.GetValue())
	}
	// Deploy-side recovery-point drain at destroy (see the header
	// note).
	if standard.ForceDestroy {
		args.ForceDestroy = pulumi.Bool(true)
	}

	createdVault, err := backup.NewVault(ctx, "vault", args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "create vault")
	}

	if standard.Lock != nil {
		lockArgs := &backup.VaultLockConfigurationArgs{
			BackupVaultName: createdVault.Name,
		}
		// Present = COMPLIANCE mode (immutable after the window);
		// absent = governance mode. Write-only at AWS - never read
		// back.
		if standard.Lock.ChangeableForDays != nil {
			lockArgs.ChangeableForDays = pulumi.Int(int(*standard.Lock.ChangeableForDays))
		}
		if standard.Lock.MinRetentionDays != nil {
			lockArgs.MinRetentionDays = pulumi.Int(int(*standard.Lock.MinRetentionDays))
		}
		if standard.Lock.MaxRetentionDays != nil {
			lockArgs.MaxRetentionDays = pulumi.Int(int(*standard.Lock.MaxRetentionDays))
		}
		if _, err := backup.NewVaultLockConfiguration(ctx, "vault-lock", lockArgs,
			pulumi.Provider(provider), pulumi.Parent(createdVault)); err != nil {
			return errors.Wrap(err, "create vault lock configuration")
		}
	}

	if standard.Policy != nil {
		policyJson, err := structToJson(standard.Policy)
		if err != nil {
			return errors.Wrap(err, "render vault policy")
		}
		if _, err := backup.NewVaultPolicy(ctx, "vault-policy", &backup.VaultPolicyArgs{
			BackupVaultName: createdVault.Name,
			Policy:          pulumi.String(policyJson),
		}, pulumi.Provider(provider), pulumi.Parent(createdVault)); err != nil {
			return errors.Wrap(err, "create vault policy")
		}
	}

	if standard.Notifications != nil {
		if _, err := backup.NewVaultNotifications(ctx, "vault-notifications", &backup.VaultNotificationsArgs{
			BackupVaultName:   createdVault.Name,
			SnsTopicArn:       pulumi.String(standard.Notifications.SnsTopicArn.GetValue()),
			BackupVaultEvents: pulumi.ToStringArray(standard.Notifications.Events),
		}, pulumi.Provider(provider), pulumi.Parent(createdVault)); err != nil {
			return errors.Wrap(err, "create vault notifications")
		}
	}

	ctx.Export(OpVaultArn, createdVault.Arn)
	ctx.Export(OpVaultName, createdVault.Name)
	return nil
}

// structToJson renders an IAM policy document (a
// google.protobuf.Struct) as the provider's normalized-JSON string.
func structToJson(in *structpb.Struct) (string, error) {
	bytes, err := json.Marshal(in.AsMap())
	if err != nil {
		return "", errors.Wrap(err, "marshal json leaf")
	}
	return string(bytes), nil
}
