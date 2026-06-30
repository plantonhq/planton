package module

import (
	"github.com/pkg/errors"
	kubernetestektonv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetestekton/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys Tekton using official release manifests.
//
// The deployment order is:
// 1. Tekton Pipelines manifests (creates namespace and CRDs)
// 2. Tekton Dashboard manifests (if enabled)
// 3. Cloud events ConfigMap patch (if configured)
// 4. Dashboard ingress resources (if enabled)
func Resources(ctx *pulumi.Context,
	stackInput *kubernetestektonv1.KubernetesTektonStackInput) error {

	locals := initializeLocals(ctx, stackInput)

	// Create Kubernetes provider from the credential in the stack-input
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// 1. Deploy Tekton Pipelines
	pipelineResources, err := deployTektonPipelines(ctx, locals, kubernetesProvider)
	if err != nil {
		return err
	}

	// 2. Deploy Tekton Dashboard (if enabled)
	var dashboardResources pulumi.Resource
	if locals.DashboardEnabled {
		dashboardResources, err = deployTektonDashboard(ctx, locals, kubernetesProvider, pipelineResources)
		if err != nil {
			return err
		}
	}

	// 3. Configure Cloud Events (if specified)
	if locals.CloudEventsSinkURL != "" {
		dependencies := []pulumi.Resource{pipelineResources}
		if dashboardResources != nil {
			dependencies = append(dependencies, dashboardResources)
		}
		if err := configureCloudEvents(ctx, locals, kubernetesProvider, dependencies); err != nil {
			return err
		}
	}

	// 4. Create Dashboard Ingress (if enabled)
	if locals.IngressEnabled && locals.DashboardEnabled {
		dependencies := []pulumi.Resource{pipelineResources}
		if dashboardResources != nil {
			dependencies = append(dependencies, dashboardResources)
		}
		if err := dashboardIngress(ctx, locals, kubernetesProvider, dependencies); err != nil {
			return err
		}
	}

	return nil
}
