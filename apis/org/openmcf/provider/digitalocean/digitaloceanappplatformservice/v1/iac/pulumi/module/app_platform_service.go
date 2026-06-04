package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	digitaloceanappplatformservicev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/digitalocean/digitaloceanappplatformservice/v1"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// appPlatformService provisions a single‑service App and exports its key outputs.
func appPlatformService(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.App, error) {

	spec := locals.DigitalOceanAppPlatformService.Spec

	// Determine instance count (use fixed count or autoscaling)
	instanceCount := spec.InstanceCount
	if instanceCount == 0 {
		instanceCount = 1 // Default to 1 if not specified
	}

	// Convert instance size slug (replace underscores with hyphens)
	instanceSizeSlug := strings.ReplaceAll(spec.InstanceSizeSlug.String(), "_", "-")

	// --------------------------------------------
	// 1. Build service/worker/job based on service_type
	// --------------------------------------------
	appSpecArgs := &digitalocean.AppSpecArgs{
		Name:   pulumi.String(spec.ServiceName),
		Region: pulumi.String(spec.Region.String()),
	}

	switch spec.ServiceType {
	case digitaloceanappplatformservicev1.DigitalOceanAppPlatformServiceType_web_service:
		serviceArgs := buildWebService(spec, instanceCount, instanceSizeSlug)
		appSpecArgs.Services = digitalocean.AppSpecServiceArray{serviceArgs}

	case digitaloceanappplatformservicev1.DigitalOceanAppPlatformServiceType_worker:
		workerArgs := buildWorker(spec, instanceCount, instanceSizeSlug)
		appSpecArgs.Workers = digitalocean.AppSpecWorkerArray{workerArgs}

	case digitaloceanappplatformservicev1.DigitalOceanAppPlatformServiceType_job:
		jobArgs := buildJob(spec, instanceSizeSlug)
		appSpecArgs.Jobs = digitalocean.AppSpecJobArray{jobArgs}

	default:
		return nil, fmt.Errorf("invalid service_type: %v", spec.ServiceType)
	}

	// --------------------------------------------
	// 2. Add custom domain if specified
	// Note: Custom domains are configured via DigitalOcean console or API after app creation
	// --------------------------------------------
	// TODO: Add domain configuration support when Pulumi SDK supports it
	// if spec.CustomDomain != nil && spec.CustomDomain.GetValue() != "" {
	// 	appSpecArgs.Domains = ...
	// }

	// --------------------------------------------
	// 3. Create the App
	// --------------------------------------------
	createdApp, err := digitalocean.NewApp(
		ctx,
		"app",
		&digitalocean.AppArgs{
			Spec: appSpecArgs,
		},
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean app")
	}

	// --------------------------------------------
	// 4. Export stack outputs
	// --------------------------------------------
	ctx.Export(OpAppId, createdApp.ID())
	ctx.Export(OpLiveUrl, createdApp.LiveUrl)
	// default_hostname: the App's default ondigitalocean.app ingress, assigned by
	// the platform regardless of any custom domain. DefaultIngress is the SDK's
	// default-hostname attribute (distinct from LiveUrl, which can be a custom domain).
	ctx.Export(OpDefaultHostname, createdApp.DefaultIngress)

	return createdApp, nil
}

// buildWebService creates a web service configuration
func buildWebService(
	spec *digitaloceanappplatformservicev1.DigitalOceanAppPlatformServiceSpec,
	instanceCount uint32,
	instanceSizeSlug string,
) *digitalocean.AppSpecServiceArgs {
	serviceArgs := &digitalocean.AppSpecServiceArgs{
		Name:             pulumi.String(spec.ServiceName),
		InstanceCount:    pulumi.Int(int(instanceCount)),
		InstanceSizeSlug: pulumi.String(instanceSizeSlug),
	}

	// Configure source (git or image)
	configureSource(spec, serviceArgs)

	// Configure environment variables
	if len(spec.Env) > 0 {
		serviceArgs.Envs = buildEnvVars(spec.Env)
	}

	// Configure autoscaling if enabled
	if spec.EnableAutoscale {
		serviceArgs.Autoscaling = &digitalocean.AppSpecServiceAutoscalingArgs{
			MinInstanceCount: pulumi.Int(int(spec.MinInstanceCount)),
			MaxInstanceCount: pulumi.Int(int(spec.MaxInstanceCount)),
			Metrics: &digitalocean.AppSpecServiceAutoscalingMetricsArgs{
				Cpu: &digitalocean.AppSpecServiceAutoscalingMetricsCpuArgs{
					Percent: pulumi.Int(80), // Default CPU threshold
				},
			},
		}
	}

	return serviceArgs
}

// buildWorker creates a worker service configuration
func buildWorker(
	spec *digitaloceanappplatformservicev1.DigitalOceanAppPlatformServiceSpec,
	instanceCount uint32,
	instanceSizeSlug string,
) *digitalocean.AppSpecWorkerArgs {
	workerArgs := &digitalocean.AppSpecWorkerArgs{
		Name:             pulumi.String(spec.ServiceName),
		InstanceCount:    pulumi.Int(int(instanceCount)),
		InstanceSizeSlug: pulumi.String(instanceSizeSlug),
	}

	// Configure source (git or image) for worker
	if spec.GetGitSource() != nil {
		g := spec.GetGitSource()
		workerArgs.Git = &digitalocean.AppSpecWorkerGitArgs{
			RepoCloneUrl: pulumi.String(g.RepoUrl),
			Branch:       pulumi.String(g.Branch),
		}
		if g.RunCommand != "" {
			workerArgs.RunCommand = pulumi.String(g.RunCommand)
		}
	} else if spec.GetImageSource() != nil {
		i := spec.GetImageSource()
		workerArgs.Image = &digitalocean.AppSpecWorkerImageArgs{
			RegistryType: pulumi.String("DOCR"),
			Registry:     pulumi.String(i.Registry.GetValue()),
			Repository:   pulumi.String(i.Repository),
			Tag:          pulumi.String(i.Tag),
		}
	}

	// Configure environment variables
	if len(spec.Env) > 0 {
		workerArgs.Envs = buildWorkerEnvVars(spec.Env)
	}

	return workerArgs
}

// buildJob creates a job configuration
func buildJob(
	spec *digitaloceanappplatformservicev1.DigitalOceanAppPlatformServiceSpec,
	instanceSizeSlug string,
) *digitalocean.AppSpecJobArgs {
	jobArgs := &digitalocean.AppSpecJobArgs{
		Name:             pulumi.String(spec.ServiceName),
		InstanceSizeSlug: pulumi.String(instanceSizeSlug),
		Kind:             pulumi.String("PRE_DEPLOY"), // Default to pre-deploy job
	}

	// Configure source (git or image) for job
	if spec.GetGitSource() != nil {
		g := spec.GetGitSource()
		jobArgs.Git = &digitalocean.AppSpecJobGitArgs{
			RepoCloneUrl: pulumi.String(g.RepoUrl),
			Branch:       pulumi.String(g.Branch),
		}
		if g.RunCommand != "" {
			jobArgs.RunCommand = pulumi.String(g.RunCommand)
		}
	} else if spec.GetImageSource() != nil {
		i := spec.GetImageSource()
		jobArgs.Image = &digitalocean.AppSpecJobImageArgs{
			RegistryType: pulumi.String("DOCR"),
			Registry:     pulumi.String(i.Registry.GetValue()),
			Repository:   pulumi.String(i.Repository),
			Tag:          pulumi.String(i.Tag),
		}
	}

	// Configure environment variables
	if len(spec.Env) > 0 {
		jobArgs.Envs = buildJobEnvVars(spec.Env)
	}

	return jobArgs
}

// configureSource sets up git or image source for a web service
func configureSource(
	spec *digitaloceanappplatformservicev1.DigitalOceanAppPlatformServiceSpec,
	serviceArgs *digitalocean.AppSpecServiceArgs,
) {
	if spec.GetGitSource() != nil {
		g := spec.GetGitSource()
		serviceArgs.Git = &digitalocean.AppSpecServiceGitArgs{
			RepoCloneUrl: pulumi.String(g.RepoUrl),
			Branch:       pulumi.String(g.Branch),
		}
		if g.BuildCommand != "" {
			serviceArgs.BuildCommand = pulumi.String(g.BuildCommand)
		}
		if g.RunCommand != "" {
			serviceArgs.RunCommand = pulumi.String(g.RunCommand)
		}
	} else if spec.GetImageSource() != nil {
		i := spec.GetImageSource()
		serviceArgs.Image = &digitalocean.AppSpecServiceImageArgs{
			RegistryType: pulumi.String("DOCR"),
			Registry:     pulumi.String(i.Registry.GetValue()),
			Repository:   pulumi.String(i.Repository),
			Tag:          pulumi.String(i.Tag),
		}
	}
}

// buildEnvVars converts env map to Pulumi env array for services
func buildEnvVars(env map[string]string) digitalocean.AppSpecServiceEnvArray {
	var envArray digitalocean.AppSpecServiceEnvArray
	for k, v := range env {
		envArray = append(envArray, digitalocean.AppSpecServiceEnvArgs{
			Key:   pulumi.String(k),
			Value: pulumi.String(v),
			Scope: pulumi.String("RUN_AND_BUILD_TIME"),
		})
	}
	return envArray
}

// buildWorkerEnvVars converts env map to Pulumi env array for workers
func buildWorkerEnvVars(env map[string]string) digitalocean.AppSpecWorkerEnvArray {
	var envArray digitalocean.AppSpecWorkerEnvArray
	for k, v := range env {
		envArray = append(envArray, digitalocean.AppSpecWorkerEnvArgs{
			Key:   pulumi.String(k),
			Value: pulumi.String(v),
			Scope: pulumi.String("RUN_TIME"),
		})
	}
	return envArray
}

// buildJobEnvVars converts env map to Pulumi env array for jobs
func buildJobEnvVars(env map[string]string) digitalocean.AppSpecJobEnvArray {
	var envArray digitalocean.AppSpecJobEnvArray
	for k, v := range env {
		envArray = append(envArray, digitalocean.AppSpecJobEnvArgs{
			Key:   pulumi.String(k),
			Value: pulumi.String(v),
			Scope: pulumi.String("RUN_TIME"),
		})
	}
	return envArray
}
