package module

import (
	"github.com/pkg/errors"
	gcpvertexaimodelgardendeploymentv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaimodelgardendeployment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := deployment(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create model garden deployment")
	}

	return nil
}
