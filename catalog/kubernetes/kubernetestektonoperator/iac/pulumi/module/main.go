package module

import (
	"github.com/pkg/errors"
	kubernetestektonoperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetestektonoperator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	pulumiyaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs the Tekton Operator from its released single-file
// manifest — the operator's OFFICIAL distribution (the in-repo Helm chart
// is unpublished, version "devel"). The manifest applies per document:
// the namespace `tekton-operator`, the 14 operator.tekton.dev CRDs, the
// operator and webhook Deployments (with the spec's typed overrides
// patched on), ConfigMaps, RBAC, Services and the webhook cert Secret.
//
// AUTO-INSTALL IS ALWAYS DISABLED: the tekton-config-defaults ConfigMap's
// AUTOINSTALL_COMPONENTS key is patched from the release's "true" to
// "false" so the operator never creates its own TektonConfig — the
// KubernetesTekton declaration kind is the single owner of the cluster's
// Tekton configuration (see autoInstallTransformation). Installing the
// operator alone deploys no Tekton components.
//
// APPLY MODE: the shared provider helper enables server-side apply — the
// TektonConfig CRD document (~70 KB) plus its siblings stay comfortably
// managed, and SSA keeps re-installs tolerant of the operator's own
// field management. The Terraform twin applies with
// server_side_apply = true.
//
// ORDERING: the manifest applies as THREE dependency-chained groups —
// namespace → workloads → CRDs — because destroy runs the reverse:
// the CRDs delete FIRST, while the operator Deployments still run. CRD
// deletion drains every CR, and the operator's runtime InstallerSets
// carry a finalizer only the LIVE operator can process (its webhook
// maintains one for itself even with zero TektonConfigs) — deleting the
// CRDs and the operator in one flat pass wedges the tektoninstallersets
// CRD in Terminating until the provider's delete await times out
// (verified live on both engines). The operator tolerates starting
// before its CRDs exist: knative-style controllers crash-retry until
// their informers sync. The Terraform twin encodes the same chain with
// depends_on.
//
// IMAGE REGISTRY: when set, every image Tekton publishes moves to it at
// the same path, tag and digest — the operator's own and, through the
// operator's IMAGE_* variables, every component it installs (see
// images.go and deploymentTransformation).
//
// DESTROY SEMANTICS: every document deletes with the resource, INCLUDING
// the CRDs — which cascade-deletes any TektonConfig on the cluster.
// Always destroy the KubernetesTekton resource FIRST while the operator
// still runs (TektonInstallerSet finalizers are operator-processed; see
// the spec's destroy note).
func Resources(ctx *pulumi.Context, stackInput *kubernetestektonoperatorv1alpha1.KubernetesTektonOperatorStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	partitions, err := fetchManifestPartitions(ManifestURL())
	if err != nil {
		return errors.Wrap(err, "failed to fetch the tekton-operator release manifest")
	}

	var images *tektonImages
	if locals.Spec.GetImageRegistry() != "" {
		images, err = loadImageTable()
		if err != nil {
			return err
		}
	}

	transformations := []pulumiyaml.Transformation{
		autoInstallTransformation(),
		deploymentTransformation(locals.Spec, images),
	}

	namespaceGroup, err := pulumiyaml.NewConfigGroup(ctx, locals.ResourceName+"-namespace",
		&pulumiyaml.ConfigGroupArgs{
			YAML:            []string{partitions.NamespaceYaml},
			Transformations: transformations,
		},
		pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "failed to apply the tekton-operator namespace")
	}

	workloadsGroup, err := pulumiyaml.NewConfigGroup(ctx, locals.ResourceName,
		&pulumiyaml.ConfigGroupArgs{
			YAML: []string{partitions.WorkloadsYaml},
			// skipAwait rides ONLY this group — see the transformation's
			// rationale (nothing webhook-shaped here can become ready
			// before the CRDs group applies).
			Transformations: append([]pulumiyaml.Transformation{skipAwaitTransformation()}, transformations...),
		},
		pulumi.Provider(kubernetesProvider),
		pulumi.DependsOn([]pulumi.Resource{namespaceGroup}))
	if err != nil {
		return errors.Wrap(err, "failed to apply the tekton-operator workloads")
	}

	crdsGroup, err := pulumiyaml.NewConfigGroup(ctx, locals.ResourceName+"-crds",
		&pulumiyaml.ConfigGroupArgs{
			YAML:            []string{partitions.CrdsYaml},
			Transformations: transformations,
		},
		pulumi.Provider(kubernetesProvider),
		pulumi.DependsOn([]pulumi.Resource{workloadsGroup}))
	if err != nil {
		return errors.Wrap(err, "failed to apply the tekton-operator CRDs")
	}

	return exportOutputs(ctx, locals, crdsGroup)
}
