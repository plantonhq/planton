package module

import (
	"github.com/pkg/errors"
	kuberneteskeycloakoperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskeycloakoperator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	pulumiyaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs the official Keycloak Operator from the
// keycloak-k8s-resources release manifests — the operator's first-party
// Kubernetes distribution (Keycloak ships NO official Helm chart). The
// bundle applies per document: ServiceAccount, ClusterRoles, bindings,
// Role, the metrics/health Service and the operator Deployment (with
// the spec's typed overrides patched on), plus the four
// k8s.keycloak.org CRDs published beside it. spec.cluster_wide selects
// the watch-scope variant: kubernetes.yml (the operator watches ONLY
// its own namespace) or cluster-wide/kubernetes.yml (per-controller
// ClusterRoleBindings and JOSDK_ALL_NAMESPACES).
//
// NAMESPACE STAMPING: the bundle ships every document WITHOUT a
// namespace field (upstream expects kustomize to set it). The module
// owns that kustomize step: metadata.namespace = <spec namespace> on
// every namespaced document, and every binding's ServiceAccount subject
// namespace rewritten to match (see namespaceTransformation). Every
// resource is FIXED-NAME (`keycloak-operator` etc. — upstream's own
// names), so exactly ONE operator install fits per namespace.
//
// APPLY MODE: the shared provider helper enables server-side apply,
// which this bundle REQUIRES — the keycloaks CRD document (~9,900
// lines) blows past the client-side last-applied-configuration
// annotation cap, and SSA keeps re-installs tolerant of the operator's
// own field management. The Terraform twin applies with
// server_side_apply = true.
//
// ORDERING: the manifests apply as dependency-chained groups —
// namespace (when create_namespace) → workloads → CRDs — because
// destroy runs the reverse: the CRDs delete FIRST, while the operator
// Deployment still runs. CRD deletion cascade-deletes every Keycloak CR
// on the cluster, and any operator-processed finalizers on those CRs
// need the LIVE operator to drain — deleting the CRDs and the operator
// in one flat pass risks wedging the drain until the provider's delete
// await times out (the tektonoperator exemplar caught this class live).
// The operator tolerates starting before its CRDs exist: the JOSDK
// operator crash-loops until they appear. The Terraform twin encodes
// the same chain with depends_on.
//
// DESTROY SEMANTICS: every document deletes with the resource,
// INCLUDING the CRDs — which cascade-deletes every Keycloak /
// KeycloakRealmImport / KeycloakOidcClient / KeycloakSamlClient CR on
// the cluster. Always destroy KubernetesKeycloak resources FIRST while
// the operator still runs.
func Resources(ctx *pulumi.Context, iacInput *kuberneteskeycloakoperatorv1alpha1.KubernetesKeycloakOperatorIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	partitions, err := fetchManifestPartitions(locals.Spec.GetClusterWide())
	if err != nil {
		return errors.Wrap(err, "failed to fetch the keycloak-operator release manifests")
	}

	transformations := []pulumiyaml.Transformation{
		namespaceTransformation(locals.Namespace),
		deploymentTransformation(locals.Spec),
	}

	// The module-authored Namespace document (the bundle ships none) —
	// created first, deleted last. Only when create_namespace: false
	// means the namespace must already exist and the module never
	// touches it.
	var workloadDependencies []pulumi.Resource
	if locals.Spec.GetCreateNamespace() {
		namespaceYaml, err := namespaceDocumentYaml(locals)
		if err != nil {
			return errors.Wrap(err, "failed to render the keycloak-operator namespace document")
		}
		namespaceGroup, err := pulumiyaml.NewConfigGroup(ctx, locals.ResourceName+"-namespace",
			&pulumiyaml.ConfigGroupArgs{
				YAML:            []string{namespaceYaml},
				Transformations: transformations,
			},
			pulumi.Provider(kubernetesProvider))
		if err != nil {
			return errors.Wrap(err, "failed to apply the keycloak-operator namespace")
		}
		workloadDependencies = append(workloadDependencies, namespaceGroup)
	}

	workloadsGroup, err := pulumiyaml.NewConfigGroup(ctx, locals.ResourceName,
		&pulumiyaml.ConfigGroupArgs{
			YAML: []string{partitions.WorkloadsYaml},
			// skipAwait rides ONLY this group — see the transformation's
			// rationale (the operator Deployment cannot become ready
			// before the CRDs group applies).
			Transformations: append([]pulumiyaml.Transformation{skipAwaitTransformation()}, transformations...),
		},
		pulumi.Provider(kubernetesProvider),
		pulumi.DependsOn(workloadDependencies))
	if err != nil {
		return errors.Wrap(err, "failed to apply the keycloak-operator workloads")
	}

	crdsGroup, err := pulumiyaml.NewConfigGroup(ctx, locals.ResourceName+"-crds",
		&pulumiyaml.ConfigGroupArgs{
			YAML:            []string{partitions.CrdsYaml},
			Transformations: transformations,
		},
		pulumi.Provider(kubernetesProvider),
		pulumi.DependsOn([]pulumi.Resource{workloadsGroup}))
	if err != nil {
		return errors.Wrap(err, "failed to apply the keycloak-operator CRDs")
	}

	return exportOutputs(ctx, locals, crdsGroup)
}
