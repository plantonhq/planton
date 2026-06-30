package module

import (
	"fmt"
	"strconv"

	kubernetesgharunnerscalesetcontrollerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesgharunnerscalesetcontroller/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesGhaRunnerScaleSetController *kubernetesgharunnerscalesetcontrollerv1.KubernetesGhaRunnerScaleSetController
	KubeLabels                            map[string]string
	Namespace                             string
	CreateNamespace                       bool
	ReleaseName                           string
	ChartVersion                          string
	ReplicaCount                          int

	// Container configuration
	CpuRequests    string
	CpuLimits      string
	MemoryRequests string
	MemoryLimits   string

	// Custom image
	ImageRepository string
	ImageTag        string
	ImagePullPolicy string

	// Flags
	LogLevel                        string
	LogFormat                       string
	WatchSingleNamespace            string
	RunnerMaxConcurrentReconciles   int
	UpdateStrategy                  string
	ExcludeLabelPropagationPrefixes []string
	K8sClientRateLimiterQPS         int
	K8sClientRateLimiterBurst       int

	// Metrics
	MetricsEnabled        bool
	ControllerManagerAddr string
	ListenerAddr          string
	ListenerEndpoint      string

	// Other
	ImagePullSecrets  []string
	PriorityClassName string
}

func initializeLocals(ctx *pulumi.Context, in *kubernetesgharunnerscalesetcontrollerv1.KubernetesGhaRunnerScaleSetControllerStackInput) *Locals {
	var l Locals
	l.KubernetesGhaRunnerScaleSetController = in.Target

	l.KubeLabels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: in.Target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesGhaRunnerScaleSetController.String(),
	}

	if id := in.Target.Metadata.Id; id != "" {
		l.KubeLabels[kuberneteslabelkeys.ResourceId] = id
	}
	if org := in.Target.Metadata.Org; org != "" {
		l.KubeLabels[kuberneteslabelkeys.Organization] = org
	}
	if env := in.Target.Metadata.Env; env != "" {
		l.KubeLabels[kuberneteslabelkeys.Environment] = env
	}

	// Namespace configuration
	if ns := in.Target.Spec.Namespace; ns != nil && ns.GetValue() != "" {
		l.Namespace = ns.GetValue()
	} else {
		l.Namespace = "arc-system" // Default namespace
	}
	l.CreateNamespace = in.Target.Spec.CreateNamespace

	// Release name - use resource name for consistency with other components
	l.ReleaseName = in.Target.Metadata.Name

	// Chart version
	if v := in.Target.Spec.HelmChartVersion; v != nil {
		l.ChartVersion = *v
	} else {
		l.ChartVersion = vars.DefaultChartVersion
	}

	// Replica count
	if r := in.Target.Spec.ReplicaCount; r != nil {
		l.ReplicaCount = int(*r)
	} else {
		l.ReplicaCount = 1
	}

	// Container resources
	if c := in.Target.Spec.Container; c != nil && c.Resources != nil {
		if c.Resources.Requests != nil {
			l.CpuRequests = c.Resources.Requests.Cpu
			l.MemoryRequests = c.Resources.Requests.Memory
		}
		if c.Resources.Limits != nil {
			l.CpuLimits = c.Resources.Limits.Cpu
			l.MemoryLimits = c.Resources.Limits.Memory
		}
		// Custom image
		if c.Image != nil {
			l.ImageRepository = c.Image.Repository
			l.ImageTag = c.Image.Tag
			l.ImagePullPolicy = c.Image.PullPolicy
		}
	}

	// Flags configuration
	if f := in.Target.Spec.Flags; f != nil {
		l.LogLevel = logLevelToString(f.LogLevel)
		l.LogFormat = logFormatToString(f.LogFormat)
		l.WatchSingleNamespace = f.WatchSingleNamespace
		if f.RunnerMaxConcurrentReconciles != nil {
			l.RunnerMaxConcurrentReconciles = int(*f.RunnerMaxConcurrentReconciles)
		}
		l.UpdateStrategy = updateStrategyToString(f.UpdateStrategy)
		l.ExcludeLabelPropagationPrefixes = f.ExcludeLabelPropagationPrefixes
		l.K8sClientRateLimiterQPS = int(f.K8SClientRateLimiterQps)
		l.K8sClientRateLimiterBurst = int(f.K8SClientRateLimiterBurst)
	}

	// Metrics configuration
	if m := in.Target.Spec.Metrics; m != nil && m.ControllerManagerAddr != "" {
		l.MetricsEnabled = true
		l.ControllerManagerAddr = m.ControllerManagerAddr
		l.ListenerAddr = m.ListenerAddr
		l.ListenerEndpoint = m.ListenerEndpoint
	}

	// Other configuration
	l.ImagePullSecrets = in.Target.Spec.ImagePullSecrets
	l.PriorityClassName = in.Target.Spec.PriorityClassName

	// Export outputs
	ctx.Export(OpNamespace, pulumi.String(l.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(l.ReleaseName))
	ctx.Export(OpChartVersion, pulumi.String(l.ChartVersion))
	ctx.Export(OpDeploymentName, pulumi.String(l.ReleaseName))
	ctx.Export(OpServiceAccountName, pulumi.String(l.ReleaseName))

	if l.MetricsEnabled {
		ctx.Export(OpMetricsEndpoint, pulumi.String(fmt.Sprintf("%s.%s.svc.cluster.local%s", l.ReleaseName, l.Namespace, l.ControllerManagerAddr)))
	}

	return &l
}

func logLevelToString(level kubernetesgharunnerscalesetcontrollerv1.KubernetesGhaRunnerScaleSetControllerFlags_LogLevel) string {
	switch level {
	case kubernetesgharunnerscalesetcontrollerv1.KubernetesGhaRunnerScaleSetControllerFlags_debug:
		return "debug"
	case kubernetesgharunnerscalesetcontrollerv1.KubernetesGhaRunnerScaleSetControllerFlags_info:
		return "info"
	case kubernetesgharunnerscalesetcontrollerv1.KubernetesGhaRunnerScaleSetControllerFlags_warn:
		return "warn"
	case kubernetesgharunnerscalesetcontrollerv1.KubernetesGhaRunnerScaleSetControllerFlags_error:
		return "error"
	default:
		return "debug"
	}
}

func logFormatToString(format kubernetesgharunnerscalesetcontrollerv1.KubernetesGhaRunnerScaleSetControllerFlags_LogFormat) string {
	switch format {
	case kubernetesgharunnerscalesetcontrollerv1.KubernetesGhaRunnerScaleSetControllerFlags_json:
		return "json"
	default:
		return "text"
	}
}

func updateStrategyToString(strategy kubernetesgharunnerscalesetcontrollerv1.KubernetesGhaRunnerScaleSetControllerFlags_UpdateStrategy) string {
	switch strategy {
	case kubernetesgharunnerscalesetcontrollerv1.KubernetesGhaRunnerScaleSetControllerFlags_eventual:
		return "eventual"
	default:
		return "immediate"
	}
}
