package module

import (
	"github.com/pkg/errors"
	gcploggingsinkv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcploggingsink/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/logging"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// loggingSink provisions the Cloud Logging sink at whichever scope the spec
// selects — one kind, four provider resources, exactly one created:
//
//	scope empty / scope.project_id  -> logging.ProjectSink
//	scope.folder_id                 -> logging.FolderSink
//	scope.organization_id           -> logging.OrganizationSink
//	scope.billing_account           -> logging.BillingAccountSink
//
// The scopes differ by design, and the spec CELs mirror those differences:
// only the project sink models writer-identity control (other scopes ALWAYS
// mint a unique writer), and only folder/org sinks carry the
// include/intercept children flags.
//
// unique_writer_identity is sent EXPLICITLY on the project sink: it is
// Optional in the provider with default true, and a spec transition
// true -> false must reach the API rather than being omitted (the
// send-true-or-omit class). The API enablement resource exists only on the
// project-scope path — folder/org/billing sinks are not project resources,
// so there is no project to enable the API in.
func loggingSink(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpLoggingSink.Spec
	scope := spec.Scope

	switch {
	case scope != nil && scope.FolderId.GetValue() != "":
		return folderSink(ctx, locals, gcpProvider)
	case scope != nil && scope.OrganizationId != "":
		return organizationSink(ctx, locals, gcpProvider)
	case scope != nil && scope.BillingAccount != "":
		return billingAccountSink(ctx, locals, gcpProvider)
	default:
		return projectSink(ctx, locals, gcpProvider)
	}
}

func projectSink(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpLoggingSink.Spec

	// Enable the Cloud Logging API so a fresh project can host the sink.
	// disable_on_destroy stays false (the provider default): tearing down
	// one sink must never disable logging for everything else in the
	// project. Matches the Terraform module.
	serviceArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("logging.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
	}
	projectId := ""
	if spec.Scope != nil {
		projectId = spec.Scope.ProjectId.GetValue()
	}
	if projectId != "" {
		serviceArgs.Project = pulumi.String(projectId)
	}
	createdProjectService, err := projects.NewService(ctx,
		"loggingsink-logging.googleapis.com", serviceArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable logging.googleapis.com api")
	}

	args := &logging.ProjectSinkArgs{
		Name:        pulumi.String(locals.SinkName),
		Destination: pulumi.String(locals.Destination),
		// Explicit send — see the function comment. Unset optional bool
		// reads true via the proto default.
		UniqueWriterIdentity: pulumi.Bool(spec.UniqueWriterIdentity == nil || spec.GetUniqueWriterIdentity()),
	}
	if spec.Filter != "" {
		args.Filter = pulumi.StringPtr(spec.Filter)
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.Disabled {
		args.Disabled = pulumi.BoolPtr(true)
	}
	if spec.CustomWriterIdentity != "" {
		args.CustomWriterIdentity = pulumi.StringPtr(spec.CustomWriterIdentity)
	}
	if bigqueryOptions := expandBigqueryOptionsProject(spec); bigqueryOptions != nil {
		args.BigqueryOptions = bigqueryOptions
	}
	if len(spec.Exclusions) > 0 {
		exclusions := logging.ProjectSinkExclusionArray{}
		for _, exclusion := range spec.Exclusions {
			exclusionArgs := &logging.ProjectSinkExclusionArgs{
				Name:   pulumi.String(exclusion.Name),
				Filter: pulumi.String(exclusion.Filter),
			}
			if exclusion.Description != "" {
				exclusionArgs.Description = pulumi.StringPtr(exclusion.Description)
			}
			if exclusion.Disabled {
				exclusionArgs.Disabled = pulumi.BoolPtr(true)
			}
			exclusions = append(exclusions, exclusionArgs)
		}
		args.Exclusions = exclusions
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}
	if projectId != "" {
		args.Project = pulumi.String(projectId)
	}

	createdSink, err := logging.NewProjectSink(ctx, "logging-sink", args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdProjectService}))
	if err != nil {
		return errors.Wrap(err, "failed to create project logging sink")
	}

	ctx.Export(OpSinkName, createdSink.Name)
	ctx.Export(OpWriterIdentity, createdSink.WriterIdentity)
	return nil
}

func folderSink(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpLoggingSink.Spec

	args := &logging.FolderSinkArgs{
		Name:        pulumi.String(locals.SinkName),
		Folder:      pulumi.String(spec.Scope.FolderId.GetValue()),
		Destination: pulumi.String(locals.Destination),
	}
	if spec.Filter != "" {
		args.Filter = pulumi.StringPtr(spec.Filter)
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.Disabled {
		args.Disabled = pulumi.BoolPtr(true)
	}
	// Children routing is the folder/org sinks' distinguishing capability;
	// both flags default false in the provider and are sent only when set.
	if spec.IncludeChildren {
		args.IncludeChildren = pulumi.BoolPtr(true)
	}
	if spec.InterceptChildren {
		args.InterceptChildren = pulumi.BoolPtr(true)
	}
	if bigqueryOptions := expandBigqueryOptionsFolder(spec); bigqueryOptions != nil {
		args.BigqueryOptions = bigqueryOptions
	}
	if len(spec.Exclusions) > 0 {
		exclusions := logging.FolderSinkExclusionArray{}
		for _, exclusion := range spec.Exclusions {
			exclusionArgs := &logging.FolderSinkExclusionArgs{
				Name:   pulumi.String(exclusion.Name),
				Filter: pulumi.String(exclusion.Filter),
			}
			if exclusion.Description != "" {
				exclusionArgs.Description = pulumi.StringPtr(exclusion.Description)
			}
			if exclusion.Disabled {
				exclusionArgs.Disabled = pulumi.BoolPtr(true)
			}
			exclusions = append(exclusions, exclusionArgs)
		}
		args.Exclusions = exclusions
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	createdSink, err := logging.NewFolderSink(ctx, "logging-sink", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create folder logging sink")
	}

	ctx.Export(OpSinkName, createdSink.Name)
	ctx.Export(OpWriterIdentity, createdSink.WriterIdentity)
	return nil
}

func organizationSink(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpLoggingSink.Spec

	args := &logging.OrganizationSinkArgs{
		Name:        pulumi.String(locals.SinkName),
		OrgId:       pulumi.String(spec.Scope.OrganizationId),
		Destination: pulumi.String(locals.Destination),
	}
	if spec.Filter != "" {
		args.Filter = pulumi.StringPtr(spec.Filter)
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.Disabled {
		args.Disabled = pulumi.BoolPtr(true)
	}
	if spec.IncludeChildren {
		args.IncludeChildren = pulumi.BoolPtr(true)
	}
	if spec.InterceptChildren {
		args.InterceptChildren = pulumi.BoolPtr(true)
	}
	if bigqueryOptions := expandBigqueryOptionsOrganization(spec); bigqueryOptions != nil {
		args.BigqueryOptions = bigqueryOptions
	}
	if len(spec.Exclusions) > 0 {
		exclusions := logging.OrganizationSinkExclusionArray{}
		for _, exclusion := range spec.Exclusions {
			exclusionArgs := &logging.OrganizationSinkExclusionArgs{
				Name:   pulumi.String(exclusion.Name),
				Filter: pulumi.String(exclusion.Filter),
			}
			if exclusion.Description != "" {
				exclusionArgs.Description = pulumi.StringPtr(exclusion.Description)
			}
			if exclusion.Disabled {
				exclusionArgs.Disabled = pulumi.BoolPtr(true)
			}
			exclusions = append(exclusions, exclusionArgs)
		}
		args.Exclusions = exclusions
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	createdSink, err := logging.NewOrganizationSink(ctx, "logging-sink", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create organization logging sink")
	}

	ctx.Export(OpSinkName, createdSink.Name)
	ctx.Export(OpWriterIdentity, createdSink.WriterIdentity)
	return nil
}

func billingAccountSink(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpLoggingSink.Spec

	args := &logging.BillingAccountSinkArgs{
		Name:           pulumi.String(locals.SinkName),
		BillingAccount: pulumi.String(spec.Scope.BillingAccount),
		Destination:    pulumi.String(locals.Destination),
	}
	if spec.Filter != "" {
		args.Filter = pulumi.StringPtr(spec.Filter)
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.Disabled {
		args.Disabled = pulumi.BoolPtr(true)
	}
	if bigqueryOptions := expandBigqueryOptionsBilling(spec); bigqueryOptions != nil {
		args.BigqueryOptions = bigqueryOptions
	}
	if len(spec.Exclusions) > 0 {
		exclusions := logging.BillingAccountSinkExclusionArray{}
		for _, exclusion := range spec.Exclusions {
			exclusionArgs := &logging.BillingAccountSinkExclusionArgs{
				Name:   pulumi.String(exclusion.Name),
				Filter: pulumi.String(exclusion.Filter),
			}
			if exclusion.Description != "" {
				exclusionArgs.Description = pulumi.StringPtr(exclusion.Description)
			}
			if exclusion.Disabled {
				exclusionArgs.Disabled = pulumi.BoolPtr(true)
			}
			exclusions = append(exclusions, exclusionArgs)
		}
		args.Exclusions = exclusions
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	createdSink, err := logging.NewBillingAccountSink(ctx, "logging-sink", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create billing account logging sink")
	}

	ctx.Export(OpSinkName, createdSink.Name)
	ctx.Export(OpWriterIdentity, createdSink.WriterIdentity)
	return nil
}

// The bigquery_options block is rendered ONLY for BigQuery destinations —
// the API rejects it elsewhere. use_partitioned_tables is the block's only
// (and required) argument, so the block's presence and the spec flag
// coincide. Each scope's SDK nests its own block type; four tiny expanders
// keep that typed without generics gymnastics.

func expandBigqueryOptionsProject(spec *gcploggingsinkv1alpha1.GcpLoggingSinkSpec) *logging.ProjectSinkBigqueryOptionsArgs {
	if spec.Destination.GetBigqueryDataset().GetValue() == "" || !spec.Destination.UsePartitionedTables {
		return nil
	}
	return &logging.ProjectSinkBigqueryOptionsArgs{UsePartitionedTables: pulumi.Bool(true)}
}

func expandBigqueryOptionsFolder(spec *gcploggingsinkv1alpha1.GcpLoggingSinkSpec) *logging.FolderSinkBigqueryOptionsArgs {
	if spec.Destination.GetBigqueryDataset().GetValue() == "" || !spec.Destination.UsePartitionedTables {
		return nil
	}
	return &logging.FolderSinkBigqueryOptionsArgs{UsePartitionedTables: pulumi.Bool(true)}
}

func expandBigqueryOptionsOrganization(spec *gcploggingsinkv1alpha1.GcpLoggingSinkSpec) *logging.OrganizationSinkBigqueryOptionsArgs {
	if spec.Destination.GetBigqueryDataset().GetValue() == "" || !spec.Destination.UsePartitionedTables {
		return nil
	}
	return &logging.OrganizationSinkBigqueryOptionsArgs{UsePartitionedTables: pulumi.Bool(true)}
}

func expandBigqueryOptionsBilling(spec *gcploggingsinkv1alpha1.GcpLoggingSinkSpec) *logging.BillingAccountSinkBigqueryOptionsArgs {
	if spec.Destination.GetBigqueryDataset().GetValue() == "" || !spec.Destination.UsePartitionedTables {
		return nil
	}
	return &logging.BillingAccountSinkBigqueryOptionsArgs{UsePartitionedTables: pulumi.Bool(true)}
}
