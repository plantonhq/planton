package module

import (
	"sort"

	"github.com/pkg/errors"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/workloadpod"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"google.golang.org/protobuf/proto"
)

// deployment renders the flagd Deployment. ConfigMap sources are mounted as
// DIRECTORIES (never subPath, which would never see an edit); flagd reads
// its sources from FLAGD_SOURCES (the module-owned Secret) and every other
// setting from its arguments. Liveness is /healthz and readiness /readyz on
// the management port - flagd reports ready only once every source has
// synced. Pod scheduling and security render through the shared workload
// pod builder. Terraform twin: kubernetes_deployment_v1.flagd.
func deployment(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource, deps []pulumi.ResourceOption) (*appsv1.Deployment, error) {
	spec := locals.Spec

	env := kubernetescorev1.EnvVarArray{
		kubernetescorev1.EnvVarArgs{
			Name: pulumi.String("FLAGD_SOURCES"),
			ValueFrom: kubernetescorev1.EnvVarSourceArgs{SecretKeyRef: kubernetescorev1.SecretKeySelectorArgs{
				Name: pulumi.String(locals.SourcesSecretName), Key: pulumi.String("sources"),
			}},
		},
	}
	for _, name := range sortedKeys(spec.GetExtraEnv()) {
		env = append(env, kubernetescorev1.EnvVarArgs{Name: pulumi.String(name), Value: pulumi.String(spec.GetExtraEnv()[name])})
	}
	secretEnvNames := make([]string, 0, len(spec.GetExtraEnvFromSecret()))
	for name := range spec.GetExtraEnvFromSecret() {
		secretEnvNames = append(secretEnvNames, name)
	}
	sort.Strings(secretEnvNames)
	for _, name := range secretEnvNames {
		ref := spec.GetExtraEnvFromSecret()[name]
		env = append(env, kubernetescorev1.EnvVarArgs{
			Name: pulumi.String(name),
			ValueFrom: kubernetescorev1.EnvVarSourceArgs{SecretKeyRef: kubernetescorev1.SecretKeySelectorArgs{
				Name: pulumi.String(ref.GetName()), Key: pulumi.String(ref.GetKey()),
			}},
		})
	}

	var mounts kubernetescorev1.VolumeMountArray
	var volumes kubernetescorev1.VolumeArray
	for _, m := range locals.Config.ConfigMapMounts {
		mounts = append(mounts, kubernetescorev1.VolumeMountArgs{Name: pulumi.String(m.Volume), MountPath: pulumi.String(m.MountPath), ReadOnly: pulumi.Bool(true)})
		volumes = append(volumes, kubernetescorev1.VolumeArgs{
			Name:      pulumi.String(m.Volume),
			ConfigMap: kubernetescorev1.ConfigMapVolumeSourceArgs{Name: pulumi.String(m.ConfigMap)},
		})
	}
	for _, m := range locals.Config.GrpcCaMounts {
		mounts = append(mounts, kubernetescorev1.VolumeMountArgs{Name: pulumi.String(m.Volume), MountPath: pulumi.String(m.MountPath), ReadOnly: pulumi.Bool(true)})
		volumes = append(volumes, secretKeyVolume(m.Volume, m.Secret, m.Key))
	}
	if tls := spec.GetServer().GetTlsSecretName(); tls != "" {
		mounts = append(mounts, kubernetescorev1.VolumeMountArgs{Name: pulumi.String("server-tls"), MountPath: pulumi.String(vars.ServerTlsMountPath), ReadOnly: pulumi.Bool(true)})
		volumes = append(volumes, kubernetescorev1.VolumeArgs{Name: pulumi.String("server-tls"), Secret: kubernetescorev1.SecretVolumeSourceArgs{SecretName: pulumi.String(tls)}})
	}
	if ca := spec.GetTelemetry().GetOtelCaCertSecret(); ca != nil {
		mounts = append(mounts, kubernetescorev1.VolumeMountArgs{Name: pulumi.String("otel-ca"), MountPath: pulumi.String(vars.OtelCaMountPath), ReadOnly: pulumi.Bool(true)})
		volumes = append(volumes, secretKeyVolume("otel-ca", ca.GetName(), ca.GetKey()))
	}
	if tls := spec.GetTelemetry().GetOtelClientTlsSecretName(); tls != "" {
		mounts = append(mounts, kubernetescorev1.VolumeMountArgs{Name: pulumi.String("otel-tls"), MountPath: pulumi.String(vars.OtelTlsMountPath), ReadOnly: pulumi.Bool(true)})
		volumes = append(volumes, kubernetescorev1.VolumeArgs{Name: pulumi.String("otel-tls"), Secret: kubernetescorev1.SecretVolumeSourceArgs{SecretName: pulumi.String(tls)}})
	}

	pullPolicy := spec.GetImage().GetPullPolicy()
	if pullPolicy == "" {
		pullPolicy = "IfNotPresent"
	}
	container := kubernetescorev1.ContainerArgs{
		Name:            pulumi.String("flagd"),
		Image:           pulumi.String(locals.Image),
		ImagePullPolicy: pulumi.String(pullPolicy),
		Args:            pulumi.ToStringArray(locals.Config.Args),
		Env:             env,
		Ports: kubernetescorev1.ContainerPortArray{
			kubernetescorev1.ContainerPortArgs{Name: pulumi.String("evaluation"), ContainerPort: pulumi.Int(locals.Port)},
			kubernetescorev1.ContainerPortArgs{Name: pulumi.String("management"), ContainerPort: pulumi.Int(locals.ManagementPort)},
			kubernetescorev1.ContainerPortArgs{Name: pulumi.String("sync"), ContainerPort: pulumi.Int(locals.SyncPort)},
			kubernetescorev1.ContainerPortArgs{Name: pulumi.String("ofrep"), ContainerPort: pulumi.Int(locals.OfrepPort)},
		},
		LivenessProbe: kubernetescorev1.ProbeArgs{
			HttpGet:             kubernetescorev1.HTTPGetActionArgs{Path: pulumi.String("/healthz"), Port: pulumi.String("management")},
			InitialDelaySeconds: pulumi.Int(5),
			PeriodSeconds:       pulumi.Int(10),
		},
		ReadinessProbe: kubernetescorev1.ProbeArgs{
			HttpGet:             kubernetescorev1.HTTPGetActionArgs{Path: pulumi.String("/readyz"), Port: pulumi.String("management")},
			InitialDelaySeconds: pulumi.Int(5),
			PeriodSeconds:       pulumi.Int(5),
		},
		VolumeMounts: mounts,
	}
	if r := resourcesArgs(spec.GetResources()); r != nil {
		container.Resources = r
	}
	if sc := containerSecurityContextArgs(spec.GetContainerSecurityContext()); sc != nil {
		container.SecurityContext = sc
	}

	podLabels := map[string]string{}
	for k, v := range spec.GetPodLabels() {
		podLabels[k] = v
	}
	for k, v := range locals.Labels {
		podLabels[k] = v
	}
	for k, v := range locals.SelectorLabels {
		podLabels[k] = v
	}

	annotations := map[string]string{"checksum/sources": locals.Config.SourcesChecksum}
	for k, v := range spec.GetPodAnnotations() {
		annotations[k] = v
	}

	pod := &kubernetesprovider.WorkloadPod{
		ServiceAccount:  &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: locals.ServiceAccountName}},
		Annotations:     annotations,
		Scheduling:      selfSpreading(spec.GetScheduling(), locals.SelectorLabels),
		SecurityContext: spec.GetPodSecurityContext(),
	}
	template := workloadpod.BuildPodTemplateSpec(pod, workloadpod.PodTemplateInputs{
		Labels:               podLabels,
		Containers:           kubernetescorev1.ContainerArray{container},
		Volumes:              volumes,
		ImagePullSecretNames: spec.GetImage().GetPullSecretNames(),
	}, "")

	deploymentSpec := &appsv1.DeploymentSpecArgs{
		Selector: kubernetesmeta.LabelSelectorArgs{MatchLabels: pulumi.ToStringMap(locals.SelectorLabels)},
		Template: template,
	}
	if !spec.GetHpa().GetEnabled() {
		replicas := 1
		if spec.Replicas != nil {
			replicas = int(spec.GetReplicas())
		}
		deploymentSpec.Replicas = pulumi.Int(replicas)
	}

	created, err := appsv1.NewDeployment(ctx, locals.Name, &appsv1.DeploymentArgs{
		Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
			Name: pulumi.String(locals.Name), Namespace: pulumi.String(locals.Namespace), Labels: pulumi.ToStringMap(locals.Labels),
		}),
		Spec: deploymentSpec,
	}, append([]pulumi.ResourceOption{pulumi.Provider(provider)}, deps...)...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create flagd deployment")
	}
	return created, nil
}

// service exposes the four flagd ports. Terraform twin:
// kubernetes_service_v1.flagd.
func service(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource, deps []pulumi.ResourceOption) error {
	serviceType := locals.Spec.GetServer().GetServiceType()
	if serviceType == "" {
		serviceType = "ClusterIP"
	}
	_, err := kubernetescorev1.NewService(ctx, locals.Name, &kubernetescorev1.ServiceArgs{
		Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
			Name: pulumi.String(locals.Name), Namespace: pulumi.String(locals.Namespace), Labels: pulumi.ToStringMap(serviceLabels(locals)),
		}),
		Spec: kubernetescorev1.ServiceSpecArgs{
			Type:     pulumi.String(serviceType),
			Selector: pulumi.ToStringMap(locals.SelectorLabels),
			Ports: kubernetescorev1.ServicePortArray{
				kubernetescorev1.ServicePortArgs{Name: pulumi.String("evaluation"), Port: pulumi.Int(locals.Port), TargetPort: pulumi.String("evaluation")},
				kubernetescorev1.ServicePortArgs{Name: pulumi.String("management"), Port: pulumi.Int(locals.ManagementPort), TargetPort: pulumi.String("management")},
				kubernetescorev1.ServicePortArgs{Name: pulumi.String("sync"), Port: pulumi.Int(locals.SyncPort), TargetPort: pulumi.String("sync")},
				kubernetescorev1.ServicePortArgs{Name: pulumi.String("ofrep"), Port: pulumi.Int(locals.OfrepPort), TargetPort: pulumi.String("ofrep")},
			},
		},
	}, append([]pulumi.ResourceOption{pulumi.Provider(provider)}, deps...)...)
	if err != nil {
		return errors.Wrap(err, "failed to create flagd service")
	}
	return nil
}

// serviceLabels are the Planton labels plus the pod selector labels: the
// ServiceMonitor selects the Service by them.
func serviceLabels(locals *Locals) map[string]string {
	labels := map[string]string{}
	for k, v := range locals.Labels {
		labels[k] = v
	}
	for k, v := range locals.SelectorLabels {
		labels[k] = v
	}
	return labels
}

// selfSpreading substitutes the flagd selector for an empty topology-spread
// label selector (the shared pod builder emits what it is given).
func selfSpreading(s *kubernetesprovider.WorkloadScheduling, selector map[string]string) *kubernetesprovider.WorkloadScheduling {
	if s == nil {
		return nil
	}
	clone := proto.Clone(s).(*kubernetesprovider.WorkloadScheduling)
	for _, c := range clone.GetTopologySpreadConstraints() {
		if len(c.GetMatchLabels()) == 0 {
			c.MatchLabels = selector
		}
	}
	return clone
}

func secretKeyVolume(volume, secret, key string) kubernetescorev1.VolumeArgs {
	return kubernetescorev1.VolumeArgs{
		Name: pulumi.String(volume),
		Secret: kubernetescorev1.SecretVolumeSourceArgs{
			SecretName: pulumi.String(secret),
			Items:      kubernetescorev1.KeyToPathArray{kubernetescorev1.KeyToPathArgs{Key: pulumi.String(key), Path: pulumi.String(key)}},
		},
	}
}

func resourcesArgs(r *kubernetesprovider.ContainerResources) *kubernetescorev1.ResourceRequirementsArgs {
	if r == nil {
		return nil
	}
	out := &kubernetescorev1.ResourceRequirementsArgs{}
	set := false
	if m := cpuMemory(r.GetRequests()); len(m) > 0 {
		out.Requests = pulumi.ToStringMap(m)
		set = true
	}
	if m := cpuMemory(r.GetLimits()); len(m) > 0 {
		out.Limits = pulumi.ToStringMap(m)
		set = true
	}
	if !set {
		return nil
	}
	return out
}

func cpuMemory(c *kubernetesprovider.CpuMemory) map[string]string {
	out := map[string]string{}
	if c.GetCpu() != "" {
		out["cpu"] = c.GetCpu()
	}
	if c.GetMemory() != "" {
		out["memory"] = c.GetMemory()
	}
	return out
}

func containerSecurityContextArgs(c *kubernetesprovider.WorkloadContainerSecurityContext) *kubernetescorev1.SecurityContextArgs {
	if c == nil {
		return nil
	}
	out := &kubernetescorev1.SecurityContextArgs{}
	if c.GetPrivileged() {
		out.Privileged = pulumi.Bool(true)
	}
	if c.RunAsUser != nil {
		out.RunAsUser = pulumi.Int(int(c.GetRunAsUser()))
	}
	if c.RunAsGroup != nil {
		out.RunAsGroup = pulumi.Int(int(c.GetRunAsGroup()))
	}
	if c.RunAsNonRoot != nil {
		out.RunAsNonRoot = pulumi.Bool(c.GetRunAsNonRoot())
	}
	if c.ReadOnlyRootFilesystem != nil {
		out.ReadOnlyRootFilesystem = pulumi.Bool(c.GetReadOnlyRootFilesystem())
	}
	if c.AllowPrivilegeEscalation != nil {
		out.AllowPrivilegeEscalation = pulumi.Bool(c.GetAllowPrivilegeEscalation())
	}
	if caps := c.GetCapabilities(); caps != nil && (len(caps.GetAdd()) > 0 || len(caps.GetDrop()) > 0) {
		out.Capabilities = kubernetescorev1.CapabilitiesArgs{
			Add:  pulumi.ToStringArray(caps.GetAdd()),
			Drop: pulumi.ToStringArray(caps.GetDrop()),
		}
	}
	if sp := c.GetSeccompProfile(); sp != nil && sp.GetType() != "" {
		profile := kubernetescorev1.SeccompProfileArgs{Type: pulumi.String(sp.GetType())}
		if sp.GetLocalhostProfile() != "" {
			profile.LocalhostProfile = pulumi.String(sp.GetLocalhostProfile())
		}
		out.SeccompProfile = profile
	}
	return out
}
