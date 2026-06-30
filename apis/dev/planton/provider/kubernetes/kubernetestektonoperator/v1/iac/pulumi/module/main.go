package module

import (
	"github.com/pkg/errors"
	kubernetestektonoperatorv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetestektonoperator/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi entry-point.
//
// The deployment order is:
// 1. Tekton Operator manifests (creates namespace and CRDs)
// 2. TektonConfig (configures which components to install, CloudEvents sink)
// 3. Dashboard ingress resources (if enabled)
//
// Note: Tekton Operator manages its own namespaces:
// - 'tekton-operator' for the operator itself
// - 'tekton-pipelines' for Tekton components (Pipelines, Triggers, Dashboard)
// These namespaces are automatically created by the Tekton Operator and cannot be customized.
func Resources(ctx *pulumi.Context,
	in *kubernetestektonoperatorv1.KubernetesTektonOperatorStackInput) error {

	locals := initializeLocals(ctx, in)

	k8sProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, in.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "setup kubernetes provider")
	}

	// 1. Deploy Tekton Operator and TektonConfig
	if err = tektonOperator(ctx, locals, k8sProvider); err != nil {
		return errors.Wrap(err, "deploy tekton operator")
	}

	// 2. Create Dashboard Ingress (if enabled and dashboard component is enabled)
	if locals.IngressEnabled && locals.EnableDashboard {
		if err = dashboardIngress(ctx, locals, k8sProvider, nil); err != nil {
			return errors.Wrap(err, "create dashboard ingress")
		}
	}

	return nil
}
