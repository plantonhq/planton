package module

import (
	"github.com/pkg/errors"
	civokubernetesclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civokubernetescluster/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/civo/pulumicivoprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—mirrors the pattern used in digital_ocean_kubernetes_cluster.
func Resources(
	ctx *pulumi.Context,
	stackInput *civokubernetesclusterv1.CivoKubernetesClusterStackInput,
) error {
	// 1. Prepare locals (metadata, spec, credentials).
	locals := initializeLocals(ctx, stackInput)

	// 2. Create a Civo provider from the supplied credential.
	civoProvider, err := pulumicivoprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup Civo provider")
	}

	// 3. Create the cluster.
	if _, err := cluster(ctx, locals, civoProvider); err != nil {
		return errors.Wrap(err, "failed to create Kubernetes cluster")
	}

	return nil
}
