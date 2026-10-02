package module

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/envsecrets"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/workflows"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// workflow provisions the Cloud Workflows workflow. Every source /
// env-var / service-account change deploys a NEW revision; executions
// already running finish on the revision they started with.
//
// `deletion_protection` is sent EXPLICITLY on every apply: it is Optional
// in the provider with default true, and a spec transition true -> false
// must reach the API rather than being omitted (the send-true-or-omit
// class silently no-ops such transitions — a destroy that should have been
// allowed keeps failing).
func workflow(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpWorkflow.Spec

	// Enable the Workflows API so a fresh project can host the workflow.
	// disable_on_destroy stays false (the provider default): tearing down
	// one workflow must never disable Workflows for everything else in the
	// project. Matches the Terraform module.
	serviceArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("workflows.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
	}
	if spec.ProjectId.GetValue() != "" {
		serviceArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdProjectService, err := projects.NewService(ctx,
		"workflow-workflows.googleapis.com", serviceArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable workflows.googleapis.com api")
	}

	// Secret values are stored in Secret Manager before the workflow
	// exists; the workflow receives each version's resource name as the
	// variable and reads the value at run time. The workflow is deployed
	// after the grants so its first execution can read them.
	storedSecrets, err := envsecrets.Store(ctx, secretPlacement(locals), secretVariables(spec), gcpProvider)
	if err != nil {
		return errors.Wrap(err, "failed to store the environment's secret values")
	}

	args := &workflows.WorkflowArgs{
		Name:           pulumi.String(locals.WorkflowName),
		SourceContents: pulumi.String(spec.SourceContents),
		Labels:         pulumi.ToStringMap(locals.GcpLabels),
		// Explicit send — see the function comment. Unset optional bool
		// reads true via the proto default.
		DeletionProtection: pulumi.Bool(spec.DeletionProtection == nil || spec.GetDeletionProtection()),
	}

	// Honor the spec contract: empty falls back to the provider's default
	// project/region (empty strings would be sent verbatim and rejected).
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.Region != "" {
		args.Region = pulumi.String(spec.Region)
	}

	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.ServiceAccount.GetValue() != "" {
		args.ServiceAccount = pulumi.String(spec.ServiceAccount.GetValue())
	}
	if spec.CryptoKey.GetValue() != "" {
		args.CryptoKeyName = pulumi.String(spec.CryptoKey.GetValue())
	}
	if spec.CallLogLevel != "" {
		args.CallLogLevel = pulumi.StringPtr(spec.CallLogLevel)
	}
	if spec.ExecutionHistoryLevel != "" {
		args.ExecutionHistoryLevel = pulumi.StringPtr(spec.ExecutionHistoryLevel)
	}
	if len(spec.UserEnvVars)+len(spec.SecretEnvVars) > 0 {
		args.UserEnvVars = userEnvVars(spec, storedSecrets.Refs)
	}
	// Resource manager tags are ForceNew: a tag change REPLACES the
	// workflow (fresh execution history).
	if len(spec.ResourceManagerTags) > 0 {
		args.Tags = pulumi.ToStringMap(spec.ResourceManagerTags)
	}
	// Unset defers to the provider default (DELETE).
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	createdWorkflow, err := workflows.NewWorkflow(ctx, "workflow", args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn(append([]pulumi.Resource{createdProjectService}, storedSecrets.Grants...)))
	if err != nil {
		return errors.Wrap(err, "failed to create workflow")
	}

	// The resource ID is the full resource name
	// (projects/{p}/locations/{region}/workflows/{name}) with the ambient
	// project/region resolved — exactly what Eventarc destinations consume.
	ctx.Export(OpWorkflowId, createdWorkflow.ID().ToStringOutput())
	ctx.Export(OpWorkflowName, createdWorkflow.Name)
	ctx.Export(OpRevisionId, createdWorkflow.RevisionId)
	ctx.Export(OpState, createdWorkflow.State)

	return nil
}
