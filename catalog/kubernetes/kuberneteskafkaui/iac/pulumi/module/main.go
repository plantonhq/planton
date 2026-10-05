package module

import (
	"github.com/pkg/errors"
	kuberneteskafkauiv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskafkaui/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs the kafbat UI console from the served kafka-ui Helm
// chart as a real Helm release. The typed spec renders into chart values
// (helm_release.go): the cluster wiring becomes yamlApplicationConfig with
// ${ENV_VAR} placeholders in every password position, each placeholder
// wired to its source Secret through envs.secretMappings; the declared
// console login password materializes as the "<name>-secrets" Secret
// (secret.go); the helm_values escape hatch merges last with Helm -f
// semantics — the exact semantic twin of the Terraform module.
//
// The release is named after metadata.name (NOT a fixed chart name):
// several consoles coexist in one cluster, each rendering its own
// Deployment and Service under the chart's fullname for that release.
func Resources(ctx *pulumi.Context, iacInput *kuberneteskafkauiv1alpha1.KubernetesKafkaUiIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// ------------------------------ namespace ----------------------------
	createdNamespace, err := namespace(ctx, iacInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// ------------------------------ console secret ------------------------
	createdConsoleSecret, err := consoleSecret(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return err
	}

	releaseDeps := namespaceDeps
	if createdConsoleSecret != nil {
		releaseDeps = append(releaseDeps, pulumi.DependsOn([]pulumi.Resource{createdConsoleSecret}))
	}

	// ------------------------------ helm release --------------------------
	if err := helmRelease(ctx, locals, kubernetesProvider, releaseDeps); err != nil {
		return err
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles: the namespace, the
// Service name (the chart's fullname for this release — mirrored, not
// overridden; locals.go), the in-cluster endpoint exposure kinds attach
// to, and the port-forward one-liner for access without any exposure.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpServiceName, pulumi.String(locals.ServiceName))
	ctx.Export(OpEndpoint, pulumi.String(locals.Endpoint))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
}
