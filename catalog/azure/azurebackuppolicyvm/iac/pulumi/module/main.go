package module

import (
	"github.com/pkg/errors"
	azurebackuppolicyvmv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurebackuppolicyvm/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/backup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurebackuppolicyvmv1alpha1.AzureBackupPolicyVmIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureBackupPolicyVm.Spec

	// Create the VM backup policy -- the schedule and retention rules
	// that govern VM backups, as an ARM child of its vault
	// (.../vaults/{vault}/backupPolicies/{name}). The policy is a free
	// configuration object; ARM carries no tags on backup policies.
	//
	// The spec's CEL contracts mirror the provider's frequency/
	// retention coupling, so a manifest that validates renders a
	// schedule ARM accepts.
	policyArgs := &backup.PolicyVMArgs{
		Name:              pulumi.String(spec.Name),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		RecoveryVaultName: pulumi.String(locals.RecoveryVaultName),
	}

	// The platform default fills policy_type ("V1", the provider's own
	// default); the nil guard keeps direct module invocations safe.
	// ForceNew on the provider -- changing the generation replaces the
	// policy.
	if spec.PolicyType != nil {
		policyArgs.PolicyType = pulumi.String(*spec.PolicyType)
	}

	// The platform default fills timezone ("UTC").
	if spec.Timezone != nil {
		policyArgs.Timezone = pulumi.String(*spec.Timezone)
	}

	backupArgs := &backup.PolicyVMBackupArgs{
		Frequency: pulumi.String(spec.Backup.Frequency),
		Time:      pulumi.String(spec.Backup.Time),
	}
	if len(spec.Backup.Weekdays) > 0 {
		backupArgs.Weekdays = pulumi.ToStringArray(spec.Backup.Weekdays)
	}
	// Hourly-only dials (spec CEL pins them to Hourly + V2).
	if spec.Backup.HourInterval != nil {
		backupArgs.HourInterval = pulumi.Int(int(*spec.Backup.HourInterval))
	}
	if spec.Backup.HourDuration != nil {
		backupArgs.HourDuration = pulumi.Int(int(*spec.Backup.HourDuration))
	}
	policyArgs.Backup = backupArgs

	// Omit when unset so the SERVICE default applies (2 days on V1,
	// 7 on V2 -- version-dependent, so the platform pins no default).
	// Azure requires vaulted daily retention > this value
	// (BMSUserErrorInstantRPRetentionExceedsVaultedRetention); the
	// spec CEL `bpv_instant_lt_daily` front-loads that, including the
	// V2-unset-defaults-to-7 case.
	if spec.InstantRestoreRetentionDays != nil {
		policyArgs.InstantRestoreRetentionDays = pulumi.Int(int(*spec.InstantRestoreRetentionDays))
	}

	if spec.InstantRestoreResourceGroup != nil {
		rgArgs := &backup.PolicyVMInstantRestoreResourceGroupArgs{
			Prefix: pulumi.String(spec.InstantRestoreResourceGroup.Prefix),
		}
		if spec.InstantRestoreResourceGroup.Suffix != "" {
			rgArgs.Suffix = pulumi.String(spec.InstantRestoreResourceGroup.Suffix)
		}
		policyArgs.InstantRestoreResourceGroup = rgArgs
	}

	if spec.TieringPolicy != nil {
		archivedArgs := &backup.PolicyVMTieringPolicyArchivedRestorePointArgs{
			Mode: pulumi.String(spec.TieringPolicy.ArchivedRestorePoint.Mode),
		}
		// TierAfter-only age (spec CEL pairs them with the mode).
		if spec.TieringPolicy.ArchivedRestorePoint.Duration != nil {
			archivedArgs.Duration = pulumi.Int(int(*spec.TieringPolicy.ArchivedRestorePoint.Duration))
		}
		if spec.TieringPolicy.ArchivedRestorePoint.DurationType != "" {
			archivedArgs.DurationType = pulumi.String(spec.TieringPolicy.ArchivedRestorePoint.DurationType)
		}
		policyArgs.TieringPolicy = &backup.PolicyVMTieringPolicyArgs{
			ArchivedRestorePoint: archivedArgs,
		}
	}

	// V2 only; omitted otherwise so the service keeps its consistency
	// default (application/file-system consistent when possible).
	if spec.ConsistencyType != "" {
		policyArgs.ConsistencyType = pulumi.String(spec.ConsistencyType)
	}

	if spec.RetentionDaily != nil {
		policyArgs.RetentionDaily = &backup.PolicyVMRetentionDailyArgs{
			Count: pulumi.Int(int(spec.RetentionDaily.Count)),
		}
	}

	if spec.RetentionWeekly != nil {
		policyArgs.RetentionWeekly = &backup.PolicyVMRetentionWeeklyArgs{
			Count:    pulumi.Int(int(spec.RetentionWeekly.Count)),
			Weekdays: pulumi.ToStringArray(spec.RetentionWeekly.Weekdays),
		}
	}

	if spec.RetentionMonthly != nil {
		monthlyArgs := &backup.PolicyVMRetentionMonthlyArgs{
			Count: pulumi.Int(int(spec.RetentionMonthly.Count)),
		}
		// The two mutually-exclusive forms (spec CEL): week-of-month
		// (weeks + weekdays) or month days (days / include_last_days).
		if len(spec.RetentionMonthly.Weeks) > 0 {
			monthlyArgs.Weeks = pulumi.ToStringArray(spec.RetentionMonthly.Weeks)
		}
		if len(spec.RetentionMonthly.Weekdays) > 0 {
			monthlyArgs.Weekdays = pulumi.ToStringArray(spec.RetentionMonthly.Weekdays)
		}
		if len(spec.RetentionMonthly.Days) > 0 {
			days := pulumi.IntArray{}
			for _, day := range spec.RetentionMonthly.Days {
				days = append(days, pulumi.Int(int(day)))
			}
			monthlyArgs.Days = days
		}
		if spec.RetentionMonthly.IncludeLastDays {
			monthlyArgs.IncludeLastDays = pulumi.Bool(true)
		}
		policyArgs.RetentionMonthly = monthlyArgs
	}

	if spec.RetentionYearly != nil {
		yearlyArgs := &backup.PolicyVMRetentionYearlyArgs{
			Count:  pulumi.Int(int(spec.RetentionYearly.Count)),
			Months: pulumi.ToStringArray(spec.RetentionYearly.Months),
		}
		if len(spec.RetentionYearly.Weeks) > 0 {
			yearlyArgs.Weeks = pulumi.ToStringArray(spec.RetentionYearly.Weeks)
		}
		if len(spec.RetentionYearly.Weekdays) > 0 {
			yearlyArgs.Weekdays = pulumi.ToStringArray(spec.RetentionYearly.Weekdays)
		}
		if len(spec.RetentionYearly.Days) > 0 {
			days := pulumi.IntArray{}
			for _, day := range spec.RetentionYearly.Days {
				days = append(days, pulumi.Int(int(day)))
			}
			yearlyArgs.Days = days
		}
		if spec.RetentionYearly.IncludeLastDays {
			yearlyArgs.IncludeLastDays = pulumi.Bool(true)
		}
		policyArgs.RetentionYearly = yearlyArgs
	}

	createdPolicy, err := backup.NewPolicyVM(ctx,
		spec.Name,
		policyArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create backup policy %s", spec.Name)
	}

	ctx.Export(OpBackupPolicyId, createdPolicy.ID())
	ctx.Export(OpBackupPolicyName, createdPolicy.Name)

	return nil
}
