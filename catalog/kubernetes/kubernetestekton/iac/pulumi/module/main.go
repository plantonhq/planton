package module

import (
	"github.com/pkg/errors"
	kubernetestektonv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetestekton/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources renders the cluster's TektonConfig — the single declaration
// the Tekton Operator (a KubernetesTektonOperator install, the registry
// prerequisite) reconciles into running components via
// TektonInstallerSet resources.
//
// THE CR NAME IS FIXED: the operator's admission webhook allows exactly
// one TektonConfig per cluster and requires the name `config` —
// metadata.name of the Planton resource keys the state identity only.
//
// UNTYPED CR: the TektonConfig CRD types its component `options` blocks
// with x-kubernetes-preserve-unknown-fields, which crd2pulumi cannot
// carry faithfully — the CR renders through apiextensions.CustomResource
// with an untyped spec body built in tekton_config.go, in byte lockstep
// with the Terraform twin's locals.
//
// DESTROY SEMANTICS (the reason this kind exists separately from the
// operator): deleting the TektonConfig triggers the operator to tear
// down every component it installed — the TektonInstallerSet finalizers
// are processed by the RUNNING operator, and this deletion blocks until
// that teardown completes (a 15-minute delete timeout covers the
// full-profile teardown). Destroying this resource BEFORE the operator
// is exactly what makes the teardown clean; the operator kind's docs
// carry the ordering contract.
func Resources(ctx *pulumi.Context, iacInput *kubernetestektonv1alpha1.KubernetesTektonIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	_, err = apiextensions.NewCustomResource(ctx, locals.ResourceName,
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String(vars.ApiVersion),
			Kind:       pulumi.String(vars.Kind),
			Metadata: &kubernetesmeta.ObjectMetaArgs{
				// Cluster-scoped, operator-required fixed name.
				Name:   pulumi.String(vars.TektonConfigName),
				Labels: pulumi.ToStringMap(locals.Labels),
			},
			OtherFields: map[string]interface{}{
				"spec": tektonConfigSpecBody(locals),
			},
		},
		pulumi.Provider(kubernetesProvider),
		// The operator-processed finalizer holds deletion until the
		// component teardown finishes; the full-profile teardown needs
		// minutes, not the default wait.
		pulumi.Timeouts(&pulumi.CustomTimeouts{Delete: "15m"}))
	if err != nil {
		return errors.Wrap(err, "failed to render the TektonConfig custom resource")
	}

	return exportOutputs(ctx, locals)
}
