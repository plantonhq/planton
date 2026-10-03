package module

import (
	"github.com/pkg/errors"
	kuberneteskarapacev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskarapace/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys one Karapace schema registry — a MODULE-OWNED-MANIFESTS
// kind (Karapace ships no Helm chart or operator), so the module renders
// core Kubernetes objects directly:
//
//  1. the namespace (optional, create_namespace),
//  2. the SASL password Secret (only when spec.kafka.sasl.password is a
//     literal — never-plaintext-env contract, see secret.go),
//  3. the registry Deployment `<metadata.name>` and its ClusterIP Service,
//  4. the REST-proxy Deployment `<metadata.name>-rest` and its Service
//     (optional, rest_proxy.enabled) — the same engine image with the role
//     flags flipped, wired to the registry Service.
//
// Both roles are configured purely through KARAPACE_* environment
// variables (pydantic env mechanism: config key X → KARAPACE_<upper X>),
// matching upstream's own compose reference. Schema storage is
// Kafka-native — no PVCs, no databases; the schemas live in a compacted
// topic on the connected cluster.
func Resources(ctx *pulumi.Context, iacInput *kuberneteskarapacev1alpha1.KubernetesKarapaceIacInput) error {
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

	var dependencies []pulumi.ResourceOption
	if createdNamespace != nil {
		dependencies = append(dependencies, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// --------------------- sasl password secret --------------------------
	createdSaslSecret, err := saslPasswordSecret(ctx, locals, kubernetesProvider, dependencies)
	if err != nil {
		return err
	}
	if createdSaslSecret != nil {
		dependencies = append(dependencies, pulumi.DependsOn([]pulumi.Resource{createdSaslSecret}))
	}

	// --------------------- registry deployment + service -----------------
	createdRegistryDeployment, err := registryDeployment(ctx, locals, kubernetesProvider, dependencies)
	if err != nil {
		return errors.Wrap(err, "failed to create registry deployment")
	}

	if err := service(ctx, locals, kubernetesProvider, createdRegistryDeployment,
		locals.RegistryName, locals.RegistryPort, locals.RegistrySelectorLabels); err != nil {
		return errors.Wrap(err, "failed to create registry service")
	}

	// --------------------- rest proxy deployment + service ---------------
	if locals.RestEnabled {
		createdRestDeployment, err := restProxyDeployment(ctx, locals, kubernetesProvider, dependencies)
		if err != nil {
			return errors.Wrap(err, "failed to create rest proxy deployment")
		}

		if err := service(ctx, locals, kubernetesProvider, createdRestDeployment,
			locals.RestName, locals.RestPort, locals.RestSelectorLabels); err != nil {
			return errors.Wrap(err, "failed to create rest proxy service")
		}
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpServiceName, pulumi.String(locals.RegistryName))
	ctx.Export(OpEndpoint, pulumi.String(locals.Endpoint))
	ctx.Export(OpRestProxyEndpoint, pulumi.String(locals.RestProxyEndpoint))
	ctx.Export(OpSchemasTopic, pulumi.String(locals.TopicName))

	return nil
}
