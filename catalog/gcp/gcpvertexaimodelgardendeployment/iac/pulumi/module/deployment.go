package module

import (
	"strconv"

	"github.com/pkg/errors"
	gcpvertexaimodelgardendeploymentv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaimodelgardendeployment/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/vertex"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// deployment deploys a Model Garden or Hugging Face model to a Vertex AI
// endpoint in one step. Every argument is immutable on Google's side, so
// any change replaces the whole deployment (undeploy, delete the endpoint,
// redeploy); deletion_policy is the only lever that updates in place.
//
// Send posture (parity with the Terraform module): optional strings and
// booleans are sent only when set; Optional+Computed replica counts are
// sent only when set so Model Garden's per-model defaults stay in charge
// and a manifest that omits deploy_config re-plans clean on either engine.
func deployment(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpVertexAiModelGardenDeployment.Spec

	// Enable the Vertex AI API first so a fresh project works on the first
	// deploy. DisableOnDestroy stays false: tearing down one deployment
	// must never disable the API for everything else in the project.
	aiplatformApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("aiplatform.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		aiplatformApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdAiplatformApi, err := projects.NewService(ctx,
		"gcpmgdep-aiplatform.googleapis.com", aiplatformApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable aiplatform.googleapis.com api")
	}

	args := &vertex.AiEndpointWithModelGardenDeploymentArgs{
		Location: pulumi.String(spec.Location),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}

	// Exactly one model source (proto-enforced).
	if spec.PublisherModelName != "" {
		args.PublisherModelName = pulumi.String(spec.PublisherModelName)
	}
	if spec.HuggingFaceModelId != "" {
		args.HuggingFaceModelId = pulumi.String(spec.HuggingFaceModelId)
	}

	if spec.ModelConfig != nil {
		args.ModelConfig = buildModelConfig(spec.ModelConfig)
	}
	if spec.DeployConfig != nil {
		args.DeployConfig = buildDeployConfig(spec.DeployConfig)
	}
	if spec.EndpointConfig != nil {
		args.EndpointConfig = buildEndpointConfig(spec.EndpointConfig)
	}

	// Engine-side destroy stance: PREVENT fails destroys, ABANDON leaves the
	// model serving. Sent only when set so the provider default stays in
	// charge otherwise.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	createdDeployment, err := vertex.NewAiEndpointWithModelGardenDeployment(ctx,
		locals.GcpVertexAiModelGardenDeployment.Metadata.Name, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdAiplatformApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create model garden deployment")
	}

	// The provider's `endpoint` is the endpoint's numeric ID segment; the
	// full resource path is rebuilt the way Google names it so the output
	// reads the same as GcpVertexAiEndpoint's endpoint_id.
	ctx.Export(OpEndpointId, pulumi.All(createdDeployment.Project, createdDeployment.Location, createdDeployment.Endpoint).ApplyT(
		func(parts []interface{}) string {
			return "projects/" + parts[0].(string) + "/locations/" + parts[1].(string) + "/endpoints/" + parts[2].(string)
		}).(pulumi.StringOutput))
	ctx.Export(OpEndpointName, createdDeployment.Endpoint)
	ctx.Export(OpDeployedModelId, createdDeployment.DeployedModelId)
	ctx.Export(OpDeployedModelDisplayName, createdDeployment.DeployedModelDisplayName)
	ctx.Export(OpLocation, createdDeployment.Location)
	return nil
}

func buildModelConfig(cfg *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentModelConfig) *vertex.AiEndpointWithModelGardenDeploymentModelConfigArgs {
	args := &vertex.AiEndpointWithModelGardenDeploymentModelConfigArgs{}
	if cfg.AcceptEula {
		args.AcceptEula = pulumi.Bool(true)
	}
	if cfg.HuggingFaceAccessToken != "" {
		args.HuggingFaceAccessToken = pulumi.String(cfg.HuggingFaceAccessToken)
	}
	if cfg.HuggingFaceCacheEnabled {
		args.HuggingFaceCacheEnabled = pulumi.Bool(true)
	}
	if cfg.ModelDisplayName != "" {
		args.ModelDisplayName = pulumi.String(cfg.ModelDisplayName)
	}
	if cfg.ContainerSpec != nil {
		args.ContainerSpec = buildContainerSpec(cfg.ContainerSpec)
	}
	return args
}

func buildContainerSpec(c *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentContainerSpec) *vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecArgs {
	args := &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecArgs{
		ImageUri: pulumi.String(c.ImageUri),
	}
	if len(c.Command) > 0 {
		args.Commands = pulumi.ToStringArray(c.Command)
	}
	if len(c.Args) > 0 {
		args.Args = pulumi.ToStringArray(c.Args)
	}
	if len(c.Env) > 0 {
		envs := vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecEnvArray{}
		for _, e := range c.Env {
			envs = append(envs, &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecEnvArgs{
				Name:  pulumi.String(e.Name),
				Value: pulumi.String(e.Value),
			})
		}
		args.Envs = envs
	}
	if len(c.Ports) > 0 {
		ports := vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecPortArray{}
		for _, p := range c.Ports {
			portArgs := &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecPortArgs{}
			if p.ContainerPort != nil {
				portArgs.ContainerPort = pulumi.Int(int(p.GetContainerPort()))
			}
			ports = append(ports, portArgs)
		}
		args.Ports = ports
	}
	if len(c.GrpcPorts) > 0 {
		ports := vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecGrpcPortArray{}
		for _, p := range c.GrpcPorts {
			portArgs := &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecGrpcPortArgs{}
			if p.ContainerPort != nil {
				portArgs.ContainerPort = pulumi.Int(int(p.GetContainerPort()))
			}
			ports = append(ports, portArgs)
		}
		args.GrpcPorts = ports
	}
	if c.PredictRoute != "" {
		args.PredictRoute = pulumi.String(c.PredictRoute)
	}
	if c.HealthRoute != "" {
		args.HealthRoute = pulumi.String(c.HealthRoute)
	}
	if c.DeploymentTimeout != "" {
		args.DeploymentTimeout = pulumi.String(c.DeploymentTimeout)
	}
	if c.SharedMemorySizeMb != nil {
		args.SharedMemorySizeMb = pulumi.String(formatInt64(c.GetSharedMemorySizeMb()))
	}

	// The three probes share one spec shape but the SDK types each one
	// separately, so each is mapped through its own small adapter below.
	if c.StartupProbe != nil {
		args.StartupProbe = buildStartupProbe(c.StartupProbe)
	}
	if c.LivenessProbe != nil {
		args.LivenessProbe = buildLivenessProbe(c.LivenessProbe)
	}
	if c.HealthProbe != nil {
		args.HealthProbe = buildHealthProbe(c.HealthProbe)
	}
	return args
}

// probeTimings returns the five timing fields as pulumi inputs, nil where
// the spec leaves them unset so Google's defaults stay in charge.
func probeTimings(p *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentProbe) (initialDelay, timeout, period, success, failure pulumi.IntPtrInput) {
	if p.InitialDelaySeconds != nil {
		initialDelay = pulumi.Int(int(p.GetInitialDelaySeconds()))
	}
	if p.TimeoutSeconds != nil {
		timeout = pulumi.Int(int(p.GetTimeoutSeconds()))
	}
	if p.PeriodSeconds != nil {
		period = pulumi.Int(int(p.GetPeriodSeconds()))
	}
	if p.SuccessThreshold != nil {
		success = pulumi.Int(int(p.GetSuccessThreshold()))
	}
	if p.FailureThreshold != nil {
		failure = pulumi.Int(int(p.GetFailureThreshold()))
	}
	return
}

func buildStartupProbe(p *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentProbe) *vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecStartupProbeArgs {
	args := &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecStartupProbeArgs{}
	args.InitialDelaySeconds, args.TimeoutSeconds, args.PeriodSeconds, args.SuccessThreshold, args.FailureThreshold = probeTimings(p)
	switch h := p.Handler.(type) {
	case *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentProbe_Exec:
		args.Exec = &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecStartupProbeExecArgs{Commands: pulumi.ToStringArray(h.Exec.Command)}
	case *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentProbe_Grpc:
		grpc := &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecStartupProbeGrpcArgs{}
		if h.Grpc.Port != nil {
			grpc.Port = pulumi.Int(int(h.Grpc.GetPort()))
		}
		if h.Grpc.Service != "" {
			grpc.Service = pulumi.String(h.Grpc.Service)
		}
		args.Grpc = grpc
	case *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentProbe_HttpGet:
		httpGet := &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecStartupProbeHttpGetArgs{}
		if h.HttpGet.Path != "" {
			httpGet.Path = pulumi.String(h.HttpGet.Path)
		}
		if h.HttpGet.Port != nil {
			httpGet.Port = pulumi.Int(int(h.HttpGet.GetPort()))
		}
		if h.HttpGet.Host != "" {
			httpGet.Host = pulumi.String(h.HttpGet.Host)
		}
		if h.HttpGet.Scheme != "" {
			httpGet.Scheme = pulumi.String(h.HttpGet.Scheme)
		}
		if len(h.HttpGet.HttpHeaders) > 0 {
			headers := vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecStartupProbeHttpGetHttpHeaderArray{}
			for _, hdr := range h.HttpGet.HttpHeaders {
				headers = append(headers, &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecStartupProbeHttpGetHttpHeaderArgs{
					Name:  pulumi.String(hdr.Name),
					Value: pulumi.String(hdr.Value),
				})
			}
			httpGet.HttpHeaders = headers
		}
		args.HttpGet = httpGet
	case *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentProbe_TcpSocket:
		tcp := &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecStartupProbeTcpSocketArgs{}
		if h.TcpSocket.Port != nil {
			tcp.Port = pulumi.Int(int(h.TcpSocket.GetPort()))
		}
		if h.TcpSocket.Host != "" {
			tcp.Host = pulumi.String(h.TcpSocket.Host)
		}
		args.TcpSocket = tcp
	}
	return args
}

func buildLivenessProbe(p *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentProbe) *vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecLivenessProbeArgs {
	args := &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecLivenessProbeArgs{}
	args.InitialDelaySeconds, args.TimeoutSeconds, args.PeriodSeconds, args.SuccessThreshold, args.FailureThreshold = probeTimings(p)
	switch h := p.Handler.(type) {
	case *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentProbe_Exec:
		args.Exec = &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecLivenessProbeExecArgs{Commands: pulumi.ToStringArray(h.Exec.Command)}
	case *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentProbe_Grpc:
		grpc := &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecLivenessProbeGrpcArgs{}
		if h.Grpc.Port != nil {
			grpc.Port = pulumi.Int(int(h.Grpc.GetPort()))
		}
		if h.Grpc.Service != "" {
			grpc.Service = pulumi.String(h.Grpc.Service)
		}
		args.Grpc = grpc
	case *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentProbe_HttpGet:
		httpGet := &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecLivenessProbeHttpGetArgs{}
		if h.HttpGet.Path != "" {
			httpGet.Path = pulumi.String(h.HttpGet.Path)
		}
		if h.HttpGet.Port != nil {
			httpGet.Port = pulumi.Int(int(h.HttpGet.GetPort()))
		}
		if h.HttpGet.Host != "" {
			httpGet.Host = pulumi.String(h.HttpGet.Host)
		}
		if h.HttpGet.Scheme != "" {
			httpGet.Scheme = pulumi.String(h.HttpGet.Scheme)
		}
		if len(h.HttpGet.HttpHeaders) > 0 {
			headers := vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecLivenessProbeHttpGetHttpHeaderArray{}
			for _, hdr := range h.HttpGet.HttpHeaders {
				headers = append(headers, &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecLivenessProbeHttpGetHttpHeaderArgs{
					Name:  pulumi.String(hdr.Name),
					Value: pulumi.String(hdr.Value),
				})
			}
			httpGet.HttpHeaders = headers
		}
		args.HttpGet = httpGet
	case *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentProbe_TcpSocket:
		tcp := &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecLivenessProbeTcpSocketArgs{}
		if h.TcpSocket.Port != nil {
			tcp.Port = pulumi.Int(int(h.TcpSocket.GetPort()))
		}
		if h.TcpSocket.Host != "" {
			tcp.Host = pulumi.String(h.TcpSocket.Host)
		}
		args.TcpSocket = tcp
	}
	return args
}

func buildHealthProbe(p *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentProbe) *vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecHealthProbeArgs {
	args := &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecHealthProbeArgs{}
	args.InitialDelaySeconds, args.TimeoutSeconds, args.PeriodSeconds, args.SuccessThreshold, args.FailureThreshold = probeTimings(p)
	switch h := p.Handler.(type) {
	case *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentProbe_Exec:
		args.Exec = &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecHealthProbeExecArgs{Commands: pulumi.ToStringArray(h.Exec.Command)}
	case *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentProbe_Grpc:
		grpc := &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecHealthProbeGrpcArgs{}
		if h.Grpc.Port != nil {
			grpc.Port = pulumi.Int(int(h.Grpc.GetPort()))
		}
		if h.Grpc.Service != "" {
			grpc.Service = pulumi.String(h.Grpc.Service)
		}
		args.Grpc = grpc
	case *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentProbe_HttpGet:
		httpGet := &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecHealthProbeHttpGetArgs{}
		if h.HttpGet.Path != "" {
			httpGet.Path = pulumi.String(h.HttpGet.Path)
		}
		if h.HttpGet.Port != nil {
			httpGet.Port = pulumi.Int(int(h.HttpGet.GetPort()))
		}
		if h.HttpGet.Host != "" {
			httpGet.Host = pulumi.String(h.HttpGet.Host)
		}
		if h.HttpGet.Scheme != "" {
			httpGet.Scheme = pulumi.String(h.HttpGet.Scheme)
		}
		if len(h.HttpGet.HttpHeaders) > 0 {
			headers := vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecHealthProbeHttpGetHttpHeaderArray{}
			for _, hdr := range h.HttpGet.HttpHeaders {
				headers = append(headers, &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecHealthProbeHttpGetHttpHeaderArgs{
					Name:  pulumi.String(hdr.Name),
					Value: pulumi.String(hdr.Value),
				})
			}
			httpGet.HttpHeaders = headers
		}
		args.HttpGet = httpGet
	case *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentProbe_TcpSocket:
		tcp := &vertex.AiEndpointWithModelGardenDeploymentModelConfigContainerSpecHealthProbeTcpSocketArgs{}
		if h.TcpSocket.Port != nil {
			tcp.Port = pulumi.Int(int(h.TcpSocket.GetPort()))
		}
		if h.TcpSocket.Host != "" {
			tcp.Host = pulumi.String(h.TcpSocket.Host)
		}
		args.TcpSocket = tcp
	}
	return args
}

func buildDeployConfig(cfg *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentDeployConfig) *vertex.AiEndpointWithModelGardenDeploymentDeployConfigArgs {
	args := &vertex.AiEndpointWithModelGardenDeploymentDeployConfigArgs{}
	if cfg.FastTryoutEnabled {
		args.FastTryoutEnabled = pulumi.Bool(true)
	}
	if len(cfg.SystemLabels) > 0 {
		args.SystemLabels = pulumi.ToStringMap(cfg.SystemLabels)
	}
	if dr := cfg.DedicatedResources; dr != nil {
		drArgs := &vertex.AiEndpointWithModelGardenDeploymentDeployConfigDedicatedResourcesArgs{
			MinReplicaCount: pulumi.Int(int(dr.MinReplicaCount)),
		}
		// Optional+Computed on Google's side: sent only when set so the
		// per-model defaults stay in charge.
		if dr.MaxReplicaCount != nil {
			drArgs.MaxReplicaCount = pulumi.Int(int(dr.GetMaxReplicaCount()))
		}
		if dr.RequiredReplicaCount != nil {
			drArgs.RequiredReplicaCount = pulumi.Int(int(dr.GetRequiredReplicaCount()))
		}
		if dr.Spot {
			drArgs.Spot = pulumi.Bool(true)
		}
		// machine_spec is required by the API even when every field is
		// left to Google; an empty block is the honest form.
		machine := &vertex.AiEndpointWithModelGardenDeploymentDeployConfigDedicatedResourcesMachineSpecArgs{}
		if ms := dr.MachineSpec; ms != nil {
			if ms.MachineType != "" {
				machine.MachineType = pulumi.String(ms.MachineType)
			}
			if ms.AcceleratorType != "" {
				machine.AcceleratorType = pulumi.String(ms.AcceleratorType)
			}
			if ms.AcceleratorCount != nil {
				machine.AcceleratorCount = pulumi.Int(int(ms.GetAcceleratorCount()))
			}
			if ms.TpuTopology != "" {
				machine.TpuTopology = pulumi.String(ms.TpuTopology)
			}
			if ms.MultihostGpuNodeCount != nil {
				machine.MultihostGpuNodeCount = pulumi.Int(int(ms.GetMultihostGpuNodeCount()))
			}
			if ra := ms.ReservationAffinity; ra != nil {
				raArgs := &vertex.AiEndpointWithModelGardenDeploymentDeployConfigDedicatedResourcesMachineSpecReservationAffinityArgs{
					ReservationAffinityType: pulumi.String(ra.ReservationAffinityType),
				}
				if ra.Key != "" {
					raArgs.Key = pulumi.String(ra.Key)
				}
				if len(ra.Values) > 0 {
					raArgs.Values = pulumi.ToStringArray(ra.Values)
				}
				machine.ReservationAffinity = raArgs
			}
		}
		drArgs.MachineSpec = machine
		if len(dr.AutoscalingMetricSpecs) > 0 {
			metrics := vertex.AiEndpointWithModelGardenDeploymentDeployConfigDedicatedResourcesAutoscalingMetricSpecArray{}
			for _, m := range dr.AutoscalingMetricSpecs {
				metricArgs := &vertex.AiEndpointWithModelGardenDeploymentDeployConfigDedicatedResourcesAutoscalingMetricSpecArgs{
					MetricName: pulumi.String(m.MetricName),
				}
				if m.Target != nil {
					metricArgs.Target = pulumi.Int(int(m.GetTarget()))
				}
				metrics = append(metrics, metricArgs)
			}
			drArgs.AutoscalingMetricSpecs = metrics
		}
		args.DedicatedResources = drArgs
	}
	return args
}

func buildEndpointConfig(cfg *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentEndpointConfig) *vertex.AiEndpointWithModelGardenDeploymentEndpointConfigArgs {
	args := &vertex.AiEndpointWithModelGardenDeploymentEndpointConfigArgs{}
	if cfg.EndpointDisplayName != "" {
		args.EndpointDisplayName = pulumi.String(cfg.EndpointDisplayName)
	}
	if cfg.DedicatedEndpointEnabled {
		args.DedicatedEndpointEnabled = pulumi.Bool(true)
	}
	if psc := cfg.PrivateServiceConnectConfig; psc != nil {
		pscArgs := &vertex.AiEndpointWithModelGardenDeploymentEndpointConfigPrivateServiceConnectConfigArgs{
			EnablePrivateServiceConnect: pulumi.Bool(psc.EnablePrivateServiceConnect),
		}
		if len(psc.ProjectAllowlist) > 0 {
			projects := []string{}
			for _, p := range psc.ProjectAllowlist {
				projects = append(projects, p.GetValue())
			}
			pscArgs.ProjectAllowlists = pulumi.ToStringArray(projects)
		}
		if auto := psc.PscAutomationConfig; auto != nil {
			pscArgs.PscAutomationConfigs = &vertex.AiEndpointWithModelGardenDeploymentEndpointConfigPrivateServiceConnectConfigPscAutomationConfigsArgs{
				ProjectId: pulumi.String(auto.ProjectId.GetValue()),
				Network:   pulumi.String(auto.Network.GetValue()),
			}
		}
		args.PrivateServiceConnectConfig = pscArgs
	}
	return args
}

// formatInt64 renders the spec's shared_memory_size_mb the way Google's
// API carries it -- an int64 as a decimal string -- so both engines send
// the same wire value.
func formatInt64(v int64) string {
	return strconv.FormatInt(v, 10)
}
