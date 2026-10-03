package module

import (
	"github.com/pkg/errors"
	kubernetesmlflowv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesmlflow/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys one MLflow tracking server + model registry — a
// MODULE-OWNED-MANIFESTS kind (MLflow publishes no Helm chart; the
// official ghcr.io/mlflow/mlflow image is the distribution), so the
// module renders core Kubernetes objects directly:
//
//  1. the namespace (optional, create_namespace),
//  2. the module-owned Secrets (backend-store URI composed at apply time
//     from the referenced database credential; the bootstrap admin
//     password; the basic-auth ini — see secrets.go),
//  3. the PVCs the sqlite/local-artifact arms need,
//  4. the server Deployment `<metadata.name>` and its Service,
//  5. the optional `mlflow gc` CronJob and ServiceMonitor.
//
// SECURED BY DEFAULT: basic authentication is ON unless the spec disables
// it — upstream's server is open by default and its auth example ships
// admin/password1234; neither ever ships from here. The exact same
// resource set renders from the Terraform module — keep them in lockstep.
func Resources(ctx *pulumi.Context, iacInput *kubernetesmlflowv1alpha1.KubernetesMlflowIacInput) error {
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

	// ------------------------- module-owned secrets ----------------------
	createdSecrets, err := mlflowSecrets(ctx, locals, kubernetesProvider, dependencies)
	if err != nil {
		return errors.Wrap(err, "failed to create module-owned secrets")
	}

	// -------------------------------- pvcs --------------------------------
	createdPvcs, err := pvcs(ctx, locals, kubernetesProvider, dependencies)
	if err != nil {
		return errors.Wrap(err, "failed to create pvcs")
	}

	deploymentDependencies := dependencies
	allDeps := make([]pulumi.Resource, 0, len(createdSecrets)+len(createdPvcs))
	allDeps = append(allDeps, createdSecrets...)
	allDeps = append(allDeps, createdPvcs...)
	if len(allDeps) > 0 {
		deploymentDependencies = append(deploymentDependencies, pulumi.DependsOn(allDeps))
	}

	// --------------------------- deployment + service ---------------------
	createdDeployment, err := serverDeployment(ctx, locals, kubernetesProvider, deploymentDependencies)
	if err != nil {
		return err
	}
	if _, err := service(ctx, locals, kubernetesProvider, dependencies); err != nil {
		return err
	}

	// ------------------------------ satellites ----------------------------
	if locals.GcEnabled {
		if err := gcCronJob(ctx, locals, kubernetesProvider, deploymentDependencies); err != nil {
			return err
		}
	}
	if locals.ServiceMonitorEnabled {
		serviceMonitorDeps := append([]pulumi.ResourceOption{}, dependencies...)
		serviceMonitorDeps = append(serviceMonitorDeps, pulumi.DependsOn([]pulumi.Resource{createdDeployment}))
		if err := serviceMonitor(ctx, locals, kubernetesProvider, serviceMonitorDeps); err != nil {
			return err
		}
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles. Credential handles are
// Secret NAMES — values stay in-cluster; empties are honest (no admin
// credential exists with auth disabled, no URI Secret exists on the
// sqlite arm).
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	adminSecretName, adminSecretKey := "", ""
	if locals.AuthEnabled {
		adminSecretName, adminSecretKey = locals.AdminSecretName, locals.AdminSecretKey
	}
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpService, pulumi.String(locals.Name))
	ctx.Export(OpTrackingEndpoint, pulumi.String(locals.TrackingEndpoint))
	ctx.Export(OpAdminPasswordSecretName, pulumi.String(adminSecretName))
	ctx.Export(OpAdminPasswordSecretKey, pulumi.String(adminSecretKey))
	ctx.Export(OpBackendStoreUriSecretName, pulumi.String(locals.BackendUriSecretName))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
}
