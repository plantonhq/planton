package module

import (
	"github.com/pkg/errors"
	alicloudfunctionv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudfunction/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/fc"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudfunctionv1.AlicloudFunctionStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AlicloudFunction.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	args := &fc.V3FunctionArgs{
		FunctionName:         pulumi.String(spec.FunctionName),
		Handler:              pulumi.String(spec.Handler),
		Runtime:              pulumi.String(spec.Runtime),
		Description:          optionalString(spec.Description),
		Cpu:                  optionalFloat64(spec.Cpu),
		MemorySize:           optionalInt(spec.MemorySize),
		Timeout:              optionalInt(spec.Timeout),
		DiskSize:             optionalInt(spec.DiskSize),
		InstanceConcurrency:  optionalInt(spec.InstanceConcurrency),
		InternetAccess:       optionalBool(spec.InternetAccess),
		EnvironmentVariables: pulumi.ToStringMap(spec.EnvironmentVariables),
		Layers:               pulumi.ToStringArray(spec.Layers),
		Tags:                 pulumi.ToStringMap(locals.Tags),
		ResourceGroupId:      optionalString(spec.ResourceGroupId),
	}

	if spec.Role != nil {
		args.Role = pulumi.StringPtr(spec.Role.GetValue())
	}

	if spec.Code != nil {
		args.Code = codeArgs(spec.Code)
	}

	if spec.VpcConfig != nil {
		args.VpcConfig = vpcConfigArgs(spec.VpcConfig)
	}

	if spec.LogConfig != nil {
		args.LogConfig = logConfigArgs(spec.LogConfig)
	}

	if spec.CustomContainerConfig != nil {
		args.CustomContainerConfig = customContainerConfigArgs(spec.CustomContainerConfig)
	}

	if spec.CustomRuntimeConfig != nil {
		args.CustomRuntimeConfig = customRuntimeConfigArgs(spec.CustomRuntimeConfig)
	}

	if spec.InstanceLifecycleConfig != nil {
		args.InstanceLifecycleConfig = instanceLifecycleConfigArgs(spec.InstanceLifecycleConfig)
	}

	if spec.NasConfig != nil {
		args.NasConfig = nasConfigArgs(spec.NasConfig)
	}

	if spec.GpuConfig != nil {
		args.GpuConfig = gpuConfigArgs(spec.GpuConfig)
	}

	function, err := fc.NewV3Function(ctx, spec.FunctionName, args,
		pulumi.Provider(alicloudProvider),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create function %s", spec.FunctionName)
	}

	ctx.Export(OpFunctionId, function.FunctionId)
	ctx.Export(OpFunctionName, function.FunctionName)
	ctx.Export(OpFunctionArn, function.FunctionArn)

	return nil
}

func codeArgs(code *alicloudfunctionv1.AlicloudFunctionCode) fc.V3FunctionCodePtrInput {
	return &fc.V3FunctionCodeArgs{
		OssBucketName: optionalString(code.OssBucketName),
		OssObjectName: optionalString(code.OssObjectName),
		ZipFile:       optionalString(code.ZipFile),
		Checksum:      optionalString(code.Checksum),
	}
}

func vpcConfigArgs(cfg *alicloudfunctionv1.AlicloudFunctionVpcConfig) fc.V3FunctionVpcConfigPtrInput {
	vswitchIds := make(pulumi.StringArray, 0, len(cfg.VswitchIds))
	for _, ref := range cfg.VswitchIds {
		vswitchIds = append(vswitchIds, pulumi.String(ref.GetValue()))
	}

	vpcArgs := &fc.V3FunctionVpcConfigArgs{
		VswitchIds: vswitchIds,
	}

	if cfg.VpcId != nil {
		vpcArgs.VpcId = pulumi.StringPtr(cfg.VpcId.GetValue())
	}

	if cfg.SecurityGroupId != nil {
		vpcArgs.SecurityGroupId = pulumi.StringPtr(cfg.SecurityGroupId.GetValue())
	}

	return vpcArgs
}

func logConfigArgs(cfg *alicloudfunctionv1.AlicloudFunctionLogConfig) fc.V3FunctionLogConfigPtrInput {
	logArgs := &fc.V3FunctionLogConfigArgs{
		Logstore: optionalString(cfg.Logstore),
	}

	if cfg.Project != nil {
		logArgs.Project = pulumi.StringPtr(cfg.Project.GetValue())
	}

	if cfg.LogBeginRule != nil {
		logArgs.LogBeginRule = pulumi.StringPtr(*cfg.LogBeginRule)
	}

	if cfg.EnableInstanceMetrics != nil {
		logArgs.EnableInstanceMetrics = pulumi.BoolPtr(*cfg.EnableInstanceMetrics)
	}

	if cfg.EnableRequestMetrics != nil {
		logArgs.EnableRequestMetrics = pulumi.BoolPtr(*cfg.EnableRequestMetrics)
	}

	return logArgs
}

func customContainerConfigArgs(cfg *alicloudfunctionv1.AlicloudFunctionCustomContainerConfig) fc.V3FunctionCustomContainerConfigPtrInput {
	containerArgs := &fc.V3FunctionCustomContainerConfigArgs{
		Image:       optionalString(cfg.Image),
		Entrypoints: pulumi.ToStringArray(cfg.Entrypoint),
		Commands:    pulumi.ToStringArray(cfg.Command),
	}

	if cfg.Port != nil {
		containerArgs.Port = pulumi.IntPtr(int(*cfg.Port))
	}

	if cfg.HealthCheckConfig != nil {
		containerArgs.HealthCheckConfig = containerHealthCheckArgs(cfg.HealthCheckConfig)
	}

	return containerArgs
}

func containerHealthCheckArgs(hc *alicloudfunctionv1.AlicloudFunctionHealthCheckConfig) fc.V3FunctionCustomContainerConfigHealthCheckConfigPtrInput {
	args := &fc.V3FunctionCustomContainerConfigHealthCheckConfigArgs{
		HttpGetUrl: optionalString(hc.HttpGetUrl),
	}

	if hc.InitialDelaySeconds != nil {
		args.InitialDelaySeconds = pulumi.IntPtr(int(*hc.InitialDelaySeconds))
	}
	if hc.TimeoutSeconds != nil {
		args.TimeoutSeconds = pulumi.IntPtr(int(*hc.TimeoutSeconds))
	}
	if hc.PeriodSeconds != nil {
		args.PeriodSeconds = pulumi.IntPtr(int(*hc.PeriodSeconds))
	}
	if hc.FailureThreshold != nil {
		args.FailureThreshold = pulumi.IntPtr(int(*hc.FailureThreshold))
	}
	if hc.SuccessThreshold != nil {
		args.SuccessThreshold = pulumi.IntPtr(int(*hc.SuccessThreshold))
	}

	return args
}

func customRuntimeConfigArgs(cfg *alicloudfunctionv1.AlicloudFunctionCustomRuntimeConfig) fc.V3FunctionCustomRuntimeConfigPtrInput {
	rtArgs := &fc.V3FunctionCustomRuntimeConfigArgs{
		Commands: pulumi.ToStringArray(cfg.Command),
		Args:     pulumi.ToStringArray(cfg.Args),
	}

	if cfg.Port != nil {
		rtArgs.Port = pulumi.IntPtr(int(*cfg.Port))
	}

	if cfg.HealthCheckConfig != nil {
		rtArgs.HealthCheckConfig = runtimeHealthCheckArgs(cfg.HealthCheckConfig)
	}

	return rtArgs
}

func runtimeHealthCheckArgs(hc *alicloudfunctionv1.AlicloudFunctionHealthCheckConfig) fc.V3FunctionCustomRuntimeConfigHealthCheckConfigPtrInput {
	args := &fc.V3FunctionCustomRuntimeConfigHealthCheckConfigArgs{
		HttpGetUrl: optionalString(hc.HttpGetUrl),
	}

	if hc.InitialDelaySeconds != nil {
		args.InitialDelaySeconds = pulumi.IntPtr(int(*hc.InitialDelaySeconds))
	}
	if hc.TimeoutSeconds != nil {
		args.TimeoutSeconds = pulumi.IntPtr(int(*hc.TimeoutSeconds))
	}
	if hc.PeriodSeconds != nil {
		args.PeriodSeconds = pulumi.IntPtr(int(*hc.PeriodSeconds))
	}
	if hc.FailureThreshold != nil {
		args.FailureThreshold = pulumi.IntPtr(int(*hc.FailureThreshold))
	}
	if hc.SuccessThreshold != nil {
		args.SuccessThreshold = pulumi.IntPtr(int(*hc.SuccessThreshold))
	}

	return args
}

func instanceLifecycleConfigArgs(cfg *alicloudfunctionv1.AlicloudFunctionInstanceLifecycleConfig) fc.V3FunctionInstanceLifecycleConfigPtrInput {
	lcArgs := &fc.V3FunctionInstanceLifecycleConfigArgs{}

	if cfg.Initializer != nil {
		initArgs := &fc.V3FunctionInstanceLifecycleConfigInitializerArgs{
			Handler:  optionalString(cfg.Initializer.Handler),
			Commands: pulumi.ToStringArray(cfg.Initializer.Command),
		}
		if cfg.Initializer.Timeout != nil {
			initArgs.Timeout = pulumi.IntPtr(int(*cfg.Initializer.Timeout))
		}
		lcArgs.Initializer = initArgs
	}

	if cfg.PreStop != nil {
		preStopArgs := &fc.V3FunctionInstanceLifecycleConfigPreStopArgs{
			Handler: optionalString(cfg.PreStop.Handler),
		}
		if cfg.PreStop.Timeout != nil {
			preStopArgs.Timeout = pulumi.IntPtr(int(*cfg.PreStop.Timeout))
		}
		lcArgs.PreStop = preStopArgs
	}

	return lcArgs
}

func nasConfigArgs(cfg *alicloudfunctionv1.AlicloudFunctionNasConfig) fc.V3FunctionNasConfigPtrInput {
	nasArgs := &fc.V3FunctionNasConfigArgs{}

	if cfg.UserId != nil {
		nasArgs.UserId = pulumi.IntPtr(int(*cfg.UserId))
	}

	if cfg.GroupId != nil {
		nasArgs.GroupId = pulumi.IntPtr(int(*cfg.GroupId))
	}

	if len(cfg.MountPoints) > 0 {
		mountPoints := make(fc.V3FunctionNasConfigMountPointArray, 0, len(cfg.MountPoints))
		for _, mp := range cfg.MountPoints {
			mpArgs := fc.V3FunctionNasConfigMountPointArgs{
				ServerAddr: optionalString(mp.ServerAddr),
				MountDir:   optionalString(mp.MountDir),
			}
			if mp.EnableTls != nil {
				mpArgs.EnableTls = pulumi.BoolPtr(*mp.EnableTls)
			}
			mountPoints = append(mountPoints, mpArgs)
		}
		nasArgs.MountPoints = mountPoints
	}

	return nasArgs
}

func gpuConfigArgs(cfg *alicloudfunctionv1.AlicloudFunctionGpuConfig) fc.V3FunctionGpuConfigPtrInput {
	return &fc.V3FunctionGpuConfigArgs{
		GpuMemorySize: pulumi.IntPtr(int(cfg.GpuMemorySize)),
		GpuType:       optionalString(cfg.GpuType),
	}
}
