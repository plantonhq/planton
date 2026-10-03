package module

import (
	"strconv"

	kubernetesgharunnerscalesetv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgharunnerscaleset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesgharunnerscalesetv1alpha1.KubernetesGhaRunnerScaleSetSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace, the materialized credential Secret — never
	// injected into the chart's own resources; Helm owns those).
	Labels map[string]string

	// Namespace the scale set installs into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// Helm release name — metadata.name. Many scale sets can coexist
	// (one per repo/org registration); each is its own release.
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// RunnerScaleSetName is the GitHub-visible fleet name — the exact
	// `runs-on:` value. spec.runner_scale_set_name, falling back to
	// metadata.name. Capped at 45 characters (fail-loud in main.go —
	// the chart's own template fails past it).
	RunnerScaleSetName string

	// GithubAuthSecretName is the Secret the chart reads the GitHub
	// credential from: the user's own Secret (existing-Secret arm) or
	// the module-materialized `<name>-github-auth` (declared arms).
	GithubAuthSecretName string

	// MaterializeAuthSecret is true on the declared PAT / GitHub App
	// arms — the module creates the Secret so credential material never
	// rides rendered chart values.
	MaterializeAuthSecret bool
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesgharunnerscalesetv1alpha1.KubernetesGhaRunnerScaleSetIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesGhaRunnerScaleSet.String(),
	}
	if target.Metadata.Id != "" {
		labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	chartVersion := spec.GetChartVersion()
	if chartVersion == "" {
		chartVersion = vars.DefaultChartVersion
	}

	runnerScaleSetName := spec.GetRunnerScaleSetName()
	if runnerScaleSetName == "" {
		runnerScaleSetName = target.Metadata.Name
	}

	githubAuthSecretName := spec.GetAuth().GetExistingSecretName()
	materialize := false
	if githubAuthSecretName == "" {
		githubAuthSecretName = target.Metadata.Name + vars.GithubAuthSecretSuffix
		materialize = true
	}

	return &Locals{
		Spec:                  spec,
		Labels:                labels,
		Namespace:             spec.Namespace.GetValue(),
		ReleaseName:           target.Metadata.Name,
		ChartVersion:          chartVersion,
		RunnerScaleSetName:    runnerScaleSetName,
		GithubAuthSecretName:  githubAuthSecretName,
		MaterializeAuthSecret: materialize,
	}
}
