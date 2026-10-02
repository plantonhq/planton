package module

import (
	"github.com/pkg/errors"
	gcpcloudarmorpolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudarmorpolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpcloudarmorpolicyv1alpha1.GcpCloudArmorPolicyStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	// Enable the Compute Engine API so a fresh project can host security
	// policies. disable_on_destroy is false: tearing down one policy must
	// never disable the API for everything else in the project.
	serviceArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("compute.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if locals.ProjectId != "" {
		serviceArgs.Project = pulumi.String(locals.ProjectId)
	}
	createdProjectService, err := projects.NewService(ctx,
		"compute-compute.googleapis.com", serviceArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable compute.googleapis.com api")
	}

	// spec.region selects the API collection: empty builds the global
	// security policy, a region name builds the regional one -- the same
	// switch the Terraform module makes with its count guards.
	if locals.IsRegional {
		if err := regionSecurityPolicy(ctx, locals, gcpProvider, createdProjectService); err != nil {
			return errors.Wrap(err, "failed to create regional cloud armor security policy")
		}
		return nil
	}

	if err := securityPolicy(ctx, locals, gcpProvider, createdProjectService); err != nil {
		return errors.Wrap(err, "failed to create cloud armor security policy")
	}

	return nil
}
