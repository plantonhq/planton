package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/logging"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// logBucket provisions the Cloud Logging bucket — one kind, four scope
// resources, exactly one created (the scope arm decides which, mirroring
// the Terraform module's count guards) — plus the bucket's log views, the
// linked BigQuery dataset, and the folder/organization settings singleton
// when the spec arms them.
//
// Provider truths the wiring honors:
//   - retention_days and location are sent EXPLICITLY (spec defaults 30 /
//     "global") so the spec's defaults are what the API applies.
//   - enable_analytics is sent ONLY when the spec explicitly sets it: the
//     provider transmits analyticsEnabled solely on explicit configuration
//     (an atomic pre-update, separate from other fields), and enabling is
//     ONE-WAY. Blanket-sending false would diverge from provider behavior.
//   - all four bucket variants ADOPT an existing bucket on bucket_id match
//     (that is how _Default is managed, and the only mode at all on
//     folder/org/billing scopes — the API creates custom buckets only
//     under projects).
func logBucket(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpLogBucket.Spec

	// The bucket's full resource name (…/locations/{l}/buckets/{b}) —
	// computed by whichever variant is created; the views and the linked
	// dataset attach to it, and it is THE composition output.
	var bucketName pulumi.StringOutput

	var bucketResource pulumi.Resource

	switch {
	case locals.IsProjectBucket:
		// The project bucket resource REQUIRES an explicit project. Honor
		// the spec contract for the empty case by reading the provider's
		// resolved default project from client config (the Pulumi
		// equivalent of the Terraform module's count-gated
		// google_client_config data source).
		var project pulumi.StringInput
		if spec.Scope.GetProjectId().GetValue() != "" {
			project = pulumi.String(spec.Scope.GetProjectId().GetValue())
		} else {
			clientConfig, err := organizations.GetClientConfig(ctx, pulumi.Provider(gcpProvider))
			if err != nil {
				return errors.Wrap(err, "failed to read provider client config for the default project")
			}
			if clientConfig.Project == "" {
				return errors.New("scope.project_id is empty and the provider has no default project configured")
			}
			project = pulumi.String(clientConfig.Project)
		}

		// Enable the Cloud Logging API so a fresh project can host the
		// bucket. Project scope only: folder/org/billing buckets are not
		// project resources. disable_on_destroy stays false — tearing down
		// one bucket must never disable logging project-wide.
		createdProjectService, err := projects.NewService(ctx, "logbucket-logging.googleapis.com",
			&projects.ServiceArgs{
				Project:                  project,
				Service:                  pulumi.String("logging.googleapis.com"),
				DisableDependentServices: pulumi.BoolPtr(true),
			}, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to enable logging.googleapis.com api")
		}

		args := &logging.ProjectBucketConfigArgs{
			Project:       project,
			BucketId:      pulumi.String(spec.BucketId),
			Location:      pulumi.String(locals.Location),
			RetentionDays: pulumi.IntPtr(locals.RetentionDays),
			// locked matches the provider default (false) when unset; a
			// live unlock is refused server-side (one-way), which is the
			// honest failure for a true -> false transition.
			Locked: pulumi.Bool(spec.Locked),
		}
		// Send analyticsEnabled only on explicit configuration — see the
		// function comment.
		if spec.EnableAnalytics != nil {
			args.EnableAnalytics = pulumi.BoolPtr(spec.GetEnableAnalytics())
		}
		if spec.Description != "" {
			args.Description = pulumi.StringPtr(spec.Description)
		}
		if spec.CmekKmsKey.GetValue() != "" {
			args.CmekSettings = &logging.ProjectBucketConfigCmekSettingsArgs{
				KmsKeyName: pulumi.String(spec.CmekKmsKey.GetValue()),
			}
		}
		if len(spec.IndexConfigs) > 0 {
			indexConfigs := logging.ProjectBucketConfigIndexConfigArray{}
			for _, indexConfig := range spec.IndexConfigs {
				indexConfigs = append(indexConfigs, &logging.ProjectBucketConfigIndexConfigArgs{
					FieldPath: pulumi.String(indexConfig.FieldPath),
					Type:      pulumi.String(indexConfig.Type),
				})
			}
			args.IndexConfigs = indexConfigs
		}
		if spec.DeletionPolicy != "" {
			args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
		}
		createdBucket, err := logging.NewProjectBucketConfig(ctx, "bucket", args,
			pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdProjectService}))
		if err != nil {
			return errors.Wrap(err, "failed to create project log bucket")
		}
		bucketName = createdBucket.Name
		bucketResource = createdBucket

	case locals.IsFolderBucket:
		args := &logging.FolderBucketConfigArgs{
			Folder:        pulumi.String(spec.Scope.FolderId.GetValue()),
			BucketId:      pulumi.String(spec.BucketId),
			Location:      pulumi.String(locals.Location),
			RetentionDays: pulumi.IntPtr(locals.RetentionDays),
		}
		if spec.Description != "" {
			args.Description = pulumi.StringPtr(spec.Description)
		}
		if spec.CmekKmsKey.GetValue() != "" {
			args.CmekSettings = &logging.FolderBucketConfigCmekSettingsArgs{
				KmsKeyName: pulumi.String(spec.CmekKmsKey.GetValue()),
			}
		}
		if len(spec.IndexConfigs) > 0 {
			indexConfigs := logging.FolderBucketConfigIndexConfigArray{}
			for _, indexConfig := range spec.IndexConfigs {
				indexConfigs = append(indexConfigs, &logging.FolderBucketConfigIndexConfigArgs{
					FieldPath: pulumi.String(indexConfig.FieldPath),
					Type:      pulumi.String(indexConfig.Type),
				})
			}
			args.IndexConfigs = indexConfigs
		}
		if spec.DeletionPolicy != "" {
			args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
		}
		createdBucket, err := logging.NewFolderBucketConfig(ctx, "bucket", args, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to create folder log bucket")
		}
		bucketName = createdBucket.Name
		bucketResource = createdBucket

	case locals.IsOrgBucket:
		args := &logging.OrganizationBucketConfigArgs{
			Organization:  pulumi.String(spec.Scope.OrganizationId),
			BucketId:      pulumi.String(spec.BucketId),
			Location:      pulumi.String(locals.Location),
			RetentionDays: pulumi.IntPtr(locals.RetentionDays),
		}
		if spec.Description != "" {
			args.Description = pulumi.StringPtr(spec.Description)
		}
		if spec.CmekKmsKey.GetValue() != "" {
			args.CmekSettings = &logging.OrganizationBucketConfigCmekSettingsArgs{
				KmsKeyName: pulumi.String(spec.CmekKmsKey.GetValue()),
			}
		}
		if len(spec.IndexConfigs) > 0 {
			indexConfigs := logging.OrganizationBucketConfigIndexConfigArray{}
			for _, indexConfig := range spec.IndexConfigs {
				indexConfigs = append(indexConfigs, &logging.OrganizationBucketConfigIndexConfigArgs{
					FieldPath: pulumi.String(indexConfig.FieldPath),
					Type:      pulumi.String(indexConfig.Type),
				})
			}
			args.IndexConfigs = indexConfigs
		}
		if spec.DeletionPolicy != "" {
			args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
		}
		createdBucket, err := logging.NewOrganizationBucketConfig(ctx, "bucket", args, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to create organization log bucket")
		}
		bucketName = createdBucket.Name
		bucketResource = createdBucket

	case locals.IsBillingBucket:
		args := &logging.BillingAccountBucketConfigArgs{
			BillingAccount: pulumi.String(spec.Scope.BillingAccount),
			BucketId:       pulumi.String(spec.BucketId),
			Location:       pulumi.String(locals.Location),
			RetentionDays:  pulumi.IntPtr(locals.RetentionDays),
		}
		if spec.Description != "" {
			args.Description = pulumi.StringPtr(spec.Description)
		}
		if spec.CmekKmsKey.GetValue() != "" {
			args.CmekSettings = &logging.BillingAccountBucketConfigCmekSettingsArgs{
				KmsKeyName: pulumi.String(spec.CmekKmsKey.GetValue()),
			}
		}
		if len(spec.IndexConfigs) > 0 {
			indexConfigs := logging.BillingAccountBucketConfigIndexConfigArray{}
			for _, indexConfig := range spec.IndexConfigs {
				indexConfigs = append(indexConfigs, &logging.BillingAccountBucketConfigIndexConfigArgs{
					FieldPath: pulumi.String(indexConfig.FieldPath),
					Type:      pulumi.String(indexConfig.Type),
				})
			}
			args.IndexConfigs = indexConfigs
		}
		if spec.DeletionPolicy != "" {
			args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
		}
		createdBucket, err := logging.NewBillingAccountBucketConfig(ctx, "bucket", args, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to create billing-account log bucket")
		}
		bucketName = createdBucket.Name
		bucketResource = createdBucket
	}

	// Log views: named, independently grantable slices of the bucket. The
	// view's location and parent derive from the bucket's full name; the
	// kind's deletion_policy fans out to every view.
	for _, logView := range spec.LogViews {
		viewArgs := &logging.LogViewArgs{
			Name:   pulumi.String(logView.ViewId),
			Bucket: bucketName,
		}
		if logView.Filter != "" {
			viewArgs.Filter = pulumi.StringPtr(logView.Filter)
		}
		if logView.Description != "" {
			viewArgs.Description = pulumi.StringPtr(logView.Description)
		}
		if spec.DeletionPolicy != "" {
			viewArgs.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
		}
		if _, err := logging.NewLogView(ctx, "log-view-"+logView.ViewId, viewArgs,
			pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{bucketResource})); err != nil {
			return errors.Wrapf(err, "failed to create log view %s", logView.ViewId)
		}
	}

	// The linked BigQuery dataset (requires analytics — proto-CEL-paired).
	if linked := spec.LinkedBigqueryDataset; linked != nil {
		linkedArgs := &logging.LinkedDatasetArgs{
			LinkId: pulumi.String(linked.LinkId),
			Bucket: bucketName,
		}
		if linked.Description != "" {
			linkedArgs.Description = pulumi.StringPtr(linked.Description)
		}
		if spec.DeletionPolicy != "" {
			linkedArgs.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
		}
		createdLinkedDataset, err := logging.NewLinkedDataset(ctx, "linked-dataset", linkedArgs,
			pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{bucketResource}))
		if err != nil {
			return errors.Wrap(err, "failed to create linked dataset")
		}
		// The BigQuery dataset id GCP assigned to the link.
		ctx.Export(OpLinkedDatasetId, createdLinkedDataset.BigqueryDatasets.Index(pulumi.Int(0)).DatasetId().Elem())
	} else {
		ctx.Export(OpLinkedDatasetId, pulumi.String(""))
	}

	// The folder/organization logging-settings singleton (scope-gated by
	// the proto CELs). Adopted on create; destroy is a state-only no-op —
	// the settings resources carry no deletion_policy by provider truth.
	if settings := spec.ScopeSettings; settings != nil {
		if locals.IsFolderBucket {
			settingsArgs := &logging.FolderSettingsArgs{
				Folder: pulumi.String(spec.Scope.FolderId.GetValue()),
				// The block's reason to exist — sent explicitly.
				DisableDefaultSink: pulumi.Bool(settings.DisableDefaultSink),
			}
			if settings.KmsKey.GetValue() != "" {
				settingsArgs.KmsKeyName = pulumi.StringPtr(settings.KmsKey.GetValue())
			}
			if settings.StorageLocation != "" {
				settingsArgs.StorageLocation = pulumi.StringPtr(settings.StorageLocation)
			}
			if _, err := logging.NewFolderSettings(ctx, "scope-settings", settingsArgs,
				pulumi.Provider(gcpProvider)); err != nil {
				return errors.Wrap(err, "failed to configure folder logging settings")
			}
		}
		if locals.IsOrgBucket {
			settingsArgs := &logging.OrganizationSettingsArgs{
				Organization:       pulumi.String(spec.Scope.OrganizationId),
				DisableDefaultSink: pulumi.Bool(settings.DisableDefaultSink),
			}
			if settings.KmsKey.GetValue() != "" {
				settingsArgs.KmsKeyName = pulumi.StringPtr(settings.KmsKey.GetValue())
			}
			if settings.StorageLocation != "" {
				settingsArgs.StorageLocation = pulumi.StringPtr(settings.StorageLocation)
			}
			if _, err := logging.NewOrganizationSettings(ctx, "scope-settings", settingsArgs,
				pulumi.Provider(gcpProvider)); err != nil {
				return errors.Wrap(err, "failed to configure organization logging settings")
			}
		}
	}

	ctx.Export(OpBucketName, bucketName)

	return nil
}
