package module

import (
	"github.com/pkg/errors"
	kubernetesgofeatureflagv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgofeatureflag/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs the GO Feature Flag relay proxy from the official chart
// as a real Helm release, with the objects the chart does not provide:
//  1. the namespace (create_namespace),
//  2. `<name>-env`, the Secret carrying every secret value as an
//     environment variable (env_secret.go),
//  3. `<name>-flag-reader`, a Role and RoleBinding per namespace granting
//     `get` on exactly the ConfigMaps the retrievers read (rbac.go),
//  4. the Helm release (helm_release.go), rendered from the typed spec with
//     the relay configuration built in relay_config.go,
//  5. the optional ServiceMonitor (service_monitor.go).
//
// The exact same resource set renders from the Terraform module - keep them
// in lockstep.
func Resources(ctx *pulumi.Context, iacInput *kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagIacInput) error {
	// NAME BUDGET, checked before anything is created (Terraform twin: the
	// lifecycle precondition on helm_release.relay_proxy).
	if len(iacInput.Target.Metadata.Name) > vars.MaxMetadataNameLength {
		return errors.Errorf(
			"metadata.name %q is %d characters; the relay's name budget allows at most %d "+
				"(the relay Service is named after the resource, and a Service name is a 63-character DNS label)",
			iacInput.Target.Metadata.Name, len(iacInput.Target.Metadata.Name), vars.MaxMetadataNameLength)
	}

	locals, err := initializeLocals(ctx, iacInput)
	if err != nil {
		return err
	}

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
	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// ------------------------------ env secret ---------------------------
	createdSecret, err := envSecret(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create env secret")
	}

	// ------------------------------ flag-reader rbac ---------------------
	rbac, err := flagReaderRbac(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return err
	}

	var releaseDeps []pulumi.Resource
	if createdNamespace != nil {
		releaseDeps = append(releaseDeps, createdNamespace)
	}
	if createdSecret != nil {
		releaseDeps = append(releaseDeps, createdSecret)
	}
	releaseDeps = append(releaseDeps, rbac...)

	// ------------------------------ helm release -------------------------
	release, err := helmRelease(ctx, locals, kubernetesProvider, []pulumi.ResourceOption{pulumi.DependsOn(releaseDeps)})
	if err != nil {
		return err
	}

	// ------------------------------ service monitor ----------------------
	if locals.Spec.GetMetrics().GetServiceMonitorEnabled() {
		if err := serviceMonitor(ctx, locals, kubernetesProvider,
			[]pulumi.ResourceOption{pulumi.DependsOn([]pulumi.Resource{release})}); err != nil {
			return err
		}
	}

	exportOutputs(ctx, locals)
	return nil
}
