package module

import (
	"github.com/pkg/errors"
	kuberneteskeycloakv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskeycloak/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys one operator-managed Keycloak server:
//
//  1. the namespace (optional, create_namespace),
//  2. the k8s.keycloak.org/v2beta1 Keycloak CR itself — the StatefulSet
//     (named exactly after this resource), `<name>-service`,
//     `<name>-discovery`, the NetworkPolicy and the create-once
//     `<name>-initial-admin` bootstrap Secret are all operator-created
//     from it. No ingress resources — the operator's own default Ingress
//     is explicitly disabled and exposure composes from Gateway API kinds
//     referencing the exported service handles.
//
// PREREQUISITE: a KubernetesKeycloakOperator watching this namespace
// (with the default namespaced watch, the operator and its Keycloak
// declarations live in the SAME namespace).
func Resources(ctx *pulumi.Context, iacInput *kuberneteskeycloakv1alpha1.KubernetesKeycloakIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// FAIL LOUDLY on names past the operator's naming budget: every
	// child derives from this name by suffixing (`-network-policy` is
	// the longest at 15 characters) and StatefulSet pod hostnames must
	// stay DNS-legal — past 48 the derived names silently break the
	// contract the exported outputs are built on. Twin: the Terraform
	// module's lifecycle precondition.
	if len(locals.ResourceName) > vars.NameBudget {
		return errors.Errorf(
			"resource name %q is %d characters — the Keycloak operator derives child names from it "+
				"(suffixes up to 15 characters) and pod hostnames must stay DNS-legal; "+
				"use a name of at most %d characters",
			locals.ResourceName, len(locals.ResourceName), vars.NameBudget)
	}

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	createdNamespace, err := namespace(ctx, iacInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	if err := keycloakCR(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create keycloak custom resource")
	}

	return exportOutputs(ctx, locals)
}
