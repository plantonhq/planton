package module

import (
	"github.com/pkg/errors"
	gcpnetworkfirewallpolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpnetworkfirewallpolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the network firewall policy on one of two provider
// resource families -- global (spec.region empty) or regional -- exactly
// as the Terraform module's count guards do. The families share every
// argument but `region`, so the spec is one shape and this switch is the
// only place the arm is decided.
func Resources(ctx *pulumi.Context, stackInput *gcpnetworkfirewallpolicyv1alpha1.GcpNetworkFirewallPolicyStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	// Enable the Compute Engine API first so a fresh project works on the
	// first deploy. disable_on_destroy stays false: tearing down one policy
	// must never disable the API for everything else in the project.
	serviceArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("compute.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if locals.ProjectId != "" {
		serviceArgs.Project = pulumi.String(locals.ProjectId)
	}
	createdProjectService, err := projects.NewService(ctx,
		"networkfirewallpolicy-compute.googleapis.com", serviceArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable compute.googleapis.com api")
	}

	if locals.IsRegional {
		if err := regionalPolicy(ctx, locals, gcpProvider, createdProjectService); err != nil {
			return errors.Wrap(err, "failed to create regional network firewall policy")
		}
		return nil
	}
	if err := globalPolicy(ctx, locals, gcpProvider, createdProjectService); err != nil {
		return errors.Wrap(err, "failed to create global network firewall policy")
	}
	return nil
}
