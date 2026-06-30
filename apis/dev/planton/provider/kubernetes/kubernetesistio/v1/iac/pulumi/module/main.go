package module

import (
	"github.com/pkg/errors"
	kubernetesistiov1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesistio/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources wires up Istio (base, control‑plane & ingress gateway) on the target
// Kubernetes cluster.  It honours spec.container.resources for the Istiod
// deployment, allowing callers to tune CPU / memory without touching Helm YAML.
func Resources(ctx *pulumi.Context, in *kubernetesistiov1.KubernetesIstioStackInput) error {
	// Initialize locals with computed values
	locals := initializeLocals(ctx, in)

	// Kubernetes provider from cluster‑credential
	kubernetesProviderConfig, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, in.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	spec := in.Target.Spec

	// ---- pick chart version ----
	// Driven by spec.version (defaulted to vars.DefaultStableVersion at manifest-load
	// time via protodefaults); fall back to the module default if somehow unset.
	chartVersion := spec.GetVersion()
	if chartVersion == "" {
		chartVersion = vars.DefaultStableVersion
	}

	// ---- conditionally create namespaces ----
	sysNSName, gwNSName, namespaceDeps, err := namespaces(ctx, in, locals, kubernetesProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create namespaces")
	}

	// convenience for repeated repo opts
	repo := helm.RepositoryOptsArgs{Repo: pulumi.String(vars.HelmRepo)}

	// ---- istio/base ----
	// Helm release name uses {metadata.name}-base to avoid conflicts when multiple instances share a namespace
	baseOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProviderConfig),
	}, namespaceDeps...)

	_, err = helm.NewRelease(ctx, locals.BaseReleaseName,
		&helm.ReleaseArgs{
			Name:            pulumi.String(locals.BaseReleaseName),
			Namespace:       sysNSName,
			Chart:           pulumi.String(vars.BaseChart),
			Version:         pulumi.String(chartVersion),
			CreateNamespace: pulumi.Bool(false),
			Atomic:          pulumi.Bool(true),
			CleanupOnFail:   pulumi.Bool(true),
			WaitForJobs:     pulumi.Bool(true),
			Timeout:         pulumi.Int(180),
			RepositoryOpts:  repo,
		},
		baseOpts...)
	if err != nil {
		return errors.Wrap(err, "installing istio/base")
	}

	// ---- istiod (control‑plane) ----
	istiodValues := pulumi.Map{}
	if res := spec.GetContainer(); res != nil && res.Resources != nil {
		// map protobuf fields -> Helm values: pilot.resources
		limits := res.Resources.GetLimits()
		requests := res.Resources.GetRequests()
		istiodValues["pilot"] = pulumi.Map{
			"resources": pulumi.Map{
				"limits": pulumi.Map{
					"cpu":    pulumi.String(limits.GetCpu()),
					"memory": pulumi.String(limits.GetMemory()),
				},
				"requests": pulumi.Map{
					"cpu":    pulumi.String(requests.GetCpu()),
					"memory": pulumi.String(requests.GetMemory()),
				},
			},
		}
	}

	// Helm release name uses {metadata.name}-istiod to avoid conflicts when multiple instances share a namespace
	istiodOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProviderConfig),
	}, namespaceDeps...)

	_, err = helm.NewRelease(ctx, locals.IstiodReleaseName,
		&helm.ReleaseArgs{
			Name:            pulumi.String(locals.IstiodReleaseName),
			Namespace:       sysNSName,
			Chart:           pulumi.String(vars.IstiodChart),
			Version:         pulumi.String(chartVersion),
			CreateNamespace: pulumi.Bool(false),
			Atomic:          pulumi.Bool(true),
			CleanupOnFail:   pulumi.Bool(true),
			WaitForJobs:     pulumi.Bool(true),
			Timeout:         pulumi.Int(180),
			Values:          istiodValues,
			RepositoryOpts:  repo,
		},
		istiodOpts...)
	if err != nil {
		return errors.Wrap(err, "installing istiod control‑plane")
	}

	// ---- ingress‑gateway ----
	// Helm release name uses {metadata.name}-gateway to avoid conflicts when multiple instances share a namespace
	gwOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProviderConfig),
	}, namespaceDeps...)

	_, err = helm.NewRelease(ctx, locals.GatewayReleaseName,
		&helm.ReleaseArgs{
			Name:            pulumi.String(locals.GatewayReleaseName),
			Namespace:       gwNSName,
			Chart:           pulumi.String(vars.GatewayChart),
			Version:         pulumi.String(chartVersion),
			CreateNamespace: pulumi.Bool(false),
			Atomic:          pulumi.Bool(true),
			CleanupOnFail:   pulumi.Bool(true),
			WaitForJobs:     pulumi.Bool(true),
			Timeout:         pulumi.Int(180),
			RepositoryOpts:  repo,
			Values: pulumi.Map{
				"service": pulumi.Map{
					"type": pulumi.String("LoadBalancer"),
				},
			},
		},
		gwOpts...)
	if err != nil {
		return errors.Wrap(err, "installing istio ingress‑gateway")
	}

	// ---- stack outputs ----
	ctx.Export(OpNamespace, sysNSName)
	ctx.Export(OpService, pulumi.String(locals.IstiodReleaseName))
	ctx.Export(OpPortForwardCommand, pulumi.Sprintf("kubectl port-forward -n %s svc/%s 15014:15014", sysNSName, locals.IstiodReleaseName))
	ctx.Export(OpKubeEndpoint, pulumi.Sprintf("%s.%s.svc.cluster.local:15012", locals.IstiodReleaseName, sysNSName))
	ctx.Export(OpIngressEndpoint, pulumi.Sprintf("%s.%s.svc.cluster.local:80", locals.GatewayReleaseName, gwNSName))

	return nil
}
