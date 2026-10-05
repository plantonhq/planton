package module

import (
	"github.com/pkg/errors"
	azuredataprotectionbackuppolicyv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredataprotectionbackuppolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/dataprotection"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates the Data Protection backup policy. Exactly one
// variant block is set in the spec (validated at admission); each
// variant creates its own provider resource -- ONE resource exists
// per deployment.
//
// EVERY policy variant is immutable after create (the provider ships
// no update path -- near-total ForceNew): changing anything replaces
// the policy.
func Resources(ctx *pulumi.Context, iacInput *azuredataprotectionbackuppolicyv1alpha1.AzureDataProtectionBackupPolicyIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureDataProtectionBackupPolicy.Spec

	switch {
	case spec.BlobStorage != nil:
		return blobStoragePolicy(ctx, locals, azureProvider)
	case spec.Disk != nil:
		return diskPolicy(ctx, locals, azureProvider)
	case spec.KubernetesCluster != nil:
		return kubernetesClusterPolicy(ctx, locals, azureProvider)
	case spec.MysqlFlexibleServer != nil:
		return mysqlFlexibleServerPolicy(ctx, locals, azureProvider)
	case spec.PostgresqlFlexibleServer != nil:
		return postgresqlFlexibleServerPolicy(ctx, locals, azureProvider)
	case spec.DataLakeStorage != nil:
		return dataLakeStoragePolicy(ctx, locals, azureProvider)
	}

	// Unreachable behind the spec's exactly-one CEL; loud beats silent
	// for a direct module invocation that skipped admission.
	return errors.New("no policy variant set -- exactly one of blob_storage, disk, kubernetes_cluster, mysql_flexible_server, postgresql_flexible_server or data_lake_storage is required")
}

func blobStoragePolicy(ctx *pulumi.Context, locals *Locals, azureProvider pulumi.ProviderResource) error {
	spec := locals.AzureDataProtectionBackupPolicy.Spec
	variant := spec.BlobStorage

	args := &dataprotection.BackupPolicyBlobStorageArgs{
		Name:    pulumi.String(spec.Name),
		VaultId: pulumi.String(locals.VaultId),
	}

	// Setting the operational duration enables the operational
	// (in-account) tier; setting the vault duration enables the vault
	// tier and requires the schedule intervals (the spec's CELs mirror
	// the provider's own AtLeastOneOf/RequiredWith lattice).
	if variant.OperationalDefaultRetentionDuration != "" {
		args.OperationalDefaultRetentionDuration = pulumi.String(variant.OperationalDefaultRetentionDuration)
	}
	if variant.VaultDefaultRetentionDuration != "" {
		args.VaultDefaultRetentionDuration = pulumi.String(variant.VaultDefaultRetentionDuration)
	}
	if len(variant.BackupRepeatingTimeIntervals) > 0 {
		args.BackupRepeatingTimeIntervals = pulumi.ToStringArray(variant.BackupRepeatingTimeIntervals)
	}
	if variant.TimeZone != "" {
		args.TimeZone = pulumi.String(variant.TimeZone)
	}

	if len(variant.RetentionRules) > 0 {
		rules := dataprotection.BackupPolicyBlobStorageRetentionRuleArray{}
		for _, rule := range variant.RetentionRules {
			criteriaArgs := &dataprotection.BackupPolicyBlobStorageRetentionRuleCriteriaArgs{}
			if rule.Criteria.AbsoluteCriteria != "" {
				criteriaArgs.AbsoluteCriteria = pulumi.String(rule.Criteria.AbsoluteCriteria)
			}
			if len(rule.Criteria.DaysOfMonth) > 0 {
				daysOfMonth := pulumi.IntArray{}
				for _, day := range rule.Criteria.DaysOfMonth {
					daysOfMonth = append(daysOfMonth, pulumi.Int(int(day)))
				}
				criteriaArgs.DaysOfMonths = daysOfMonth
			}
			if len(rule.Criteria.DaysOfWeek) > 0 {
				criteriaArgs.DaysOfWeeks = pulumi.ToStringArray(rule.Criteria.DaysOfWeek)
			}
			if len(rule.Criteria.MonthsOfYear) > 0 {
				criteriaArgs.MonthsOfYears = pulumi.ToStringArray(rule.Criteria.MonthsOfYear)
			}
			if len(rule.Criteria.ScheduledBackupTimes) > 0 {
				criteriaArgs.ScheduledBackupTimes = pulumi.ToStringArray(rule.Criteria.ScheduledBackupTimes)
			}
			if len(rule.Criteria.WeeksOfMonth) > 0 {
				criteriaArgs.WeeksOfMonths = pulumi.ToStringArray(rule.Criteria.WeeksOfMonth)
			}

			rules = append(rules, &dataprotection.BackupPolicyBlobStorageRetentionRuleArgs{
				Name:     pulumi.String(rule.Name),
				Priority: pulumi.Int(int(rule.GetPriority())),
				Criteria: criteriaArgs,
				LifeCycle: &dataprotection.BackupPolicyBlobStorageRetentionRuleLifeCycleArgs{
					DataStoreType: pulumi.String(rule.LifeCycle.DataStoreType),
					Duration:      pulumi.String(rule.LifeCycle.Duration),
				},
			})
		}
		args.RetentionRules = rules
	}

	createdPolicy, err := dataprotection.NewBackupPolicyBlobStorage(ctx,
		spec.Name,
		args,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create blob storage backup policy %s", spec.Name)
	}

	exportOutputs(ctx, createdPolicy.ID(), spec.Name)
	return nil
}

func diskPolicy(ctx *pulumi.Context, locals *Locals, azureProvider pulumi.ProviderResource) error {
	spec := locals.AzureDataProtectionBackupPolicy.Spec
	variant := spec.Disk

	args := &dataprotection.BackupPolicyDiskArgs{
		Name:                         pulumi.String(spec.Name),
		VaultId:                      pulumi.String(locals.VaultId),
		BackupRepeatingTimeIntervals: pulumi.ToStringArray(variant.BackupRepeatingTimeIntervals),
		DefaultRetentionDuration:     pulumi.String(variant.DefaultRetentionDuration),
	}
	if variant.TimeZone != "" {
		args.TimeZone = pulumi.String(variant.TimeZone)
	}

	if len(variant.RetentionRules) > 0 {
		rules := dataprotection.BackupPolicyDiskRetentionRuleArray{}
		for _, rule := range variant.RetentionRules {
			criteriaArgs := &dataprotection.BackupPolicyDiskRetentionRuleCriteriaArgs{}
			if rule.Criteria.AbsoluteCriteria != "" {
				criteriaArgs.AbsoluteCriteria = pulumi.String(rule.Criteria.AbsoluteCriteria)
			}
			rules = append(rules, &dataprotection.BackupPolicyDiskRetentionRuleArgs{
				Name:     pulumi.String(rule.Name),
				Duration: pulumi.String(rule.Duration),
				Priority: pulumi.Int(int(rule.GetPriority())),
				Criteria: criteriaArgs,
			})
		}
		args.RetentionRules = rules
	}

	createdPolicy, err := dataprotection.NewBackupPolicyDisk(ctx,
		spec.Name,
		args,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create disk backup policy %s", spec.Name)
	}

	exportOutputs(ctx, createdPolicy.ID(), spec.Name)
	return nil
}

// The one variant the provider addresses by vault NAME + resource
// group instead of vault ID -- both derived in locals from the spec's
// single vault_id reference (ARM IDs are structured).
func kubernetesClusterPolicy(ctx *pulumi.Context, locals *Locals, azureProvider pulumi.ProviderResource) error {
	spec := locals.AzureDataProtectionBackupPolicy.Spec
	variant := spec.KubernetesCluster

	defaultLifeCycles := dataprotection.BackupPolicyKubernetesClusterDefaultRetentionRuleLifeCycleArray{}
	for _, lifeCycle := range variant.DefaultRetentionRule.LifeCycles {
		defaultLifeCycles = append(defaultLifeCycles, &dataprotection.BackupPolicyKubernetesClusterDefaultRetentionRuleLifeCycleArgs{
			DataStoreType: pulumi.String(lifeCycle.DataStoreType),
			Duration:      pulumi.String(lifeCycle.Duration),
		})
	}

	args := &dataprotection.BackupPolicyKubernetesClusterArgs{
		Name:                         pulumi.String(spec.Name),
		ResourceGroupName:            pulumi.String(locals.VaultResourceGroupName),
		VaultName:                    pulumi.String(locals.VaultName),
		BackupRepeatingTimeIntervals: pulumi.ToStringArray(variant.BackupRepeatingTimeIntervals),
		DefaultRetentionRule: &dataprotection.BackupPolicyKubernetesClusterDefaultRetentionRuleArgs{
			LifeCycles: defaultLifeCycles,
		},
	}
	if variant.TimeZone != "" {
		args.TimeZone = pulumi.String(variant.TimeZone)
	}

	if len(variant.RetentionRules) > 0 {
		rules := dataprotection.BackupPolicyKubernetesClusterRetentionRuleArray{}
		for _, rule := range variant.RetentionRules {
			criteriaArgs := &dataprotection.BackupPolicyKubernetesClusterRetentionRuleCriteriaArgs{}
			if rule.Criteria.AbsoluteCriteria != "" {
				criteriaArgs.AbsoluteCriteria = pulumi.String(rule.Criteria.AbsoluteCriteria)
			}
			if len(rule.Criteria.DaysOfWeek) > 0 {
				criteriaArgs.DaysOfWeeks = pulumi.ToStringArray(rule.Criteria.DaysOfWeek)
			}
			if len(rule.Criteria.MonthsOfYear) > 0 {
				criteriaArgs.MonthsOfYears = pulumi.ToStringArray(rule.Criteria.MonthsOfYear)
			}
			if len(rule.Criteria.ScheduledBackupTimes) > 0 {
				criteriaArgs.ScheduledBackupTimes = pulumi.ToStringArray(rule.Criteria.ScheduledBackupTimes)
			}
			if len(rule.Criteria.WeeksOfMonth) > 0 {
				criteriaArgs.WeeksOfMonths = pulumi.ToStringArray(rule.Criteria.WeeksOfMonth)
			}

			lifeCycles := dataprotection.BackupPolicyKubernetesClusterRetentionRuleLifeCycleArray{}
			for _, lifeCycle := range rule.LifeCycles {
				lifeCycles = append(lifeCycles, &dataprotection.BackupPolicyKubernetesClusterRetentionRuleLifeCycleArgs{
					DataStoreType: pulumi.String(lifeCycle.DataStoreType),
					Duration:      pulumi.String(lifeCycle.Duration),
				})
			}

			rules = append(rules, &dataprotection.BackupPolicyKubernetesClusterRetentionRuleArgs{
				Name:       pulumi.String(rule.Name),
				Priority:   pulumi.Int(int(rule.GetPriority())),
				Criteria:   criteriaArgs,
				LifeCycles: lifeCycles,
			})
		}
		args.RetentionRules = rules
	}

	createdPolicy, err := dataprotection.NewBackupPolicyKubernetesCluster(ctx,
		spec.Name,
		args,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create kubernetes cluster backup policy %s", spec.Name)
	}

	exportOutputs(ctx, createdPolicy.ID(), spec.Name)
	return nil
}

func mysqlFlexibleServerPolicy(ctx *pulumi.Context, locals *Locals, azureProvider pulumi.ProviderResource) error {
	spec := locals.AzureDataProtectionBackupPolicy.Spec
	variant := spec.MysqlFlexibleServer

	defaultLifeCycles := dataprotection.BackupPolicyMysqlFlexibleServerDefaultRetentionRuleLifeCycleArray{}
	for _, lifeCycle := range variant.DefaultRetentionRule.LifeCycles {
		defaultLifeCycles = append(defaultLifeCycles, &dataprotection.BackupPolicyMysqlFlexibleServerDefaultRetentionRuleLifeCycleArgs{
			DataStoreType: pulumi.String(lifeCycle.DataStoreType),
			Duration:      pulumi.String(lifeCycle.Duration),
		})
	}

	args := &dataprotection.BackupPolicyMysqlFlexibleServerArgs{
		Name:                         pulumi.String(spec.Name),
		VaultId:                      pulumi.String(locals.VaultId),
		BackupRepeatingTimeIntervals: pulumi.ToStringArray(variant.BackupRepeatingTimeIntervals),
		DefaultRetentionRule: &dataprotection.BackupPolicyMysqlFlexibleServerDefaultRetentionRuleArgs{
			LifeCycles: defaultLifeCycles,
		},
	}
	if variant.TimeZone != "" {
		args.TimeZone = pulumi.String(variant.TimeZone)
	}

	if len(variant.RetentionRules) > 0 {
		rules := dataprotection.BackupPolicyMysqlFlexibleServerRetentionRuleArray{}
		for _, rule := range variant.RetentionRules {
			criteriaArgs := &dataprotection.BackupPolicyMysqlFlexibleServerRetentionRuleCriteriaArgs{}
			if rule.Criteria.AbsoluteCriteria != "" {
				criteriaArgs.AbsoluteCriteria = pulumi.String(rule.Criteria.AbsoluteCriteria)
			}
			if len(rule.Criteria.DaysOfWeek) > 0 {
				criteriaArgs.DaysOfWeeks = pulumi.ToStringArray(rule.Criteria.DaysOfWeek)
			}
			if len(rule.Criteria.MonthsOfYear) > 0 {
				criteriaArgs.MonthsOfYears = pulumi.ToStringArray(rule.Criteria.MonthsOfYear)
			}
			if len(rule.Criteria.ScheduledBackupTimes) > 0 {
				criteriaArgs.ScheduledBackupTimes = pulumi.ToStringArray(rule.Criteria.ScheduledBackupTimes)
			}
			if len(rule.Criteria.WeeksOfMonth) > 0 {
				criteriaArgs.WeeksOfMonths = pulumi.ToStringArray(rule.Criteria.WeeksOfMonth)
			}

			lifeCycles := dataprotection.BackupPolicyMysqlFlexibleServerRetentionRuleLifeCycleArray{}
			for _, lifeCycle := range rule.LifeCycles {
				lifeCycles = append(lifeCycles, &dataprotection.BackupPolicyMysqlFlexibleServerRetentionRuleLifeCycleArgs{
					DataStoreType: pulumi.String(lifeCycle.DataStoreType),
					Duration:      pulumi.String(lifeCycle.Duration),
				})
			}

			rules = append(rules, &dataprotection.BackupPolicyMysqlFlexibleServerRetentionRuleArgs{
				Name:       pulumi.String(rule.Name),
				Priority:   pulumi.Int(int(rule.GetPriority())),
				Criteria:   criteriaArgs,
				LifeCycles: lifeCycles,
			})
		}
		args.RetentionRules = rules
	}

	createdPolicy, err := dataprotection.NewBackupPolicyMysqlFlexibleServer(ctx,
		spec.Name,
		args,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create mysql flexible server backup policy %s", spec.Name)
	}

	exportOutputs(ctx, createdPolicy.ID(), spec.Name)
	return nil
}

func postgresqlFlexibleServerPolicy(ctx *pulumi.Context, locals *Locals, azureProvider pulumi.ProviderResource) error {
	spec := locals.AzureDataProtectionBackupPolicy.Spec
	variant := spec.PostgresqlFlexibleServer

	defaultLifeCycles := dataprotection.BackupPolicyPostgresqlFlexibleServerDefaultRetentionRuleLifeCycleArray{}
	for _, lifeCycle := range variant.DefaultRetentionRule.LifeCycles {
		defaultLifeCycles = append(defaultLifeCycles, &dataprotection.BackupPolicyPostgresqlFlexibleServerDefaultRetentionRuleLifeCycleArgs{
			DataStoreType: pulumi.String(lifeCycle.DataStoreType),
			Duration:      pulumi.String(lifeCycle.Duration),
		})
	}

	args := &dataprotection.BackupPolicyPostgresqlFlexibleServerArgs{
		Name:                         pulumi.String(spec.Name),
		VaultId:                      pulumi.String(locals.VaultId),
		BackupRepeatingTimeIntervals: pulumi.ToStringArray(variant.BackupRepeatingTimeIntervals),
		DefaultRetentionRule: &dataprotection.BackupPolicyPostgresqlFlexibleServerDefaultRetentionRuleArgs{
			LifeCycles: defaultLifeCycles,
		},
	}
	if variant.TimeZone != "" {
		args.TimeZone = pulumi.String(variant.TimeZone)
	}

	if len(variant.RetentionRules) > 0 {
		rules := dataprotection.BackupPolicyPostgresqlFlexibleServerRetentionRuleArray{}
		for _, rule := range variant.RetentionRules {
			criteriaArgs := &dataprotection.BackupPolicyPostgresqlFlexibleServerRetentionRuleCriteriaArgs{}
			if rule.Criteria.AbsoluteCriteria != "" {
				criteriaArgs.AbsoluteCriteria = pulumi.String(rule.Criteria.AbsoluteCriteria)
			}
			if len(rule.Criteria.DaysOfWeek) > 0 {
				criteriaArgs.DaysOfWeeks = pulumi.ToStringArray(rule.Criteria.DaysOfWeek)
			}
			if len(rule.Criteria.MonthsOfYear) > 0 {
				criteriaArgs.MonthsOfYears = pulumi.ToStringArray(rule.Criteria.MonthsOfYear)
			}
			if len(rule.Criteria.ScheduledBackupTimes) > 0 {
				criteriaArgs.ScheduledBackupTimes = pulumi.ToStringArray(rule.Criteria.ScheduledBackupTimes)
			}
			if len(rule.Criteria.WeeksOfMonth) > 0 {
				criteriaArgs.WeeksOfMonths = pulumi.ToStringArray(rule.Criteria.WeeksOfMonth)
			}

			lifeCycles := dataprotection.BackupPolicyPostgresqlFlexibleServerRetentionRuleLifeCycleArray{}
			for _, lifeCycle := range rule.LifeCycles {
				lifeCycles = append(lifeCycles, &dataprotection.BackupPolicyPostgresqlFlexibleServerRetentionRuleLifeCycleArgs{
					DataStoreType: pulumi.String(lifeCycle.DataStoreType),
					Duration:      pulumi.String(lifeCycle.Duration),
				})
			}

			rules = append(rules, &dataprotection.BackupPolicyPostgresqlFlexibleServerRetentionRuleArgs{
				Name:       pulumi.String(rule.Name),
				Priority:   pulumi.Int(int(rule.GetPriority())),
				Criteria:   criteriaArgs,
				LifeCycles: lifeCycles,
			})
		}
		args.RetentionRules = rules
	}

	createdPolicy, err := dataprotection.NewBackupPolicyPostgresqlFlexibleServer(ctx,
		spec.Name,
		args,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create postgresql flexible server backup policy %s", spec.Name)
	}

	exportOutputs(ctx, createdPolicy.ID(), spec.Name)
	return nil
}

// Retention rules are FLAT here (criteria fields on the rule -- the
// provider's own shape) and priorities are assigned by ORDER: the
// provider stamps rule N with priority N+1 (there is no priority
// argument).
func dataLakeStoragePolicy(ctx *pulumi.Context, locals *Locals, azureProvider pulumi.ProviderResource) error {
	spec := locals.AzureDataProtectionBackupPolicy.Spec
	variant := spec.DataLakeStorage

	args := &dataprotection.BackupPolicyDataLakeStorageArgs{
		Name:                        pulumi.String(spec.Name),
		DataProtectionBackupVaultId: pulumi.String(locals.VaultId),
		BackupSchedules:             pulumi.ToStringArray(variant.BackupSchedule),
		DefaultRetentionDuration:    pulumi.String(variant.DefaultRetentionDuration),
	}
	if variant.TimeZone != "" {
		args.TimeZone = pulumi.String(variant.TimeZone)
	}

	if len(variant.RetentionRules) > 0 {
		rules := dataprotection.BackupPolicyDataLakeStorageRetentionRuleArray{}
		for _, rule := range variant.RetentionRules {
			ruleArgs := &dataprotection.BackupPolicyDataLakeStorageRetentionRuleArgs{
				Name:     pulumi.String(rule.Name),
				Duration: pulumi.String(rule.Duration),
			}
			if rule.AbsoluteCriteria != "" {
				ruleArgs.AbsoluteCriteria = pulumi.String(rule.AbsoluteCriteria)
			}
			if len(rule.DaysOfWeek) > 0 {
				ruleArgs.DaysOfWeeks = pulumi.ToStringArray(rule.DaysOfWeek)
			}
			if len(rule.MonthsOfYear) > 0 {
				ruleArgs.MonthsOfYears = pulumi.ToStringArray(rule.MonthsOfYear)
			}
			if len(rule.ScheduledBackupTimes) > 0 {
				ruleArgs.ScheduledBackupTimes = pulumi.ToStringArray(rule.ScheduledBackupTimes)
			}
			if len(rule.WeeksOfMonth) > 0 {
				ruleArgs.WeeksOfMonths = pulumi.ToStringArray(rule.WeeksOfMonth)
			}
			rules = append(rules, ruleArgs)
		}
		args.RetentionRules = rules
	}

	createdPolicy, err := dataprotection.NewBackupPolicyDataLakeStorage(ctx,
		spec.Name,
		args,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create data lake storage backup policy %s", spec.Name)
	}

	exportOutputs(ctx, createdPolicy.ID(), spec.Name)
	return nil
}

// exportOutputs keeps the six variant branches' output shapes
// identical -- one policy ID, one name, whichever variant ran.
func exportOutputs(ctx *pulumi.Context, policyId pulumi.IDOutput, policyName string) {
	ctx.Export(OpBackupPolicyId, policyId)
	ctx.Export(OpBackupPolicyName, pulumi.String(policyName))
}
