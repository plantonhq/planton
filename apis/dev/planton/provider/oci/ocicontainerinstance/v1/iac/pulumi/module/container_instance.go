package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	ocicontainerinstancev1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocicontainerinstance/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/containerengine"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func containerInstance(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciContainerInstance.Spec

	args := &containerengine.ContainerInstanceArgs{
		CompartmentId:      pulumi.String(spec.CompartmentId.GetValue()),
		AvailabilityDomain: pulumi.String(spec.AvailabilityDomain),
		Shape:              pulumi.String(spec.Shape),
		DisplayName:        pulumi.StringPtr(locals.DisplayName),
		FreeformTags:       pulumi.ToStringMap(locals.FreeformTags),
		ShapeConfig:        buildShapeConfig(spec.ShapeConfig),
		Containers:         buildContainers(spec.Containers),
		Vnics:              buildVnics(spec.Vnics),
	}

	if spec.ContainerRestartPolicy != ocicontainerinstancev1.OciContainerInstanceSpec_restart_policy_unspecified {
		args.ContainerRestartPolicy = pulumi.StringPtr(strings.ToUpper(spec.ContainerRestartPolicy.String()))
	}

	if spec.FaultDomain != "" {
		args.FaultDomain = pulumi.StringPtr(spec.FaultDomain)
	}

	if spec.GracefulShutdownTimeoutInSeconds > 0 {
		args.GracefulShutdownTimeoutInSeconds = pulumi.StringPtr(fmt.Sprintf("%d", spec.GracefulShutdownTimeoutInSeconds))
	}

	if spec.DnsConfig != nil {
		args.DnsConfig = buildDnsConfig(spec.DnsConfig)
	}

	if len(spec.ImagePullSecrets) > 0 {
		args.ImagePullSecrets = buildImagePullSecrets(spec.ImagePullSecrets)
	}

	if len(spec.Volumes) > 0 {
		args.Volumes = buildVolumes(spec.Volumes)
	}

	created, err := containerengine.NewContainerInstance(ctx, locals.DisplayName, args, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create oci container instance")
	}

	ctx.Export(OpContainerInstanceId, created.ID())

	ctx.Export(OpContainerIds, created.Containers.ApplyT(func(containers []containerengine.ContainerInstanceContainer) string {
		ids := make([]string, 0, len(containers))
		for _, c := range containers {
			if c.ContainerId != nil {
				ids = append(ids, *c.ContainerId)
			}
		}
		return strings.Join(ids, ",")
	}).(pulumi.StringOutput))

	return nil
}

func buildShapeConfig(sc *ocicontainerinstancev1.OciContainerInstanceSpec_ShapeConfig) *containerengine.ContainerInstanceShapeConfigArgs {
	args := &containerengine.ContainerInstanceShapeConfigArgs{
		Ocpus: pulumi.Float64(float64(sc.Ocpus)),
	}
	if sc.MemoryInGbs > 0 {
		args.MemoryInGbs = pulumi.Float64Ptr(float64(sc.MemoryInGbs))
	}
	return args
}

func buildContainers(containers []*ocicontainerinstancev1.OciContainerInstanceSpec_Container) containerengine.ContainerInstanceContainerArray {
	result := make(containerengine.ContainerInstanceContainerArray, len(containers))
	for i, c := range containers {
		result[i] = buildContainer(c)
	}
	return result
}

func buildContainer(c *ocicontainerinstancev1.OciContainerInstanceSpec_Container) *containerengine.ContainerInstanceContainerArgs {
	args := &containerengine.ContainerInstanceContainerArgs{
		ImageUrl: pulumi.String(c.ImageUrl),
	}

	if c.DisplayName != "" {
		args.DisplayName = pulumi.StringPtr(c.DisplayName)
	}

	if len(c.Command) > 0 {
		args.Commands = pulumi.ToStringArray(c.Command)
	}

	if len(c.Arguments) > 0 {
		args.Arguments = pulumi.ToStringArray(c.Arguments)
	}

	if len(c.EnvironmentVariables) > 0 {
		args.EnvironmentVariables = pulumi.ToStringMap(c.EnvironmentVariables)
	}

	if c.WorkingDirectory != "" {
		args.WorkingDirectory = pulumi.StringPtr(c.WorkingDirectory)
	}

	if c.IsResourcePrincipalDisabled {
		args.IsResourcePrincipalDisabled = pulumi.BoolPtr(true)
	}

	if c.ResourceConfig != nil {
		args.ResourceConfig = buildContainerResourceConfig(c.ResourceConfig)
	}

	if len(c.HealthChecks) > 0 {
		args.HealthChecks = buildHealthChecks(c.HealthChecks)
	}

	if c.SecurityContext != nil {
		args.SecurityContext = buildSecurityContext(c.SecurityContext)
	}

	if len(c.VolumeMounts) > 0 {
		args.VolumeMounts = buildVolumeMounts(c.VolumeMounts)
	}

	return args
}

func buildContainerResourceConfig(rc *ocicontainerinstancev1.OciContainerInstanceSpec_ContainerResourceConfig) *containerengine.ContainerInstanceContainerResourceConfigArgs {
	args := &containerengine.ContainerInstanceContainerResourceConfigArgs{}
	if rc.MemoryLimitInGbs > 0 {
		args.MemoryLimitInGbs = pulumi.Float64Ptr(float64(rc.MemoryLimitInGbs))
	}
	if rc.VcpusLimit > 0 {
		args.VcpusLimit = pulumi.Float64Ptr(float64(rc.VcpusLimit))
	}
	return args
}

func buildHealthChecks(checks []*ocicontainerinstancev1.OciContainerInstanceSpec_HealthCheck) containerengine.ContainerInstanceContainerHealthCheckArray {
	result := make(containerengine.ContainerInstanceContainerHealthCheckArray, len(checks))
	for i, hc := range checks {
		args := &containerengine.ContainerInstanceContainerHealthCheckArgs{
			HealthCheckType: pulumi.String(strings.ToUpper(hc.HealthCheckType.String())),
			Port:            pulumi.Int(int(hc.Port)),
		}

		if hc.Name != "" {
			args.Name = pulumi.StringPtr(hc.Name)
		}
		if hc.Path != "" {
			args.Path = pulumi.StringPtr(hc.Path)
		}
		if hc.FailureAction != ocicontainerinstancev1.OciContainerInstanceSpec_failure_action_unspecified {
			args.FailureAction = pulumi.StringPtr(strings.ToUpper(hc.FailureAction.String()))
		}
		if hc.FailureThreshold > 0 {
			args.FailureThreshold = pulumi.IntPtr(int(hc.FailureThreshold))
		}
		if hc.SuccessThreshold > 0 {
			args.SuccessThreshold = pulumi.IntPtr(int(hc.SuccessThreshold))
		}
		if hc.InitialDelayInSeconds > 0 {
			args.InitialDelayInSeconds = pulumi.IntPtr(int(hc.InitialDelayInSeconds))
		}
		if hc.IntervalInSeconds > 0 {
			args.IntervalInSeconds = pulumi.IntPtr(int(hc.IntervalInSeconds))
		}
		if hc.TimeoutInSeconds > 0 {
			args.TimeoutInSeconds = pulumi.IntPtr(int(hc.TimeoutInSeconds))
		}
		if len(hc.Headers) > 0 {
			headers := make(containerengine.ContainerInstanceContainerHealthCheckHeaderArray, len(hc.Headers))
			for j, h := range hc.Headers {
				headers[j] = &containerengine.ContainerInstanceContainerHealthCheckHeaderArgs{
					Name:  pulumi.String(h.Name),
					Value: pulumi.String(h.Value),
				}
			}
			args.Headers = headers
		}

		result[i] = args
	}
	return result
}

func buildSecurityContext(sc *ocicontainerinstancev1.OciContainerInstanceSpec_SecurityContext) *containerengine.ContainerInstanceContainerSecurityContextArgs {
	args := &containerengine.ContainerInstanceContainerSecurityContextArgs{
		SecurityContextType: pulumi.StringPtr("LINUX"),
	}

	if sc.IsNonRootUserCheckEnabled {
		args.IsNonRootUserCheckEnabled = pulumi.BoolPtr(true)
	}
	if sc.IsRootFileSystemReadonly {
		args.IsRootFileSystemReadonly = pulumi.BoolPtr(true)
	}
	if sc.RunAsUser > 0 {
		args.RunAsUser = pulumi.IntPtr(int(sc.RunAsUser))
	}
	if sc.RunAsGroup > 0 {
		args.RunAsGroup = pulumi.IntPtr(int(sc.RunAsGroup))
	}
	if sc.Capabilities != nil {
		caps := &containerengine.ContainerInstanceContainerSecurityContextCapabilitiesArgs{}
		if len(sc.Capabilities.AddCapabilities) > 0 {
			caps.AddCapabilities = pulumi.ToStringArray(sc.Capabilities.AddCapabilities)
		}
		if len(sc.Capabilities.DropCapabilities) > 0 {
			caps.DropCapabilities = pulumi.ToStringArray(sc.Capabilities.DropCapabilities)
		}
		args.Capabilities = caps
	}

	return args
}

func buildVolumeMounts(mounts []*ocicontainerinstancev1.OciContainerInstanceSpec_VolumeMount) containerengine.ContainerInstanceContainerVolumeMountArray {
	result := make(containerengine.ContainerInstanceContainerVolumeMountArray, len(mounts))
	for i, vm := range mounts {
		args := &containerengine.ContainerInstanceContainerVolumeMountArgs{
			MountPath:  pulumi.String(vm.MountPath),
			VolumeName: pulumi.String(vm.VolumeName),
		}
		if vm.IsReadOnly {
			args.IsReadOnly = pulumi.BoolPtr(true)
		}
		if vm.Partition > 0 {
			args.Partition = pulumi.IntPtr(int(vm.Partition))
		}
		if vm.SubPath != "" {
			args.SubPath = pulumi.StringPtr(vm.SubPath)
		}
		result[i] = args
	}
	return result
}

func buildVnics(vnics []*ocicontainerinstancev1.OciContainerInstanceSpec_Vnic) containerengine.ContainerInstanceVnicArray {
	result := make(containerengine.ContainerInstanceVnicArray, len(vnics))
	for i, v := range vnics {
		args := &containerengine.ContainerInstanceVnicArgs{
			SubnetId: pulumi.String(v.SubnetId.GetValue()),
		}
		if v.DisplayName != "" {
			args.DisplayName = pulumi.StringPtr(v.DisplayName)
		}
		if v.HostnameLabel != "" {
			args.HostnameLabel = pulumi.StringPtr(v.HostnameLabel)
		}
		if v.IsPublicIpAssigned != nil {
			args.IsPublicIpAssigned = pulumi.BoolPtr(*v.IsPublicIpAssigned)
		}
		if len(v.NsgIds) > 0 {
			nsgIds := make(pulumi.StringArray, len(v.NsgIds))
			for j, nsg := range v.NsgIds {
				nsgIds[j] = pulumi.String(nsg.GetValue())
			}
			args.NsgIds = nsgIds
		}
		if v.PrivateIp != "" {
			args.PrivateIp = pulumi.StringPtr(v.PrivateIp)
		}
		if v.SkipSourceDestCheck {
			args.SkipSourceDestCheck = pulumi.BoolPtr(true)
		}
		result[i] = args
	}
	return result
}

func buildDnsConfig(dns *ocicontainerinstancev1.OciContainerInstanceSpec_DnsConfig) *containerengine.ContainerInstanceDnsConfigArgs {
	args := &containerengine.ContainerInstanceDnsConfigArgs{}
	if len(dns.Nameservers) > 0 {
		args.Nameservers = pulumi.ToStringArray(dns.Nameservers)
	}
	if len(dns.Options) > 0 {
		args.Options = pulumi.ToStringArray(dns.Options)
	}
	if len(dns.Searches) > 0 {
		args.Searches = pulumi.ToStringArray(dns.Searches)
	}
	return args
}

func buildImagePullSecrets(secrets []*ocicontainerinstancev1.OciContainerInstanceSpec_ImagePullSecret) containerengine.ContainerInstanceImagePullSecretArray {
	result := make(containerengine.ContainerInstanceImagePullSecretArray, len(secrets))
	for i, s := range secrets {
		args := &containerengine.ContainerInstanceImagePullSecretArgs{
			RegistryEndpoint: pulumi.String(s.RegistryEndpoint),
			SecretType:       pulumi.String(strings.ToUpper(s.SecretType.String())),
		}
		if s.Username != "" {
			args.Username = pulumi.StringPtr(s.Username)
		}
		if s.Password != "" {
			args.Password = pulumi.StringPtr(s.Password)
		}
		if s.SecretId != nil && s.SecretId.GetValue() != "" {
			args.SecretId = pulumi.StringPtr(s.SecretId.GetValue())
		}
		result[i] = args
	}
	return result
}

func buildVolumes(volumes []*ocicontainerinstancev1.OciContainerInstanceSpec_Volume) containerengine.ContainerInstanceVolumeArray {
	result := make(containerengine.ContainerInstanceVolumeArray, len(volumes))
	for i, v := range volumes {
		args := &containerengine.ContainerInstanceVolumeArgs{
			Name:       pulumi.String(v.Name),
			VolumeType: pulumi.String(strings.ToUpper(v.VolumeType.String())),
		}
		if v.BackingStore != "" {
			args.BackingStore = pulumi.StringPtr(v.BackingStore)
		}
		if len(v.Configs) > 0 {
			configs := make(containerengine.ContainerInstanceVolumeConfigArray, len(v.Configs))
			for j, c := range v.Configs {
				configArgs := &containerengine.ContainerInstanceVolumeConfigArgs{}
				if c.Data != "" {
					configArgs.Data = pulumi.StringPtr(c.Data)
				}
				if c.FileName != "" {
					configArgs.FileName = pulumi.StringPtr(c.FileName)
				}
				if c.Path != "" {
					configArgs.Path = pulumi.StringPtr(c.Path)
				}
				configs[j] = configArgs
			}
			args.Configs = configs
		}
		result[i] = args
	}
	return result
}
