package module

import (
	"github.com/pkg/errors"
	kubernetesingressnginxv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesingressnginx/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs the ingress-nginx controller from the official Helm
// chart as a real Helm release. The typed spec renders into chart values
// (values.go); the helm_values escape hatch merges last with Helm -f
// semantics — the exact semantic twin of the Terraform module's
// helm_release with values = [typed, helm_values].
//
// The release is named after metadata.name (NOT a fixed chart name):
// multiple controller instances per cluster — public + internal traffic
// splits, each owning its own IngressClass — are a first-class upstream
// pattern.
func Resources(ctx *pulumi.Context, iacInput *kubernetesingressnginxv1alpha1.KubernetesIngressNginxIacInput) error {
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

	var releaseDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		releaseDeps = append(releaseDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// ------------------------------ helm release --------------------------
	mergedValues, err := buildHelmValues(locals)
	if err != nil {
		return errors.Wrap(err, "failed to build helm values")
	}

	releaseArgs := &helmv3.ReleaseArgs{
		Name:      pulumi.String(locals.ReleaseName),
		Namespace: pulumi.String(locals.Namespace),
		Chart:     pulumi.String(vars.HelmChartName),
		Version:   pulumi.String(locals.ChartVersion),
		RepositoryOpts: &helmv3.RepositoryOptsArgs{
			Repo: pulumi.String(vars.HelmChartRepo),
		},
		Values: pulumi.ToMap(mergedValues),
		// The module owns namespace creation (create_namespace flag).
		CreateNamespace: pulumi.Bool(false),
		// Wait for the release's resources to become ready — a controller
		// that never starts (bad image, unschedulable pod, webhook certgen
		// failure) should fail THIS deploy, not the first Ingress. Helm's
		// readiness check on a LoadBalancer-type Service also waits for
		// the cloud LB address, so on clusters WITHOUT a cloud LB
		// controller (kind, bare metal) a LoadBalancer service type times
		// out loudly here — deliberate: use node_port/host access on such
		// clusters, and the failure names the real problem instead of
		// leaving a silently Pending entry point.
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(300),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)

	createdRelease, err := helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install ingress-nginx helm release")
	}

	// ---------------------- load-balancer address read ----------------------
	// For LoadBalancer-type services the release wait above guarantees the
	// address exists by the time we read it. The read is skipped entirely
	// for node_port/cluster_ip (there is no LB status to read) — those
	// deploys export empty address outputs by design. Reads run through the
	// provider's read path (no awaiters), so this never blocks.
	loadBalancerIp := pulumi.String("").ToStringOutput()
	loadBalancerHostname := pulumi.String("").ToStringOutput()
	if locals.Spec.GetService().GetType() == kubernetesingressnginxv1alpha1.KubernetesIngressNginxServiceType_load_balancer {
		controllerService, err := kubernetescorev1.GetService(ctx,
			locals.ControllerServiceName,
			pulumi.ID(locals.Namespace+"/"+locals.ControllerServiceName),
			nil,
			pulumi.Provider(kubernetesProvider),
			pulumi.DependsOn([]pulumi.Resource{createdRelease}))
		if err != nil {
			return errors.Wrap(err, "failed to read controller service for load-balancer address")
		}
		ingress := controllerService.Status.LoadBalancer().Ingress()
		loadBalancerIp = ingress.Index(pulumi.Int(0)).Ip().ApplyT(func(v *string) string {
			if v == nil {
				return ""
			}
			return *v
		}).(pulumi.StringOutput)
		loadBalancerHostname = ingress.Index(pulumi.Int(0)).Hostname().ApplyT(func(v *string) string {
			if v == nil {
				return ""
			}
			return *v
		}).(pulumi.StringOutput)
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpIngressClassName, pulumi.String(locals.IngressClassName))
	ctx.Export(OpControllerServiceName, pulumi.String(locals.ControllerServiceName))
	ctx.Export(OpInternalServiceName, pulumi.String(locals.InternalServiceName))
	ctx.Export(OpLoadBalancerIp, loadBalancerIp)
	ctx.Export(OpLoadBalancerHostname, loadBalancerHostname)

	return nil
}
