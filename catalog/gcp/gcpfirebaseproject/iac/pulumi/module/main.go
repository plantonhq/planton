package module

import (
	"github.com/pkg/errors"
	gcpfirebaseprojectv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpfirebaseproject/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpfirebaseprojectv1alpha1.GcpFirebaseProjectStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// The Firebase Management API attributes quota to the caller's project
	// on user-credential calls -- Google's own Firebase provider docs say to
	// set user_project_override, and under plain ADC (`gcloud auth
	// application-default login`) a create without it fails with 403
	// "requires a quota project". The override attributes quota to the
	// project being enabled under every credential mode.
	//
	// One provider serves everything here: the Pulumi SDK is bridged from
	// Google's beta Terraform provider, so the beta-only Firebase resources
	// (the project enablement, the default bucket) and the GA App Check
	// resources all come from the same instance. The Terraform module has
	// to attach `provider = google-beta` to the beta-only resources
	// explicitly; that asymmetry is provider packaging, not a behavioral
	// divergence.
	gcpProvider, err := pulumigoogleprovider.GetWithQuotaProject(ctx, stackInput.ProviderConfig,
		stackInput.Target.GetSpec().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := firebaseProject(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to enable firebase on the project")
	}

	return nil
}
