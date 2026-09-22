package module

import (
	"github.com/pkg/errors"
	gcpvertexaiagentenginev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaiagentengine/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/vertex"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// agentEngine creates the Vertex AI Agent Engine instance (Google's
// ReasoningEngine): the agent's code source, identity, deployment shape,
// and optional Memory Bank. Location and the encryption key are immutable;
// everything else updates in place, and a new source archive redeploys the
// agent's code.
//
// Send posture (parity with the Terraform module): optional strings,
// booleans, and Optional+Computed numbers are sent only when set so
// Google's defaults stay in charge and a manifest that never mentions them
// re-plans clean on either engine.
func agentEngine(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpVertexAiAgentEngine.Spec

	// Enable the Vertex AI API first so a fresh project works on the first
	// deploy. DisableOnDestroy stays false: tearing down one agent must
	// never disable the API for everything else in the project.
	aiplatformApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("aiplatform.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		aiplatformApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdAiplatformApi, err := projects.NewService(ctx,
		"gcpagent-aiplatform.googleapis.com", aiplatformApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable aiplatform.googleapis.com api")
	}

	args := &vertex.AiReasoningEngineArgs{
		DisplayName: pulumi.String(locals.DisplayName),
		Region:      pulumi.String(spec.Location),
		Labels:      pulumi.ToStringMap(locals.GcpLabels),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}
	if spec.KmsKeyName.GetValue() != "" {
		args.EncryptionSpec = &vertex.AiReasoningEngineEncryptionSpecArgs{
			KmsKeyName: pulumi.String(spec.KmsKeyName.GetValue()),
		}
	}
	if spec.Spec != nil {
		args.Spec = buildSpec(spec.Spec)
	}
	if spec.ContextSpec != nil && spec.ContextSpec.MemoryBankConfig != nil {
		args.ContextSpec = &vertex.AiReasoningEngineContextSpecArgs{
			MemoryBankConfig: buildMemoryBank(spec.ContextSpec.MemoryBankConfig),
		}
	}

	// Engine-side destroy stance: PREVENT fails destroys, ABANDON leaves the
	// agent running. Sent only when set so the provider default stays in
	// charge otherwise.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	createdEngine, err := vertex.NewAiReasoningEngine(ctx,
		locals.GcpVertexAiAgentEngine.Metadata.Name, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdAiplatformApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create agent engine")
	}

	// Google's `name` is the numeric ID; the full resource path is what the
	// SDK and a :query call take, so both are exported.
	ctx.Export(OpName, createdEngine.ID().ToStringOutput())
	ctx.Export(OpReasoningEngineId, createdEngine.Name)
	ctx.Export(OpLocation, createdEngine.Region)
	ctx.Export(OpCreateTime, createdEngine.CreateTime)
	ctx.Export(OpUpdateTime, createdEngine.UpdateTime)
	return nil
}

// buildSpec maps the agent's code source, identity, and deployment shape.
// build_spec.service_account is not mapped: the pinned SDK lacks it, and
// an argument one engine cannot send is never a one-engine field.
func buildSpec(s *gcpvertexaiagentenginev1alpha1.GcpVertexAiAgentEngineSpecConfig) *vertex.AiReasoningEngineSpecArgs {
	args := &vertex.AiReasoningEngineSpecArgs{}
	if s.AgentFramework != "" {
		args.AgentFramework = pulumi.String(s.AgentFramework)
	}
	if s.ClassMethods != "" {
		args.ClassMethods = pulumi.String(s.ClassMethods)
	}
	if s.IdentityType != "" {
		args.IdentityType = pulumi.String(s.IdentityType)
	}
	if s.ServiceAccount.GetValue() != "" {
		args.ServiceAccount = pulumi.String(s.ServiceAccount.GetValue())
	}
	if c := s.ContainerSpec; c != nil {
		container := &vertex.AiReasoningEngineSpecContainerSpecArgs{
			ImageUri: pulumi.String(c.ImageUri),
		}
		if c.Port != nil {
			container.Port = pulumi.Int(int(c.GetPort()))
		}
		args.ContainerSpec = container
	}
	if src := s.SourceCodeSpec; src != nil {
		args.SourceCodeSpec = buildSourceCodeSpec(src)
	}
	if p := s.PackageSpec; p != nil {
		pkg := &vertex.AiReasoningEngineSpecPackageSpecArgs{}
		if p.PickleObjectGcsUri != "" {
			pkg.PickleObjectGcsUri = pulumi.String(p.PickleObjectGcsUri)
		}
		if p.DependencyFilesGcsUri != "" {
			pkg.DependencyFilesGcsUri = pulumi.String(p.DependencyFilesGcsUri)
		}
		if p.RequirementsGcsUri != "" {
			pkg.RequirementsGcsUri = pulumi.String(p.RequirementsGcsUri)
		}
		if p.PythonVersion != "" {
			pkg.PythonVersion = pulumi.String(p.PythonVersion)
		}
		args.PackageSpec = pkg
	}
	if b := s.BuildSpec; b != nil && b.WorkerPool != "" {
		args.BuildSpec = &vertex.AiReasoningEngineSpecBuildSpecArgs{
			WorkerPool: pulumi.String(b.WorkerPool),
		}
	}
	if d := s.DeploymentSpec; d != nil {
		args.DeploymentSpec = buildDeploymentSpec(d)
	}
	return args
}

func buildSourceCodeSpec(src *gcpvertexaiagentenginev1alpha1.GcpVertexAiAgentEngineSourceCodeSpec) *vertex.AiReasoningEngineSpecSourceCodeSpecArgs {
	args := &vertex.AiReasoningEngineSpecSourceCodeSpecArgs{}
	if src.InlineSource != nil {
		args.InlineSource = &vertex.AiReasoningEngineSpecSourceCodeSpecInlineSourceArgs{
			SourceArchive: pulumi.String(src.InlineSource.SourceArchive),
		}
	}
	if dc := src.DeveloperConnectSource; dc != nil && dc.Config != nil {
		args.DeveloperConnectSource = &vertex.AiReasoningEngineSpecSourceCodeSpecDeveloperConnectSourceArgs{
			Config: &vertex.AiReasoningEngineSpecSourceCodeSpecDeveloperConnectSourceConfigArgs{
				GitRepositoryLink: pulumi.String(dc.Config.GitRepositoryLink),
				Dir:               pulumi.String(dc.Config.Dir),
				Revision:          pulumi.String(dc.Config.Revision),
			},
		}
	}
	if ac := src.AgentConfigSource; ac != nil {
		agentConfig := &vertex.AiReasoningEngineSpecSourceCodeSpecAgentConfigSourceArgs{}
		if ac.AdkConfig != nil {
			agentConfig.AdkConfig = &vertex.AiReasoningEngineSpecSourceCodeSpecAgentConfigSourceAdkConfigArgs{
				JsonConfig: pulumi.String(ac.AdkConfig.JsonConfig),
			}
		}
		if ac.InlineSource != nil {
			agentConfig.InlineSource = &vertex.AiReasoningEngineSpecSourceCodeSpecAgentConfigSourceInlineSourceArgs{
				SourceArchive: pulumi.String(ac.InlineSource.SourceArchive),
			}
		}
		args.AgentConfigSource = agentConfig
	}
	if py := src.PythonSpec; py != nil {
		python := &vertex.AiReasoningEngineSpecSourceCodeSpecPythonSpecArgs{}
		if py.Version != "" {
			python.Version = pulumi.String(py.Version)
		}
		if py.EntrypointModule != "" {
			python.EntrypointModule = pulumi.String(py.EntrypointModule)
		}
		if py.EntrypointObject != "" {
			python.EntrypointObject = pulumi.String(py.EntrypointObject)
		}
		if py.RequirementsFile != "" {
			python.RequirementsFile = pulumi.String(py.RequirementsFile)
		}
		args.PythonSpec = python
	}
	if img := src.ImageSpec; img != nil {
		image := &vertex.AiReasoningEngineSpecSourceCodeSpecImageSpecArgs{}
		if len(img.BuildArgs) > 0 {
			image.BuildArgs = pulumi.ToStringMap(img.BuildArgs)
		}
		args.ImageSpec = image
	}
	return args
}

func buildDeploymentSpec(d *gcpvertexaiagentenginev1alpha1.GcpVertexAiAgentEngineDeploymentSpec) *vertex.AiReasoningEngineSpecDeploymentSpecArgs {
	args := &vertex.AiReasoningEngineSpecDeploymentSpecArgs{}
	if len(d.Env) > 0 {
		envs := vertex.AiReasoningEngineSpecDeploymentSpecEnvArray{}
		for _, e := range d.Env {
			envs = append(envs, &vertex.AiReasoningEngineSpecDeploymentSpecEnvArgs{
				Name:  pulumi.String(e.Name),
				Value: pulumi.String(e.Value),
			})
		}
		args.Envs = envs
	}
	if len(d.SecretEnv) > 0 {
		secrets := vertex.AiReasoningEngineSpecDeploymentSpecSecretEnvArray{}
		for _, s := range d.SecretEnv {
			ref := &vertex.AiReasoningEngineSpecDeploymentSpecSecretEnvSecretRefArgs{
				Secret: pulumi.String(s.SecretRef.Secret.GetValue()),
			}
			if s.SecretRef.Version != "" {
				ref.Version = pulumi.String(s.SecretRef.Version)
			}
			secrets = append(secrets, &vertex.AiReasoningEngineSpecDeploymentSpecSecretEnvArgs{
				Name:      pulumi.String(s.Name),
				SecretRef: ref,
			})
		}
		args.SecretEnvs = secrets
	}
	// Optional+Computed on Google's side: sent only when set.
	if d.MinInstances != nil {
		args.MinInstances = pulumi.Int(int(d.GetMinInstances()))
	}
	if d.MaxInstances != nil {
		args.MaxInstances = pulumi.Int(int(d.GetMaxInstances()))
	}
	if d.ContainerConcurrency != nil {
		args.ContainerConcurrency = pulumi.Int(int(d.GetContainerConcurrency()))
	}
	if len(d.ResourceLimits) > 0 {
		args.ResourceLimits = pulumi.ToStringMap(d.ResourceLimits)
	}
	if psc := d.PscInterfaceConfig; psc != nil {
		pscArgs := &vertex.AiReasoningEngineSpecDeploymentSpecPscInterfaceConfigArgs{}
		if psc.NetworkAttachment != "" {
			pscArgs.NetworkAttachment = pulumi.String(psc.NetworkAttachment)
		}
		if len(psc.DnsPeeringConfigs) > 0 {
			peerings := vertex.AiReasoningEngineSpecDeploymentSpecPscInterfaceConfigDnsPeeringConfigArray{}
			for _, p := range psc.DnsPeeringConfigs {
				peerings = append(peerings, &vertex.AiReasoningEngineSpecDeploymentSpecPscInterfaceConfigDnsPeeringConfigArgs{
					Domain:        pulumi.String(p.Domain),
					TargetProject: pulumi.String(p.TargetProject.GetValue()),
					TargetNetwork: pulumi.String(p.TargetNetwork.GetValue()),
				})
			}
			pscArgs.DnsPeeringConfigs = peerings
		}
		args.PscInterfaceConfig = pscArgs
	}
	if gw := d.AgentGatewayConfig; gw != nil {
		gwArgs := &vertex.AiReasoningEngineSpecDeploymentSpecAgentGatewayConfigArgs{}
		if gw.ClientToAgentConfig != nil {
			gwArgs.ClientToAgentConfig = &vertex.AiReasoningEngineSpecDeploymentSpecAgentGatewayConfigClientToAgentConfigArgs{
				AgentGateway: pulumi.String(gw.ClientToAgentConfig.AgentGateway),
			}
		}
		if gw.AgentToAnywhereConfig != nil {
			gwArgs.AgentToAnywhereConfig = &vertex.AiReasoningEngineSpecDeploymentSpecAgentGatewayConfigAgentToAnywhereConfigArgs{
				AgentGateway: pulumi.String(gw.AgentToAnywhereConfig.AgentGateway),
			}
		}
		args.AgentGatewayConfig = gwArgs
	}
	return args
}
