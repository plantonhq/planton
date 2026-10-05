package module

import (
	"github.com/pkg/errors"
	kubernetesflagdv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesflagd/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources runs flagd as module-owned manifests (flagd publishes no Helm
// chart):
//  1. the namespace (create_namespace),
//  2. `<name>-sources`, the Secret holding flagd's SourceConfig array,
//  3. the ServiceAccount (unless an existing one is named) and, for
//     FeatureFlag sources, `<name>-flag-reader` Roles and RoleBindings,
//  4. the Deployment and the Service exposing the evaluation, management,
//     sync and OFREP ports,
//  5. the optional HorizontalPodAutoscaler, PodDisruptionBudget and
//     ServiceMonitor.
//
// The exact same resource set renders from the Terraform module - keep them
// in lockstep.
func Resources(ctx *pulumi.Context, iacInput *kubernetesflagdv1alpha1.KubernetesFlagdIacInput) error {
	if len(iacInput.Target.Metadata.Name) > vars.MaxMetadataNameLength {
		return errors.Errorf(
			"metadata.name %q is %d characters; flagd's name budget allows at most %d "+
				"(the Service is named after the resource, and a Service name is a 63-character DNS label)",
			iacInput.Target.Metadata.Name, len(iacInput.Target.Metadata.Name), vars.MaxMetadataNameLength)
	}

	locals, err := initializeLocals(iacInput)
	if err != nil {
		return err
	}

	provider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	createdNamespace, err := namespace(ctx, iacInput, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}
	var nsDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		nsDeps = append(nsDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	secret, err := sourcesSecret(ctx, locals, provider, nsDeps)
	if err != nil {
		return err
	}
	sa, err := serviceAccount(ctx, locals, provider, nsDeps)
	if err != nil {
		return err
	}
	rbac, err := featureFlagReaderRbac(ctx, locals, provider, nsDeps)
	if err != nil {
		return err
	}

	workloadDeps := []pulumi.Resource{secret}
	if createdNamespace != nil {
		workloadDeps = append(workloadDeps, createdNamespace)
	}
	if sa != nil {
		workloadDeps = append(workloadDeps, sa)
	}
	workloadDeps = append(workloadDeps, rbac...)

	created, err := deployment(ctx, locals, provider, []pulumi.ResourceOption{pulumi.DependsOn(workloadDeps)})
	if err != nil {
		return err
	}
	if err := service(ctx, locals, provider, nsDeps); err != nil {
		return err
	}
	afterDeployment := []pulumi.ResourceOption{pulumi.DependsOn([]pulumi.Resource{created})}
	if err := hpa(ctx, locals, provider, afterDeployment); err != nil {
		return err
	}
	if err := pdb(ctx, locals, provider, afterDeployment); err != nil {
		return err
	}
	if err := serviceMonitor(ctx, locals, provider, afterDeployment); err != nil {
		return err
	}

	exportOutputs(ctx, locals)
	return nil
}
