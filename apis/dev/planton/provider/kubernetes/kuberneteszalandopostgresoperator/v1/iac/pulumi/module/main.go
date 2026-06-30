package module

import (
	"github.com/pkg/errors"
	kuberneteszalandopostgresoperatorv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kuberneteszalandopostgresoperator/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi entry‑point invoked by the Project‑Planton CLI.
func Resources(ctx *pulumi.Context, stackInput *kuberneteszalandopostgresoperatorv1.KubernetesZalandoPostgresOperatorStackInput) error {
	// Translate incoming protobuf‑generated types into helper data we
	//                need throughout the module (labels, metadata, etc.).
	locals := initializeLocals(ctx, stackInput)

	// Create a Kubernetes provider from the supplied cluster credential.
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx,
		stackInput.ProviderConfig,
		"kubernetes",
	)
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	// ------------------------------ namespace ----------------------------
	// Conditionally create namespace based on create_namespace flag
	createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	// Build conditional namespace dependency (Pulumi equivalent of Terraform depends_on).
	// When create_namespace is false, createdNamespace is nil and namespaceDeps is empty.
	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// Install / upgrade the Zalando Postgres‑Operator.
	if err := postgresOperator(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to install postgres-operator resources")
	}

	return nil
}
