package module

import (
	"github.com/pkg/errors"
	alicloudsaeapplicationv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudsaeapplication/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/sae"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudsaeapplicationv1.AlicloudSaeApplicationStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AlicloudSaeApplication.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	args := &sae.ApplicationArgs{
		AppName:     pulumi.String(spec.AppName),
		PackageType: pulumi.String(spec.PackageType),
		Replicas:    pulumi.Int(int(spec.Replicas)),
		Cpu:         pulumi.IntPtr(int(spec.Cpu)),
		Memory:      pulumi.IntPtr(int(spec.Memory)),
		Tags:        pulumi.ToStringMap(locals.Tags),
	}

	if spec.AppDescription != "" {
		args.AppDescription = pulumi.String(spec.AppDescription)
	}

	if spec.VpcId != nil {
		args.VpcId = pulumi.StringPtr(spec.VpcId.GetValue())
	}

	if spec.VswitchId != nil {
		args.VswitchId = pulumi.StringPtr(spec.VswitchId.GetValue())
	}

	if spec.SecurityGroupId != nil {
		args.SecurityGroupId = pulumi.StringPtr(spec.SecurityGroupId.GetValue())
	}

	if spec.NamespaceId != "" {
		args.NamespaceId = pulumi.String(spec.NamespaceId)
	}

	if spec.ImageUrl != "" {
		args.ImageUrl = pulumi.String(spec.ImageUrl)
	}

	if spec.PackageUrl != "" {
		args.PackageUrl = pulumi.String(spec.PackageUrl)
	}

	if spec.PackageVersion != "" {
		args.PackageVersion = pulumi.String(spec.PackageVersion)
	}

	if spec.Command != "" {
		args.Command = pulumi.String(spec.Command)
	}

	if len(spec.CommandArgs) > 0 {
		args.CommandArgsV2s = pulumi.ToStringArray(spec.CommandArgs)
	}

	if envsJSON := envsToJSON(spec.Envs); envsJSON != "" {
		args.Envs = pulumi.String(envsJSON)
	}

	if spec.Jdk != "" {
		args.Jdk = pulumi.String(spec.Jdk)
	}

	if spec.JarStartOptions != "" {
		args.JarStartOptions = pulumi.String(spec.JarStartOptions)
	}

	if spec.JarStartArgs != "" {
		args.JarStartArgs = pulumi.String(spec.JarStartArgs)
	}

	args.ProgrammingLanguage = optionalStringPtr(spec.ProgrammingLanguage)

	if spec.Timezone != "" {
		args.Timezone = pulumi.String(spec.Timezone)
	}

	args.TerminationGracePeriodSeconds = optionalInt(spec.TerminationGracePeriodSeconds)
	args.MinReadyInstances = optionalInt(spec.MinReadyInstances)

	if spec.AcrInstanceId != "" {
		args.AcrInstanceId = pulumi.String(spec.AcrInstanceId)
	}

	if spec.Liveness != nil {
		args.LivenessV2 = healthCheckArgs(spec.Liveness)
	}

	if spec.Readiness != nil {
		args.ReadinessV2 = readinessCheckArgs(spec.Readiness)
	}

	if len(spec.CustomHostAliases) > 0 {
		args.CustomHostAliasV2s = customHostAliasArgs(spec.CustomHostAliases)
	}

	if spec.UpdateStrategy != nil {
		args.UpdateStrategyV2 = updateStrategyArgs(spec.UpdateStrategy)
	}

	if spec.SlsConfigs != "" {
		args.SlsConfigs = pulumi.String(spec.SlsConfigs)
	}

	app, err := sae.NewApplication(ctx, spec.AppName, args,
		pulumi.Provider(alicloudProvider),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create SAE application %s", spec.AppName)
	}

	ctx.Export(OpAppId, app.ID())
	ctx.Export(OpAppName, pulumi.String(spec.AppName))

	return nil
}

func healthCheckArgs(hc *alicloudsaeapplicationv1.AlicloudSaeApplicationHealthCheck) sae.ApplicationLivenessV2PtrInput {
	args := &sae.ApplicationLivenessV2Args{}

	if hc.HttpGet != nil {
		args.HttpGet = &sae.ApplicationLivenessV2HttpGetArgs{
			Path: optionalString(hc.HttpGet.Path),
			Port: pulumi.IntPtr(int(hc.HttpGet.Port)),
		}
	}

	if hc.TcpSocket != nil {
		args.TcpSocket = &sae.ApplicationLivenessV2TcpSocketArgs{
			Port: pulumi.IntPtr(int(hc.TcpSocket.Port)),
		}
	}

	if hc.Exec != nil {
		args.Exec = &sae.ApplicationLivenessV2ExecArgs{
			Commands: pulumi.ToStringArray([]string{hc.Exec.Command}),
		}
	}

	if hc.InitialDelaySeconds != nil {
		args.InitialDelaySeconds = pulumi.IntPtr(int(*hc.InitialDelaySeconds))
	}
	if hc.PeriodSeconds != nil {
		args.PeriodSeconds = pulumi.IntPtr(int(*hc.PeriodSeconds))
	}
	if hc.TimeoutSeconds != nil {
		args.TimeoutSeconds = pulumi.IntPtr(int(*hc.TimeoutSeconds))
	}
	if hc.FailureThreshold != nil {
		args.FailureThreshold = pulumi.IntPtr(int(*hc.FailureThreshold))
	}

	return args
}

func readinessCheckArgs(hc *alicloudsaeapplicationv1.AlicloudSaeApplicationHealthCheck) sae.ApplicationReadinessV2PtrInput {
	args := &sae.ApplicationReadinessV2Args{}

	if hc.HttpGet != nil {
		args.HttpGet = &sae.ApplicationReadinessV2HttpGetArgs{
			Path: optionalString(hc.HttpGet.Path),
			Port: pulumi.IntPtr(int(hc.HttpGet.Port)),
		}
	}

	if hc.TcpSocket != nil {
		args.TcpSocket = &sae.ApplicationReadinessV2TcpSocketArgs{
			Port: pulumi.IntPtr(int(hc.TcpSocket.Port)),
		}
	}

	if hc.Exec != nil {
		args.Exec = &sae.ApplicationReadinessV2ExecArgs{
			Commands: pulumi.ToStringArray([]string{hc.Exec.Command}),
		}
	}

	if hc.InitialDelaySeconds != nil {
		args.InitialDelaySeconds = pulumi.IntPtr(int(*hc.InitialDelaySeconds))
	}
	if hc.PeriodSeconds != nil {
		args.PeriodSeconds = pulumi.IntPtr(int(*hc.PeriodSeconds))
	}
	if hc.TimeoutSeconds != nil {
		args.TimeoutSeconds = pulumi.IntPtr(int(*hc.TimeoutSeconds))
	}
	if hc.FailureThreshold != nil {
		args.FailureThreshold = pulumi.IntPtr(int(*hc.FailureThreshold))
	}

	return args
}

func customHostAliasArgs(aliases []*alicloudsaeapplicationv1.AlicloudSaeApplicationCustomHostAlias) sae.ApplicationCustomHostAliasV2ArrayInput {
	result := sae.ApplicationCustomHostAliasV2Array{}
	for _, alias := range aliases {
		result = append(result, &sae.ApplicationCustomHostAliasV2Args{
			HostName: optionalString(alias.HostName),
			Ip:       optionalString(alias.Ip),
		})
	}
	return result
}

func updateStrategyArgs(strategy *alicloudsaeapplicationv1.AlicloudSaeApplicationUpdateStrategy) sae.ApplicationUpdateStrategyV2PtrInput {
	args := &sae.ApplicationUpdateStrategyV2Args{
		Type: optionalStringPtr(strategy.Type),
	}

	if strategy.BatchUpdate != nil {
		batchArgs := &sae.ApplicationUpdateStrategyV2BatchUpdateArgs{
			Batch:         optionalInt(strategy.BatchUpdate.Batch),
			BatchWaitTime: optionalInt(strategy.BatchUpdate.BatchWaitTime),
			ReleaseType:   optionalStringPtr(strategy.BatchUpdate.ReleaseType),
		}
		args.BatchUpdate = batchArgs
	}

	return args
}
