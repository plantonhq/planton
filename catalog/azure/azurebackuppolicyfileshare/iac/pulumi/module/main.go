package module

import (
	"github.com/pkg/errors"
	azurebackuppolicyfilesharev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurebackuppolicyfileshare/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/backup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurebackuppolicyfilesharev1alpha1.AzureBackupPolicyFileShareIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureBackupPolicyFileShare.Spec

	// Create the file-share backup policy -- the schedule and
	// retention rules that govern Azure Files share backups, as an ARM
	// child of its vault (.../vaults/{vault}/backupPolicies/{name}).
	// The policy is a free configuration object; ARM carries no tags
	// on backup policies.
	//
	// The spec's CEL contracts mirror the provider's schedule-shape
	// and vault-standard contracts, so a manifest that validates
	// renders a schedule ARM accepts.
	policyArgs := &backup.PolicyFileShareArgs{
		Name:              pulumi.String(spec.Name),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		RecoveryVaultName: pulumi.String(locals.RecoveryVaultName),
	}

	// The platform default fills timezone ("UTC").
	if spec.Timezone != nil {
		policyArgs.Timezone = pulumi.String(*spec.Timezone)
	}

	backupArgs := &backup.PolicyFileShareBackupArgs{
		Frequency: pulumi.String(spec.Backup.Frequency),
	}
	// Daily-only (spec CEL pins the shape to the frequency): the time
	// of day the backup runs. Omitted otherwise so the provider's
	// time/hourly exactly-one contract holds.
	if spec.Backup.Time != "" {
		backupArgs.Time = pulumi.String(spec.Backup.Time)
	}
	// Hourly-only (spec CEL): the backup window.
	if spec.Backup.Hourly != nil {
		backupArgs.Hourly = &backup.PolicyFileShareBackupHourlyArgs{
			Interval:       pulumi.Int(int(spec.Backup.Hourly.Interval)),
			StartTime:      pulumi.String(spec.Backup.Hourly.StartTime),
			WindowDuration: pulumi.Int(int(spec.Backup.Hourly.WindowDuration)),
		}
	}
	policyArgs.Backup = backupArgs

	// The platform default fills backup_tier ("snapshot", the
	// provider's own default); the nil guard keeps direct module
	// invocations safe.
	if spec.BackupTier != nil {
		policyArgs.BackupTier = pulumi.String(*spec.BackupTier)
	}

	// vault-standard only (spec CEL); omitted otherwise so the service
	// manages local snapshot retention.
	if spec.SnapshotRetentionInDays != nil {
		policyArgs.SnapshotRetentionInDays = pulumi.Int(int(*spec.SnapshotRetentionInDays))
	}

	// ALWAYS required (the provider's own contract) -- the base
	// retention layer for both Daily and Hourly schedules.
	policyArgs.RetentionDaily = &backup.PolicyFileShareRetentionDailyArgs{
		Count: pulumi.Int(int(spec.RetentionDaily.Count)),
	}

	if spec.RetentionWeekly != nil {
		policyArgs.RetentionWeekly = &backup.PolicyFileShareRetentionWeeklyArgs{
			Count:    pulumi.Int(int(spec.RetentionWeekly.Count)),
			Weekdays: pulumi.ToStringArray(spec.RetentionWeekly.Weekdays),
		}
	}

	if spec.RetentionMonthly != nil {
		monthlyArgs := &backup.PolicyFileShareRetentionMonthlyArgs{
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
		yearlyArgs := &backup.PolicyFileShareRetentionYearlyArgs{
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

	createdPolicy, err := backup.NewPolicyFileShare(ctx,
		spec.Name,
		policyArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create file-share backup policy %s", spec.Name)
	}

	ctx.Export(OpBackupPolicyId, createdPolicy.ID())
	ctx.Export(OpBackupPolicyName, createdPolicy.Name)

	return nil
}
