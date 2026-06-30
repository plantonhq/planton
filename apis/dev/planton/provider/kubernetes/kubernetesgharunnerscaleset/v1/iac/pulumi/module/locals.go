package module

import (
	"encoding/base64"
	"strconv"

	kubernetesgharunnerscalesetv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesgharunnerscaleset/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds all derived configuration for the runner scale set deployment.
type Locals struct {
	KubernetesGhaRunnerScaleSet *kubernetesgharunnerscalesetv1.KubernetesGhaRunnerScaleSet
	KubeLabels                  map[string]string
	Namespace                   string
	CreateNamespace             bool
	ReleaseName                 string
	ChartVersion                string

	// GitHub configuration
	GitHubConfigURL     string
	GitHubSecretName    string
	UseExistingSecret   bool
	PatToken            string
	GitHubAppID         string
	GitHubAppInstallID  string
	GitHubAppPrivateKey string

	// Scaling
	MinRunners int32
	MaxRunners int32

	// Runner group and name
	RunnerGroup        string
	RunnerScaleSetName string

	// Container mode
	ContainerModeType           string
	WorkVolumeClaimStorageClass string
	WorkVolumeClaimSize         string
	WorkVolumeClaimAccessModes  []string

	// Runner container
	RunnerImageRepository string
	RunnerImageTag        string
	RunnerImagePullPolicy string
	RunnerCpuRequests     string
	RunnerCpuLimits       string
	RunnerMemoryRequests  string
	RunnerMemoryLimits    string
	RunnerEnvVars         map[string]string
	RunnerVolumeMounts    []*kubernetesgharunnerscalesetv1.KubernetesGhaRunnerScaleSetVolumeMount

	// Persistent volumes
	PersistentVolumes []*kubernetesgharunnerscalesetv1.KubernetesGhaRunnerScaleSetPersistentVolume

	// Controller service account
	ControllerServiceAccountName      string
	ControllerServiceAccountNamespace string

	// Other
	ImagePullSecrets []string
	Labels           map[string]string
	Annotations      map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *kubernetesgharunnerscalesetv1.KubernetesGhaRunnerScaleSetStackInput) *Locals {
	var l Locals
	l.KubernetesGhaRunnerScaleSet = in.Target

	// Initialize labels
	l.KubeLabels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: in.Target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesGhaRunnerScaleSet.String(),
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
		l.Namespace = "gha-runners"
	}
	l.CreateNamespace = in.Target.Spec.CreateNamespace

	// Release name
	l.ReleaseName = in.Target.Metadata.Name

	// Chart version
	if v := in.Target.Spec.HelmChartVersion; v != nil {
		l.ChartVersion = *v
	} else {
		l.ChartVersion = DefaultChartVersion
	}

	// GitHub configuration
	if gh := in.Target.Spec.Github; gh != nil {
		l.GitHubConfigURL = gh.ConfigUrl

		switch auth := gh.Auth.(type) {
		case *kubernetesgharunnerscalesetv1.KubernetesGhaRunnerScaleSetGitHubConfig_PatToken:
			if auth.PatToken != nil {
				l.PatToken = auth.PatToken.Token
			}
		case *kubernetesgharunnerscalesetv1.KubernetesGhaRunnerScaleSetGitHubConfig_GithubApp:
			if auth.GithubApp != nil {
				l.GitHubAppID = auth.GithubApp.AppId
				l.GitHubAppInstallID = auth.GithubApp.InstallationId
				l.GitHubAppPrivateKey = decodeBase64(auth.GithubApp.PrivateKeyBase64)
			}
		case *kubernetesgharunnerscalesetv1.KubernetesGhaRunnerScaleSetGitHubConfig_ExistingSecretName:
			l.GitHubSecretName = auth.ExistingSecretName
			l.UseExistingSecret = true
		}
	}

	// Generate secret name if not using existing
	if !l.UseExistingSecret {
		l.GitHubSecretName = l.ReleaseName + "-github-secret"
	}

	// Scaling configuration
	l.MinRunners = 0
	l.MaxRunners = 5
	if s := in.Target.Spec.Scaling; s != nil {
		if s.MinRunners != nil {
			l.MinRunners = *s.MinRunners
		}
		if s.MaxRunners != nil {
			l.MaxRunners = *s.MaxRunners
		}
	}

	// Runner group and name
	l.RunnerGroup = in.Target.Spec.GetRunnerGroup()
	l.RunnerScaleSetName = in.Target.Spec.RunnerScaleSetName
	if l.RunnerScaleSetName == "" {
		l.RunnerScaleSetName = l.ReleaseName
	}

	// Container mode
	if cm := in.Target.Spec.ContainerMode; cm != nil {
		l.ContainerModeType = containerModeTypeToString(cm.Type)

		if cm.WorkVolumeClaim != nil {
			l.WorkVolumeClaimStorageClass = cm.WorkVolumeClaim.StorageClass
			l.WorkVolumeClaimSize = cm.WorkVolumeClaim.Size
			l.WorkVolumeClaimAccessModes = cm.WorkVolumeClaim.AccessModes
			if len(l.WorkVolumeClaimAccessModes) == 0 {
				l.WorkVolumeClaimAccessModes = []string{"ReadWriteOnce"}
			}
		}
	}

	// Runner configuration
	if r := in.Target.Spec.Runner; r != nil {
		if r.Image != nil {
			l.RunnerImageRepository = r.Image.GetRepository()
			l.RunnerImageTag = r.Image.GetTag()
			l.RunnerImagePullPolicy = r.Image.GetPullPolicy()
		}
		if r.Resources != nil {
			if r.Resources.Requests != nil {
				l.RunnerCpuRequests = r.Resources.Requests.Cpu
				l.RunnerMemoryRequests = r.Resources.Requests.Memory
			}
			if r.Resources.Limits != nil {
				l.RunnerCpuLimits = r.Resources.Limits.Cpu
				l.RunnerMemoryLimits = r.Resources.Limits.Memory
			}
		}
		if len(r.Env) > 0 {
			l.RunnerEnvVars = make(map[string]string)
			for _, e := range r.Env {
				l.RunnerEnvVars[e.Name] = e.Value
			}
		}
		l.RunnerVolumeMounts = r.VolumeMounts
	}

	// Persistent volumes
	l.PersistentVolumes = in.Target.Spec.PersistentVolumes

	// Controller service account
	if csa := in.Target.Spec.ControllerServiceAccount; csa != nil {
		l.ControllerServiceAccountName = csa.Name
		l.ControllerServiceAccountNamespace = csa.Namespace
	}

	// Other configuration
	l.ImagePullSecrets = in.Target.Spec.ImagePullSecrets
	l.Labels = in.Target.Spec.Labels
	l.Annotations = in.Target.Spec.Annotations

	// Export outputs
	ctx.Export(OpNamespace, pulumi.String(l.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(l.ReleaseName))
	ctx.Export(OpChartVersion, pulumi.String(l.ChartVersion))
	ctx.Export(OpRunnerScaleSetName, pulumi.String(l.RunnerScaleSetName))
	ctx.Export(OpGitHubConfigURL, pulumi.String(l.GitHubConfigURL))
	ctx.Export(OpGitHubSecretName, pulumi.String(l.GitHubSecretName))
	ctx.Export(OpMinRunners, pulumi.Int(l.MinRunners))
	ctx.Export(OpMaxRunners, pulumi.Int(l.MaxRunners))
	ctx.Export(OpContainerMode, pulumi.String(l.ContainerModeType))

	// Export PVC names
	var pvcNames []string
	for _, pv := range l.PersistentVolumes {
		pvcNames = append(pvcNames, l.ReleaseName+"-"+pv.Name)
	}
	ctx.Export(OpPvcNames, pulumi.ToStringArray(pvcNames))

	return &l
}

func containerModeTypeToString(t kubernetesgharunnerscalesetv1.KubernetesGhaRunnerScaleSetContainerMode_ContainerModeType) string {
	switch t {
	case kubernetesgharunnerscalesetv1.KubernetesGhaRunnerScaleSetContainerMode_DIND:
		return "dind"
	case kubernetesgharunnerscalesetv1.KubernetesGhaRunnerScaleSetContainerMode_KUBERNETES:
		return "kubernetes"
	case kubernetesgharunnerscalesetv1.KubernetesGhaRunnerScaleSetContainerMode_KUBERNETES_NO_VOLUME:
		return "kubernetes-novolume"
	case kubernetesgharunnerscalesetv1.KubernetesGhaRunnerScaleSetContainerMode_DEFAULT:
		return ""
	default:
		return ""
	}
}

// decodeBase64 decodes a base64 encoded string.
// Returns empty string if decoding fails.
func decodeBase64(encoded string) string {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return ""
	}
	return string(decoded)
}
