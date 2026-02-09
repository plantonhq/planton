package module

import (
	"github.com/pkg/errors"
	kubernetessolroperatorv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetessolroperator/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	pulumiyaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates all Pulumi resources for the Apache Solr Operator Kubernetes add‑on.
func Resources(ctx *pulumi.Context, stackInput *kubernetessolroperatorv1.KubernetesSolrOperatorStackInput) error {
	// Initialize locals with computed values
	locals := initializeLocals(ctx, stackInput)

	// Set up kubernetes provider from the supplied cluster credential
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
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

	// --------------------------------------------------------------------
	// 2. Apply CRDs required by the operator
	// Uses computed CrdsResourceName to avoid conflicts when multiple
	// instances share a namespace.
	// --------------------------------------------------------------------
	//
	// Note on CRD Deletion:
	// CRDs are cluster-scoped and have built-in protection preventing deletion while
	// CustomResources of that type exist. During `pulumi destroy`, CRDs will wait
	// until all CRs are removed. The namespace background deletion policy (above)
	// ensures the operator stops running quickly, which allows CRs to be garbage
	// collected, unblocking CRD deletion.
	//
	// We intentionally avoid using ConfigFile transformations here because they
	// cause Pulumi to recompute diffs on every operation, leading to massive
	// (180MB+) diff sizes due to the embedded OpenAPI schemas in the CRDs.
	crdsOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	crds, err := pulumiyaml.NewConfigFile(ctx, locals.CrdsResourceName,
		&pulumiyaml.ConfigFileArgs{
			File: locals.CrdManifestURL,
		},
		crdsOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to apply CRDs")
	}

	// --------------------------------------------------------------------
	// 3. Deploy the operator via Helm
	// Uses computed HelmReleaseName to avoid conflicts when multiple
	// instances share a namespace.
	// --------------------------------------------------------------------
	helmOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
		pulumi.DependsOn([]pulumi.Resource{crds}),
		pulumi.IgnoreChanges([]string{"status", "description", "resourceNames"}),
	}, namespaceDeps...)
	_, err = helm.NewRelease(ctx, locals.HelmReleaseName,
		&helm.ReleaseArgs{
			Name:            pulumi.String(locals.HelmReleaseName),
			Namespace:       pulumi.String(locals.Namespace),
			Chart:           pulumi.String(vars.HelmChartName),
			Version:         pulumi.String(locals.ChartVersion),
			CreateNamespace: pulumi.Bool(false),
			Atomic:          pulumi.Bool(true),
			CleanupOnFail:   pulumi.Bool(true),
			WaitForJobs:     pulumi.Bool(true),
			Timeout:         pulumi.Int(180),
			Values:          pulumi.Map{}, // no extra values at this time
			RepositoryOpts: helm.RepositoryOptsArgs{
				Repo: pulumi.String(vars.HelmChartRepo),
			},
		},
		helmOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to install solr‑operator helm release")
	}

	return nil
}
