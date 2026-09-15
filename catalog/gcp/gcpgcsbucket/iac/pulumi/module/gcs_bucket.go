package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func gcsBucket(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) (*storage.Bucket, error) {
	spec := locals.GcpGcsBucket.Spec

	// Enable the Cloud Storage API — the control plane that owns buckets.
	// DisableOnDestroy stays false: tearing down one bucket must never
	// disable the API for everything else in the project (other buckets
	// keep serving).
	storageApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("storage.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	// Honor the spec contract: an empty project_id falls back to the
	// provider's default project.
	if spec.ProjectId.GetValue() != "" {
		storageApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdStorageApi, err := projects.NewService(ctx,
		"gcsbkt-storage.googleapis.com", storageApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to enable storage.googleapis.com api")
	}

	// The GCS bucket. Sharp edges, all taught by the API rather than
	// invented here:
	//
	//   - name, location, project, custom placement, hierarchical
	//     namespace, and enable_object_retention are immutable — changing
	//     any of them replaces the bucket and everything in it.
	//
	//   - force_destroy defaults to false: destroying a non-empty bucket
	//     fails instead of silently erasing data. When true, the engine
	//     deletes every object version first (which can take a long time
	//     on large buckets and refuses objects still under a locked
	//     retention policy).
	//
	//   - A locked retention policy is irreversible; any attempt to
	//     unlock forces bucket re-creation.
	//
	//   - Soft delete is a server-side default (7 days) even when the
	//     block is omitted; the module sends the block only when the spec
	//     sets it, so unset specs follow GCP's default without a
	//     perpetual diff.
	args := &storage.BucketArgs{
		Name:                     pulumi.String(spec.BucketName),
		Location:                 pulumi.String(spec.Location),
		Labels:                   pulumi.ToStringMap(locals.GcpLabels),
		ForceDestroy:             pulumi.Bool(spec.ForceDestroy),
		UniformBucketLevelAccess: pulumi.Bool(spec.UniformBucketLevelAccessEnabled),
	}

	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
	}

	// Storage class defaults to STANDARD client-side on Terraform; send
	// only when set and let the API default otherwise — the realized
	// bucket is STANDARD either way.
	if spec.StorageClass != "" {
		args.StorageClass = pulumi.StringPtr(spec.StorageClass)
	}

	if spec.PublicAccessPrevention != "" {
		args.PublicAccessPrevention = pulumi.StringPtr(spec.PublicAccessPrevention)
	}

	if spec.RequesterPays {
		args.RequesterPays = pulumi.BoolPtr(true)
	}

	if spec.DefaultEventBasedHold {
		args.DefaultEventBasedHold = pulumi.BoolPtr(true)
	}

	// Create-time-only surface (immutable).
	if spec.EnableObjectRetention {
		args.EnableObjectRetention = pulumi.BoolPtr(true)
	}

	if spec.HierarchicalNamespaceEnabled {
		args.HierarchicalNamespace = &storage.BucketHierarchicalNamespaceArgs{
			Enabled: pulumi.Bool(true),
		}
	}

	if spec.VersioningEnabled {
		args.Versioning = &storage.BucketVersioningArgs{
			Enabled: pulumi.Bool(true),
		}
	}

	// Autoclass: GCS moves each object between classes on observed access.
	// The spec's CEL guard already rejects combining it with
	// SetStorageClass lifecycle rules.
	if spec.Autoclass != nil {
		autoclassArgs := &storage.BucketAutoclassArgs{
			Enabled: pulumi.Bool(spec.Autoclass.GetEnabled()),
		}
		if spec.Autoclass.TerminalStorageClass != "" {
			autoclassArgs.TerminalStorageClass = pulumi.StringPtr(spec.Autoclass.TerminalStorageClass)
		}
		args.Autoclass = autoclassArgs
	}

	// Nullable numeric conditions ride on presence: nil means "criterion
	// unset" while an explicit 0 is a meaningful value — the *_if_zero
	// flags tell the provider to send the zero.
	if len(spec.LifecycleRules) > 0 {
		rules := storage.BucketLifecycleRuleArray{}
		for _, rule := range spec.LifecycleRules {
			actionArgs := &storage.BucketLifecycleRuleActionArgs{
				Type: pulumi.String(rule.Action.Type),
			}
			if rule.Action.StorageClass != "" {
				actionArgs.StorageClass = pulumi.StringPtr(rule.Action.StorageClass)
			}

			conditionArgs := &storage.BucketLifecycleRuleConditionArgs{}
			if rule.Condition.AgeDays != nil {
				conditionArgs.Age = pulumi.IntPtr(int(*rule.Condition.AgeDays))
				if *rule.Condition.AgeDays == 0 {
					conditionArgs.SendAgeIfZero = pulumi.BoolPtr(true)
				}
			}
			if rule.Condition.CreatedBefore != "" {
				conditionArgs.CreatedBefore = pulumi.StringPtr(rule.Condition.CreatedBefore)
			}
			if rule.Condition.WithState != "" {
				conditionArgs.WithState = pulumi.StringPtr(rule.Condition.WithState)
			}
			if len(rule.Condition.MatchesStorageClass) > 0 {
				conditionArgs.MatchesStorageClasses = pulumi.ToStringArray(rule.Condition.MatchesStorageClass)
			}
			if len(rule.Condition.MatchesPrefix) > 0 {
				conditionArgs.MatchesPrefixes = pulumi.ToStringArray(rule.Condition.MatchesPrefix)
			}
			if len(rule.Condition.MatchesSuffix) > 0 {
				conditionArgs.MatchesSuffixes = pulumi.ToStringArray(rule.Condition.MatchesSuffix)
			}
			if rule.Condition.NumNewerVersions != nil {
				conditionArgs.NumNewerVersions = pulumi.IntPtr(int(*rule.Condition.NumNewerVersions))
				if *rule.Condition.NumNewerVersions == 0 {
					conditionArgs.SendNumNewerVersionsIfZero = pulumi.BoolPtr(true)
				}
			}
			if rule.Condition.DaysSinceNoncurrentTime != nil {
				conditionArgs.DaysSinceNoncurrentTime = pulumi.IntPtr(int(*rule.Condition.DaysSinceNoncurrentTime))
				if *rule.Condition.DaysSinceNoncurrentTime == 0 {
					conditionArgs.SendDaysSinceNoncurrentTimeIfZero = pulumi.BoolPtr(true)
				}
			}
			if rule.Condition.NoncurrentTimeBefore != "" {
				conditionArgs.NoncurrentTimeBefore = pulumi.StringPtr(rule.Condition.NoncurrentTimeBefore)
			}
			if rule.Condition.DaysSinceCustomTime != nil {
				conditionArgs.DaysSinceCustomTime = pulumi.IntPtr(int(*rule.Condition.DaysSinceCustomTime))
				if *rule.Condition.DaysSinceCustomTime == 0 {
					conditionArgs.SendDaysSinceCustomTimeIfZero = pulumi.BoolPtr(true)
				}
			}
			if rule.Condition.CustomTimeBefore != "" {
				conditionArgs.CustomTimeBefore = pulumi.StringPtr(rule.Condition.CustomTimeBefore)
			}
			// Size-band conditions: explicit presence, so a set 0 (matches
			// every object) is distinguishable from unset.
			if rule.Condition.SizeAboveBytes != nil {
				conditionArgs.SizeAboveBytes = pulumi.IntPtr(int(*rule.Condition.SizeAboveBytes))
			}
			if rule.Condition.SizeBelowBytes != nil {
				conditionArgs.SizeBelowBytes = pulumi.IntPtr(int(*rule.Condition.SizeBelowBytes))
			}

			rules = append(rules, &storage.BucketLifecycleRuleArgs{
				Action:    actionArgs,
				Condition: conditionArgs,
			})
		}
		args.LifecycleRules = rules
	}

	// WORM retention. Locking cannot happen at create — GCP locks in a
	// follow-up call after the policy exists.
	if spec.RetentionPolicy != nil {
		args.RetentionPolicy = &storage.BucketRetentionPolicyArgs{
			RetentionPeriod: pulumi.String(fmt.Sprintf("%d", spec.RetentionPolicy.RetentionPeriodSeconds)),
			IsLocked:        pulumi.BoolPtr(spec.RetentionPolicy.IsLocked),
		}
	}

	// Sent only when the spec sets it; an omitted block follows GCP's
	// 7-day default. A set 0 disables soft delete.
	if spec.SoftDeletePolicy != nil && spec.SoftDeletePolicy.RetentionDurationSeconds != nil {
		args.SoftDeletePolicy = &storage.BucketSoftDeletePolicyArgs{
			RetentionDurationSeconds: pulumi.IntPtr(int(*spec.SoftDeletePolicy.RetentionDurationSeconds)),
		}
	}

	// One provider block carries both the default CMEK key and the
	// per-encryption-type enforcement for new objects, so the module emits
	// it when either half is configured.
	//
	// Default CMEK: the GCS service agent must hold
	// roles/cloudkms.cryptoKeyEncrypterDecrypter on this key before create.
	// Enforcement changes apply to NEW objects only.
	if spec.KmsKeyName.GetValue() != "" || spec.EncryptionEnforcement != nil {
		encryptionArgs := &storage.BucketEncryptionArgs{}
		if spec.KmsKeyName.GetValue() != "" {
			encryptionArgs.DefaultKmsKeyName = pulumi.StringPtr(spec.KmsKeyName.GetValue())
		}
		if spec.EncryptionEnforcement != nil {
			if spec.EncryptionEnforcement.GoogleManagedRestrictionMode != "" {
				encryptionArgs.GoogleManagedEncryptionEnforcementConfig = &storage.BucketEncryptionGoogleManagedEncryptionEnforcementConfigArgs{
					RestrictionMode: pulumi.String(spec.EncryptionEnforcement.GoogleManagedRestrictionMode),
				}
			}
			if spec.EncryptionEnforcement.CustomerManagedRestrictionMode != "" {
				encryptionArgs.CustomerManagedEncryptionEnforcementConfig = &storage.BucketEncryptionCustomerManagedEncryptionEnforcementConfigArgs{
					RestrictionMode: pulumi.String(spec.EncryptionEnforcement.CustomerManagedRestrictionMode),
				}
			}
			if spec.EncryptionEnforcement.CustomerSuppliedRestrictionMode != "" {
				encryptionArgs.CustomerSuppliedEncryptionEnforcementConfig = &storage.BucketEncryptionCustomerSuppliedEncryptionEnforcementConfigArgs{
					RestrictionMode: pulumi.String(spec.EncryptionEnforcement.CustomerSuppliedRestrictionMode),
				}
			}
		}
		args.Encryption = encryptionArgs
	}

	if spec.Website != nil {
		websiteArgs := &storage.BucketWebsiteArgs{}
		if spec.Website.MainPageSuffix != "" {
			websiteArgs.MainPageSuffix = pulumi.StringPtr(spec.Website.MainPageSuffix)
		}
		if spec.Website.NotFoundPage != "" {
			websiteArgs.NotFoundPage = pulumi.StringPtr(spec.Website.NotFoundPage)
		}
		args.Website = websiteArgs
	}

	if len(spec.CorsRules) > 0 {
		corsRules := storage.BucketCorArray{}
		for _, corsRule := range spec.CorsRules {
			corArgs := &storage.BucketCorArgs{
				Origins: pulumi.ToStringArray(corsRule.Origins),
				Methods: pulumi.ToStringArray(corsRule.Methods),
			}
			if len(corsRule.ResponseHeaders) > 0 {
				corArgs.ResponseHeaders = pulumi.ToStringArray(corsRule.ResponseHeaders)
			}
			if corsRule.MaxAgeSeconds > 0 {
				corArgs.MaxAgeSeconds = pulumi.IntPtr(int(corsRule.MaxAgeSeconds))
			}
			corsRules = append(corsRules, corArgs)
		}
		args.Cors = corsRules
	}

	if spec.Logging != nil {
		loggingArgs := &storage.BucketLoggingArgs{
			LogBucket: pulumi.String(spec.Logging.LogBucket.GetValue()),
		}
		if spec.Logging.LogObjectPrefix != "" {
			loggingArgs.LogObjectPrefix = pulumi.StringPtr(spec.Logging.LogObjectPrefix)
		}
		args.Logging = loggingArgs
	}

	// Custom dual-region placement (exactly two regions, enforced
	// pre-deploy).
	if spec.CustomPlacementConfig != nil {
		args.CustomPlacementConfig = &storage.BucketCustomPlacementConfigArgs{
			DataLocations: pulumi.ToStringArray(spec.CustomPlacementConfig.DataLocations),
		}
	}

	if spec.Rpo != "" {
		args.Rpo = pulumi.StringPtr(spec.Rpo)
	}

	// Destroy-time guard: PREVENT fails the destroy; ABANDON unmanages
	// the bucket without deleting it. Unset falls back to the provider
	// default (DELETE). Orthogonal to force_destroy, which governs
	// whether a permitted deletion may erase contained objects.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	// Network-layer IP filtering: which CIDR ranges / VPC networks may
	// reach the bucket at all, evaluated before IAM. The spec's CEL guard
	// rejects an Enabled filter with no sources pre-deploy.
	if spec.IpFilter != nil {
		ipFilterArgs := &storage.BucketIpFilterArgs{
			Mode: pulumi.String(spec.IpFilter.Mode),
		}
		if spec.IpFilter.AllowCrossOrgVpcs {
			ipFilterArgs.AllowCrossOrgVpcs = pulumi.BoolPtr(true)
		}
		if spec.IpFilter.AllowAllServiceAgentAccess {
			ipFilterArgs.AllowAllServiceAgentAccess = pulumi.BoolPtr(true)
		}
		if spec.IpFilter.PublicNetworkSource != nil {
			ipFilterArgs.PublicNetworkSource = &storage.BucketIpFilterPublicNetworkSourceArgs{
				AllowedIpCidrRanges: pulumi.ToStringArray(spec.IpFilter.PublicNetworkSource.AllowedIpCidrRanges),
			}
		}
		if len(spec.IpFilter.VpcNetworkSources) > 0 {
			sources := storage.BucketIpFilterVpcNetworkSourceArray{}
			for _, source := range spec.IpFilter.VpcNetworkSources {
				sources = append(sources, &storage.BucketIpFilterVpcNetworkSourceArgs{
					Network:             pulumi.String(source.Network.GetValue()),
					AllowedIpCidrRanges: pulumi.ToStringArray(source.AllowedIpCidrRanges),
				})
			}
			ipFilterArgs.VpcNetworkSources = sources
		}
		args.IpFilter = ipFilterArgs
	}

	createdBucket, err := storage.NewBucket(ctx, "bucket", args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdStorageApi}))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create bucket")
	}

	// Additive IAM grants: one (role, member) pair per resource, merging
	// into the bucket's policy without touching grants made elsewhere —
	// authoritative bindings/policies are deliberately not used. Resource
	// names key on (role, member) — the grant's identity — matching the
	// Terraform module's for_each keys.
	for _, iamMember := range spec.IamMembers {
		member := iamMember.Member.GetValue()
		iamArgs := &storage.BucketIAMMemberArgs{
			Bucket: createdBucket.Name,
			Role:   pulumi.String(iamMember.Role),
			Member: pulumi.String(member),
		}
		if iamMember.Condition != nil {
			conditionArgs := &storage.BucketIAMMemberConditionArgs{
				Title:      pulumi.String(iamMember.Condition.Title),
				Expression: pulumi.String(iamMember.Condition.Expression),
			}
			if iamMember.Condition.Description != "" {
				conditionArgs.Description = pulumi.StringPtr(iamMember.Condition.Description)
			}
			iamArgs.Condition = conditionArgs
		}
		_, err := storage.NewBucketIAMMember(ctx,
			fmt.Sprintf("iam-member-%s-%s", iamMember.Role, member), iamArgs,
			pulumi.Provider(gcpProvider),
			pulumi.Parent(createdBucket))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to grant %s to %s on the bucket", iamMember.Role, member)
		}
	}

	// For GCS the resource ID equals the globally unique bucket name — the
	// value every consumer (backend buckets, function sources, Dataproc
	// staging, Pub/Sub sinks) references.
	ctx.Export(OpBucketId, createdBucket.ID())
	ctx.Export(OpBucketName, createdBucket.Name)
	ctx.Export(OpUrl, createdBucket.Url)
	ctx.Export(OpSelfLink, createdBucket.SelfLink)
	// GCS reports location upper-cased regardless of input case.
	ctx.Export(OpLocation, createdBucket.Location)
	ctx.Export(OpProjectNumber, createdBucket.ProjectNumber)

	return createdBucket, nil
}
