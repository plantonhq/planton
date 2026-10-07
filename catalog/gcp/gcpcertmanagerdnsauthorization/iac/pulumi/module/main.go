package module

import (
	"github.com/pkg/errors"
	gcpcertmanagerdnsauthorizationv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcertmanagerdnsauthorization/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpcertmanagerdnsauthorizationv1alpha1.GcpCertManagerDnsAuthorizationIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to setup gcp provider")
	}

	// Enable the Certificate Manager API so a fresh project can host DNS
	// authorizations. disable_on_destroy is false: tearing down one
	// authorization must never disable the API for everything else in the
	// project.
	serviceArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("certificatemanager.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if locals.ProjectId != "" {
		serviceArgs.Project = pulumi.String(locals.ProjectId)
	}
	createdProjectService, err := projects.NewService(ctx,
		"certificatemanager-certificatemanager.googleapis.com", serviceArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable certificatemanager.googleapis.com api")
	}

	return dnsAuthorization(ctx, locals, gcpProvider, createdProjectService)
}
