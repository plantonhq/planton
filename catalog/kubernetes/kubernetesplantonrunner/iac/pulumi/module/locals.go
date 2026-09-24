package module

import (
	"strconv"

	kubernetesplantonrunnerv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesplantonrunner/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the stack input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesplantonrunnerv1alpha1.KubernetesPlantonRunnerSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace and the token Secret — never injected into the
	// chart's own resources; Helm owns those).
	Labels map[string]string

	// Namespace the runner installs into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// Helm release name — metadata.name. Many runners can coexist (each
	// is its own enrollment); each is its own release.
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// Chart repository resolved to the pinned default when unset, for the
	// same reason.
	ChartRepository string

	// RunnerName is the name the runner registers itself under when it
	// joins the control plane. spec.runner_name, falling back to
	// "<env>-<metadata.name>" (metadata.name outside an environment) —
	// the SAME derivation the platform uses for records that reference
	// this runner (its minted token, its managed destroy); changing this
	// formula breaks arrival attribution and managed teardown.
	RunnerName string

	// TokenSecretName is the module-created Secret the chart reads the
	// runner token from — the token never rides rendered chart values.
	TokenSecretName string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, stackInput *kubernetesplantonrunnerv1alpha1.KubernetesPlantonRunnerStackInput) *Locals {
	target := stackInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesPlantonRunner.String(),
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

	chartRepository := spec.GetChartRepository()
	if chartRepository == "" {
		chartRepository = vars.DefaultChartRepository
	}

	runnerName := spec.GetRunnerName()
	if runnerName == "" {
		runnerName = target.Metadata.Name
		if target.Metadata.Env != "" {
			runnerName = target.Metadata.Env + "-" + target.Metadata.Name
		}
	}

	return &Locals{
		Spec:            spec,
		Labels:          labels,
		Namespace:       spec.Namespace.GetValue(),
		ReleaseName:     target.Metadata.Name,
		ChartVersion:    chartVersion,
		ChartRepository: chartRepository,
		RunnerName:      runnerName,
		TokenSecretName: target.Metadata.Name + vars.TokenSecretSuffix,
	}
}
