package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ghaRunnerScaleSetController deploys the GitHub Actions Runner Scale Set Controller
// using the official Helm chart.
func ghaRunnerScaleSetController(ctx *pulumi.Context, locals *Locals, k8sProvider *kubernetes.Provider, namespaceDeps []pulumi.ResourceOption) error {
	// Build Helm values
	helmValues := buildHelmValues(locals)

	// Deploy Helm chart
	// For OCI charts, the full URL must be passed as the Chart parameter
	// (RepositoryOpts.Repo doesn't work with OCI registries in Pulumi)
	helmOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(k8sProvider),
	}, namespaceDeps...)

	_, err := helmv3.NewRelease(ctx, locals.ReleaseName, &helmv3.ReleaseArgs{
		Name:            pulumi.String(locals.ReleaseName),
		Namespace:       pulumi.String(locals.Namespace),
		CreateNamespace: pulumi.Bool(false), // We handle namespace creation ourselves
		Chart:           pulumi.String(vars.HelmChartOCI),
		Version:         pulumi.String(locals.ChartVersion),
		Values:          helmValues,
	}, helmOpts...)
	if err != nil {
		return errors.Wrap(err, "deploy helm release")
	}

	return nil
}

// buildHelmValues constructs the Helm values map from the Locals struct.
func buildHelmValues(locals *Locals) pulumi.Map {
	values := pulumi.Map{
		"replicaCount": pulumi.Int(locals.ReplicaCount),
		"labels":       pulumi.ToStringMap(locals.KubeLabels),
	}

	// Container resources
	resources := pulumi.Map{}
	if locals.CpuRequests != "" || locals.MemoryRequests != "" {
		requests := pulumi.Map{}
		if locals.CpuRequests != "" {
			requests["cpu"] = pulumi.String(locals.CpuRequests)
		}
		if locals.MemoryRequests != "" {
			requests["memory"] = pulumi.String(locals.MemoryRequests)
		}
		resources["requests"] = requests
	}
	if locals.CpuLimits != "" || locals.MemoryLimits != "" {
		limits := pulumi.Map{}
		if locals.CpuLimits != "" {
			limits["cpu"] = pulumi.String(locals.CpuLimits)
		}
		if locals.MemoryLimits != "" {
			limits["memory"] = pulumi.String(locals.MemoryLimits)
		}
		resources["limits"] = limits
	}
	if len(resources) > 0 {
		values["resources"] = resources
	}

	// Custom image
	if locals.ImageRepository != "" || locals.ImageTag != "" || locals.ImagePullPolicy != "" {
		image := pulumi.Map{}
		if locals.ImageRepository != "" {
			image["repository"] = pulumi.String(locals.ImageRepository)
		}
		if locals.ImageTag != "" {
			image["tag"] = pulumi.String(locals.ImageTag)
		}
		if locals.ImagePullPolicy != "" {
			image["pullPolicy"] = pulumi.String(locals.ImagePullPolicy)
		}
		values["image"] = image
	}

	// Flags
	flags := pulumi.Map{}
	if locals.LogLevel != "" {
		flags["logLevel"] = pulumi.String(locals.LogLevel)
	}
	if locals.LogFormat != "" {
		flags["logFormat"] = pulumi.String(locals.LogFormat)
	}
	if locals.WatchSingleNamespace != "" {
		flags["watchSingleNamespace"] = pulumi.String(locals.WatchSingleNamespace)
	}
	if locals.RunnerMaxConcurrentReconciles > 0 {
		flags["runnerMaxConcurrentReconciles"] = pulumi.Int(locals.RunnerMaxConcurrentReconciles)
	}
	if locals.UpdateStrategy != "" {
		flags["updateStrategy"] = pulumi.String(locals.UpdateStrategy)
	}
	if len(locals.ExcludeLabelPropagationPrefixes) > 0 {
		flags["excludeLabelPropagationPrefixes"] = pulumi.ToStringArray(locals.ExcludeLabelPropagationPrefixes)
	}
	if locals.K8sClientRateLimiterQPS > 0 {
		flags["k8sClientRateLimiterQPS"] = pulumi.Int(locals.K8sClientRateLimiterQPS)
	}
	if locals.K8sClientRateLimiterBurst > 0 {
		flags["k8sClientRateLimiterBurst"] = pulumi.Int(locals.K8sClientRateLimiterBurst)
	}
	if len(flags) > 0 {
		values["flags"] = flags
	}

	// Metrics
	if locals.MetricsEnabled {
		values["metrics"] = pulumi.Map{
			"controllerManagerAddr": pulumi.String(locals.ControllerManagerAddr),
			"listenerAddr":          pulumi.String(locals.ListenerAddr),
			"listenerEndpoint":      pulumi.String(locals.ListenerEndpoint),
		}
	}

	// Image pull secrets
	if len(locals.ImagePullSecrets) > 0 {
		secrets := pulumi.Array{}
		for _, secret := range locals.ImagePullSecrets {
			secrets = append(secrets, pulumi.Map{"name": pulumi.String(secret)})
		}
		values["imagePullSecrets"] = secrets
	}

	// Priority class
	if locals.PriorityClassName != "" {
		values["priorityClassName"] = pulumi.String(locals.PriorityClassName)
	}

	return values
}
