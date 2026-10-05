package module

import (
	"github.com/pkg/errors"
	gcpcomputemigv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcomputemig/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi program entry-point for the GcpComputeMig
// component. It creates the group's whole stack in reference-linked
// order: instance template -> instance group manager -> autoscaler /
// per-instance configs / resize requests. Zonal vs regional resource
// selection follows the spec's zone-XOR-region selector.
func Resources(ctx *pulumi.Context, iacInput *gcpcomputemigv1alpha1.GcpComputeMigIacInput) error {
	locals := initializeLocals(iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	spec := locals.GcpComputeMig.Spec

	// Enable the Compute Engine API — the control plane that owns every
	// resource in this kind. DisableOnDestroy stays false: tearing down
	// one group must never disable the API for everything else in the
	// project.
	computeApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("compute.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		computeApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdComputeApi, err := projects.NewService(ctx,
		"gcpmig-compute.googleapis.com", computeApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable compute.googleapis.com api")
	}

	createdTemplate, err := instanceTemplate(ctx, locals, gcpProvider, createdComputeApi)
	if err != nil {
		return errors.Wrap(err, "failed to create instance template")
	}

	createdGroupManager, err := instanceGroupManager(ctx, locals, gcpProvider, createdTemplate)
	if err != nil {
		return errors.Wrap(err, "failed to create instance group manager")
	}

	if spec.Autoscaler != nil {
		if err := autoscaler(ctx, locals, gcpProvider, createdGroupManager); err != nil {
			return errors.Wrap(err, "failed to create autoscaler")
		}
	}

	if err := perInstanceConfigs(ctx, locals, gcpProvider, createdGroupManager); err != nil {
		return errors.Wrap(err, "failed to create per-instance configs")
	}

	if err := resizeRequests(ctx, locals, gcpProvider, createdGroupManager); err != nil {
		return errors.Wrap(err, "failed to create resize requests")
	}

	// Semantic outputs — names and shapes byte-identical to the
	// Terraform module's outputs.
	ctx.Export(OpInstanceGroup, createdGroupManager.InstanceGroup)
	ctx.Export(OpSelfLink, createdGroupManager.SelfLink)
	ctx.Export(OpCurrentTemplateSelfLink, createdTemplate.TemplateRef)
	ctx.Export(OpMigName, createdGroupManager.Name)
	ctx.Export(OpLocation, pulumi.String(locals.Location))

	return nil
}
