package module

import (
	"github.com/pkg/errors"
	do "github.com/plantonhq/planton/catalog/digitalocean"
	digitaloceanappv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanapp/v1alpha1"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func app(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.App, error) {
	spec := locals.DigitalOceanApp.Spec

	// Two arms Terraform wires that the Pulumi bridge still cannot express at
	// pulumi-digitalocean v4.53.0 (re-verified against the SDK on disk: no
	// SecureHeader on AppSpecIngressRuleArgs, no LivenessHealthCheck on the
	// service or worker args). Failing loudly on a meaningful set keeps the two
	// engines honest with each other -- a silent drop would deploy a different
	// app than the manifest describes. Re-evaluate on every SDK pin bump; the
	// maintenance, vpc, ingress authority, and alert-destination arms that used
	// to sit in this list closed at v4.53.0 and are wired below.
	if spec.GetIngress() != nil && spec.GetIngress().GetSecureHeader() != nil {
		return nil, errors.New("PARITY-EXCEPTION: spec.ingress.secure_header is modeled and Terraform wires it; the Pulumi DigitalOcean SDK v4.53.0 has no secure_header on ingress rules. Re-evaluate when the SDK exposes ingress.secure_header.")
	}
	for _, s := range spec.GetServices() {
		if s.GetLivenessHealthCheck() != nil {
			return nil, errors.New("PARITY-EXCEPTION: service liveness_health_check is modeled and Terraform wires it; the Pulumi DigitalOcean SDK v4.53.0 has no liveness_health_check on services. Re-evaluate when the SDK exposes service.liveness_health_check.")
		}
	}
	for _, w := range spec.GetWorkers() {
		if w.GetLivenessHealthCheck() != nil {
			return nil, errors.New("PARITY-EXCEPTION: worker liveness_health_check is modeled and Terraform wires it; the Pulumi DigitalOcean SDK v4.53.0 has no liveness_health_check on workers. Re-evaluate when the SDK exposes worker.liveness_health_check.")
		}
	}

	appSpec := &digitalocean.AppSpecArgs{
		Name:                         pulumi.String(spec.GetAppName()),
		Region:                       pulumi.String(spec.GetRegion().String()),
		DisableEdgeCache:             pulumi.Bool(spec.GetDisableEdgeCache()),
		DisableEmailObfuscation:      pulumi.Bool(spec.GetDisableEmailObfuscation()),
		EnhancedThreatControlEnabled: pulumi.Bool(spec.GetEnhancedThreatControlEnabled()),
		Envs:                         appEnvs(spec.GetEnvs()),
	}
	if feats := spec.GetFeatures(); len(feats) > 0 {
		appSpec.Features = pulumi.ToStringArray(feats)
	}
	if m := spec.GetMaintenance(); m != nil {
		appSpec.Maintenance = &digitalocean.AppSpecMaintenanceArgs{
			Enabled:        pulumi.Bool(m.GetEnabled()),
			Archive:        pulumi.Bool(m.GetArchive()),
			OfflinePageUrl: strPtr(m.GetOfflinePageUrl()),
		}
	}
	// The SDK models VPC placement as a list; the API and Terraform accept
	// exactly one VPC per app, so the spec's single reference becomes a
	// one-element array here (the same shape Terraform's single vpc block
	// serializes to).
	if v := spec.GetVpc(); v != nil && v.GetValue() != "" {
		appSpec.Vpcs = digitalocean.AppSpecVpcArray{
			digitalocean.AppSpecVpcArgs{Id: pulumi.String(v.GetValue())},
		}
	}

	services, err := buildServices(spec.GetServices())
	if err != nil {
		return nil, err
	}
	appSpec.Services = services

	workers, err := buildWorkers(spec.GetWorkers())
	if err != nil {
		return nil, err
	}
	appSpec.Workers = workers

	jobs, err := buildJobs(spec.GetJobs())
	if err != nil {
		return nil, err
	}
	appSpec.Jobs = jobs

	sites, err := buildStaticSites(spec.GetStaticSites())
	if err != nil {
		return nil, err
	}
	appSpec.StaticSites = sites

	fns, err := buildFunctions(spec.GetFunctions())
	if err != nil {
		return nil, err
	}
	appSpec.Functions = fns

	appSpec.Databases = buildDatabases(spec.GetDatabases())
	appSpec.DomainNames = buildDomains(spec.GetDomains())
	appSpec.Alerts = buildAppAlerts(spec.GetAlerts())

	if spec.GetEgress() != do.DigitalOceanAppEgressType_digital_ocean_app_egress_type_unspecified {
		appSpec.Egresses = digitalocean.AppSpecEgressArray{
			digitalocean.AppSpecEgressArgs{Type: pulumi.String(providerEnum(spec.GetEgress().String()))},
		}
	}
	if spec.GetIngress() != nil {
		appSpec.Ingress = buildIngress(spec.GetIngress())
	}

	args := &digitalocean.AppArgs{Spec: appSpec}
	if spec.GetProjectId() != "" {
		args.ProjectId = pulumi.String(spec.GetProjectId())
	}

	created, err := digitalocean.NewApp(ctx, "app", args, pulumi.Provider(digitalOceanProvider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean app")
	}

	ctx.Export(OpAppId, created.ID())
	ctx.Export(OpLiveUrl, created.LiveUrl)
	ctx.Export(OpLiveDomain, created.LiveDomain)
	ctx.Export(OpActiveDeploymentId, created.ActiveDeploymentId)
	ctx.Export(OpDefaultHostname, created.DefaultIngress.ApplyT(func(u string) string {
		return stripScheme(u)
	}).(pulumi.StringOutput))

	return created, nil
}

func buildServices(in []*digitaloceanappv1alpha1.DigitalOceanAppService) (digitalocean.AppSpecServiceArray, error) {
	out := digitalocean.AppSpecServiceArray{}
	for _, s := range in {
		args := digitalocean.AppSpecServiceArgs{
			Name:             pulumi.String(s.GetName()),
			SourceDir:        strPtr(s.GetSourceDir()),
			EnvironmentSlug:  strPtr(s.GetEnvironmentSlug()),
			DockerfilePath:   strPtr(s.GetDockerfilePath()),
			BuildCommand:     strPtr(s.GetBuildCommand()),
			RunCommand:       strPtr(s.GetRunCommand()),
			InstanceSizeSlug: strPtr(s.GetInstanceSizeSlug()),
			Envs:             serviceEnvs(s.GetEnvs()),
		}
		if s.GetInstanceCount() > 0 {
			args.InstanceCount = pulumi.Int(int(s.GetInstanceCount()))
		}
		if s.HttpPort != nil {
			args.HttpPort = pulumi.Int(int(*s.HttpPort))
		}
		if ports := s.GetInternalPorts(); len(ports) > 0 {
			ints := make([]int, len(ports))
			for i, p := range ports {
				ints[i] = int(p)
			}
			args.InternalPorts = pulumi.ToIntArray(ints)
		}
		applyServiceSource(&args, s)
		if hc := healthCheck(s.GetHealthCheck()); hc != nil {
			args.HealthCheck = hc
		}
		if s.GetAutoscaling() != nil {
			a := s.GetAutoscaling()
			args.Autoscaling = &digitalocean.AppSpecServiceAutoscalingArgs{
				MinInstanceCount: pulumi.Int(int(a.GetMinInstanceCount())),
				MaxInstanceCount: pulumi.Int(int(a.GetMaxInstanceCount())),
				Metrics: digitalocean.AppSpecServiceAutoscalingMetricsArgs{
					Cpu: digitalocean.AppSpecServiceAutoscalingMetricsCpuArgs{
						Percent: pulumi.Int(int(a.GetCpuPercent())),
					},
				},
			}
		}
		if s.GetTermination() != nil {
			args.Termination = &digitalocean.AppSpecServiceTerminationArgs{
				GracePeriodSeconds: intPtrFromUint32(s.GetTermination().GracePeriodSeconds),
				DrainSeconds:       intPtrFromUint32(s.GetTermination().DrainSeconds),
			}
		}
		args.Alerts = serviceAlerts(s.GetAlerts())
		args.LogDestinations = serviceLogs(s.GetLogDestinations())
		out = append(out, args)
	}
	return out, nil
}

func applyServiceSource(args *digitalocean.AppSpecServiceArgs, s *digitaloceanappv1alpha1.DigitalOceanAppService) {
	if g := s.GetGit(); g != nil {
		args.Git = &digitalocean.AppSpecServiceGitArgs{
			RepoCloneUrl: pulumi.String(g.GetRepoCloneUrl()),
			Branch:       pulumi.String(g.GetBranch()),
		}
	}
	if g := s.GetGithub(); g != nil {
		args.Github = &digitalocean.AppSpecServiceGithubArgs{
			Repo:         pulumi.String(g.GetRepo()),
			Branch:       pulumi.String(g.GetBranch()),
			DeployOnPush: pulumi.Bool(g.GetDeployOnPush()),
		}
	}
	if g := s.GetGitlab(); g != nil {
		args.Gitlab = &digitalocean.AppSpecServiceGitlabArgs{
			Repo:         pulumi.String(g.GetRepo()),
			Branch:       pulumi.String(g.GetBranch()),
			DeployOnPush: pulumi.Bool(g.GetDeployOnPush()),
		}
	}
	if g := s.GetBitbucket(); g != nil {
		args.Bitbucket = &digitalocean.AppSpecServiceBitbucketArgs{
			Repo:         pulumi.String(g.GetRepo()),
			Branch:       pulumi.String(g.GetBranch()),
			DeployOnPush: pulumi.Bool(g.GetDeployOnPush()),
		}
	}
	if img := s.GetImage(); img != nil {
		args.Image = serviceImage(img)
	}
}

func serviceImage(img *do.DigitalOceanAppImageSource) *digitalocean.AppSpecServiceImageArgs {
	args := &digitalocean.AppSpecServiceImageArgs{
		RegistryType:        pulumi.String(providerEnum(img.GetRegistryType().String())),
		Registry:            strPtr(img.GetRegistry()),
		Repository:          pulumi.String(img.GetRepository()),
		Tag:                 strPtr(img.GetTag()),
		Digest:              strPtr(img.GetDigest()),
		RegistryCredentials: strPtr(img.GetRegistryCredentials()),
	}
	if img.GetDeployOnPush() {
		args.DeployOnPushes = digitalocean.AppSpecServiceImageDeployOnPushArray{
			digitalocean.AppSpecServiceImageDeployOnPushArgs{Enabled: pulumi.Bool(true)},
		}
	}
	return args
}

func serviceAlerts(in []*do.DigitalOceanAppComponentAlert) digitalocean.AppSpecServiceAlertArray {
	out := digitalocean.AppSpecServiceAlertArray{}
	for _, a := range in {
		args := digitalocean.AppSpecServiceAlertArgs{
			Rule:     pulumi.String(providerEnum(a.GetRule().String())),
			Operator: pulumi.String(providerEnum(a.GetOperator().String())),
			Window:   pulumi.String(providerEnum(a.GetWindow().String())),
			Value:    pulumi.Float64(a.GetValue()),
			Disabled: pulumi.Bool(a.GetDisabled()),
		}
		if d := a.GetDestinations(); destinationsSet(d) {
			hooks := digitalocean.AppSpecServiceAlertDestinationsSlackWebhookArray{}
			for _, h := range d.GetSlackWebhooks() {
				hooks = append(hooks, digitalocean.AppSpecServiceAlertDestinationsSlackWebhookArgs{
					Channel: pulumi.String(h.GetChannel()), Url: secretString(h.GetUrl()),
				})
			}
			args.Destinations = &digitalocean.AppSpecServiceAlertDestinationsArgs{
				Emails: pulumi.ToStringArray(d.GetEmails()), SlackWebhooks: hooks,
			}
		}
		out = append(out, args)
	}
	return out
}

func serviceLogs(in []*do.DigitalOceanAppLogDestination) digitalocean.AppSpecServiceLogDestinationArray {
	out := digitalocean.AppSpecServiceLogDestinationArray{}
	for _, d := range in {
		args := digitalocean.AppSpecServiceLogDestinationArgs{Name: pulumi.String(d.GetName())}
		if p := d.GetPapertrail(); p != nil {
			args.Papertrail = &digitalocean.AppSpecServiceLogDestinationPapertrailArgs{
				Endpoint: pulumi.String(p.GetEndpoint()),
			}
		}
		if dd := d.GetDatadog(); dd != nil {
			args.Datadog = &digitalocean.AppSpecServiceLogDestinationDatadogArgs{
				ApiKey:   pulumi.String(dd.GetApiKey()),
				Endpoint: strPtr(dd.GetEndpoint()),
			}
		}
		if l := d.GetLogtail(); l != nil {
			args.Logtail = &digitalocean.AppSpecServiceLogDestinationLogtailArgs{
				Token: pulumi.String(l.GetToken()),
			}
		}
		if o := d.GetOpenSearch(); o != nil {
			osArgs := &digitalocean.AppSpecServiceLogDestinationOpenSearchArgs{
				Endpoint:    strPtr(o.GetEndpoint()),
				IndexName:   strPtr(o.GetIndexName()),
				ClusterName: strPtr(o.GetClusterName()),
			}
			ba := o.GetBasicAuth()
			user, pass := "", ""
			if ba != nil {
				user, pass = ba.GetUser(), ba.GetPassword()
			}
			osArgs.BasicAuth = digitalocean.AppSpecServiceLogDestinationOpenSearchBasicAuthArgs{
				User:     strPtr(user),
				Password: strPtr(pass),
			}
			args.OpenSearch = osArgs
		}
		out = append(out, args)
	}
	return out
}

func buildWorkers(in []*digitaloceanappv1alpha1.DigitalOceanAppWorker) (digitalocean.AppSpecWorkerArray, error) {
	out := digitalocean.AppSpecWorkerArray{}
	for _, w := range in {
		args := digitalocean.AppSpecWorkerArgs{
			Name:             pulumi.String(w.GetName()),
			SourceDir:        strPtr(w.GetSourceDir()),
			EnvironmentSlug:  strPtr(w.GetEnvironmentSlug()),
			DockerfilePath:   strPtr(w.GetDockerfilePath()),
			BuildCommand:     strPtr(w.GetBuildCommand()),
			RunCommand:       strPtr(w.GetRunCommand()),
			InstanceSizeSlug: strPtr(w.GetInstanceSizeSlug()),
			Envs:             workerEnvs(w.GetEnvs()),
		}
		if w.GetInstanceCount() > 0 {
			args.InstanceCount = pulumi.Int(int(w.GetInstanceCount()))
		}
		if g := w.GetGit(); g != nil {
			args.Git = &digitalocean.AppSpecWorkerGitArgs{
				RepoCloneUrl: pulumi.String(g.GetRepoCloneUrl()),
				Branch:       pulumi.String(g.GetBranch()),
			}
		}
		if g := w.GetGithub(); g != nil {
			args.Github = &digitalocean.AppSpecWorkerGithubArgs{
				Repo:         pulumi.String(g.GetRepo()),
				Branch:       pulumi.String(g.GetBranch()),
				DeployOnPush: pulumi.Bool(g.GetDeployOnPush()),
			}
		}
		if g := w.GetGitlab(); g != nil {
			args.Gitlab = &digitalocean.AppSpecWorkerGitlabArgs{
				Repo:         pulumi.String(g.GetRepo()),
				Branch:       pulumi.String(g.GetBranch()),
				DeployOnPush: pulumi.Bool(g.GetDeployOnPush()),
			}
		}
		if g := w.GetBitbucket(); g != nil {
			args.Bitbucket = &digitalocean.AppSpecWorkerBitbucketArgs{
				Repo:         pulumi.String(g.GetRepo()),
				Branch:       pulumi.String(g.GetBranch()),
				DeployOnPush: pulumi.Bool(g.GetDeployOnPush()),
			}
		}
		if img := w.GetImage(); img != nil {
			args.Image = workerImage(img)
		}
		if w.GetAutoscaling() != nil {
			a := w.GetAutoscaling()
			args.Autoscaling = &digitalocean.AppSpecWorkerAutoscalingArgs{
				MinInstanceCount: pulumi.Int(int(a.GetMinInstanceCount())),
				MaxInstanceCount: pulumi.Int(int(a.GetMaxInstanceCount())),
				Metrics: digitalocean.AppSpecWorkerAutoscalingMetricsArgs{
					Cpu: digitalocean.AppSpecWorkerAutoscalingMetricsCpuArgs{
						Percent: pulumi.Int(int(a.GetCpuPercent())),
					},
				},
			}
		}
		if w.GetTermination() != nil {
			args.Termination = &digitalocean.AppSpecWorkerTerminationArgs{
				GracePeriodSeconds: intPtrFromUint32(w.GetTermination().GracePeriodSeconds),
			}
		}
		args.Alerts = workerAlerts(w.GetAlerts())
		args.LogDestinations = workerLogs(w.GetLogDestinations())
		out = append(out, args)
	}
	return out, nil
}

func workerImage(img *do.DigitalOceanAppImageSource) *digitalocean.AppSpecWorkerImageArgs {
	args := &digitalocean.AppSpecWorkerImageArgs{
		RegistryType:        pulumi.String(providerEnum(img.GetRegistryType().String())),
		Registry:            strPtr(img.GetRegistry()),
		Repository:          pulumi.String(img.GetRepository()),
		Tag:                 strPtr(img.GetTag()),
		Digest:              strPtr(img.GetDigest()),
		RegistryCredentials: strPtr(img.GetRegistryCredentials()),
	}
	if img.GetDeployOnPush() {
		args.DeployOnPushes = digitalocean.AppSpecWorkerImageDeployOnPushArray{
			digitalocean.AppSpecWorkerImageDeployOnPushArgs{Enabled: pulumi.Bool(true)},
		}
	}
	return args
}

func workerAlerts(in []*do.DigitalOceanAppComponentAlert) digitalocean.AppSpecWorkerAlertArray {
	out := digitalocean.AppSpecWorkerAlertArray{}
	for _, a := range in {
		args := digitalocean.AppSpecWorkerAlertArgs{
			Rule:     pulumi.String(providerEnum(a.GetRule().String())),
			Operator: pulumi.String(providerEnum(a.GetOperator().String())),
			Window:   pulumi.String(providerEnum(a.GetWindow().String())),
			Value:    pulumi.Float64(a.GetValue()),
			Disabled: pulumi.Bool(a.GetDisabled()),
		}
		if d := a.GetDestinations(); destinationsSet(d) {
			hooks := digitalocean.AppSpecWorkerAlertDestinationsSlackWebhookArray{}
			for _, h := range d.GetSlackWebhooks() {
				hooks = append(hooks, digitalocean.AppSpecWorkerAlertDestinationsSlackWebhookArgs{
					Channel: pulumi.String(h.GetChannel()), Url: secretString(h.GetUrl()),
				})
			}
			args.Destinations = &digitalocean.AppSpecWorkerAlertDestinationsArgs{
				Emails: pulumi.ToStringArray(d.GetEmails()), SlackWebhooks: hooks,
			}
		}
		out = append(out, args)
	}
	return out
}

func workerLogs(in []*do.DigitalOceanAppLogDestination) digitalocean.AppSpecWorkerLogDestinationArray {
	out := digitalocean.AppSpecWorkerLogDestinationArray{}
	for _, d := range in {
		args := digitalocean.AppSpecWorkerLogDestinationArgs{Name: pulumi.String(d.GetName())}
		if p := d.GetPapertrail(); p != nil {
			args.Papertrail = &digitalocean.AppSpecWorkerLogDestinationPapertrailArgs{Endpoint: pulumi.String(p.GetEndpoint())}
		}
		if dd := d.GetDatadog(); dd != nil {
			args.Datadog = &digitalocean.AppSpecWorkerLogDestinationDatadogArgs{
				ApiKey: pulumi.String(dd.GetApiKey()), Endpoint: strPtr(dd.GetEndpoint()),
			}
		}
		if l := d.GetLogtail(); l != nil {
			args.Logtail = &digitalocean.AppSpecWorkerLogDestinationLogtailArgs{Token: pulumi.String(l.GetToken())}
		}
		if o := d.GetOpenSearch(); o != nil {
			user, pass := "", ""
			if ba := o.GetBasicAuth(); ba != nil {
				user, pass = ba.GetUser(), ba.GetPassword()
			}
			args.OpenSearch = &digitalocean.AppSpecWorkerLogDestinationOpenSearchArgs{
				Endpoint:    strPtr(o.GetEndpoint()),
				IndexName:   strPtr(o.GetIndexName()),
				ClusterName: strPtr(o.GetClusterName()),
				BasicAuth: digitalocean.AppSpecWorkerLogDestinationOpenSearchBasicAuthArgs{
					User: strPtr(user), Password: strPtr(pass),
				},
			}
		}
		out = append(out, args)
	}
	return out
}

func buildJobs(in []*digitaloceanappv1alpha1.DigitalOceanAppJob) (digitalocean.AppSpecJobArray, error) {
	out := digitalocean.AppSpecJobArray{}
	for _, j := range in {
		args := digitalocean.AppSpecJobArgs{
			Name:             pulumi.String(j.GetName()),
			SourceDir:        strPtr(j.GetSourceDir()),
			EnvironmentSlug:  strPtr(j.GetEnvironmentSlug()),
			DockerfilePath:   strPtr(j.GetDockerfilePath()),
			BuildCommand:     strPtr(j.GetBuildCommand()),
			RunCommand:       strPtr(j.GetRunCommand()),
			InstanceSizeSlug: strPtr(j.GetInstanceSizeSlug()),
			Envs:             jobEnvs(j.GetEnvs()),
		}
		if j.GetInstanceCount() > 0 {
			args.InstanceCount = pulumi.Int(int(j.GetInstanceCount()))
		}
		if k := providerEnum(j.GetKind().String()); k != "" {
			args.Kind = pulumi.String(k)
		}
		if g := j.GetGit(); g != nil {
			args.Git = &digitalocean.AppSpecJobGitArgs{
				RepoCloneUrl: pulumi.String(g.GetRepoCloneUrl()),
				Branch:       pulumi.String(g.GetBranch()),
			}
		}
		if g := j.GetGithub(); g != nil {
			args.Github = &digitalocean.AppSpecJobGithubArgs{
				Repo: pulumi.String(g.GetRepo()), Branch: pulumi.String(g.GetBranch()), DeployOnPush: pulumi.Bool(g.GetDeployOnPush()),
			}
		}
		if g := j.GetGitlab(); g != nil {
			args.Gitlab = &digitalocean.AppSpecJobGitlabArgs{
				Repo: pulumi.String(g.GetRepo()), Branch: pulumi.String(g.GetBranch()), DeployOnPush: pulumi.Bool(g.GetDeployOnPush()),
			}
		}
		if g := j.GetBitbucket(); g != nil {
			args.Bitbucket = &digitalocean.AppSpecJobBitbucketArgs{
				Repo: pulumi.String(g.GetRepo()), Branch: pulumi.String(g.GetBranch()), DeployOnPush: pulumi.Bool(g.GetDeployOnPush()),
			}
		}
		if img := j.GetImage(); img != nil {
			args.Image = jobImage(img)
		}
		if j.GetTermination() != nil {
			args.Termination = &digitalocean.AppSpecJobTerminationArgs{
				GracePeriodSeconds: intPtrFromUint32(j.GetTermination().GracePeriodSeconds),
			}
		}
		args.Alerts = jobAlerts(j.GetAlerts())
		args.LogDestinations = jobLogs(j.GetLogDestinations())
		out = append(out, args)
	}
	return out, nil
}

func jobImage(img *do.DigitalOceanAppImageSource) *digitalocean.AppSpecJobImageArgs {
	args := &digitalocean.AppSpecJobImageArgs{
		RegistryType:        pulumi.String(providerEnum(img.GetRegistryType().String())),
		Registry:            strPtr(img.GetRegistry()),
		Repository:          pulumi.String(img.GetRepository()),
		Tag:                 strPtr(img.GetTag()),
		Digest:              strPtr(img.GetDigest()),
		RegistryCredentials: strPtr(img.GetRegistryCredentials()),
	}
	if img.GetDeployOnPush() {
		args.DeployOnPushes = digitalocean.AppSpecJobImageDeployOnPushArray{
			digitalocean.AppSpecJobImageDeployOnPushArgs{Enabled: pulumi.Bool(true)},
		}
	}
	return args
}

func jobAlerts(in []*do.DigitalOceanAppComponentAlert) digitalocean.AppSpecJobAlertArray {
	out := digitalocean.AppSpecJobAlertArray{}
	for _, a := range in {
		args := digitalocean.AppSpecJobAlertArgs{
			Rule:     pulumi.String(providerEnum(a.GetRule().String())),
			Operator: pulumi.String(providerEnum(a.GetOperator().String())),
			Window:   pulumi.String(providerEnum(a.GetWindow().String())),
			Value:    pulumi.Float64(a.GetValue()),
			Disabled: pulumi.Bool(a.GetDisabled()),
		}
		if d := a.GetDestinations(); destinationsSet(d) {
			hooks := digitalocean.AppSpecJobAlertDestinationsSlackWebhookArray{}
			for _, h := range d.GetSlackWebhooks() {
				hooks = append(hooks, digitalocean.AppSpecJobAlertDestinationsSlackWebhookArgs{
					Channel: pulumi.String(h.GetChannel()), Url: secretString(h.GetUrl()),
				})
			}
			args.Destinations = &digitalocean.AppSpecJobAlertDestinationsArgs{
				Emails: pulumi.ToStringArray(d.GetEmails()), SlackWebhooks: hooks,
			}
		}
		out = append(out, args)
	}
	return out
}

func jobLogs(in []*do.DigitalOceanAppLogDestination) digitalocean.AppSpecJobLogDestinationArray {
	out := digitalocean.AppSpecJobLogDestinationArray{}
	for _, d := range in {
		args := digitalocean.AppSpecJobLogDestinationArgs{Name: pulumi.String(d.GetName())}
		if p := d.GetPapertrail(); p != nil {
			args.Papertrail = &digitalocean.AppSpecJobLogDestinationPapertrailArgs{Endpoint: pulumi.String(p.GetEndpoint())}
		}
		if dd := d.GetDatadog(); dd != nil {
			args.Datadog = &digitalocean.AppSpecJobLogDestinationDatadogArgs{
				ApiKey: pulumi.String(dd.GetApiKey()), Endpoint: strPtr(dd.GetEndpoint()),
			}
		}
		if l := d.GetLogtail(); l != nil {
			args.Logtail = &digitalocean.AppSpecJobLogDestinationLogtailArgs{Token: pulumi.String(l.GetToken())}
		}
		if o := d.GetOpenSearch(); o != nil {
			user, pass := "", ""
			if ba := o.GetBasicAuth(); ba != nil {
				user, pass = ba.GetUser(), ba.GetPassword()
			}
			args.OpenSearch = &digitalocean.AppSpecJobLogDestinationOpenSearchArgs{
				Endpoint: strPtr(o.GetEndpoint()), IndexName: strPtr(o.GetIndexName()), ClusterName: strPtr(o.GetClusterName()),
				BasicAuth: digitalocean.AppSpecJobLogDestinationOpenSearchBasicAuthArgs{User: strPtr(user), Password: strPtr(pass)},
			}
		}
		out = append(out, args)
	}
	return out
}

func buildStaticSites(in []*digitaloceanappv1alpha1.DigitalOceanAppStaticSite) (digitalocean.AppSpecStaticSiteArray, error) {
	out := digitalocean.AppSpecStaticSiteArray{}
	for _, s := range in {
		args := digitalocean.AppSpecStaticSiteArgs{
			Name:             pulumi.String(s.GetName()),
			SourceDir:        strPtr(s.GetSourceDir()),
			EnvironmentSlug:  strPtr(s.GetEnvironmentSlug()),
			DockerfilePath:   strPtr(s.GetDockerfilePath()),
			BuildCommand:     strPtr(s.GetBuildCommand()),
			OutputDir:        strPtr(s.GetOutputDir()),
			IndexDocument:    strPtr(s.GetIndexDocument()),
			ErrorDocument:    strPtr(s.GetErrorDocument()),
			CatchallDocument: strPtr(s.GetCatchallDocument()),
			Envs:             staticSiteEnvs(s.GetEnvs()),
		}
		if g := s.GetGit(); g != nil {
			args.Git = &digitalocean.AppSpecStaticSiteGitArgs{
				RepoCloneUrl: pulumi.String(g.GetRepoCloneUrl()), Branch: pulumi.String(g.GetBranch()),
			}
		}
		if g := s.GetGithub(); g != nil {
			args.Github = &digitalocean.AppSpecStaticSiteGithubArgs{
				Repo: pulumi.String(g.GetRepo()), Branch: pulumi.String(g.GetBranch()), DeployOnPush: pulumi.Bool(g.GetDeployOnPush()),
			}
		}
		if g := s.GetGitlab(); g != nil {
			args.Gitlab = &digitalocean.AppSpecStaticSiteGitlabArgs{
				Repo: pulumi.String(g.GetRepo()), Branch: pulumi.String(g.GetBranch()), DeployOnPush: pulumi.Bool(g.GetDeployOnPush()),
			}
		}
		if g := s.GetBitbucket(); g != nil {
			args.Bitbucket = &digitalocean.AppSpecStaticSiteBitbucketArgs{
				Repo: pulumi.String(g.GetRepo()), Branch: pulumi.String(g.GetBranch()), DeployOnPush: pulumi.Bool(g.GetDeployOnPush()),
			}
		}
		out = append(out, args)
	}
	return out, nil
}

func buildFunctions(in []*digitaloceanappv1alpha1.DigitalOceanAppFunctionComponent) (digitalocean.AppSpecFunctionArray, error) {
	out := digitalocean.AppSpecFunctionArray{}
	for _, f := range in {
		args := digitalocean.AppSpecFunctionArgs{
			Name:      pulumi.String(f.GetName()),
			SourceDir: strPtr(f.GetSourceDir()),
			Envs:      functionEnvs(f.GetEnvs()),
			Alerts:    functionAlerts(f.GetAlerts()),
		}
		if g := f.GetGit(); g != nil {
			args.Git = &digitalocean.AppSpecFunctionGitArgs{
				RepoCloneUrl: pulumi.String(g.GetRepoCloneUrl()), Branch: pulumi.String(g.GetBranch()),
			}
		}
		if g := f.GetGithub(); g != nil {
			args.Github = &digitalocean.AppSpecFunctionGithubArgs{
				Repo: pulumi.String(g.GetRepo()), Branch: pulumi.String(g.GetBranch()), DeployOnPush: pulumi.Bool(g.GetDeployOnPush()),
			}
		}
		if g := f.GetGitlab(); g != nil {
			args.Gitlab = &digitalocean.AppSpecFunctionGitlabArgs{
				Repo: pulumi.String(g.GetRepo()), Branch: pulumi.String(g.GetBranch()), DeployOnPush: pulumi.Bool(g.GetDeployOnPush()),
			}
		}
		if g := f.GetBitbucket(); g != nil {
			args.Bitbucket = &digitalocean.AppSpecFunctionBitbucketArgs{
				Repo: pulumi.String(g.GetRepo()), Branch: pulumi.String(g.GetBranch()), DeployOnPush: pulumi.Bool(g.GetDeployOnPush()),
			}
		}
		args.LogDestinations = functionLogs(f.GetLogDestinations())
		out = append(out, args)
	}
	return out, nil
}

func functionAlerts(in []*do.DigitalOceanAppComponentAlert) digitalocean.AppSpecFunctionAlertArray {
	out := digitalocean.AppSpecFunctionAlertArray{}
	for _, a := range in {
		args := digitalocean.AppSpecFunctionAlertArgs{
			Rule:     pulumi.String(providerEnum(a.GetRule().String())),
			Operator: pulumi.String(providerEnum(a.GetOperator().String())),
			Window:   pulumi.String(providerEnum(a.GetWindow().String())),
			Value:    pulumi.Float64(a.GetValue()),
			Disabled: pulumi.Bool(a.GetDisabled()),
		}
		if d := a.GetDestinations(); destinationsSet(d) {
			hooks := digitalocean.AppSpecFunctionAlertDestinationsSlackWebhookArray{}
			for _, h := range d.GetSlackWebhooks() {
				hooks = append(hooks, digitalocean.AppSpecFunctionAlertDestinationsSlackWebhookArgs{
					Channel: pulumi.String(h.GetChannel()), Url: secretString(h.GetUrl()),
				})
			}
			args.Destinations = &digitalocean.AppSpecFunctionAlertDestinationsArgs{
				Emails: pulumi.ToStringArray(d.GetEmails()), SlackWebhooks: hooks,
			}
		}
		out = append(out, args)
	}
	return out
}

func functionLogs(in []*do.DigitalOceanAppLogDestination) digitalocean.AppSpecFunctionLogDestinationArray {
	out := digitalocean.AppSpecFunctionLogDestinationArray{}
	for _, d := range in {
		args := digitalocean.AppSpecFunctionLogDestinationArgs{Name: pulumi.String(d.GetName())}
		if p := d.GetPapertrail(); p != nil {
			args.Papertrail = &digitalocean.AppSpecFunctionLogDestinationPapertrailArgs{Endpoint: pulumi.String(p.GetEndpoint())}
		}
		if dd := d.GetDatadog(); dd != nil {
			args.Datadog = &digitalocean.AppSpecFunctionLogDestinationDatadogArgs{
				ApiKey: pulumi.String(dd.GetApiKey()), Endpoint: strPtr(dd.GetEndpoint()),
			}
		}
		if l := d.GetLogtail(); l != nil {
			args.Logtail = &digitalocean.AppSpecFunctionLogDestinationLogtailArgs{Token: pulumi.String(l.GetToken())}
		}
		if o := d.GetOpenSearch(); o != nil {
			user, pass := "", ""
			if ba := o.GetBasicAuth(); ba != nil {
				user, pass = ba.GetUser(), ba.GetPassword()
			}
			args.OpenSearch = &digitalocean.AppSpecFunctionLogDestinationOpenSearchArgs{
				Endpoint: strPtr(o.GetEndpoint()), IndexName: strPtr(o.GetIndexName()), ClusterName: strPtr(o.GetClusterName()),
				BasicAuth: digitalocean.AppSpecFunctionLogDestinationOpenSearchBasicAuthArgs{User: strPtr(user), Password: strPtr(pass)},
			}
		}
		out = append(out, args)
	}
	return out
}

func buildDatabases(in []*do.DigitalOceanAppDatabase) digitalocean.AppSpecDatabaseArray {
	out := digitalocean.AppSpecDatabaseArray{}
	for _, d := range in {
		args := digitalocean.AppSpecDatabaseArgs{
			Name:       strPtr(d.GetName()),
			Version:    strPtr(d.GetVersion()),
			Production: pulumi.Bool(d.GetProduction()),
			DbName:     strPtr(d.GetDbName()),
			DbUser:     strPtr(d.GetDbUser()),
		}
		if e := providerEnum(d.GetEngine().String()); e != "" {
			args.Engine = pulumi.String(e)
		}
		if d.GetClusterName() != nil && d.GetClusterName().GetValue() != "" {
			args.ClusterName = pulumi.String(d.GetClusterName().GetValue())
		}
		out = append(out, args)
	}
	return out
}

func buildDomains(in []*do.DigitalOceanAppDomain) digitalocean.AppSpecDomainNameArray {
	out := digitalocean.AppSpecDomainNameArray{}
	for _, d := range in {
		args := digitalocean.AppSpecDomainNameArgs{
			Name:     pulumi.String(d.GetName()),
			Wildcard: pulumi.Bool(d.GetWildcard()),
		}
		if d.GetType() != "" {
			args.Type = pulumi.String(d.GetType())
		}
		if d.GetZone() != nil && d.GetZone().GetValue() != "" {
			args.Zone = pulumi.String(d.GetZone().GetValue())
		}
		out = append(out, args)
	}
	return out
}

// buildAppAlerts wires app-level alert rules (deployment, domain, autoscale
// events). Destinations ride the same block: the provider applies them through
// a side-channel call after the app exists and reads them back from
// ListAlerts, so both engines carry them in state the same way.
func buildAppAlerts(in []*do.DigitalOceanAppAlert) digitalocean.AppSpecAlertArray {
	out := digitalocean.AppSpecAlertArray{}
	for _, a := range in {
		args := digitalocean.AppSpecAlertArgs{
			Rule:     pulumi.String(providerEnum(a.GetRule().String())),
			Disabled: pulumi.Bool(a.GetDisabled()),
		}
		if d := a.GetDestinations(); destinationsSet(d) {
			hooks := digitalocean.AppSpecAlertDestinationsSlackWebhookArray{}
			for _, h := range d.GetSlackWebhooks() {
				hooks = append(hooks, digitalocean.AppSpecAlertDestinationsSlackWebhookArgs{
					Channel: pulumi.String(h.GetChannel()), Url: secretString(h.GetUrl()),
				})
			}
			args.Destinations = &digitalocean.AppSpecAlertDestinationsArgs{
				Emails: pulumi.ToStringArray(d.GetEmails()), SlackWebhooks: hooks,
			}
		}
		out = append(out, args)
	}
	return out
}

func buildIngress(ing *do.DigitalOceanAppIngress) *digitalocean.AppSpecIngressArgs {
	rules := digitalocean.AppSpecIngressRuleArray{}
	for _, r := range ing.GetRules() {
		rule := digitalocean.AppSpecIngressRuleArgs{}
		if m := r.GetMatch(); m != nil && (m.GetPathPrefix() != "" || m.GetAuthorityExact() != "") {
			match := &digitalocean.AppSpecIngressRuleMatchArgs{}
			if m.GetPathPrefix() != "" {
				match.Path = &digitalocean.AppSpecIngressRuleMatchPathArgs{
					Prefix: pulumi.String(m.GetPathPrefix()),
				}
			}
			if m.GetAuthorityExact() != "" {
				match.Authority = &digitalocean.AppSpecIngressRuleMatchAuthorityArgs{
					Exact: pulumi.String(m.GetAuthorityExact()),
				}
			}
			rule.Match = match
		}
		if c := r.GetComponent(); c != nil {
			rule.Component = &digitalocean.AppSpecIngressRuleComponentArgs{
				Name:               pulumi.String(c.GetName()),
				PreservePathPrefix: pulumi.Bool(c.GetPreservePathPrefix()),
				Rewrite:            strPtr(c.GetRewrite()),
			}
		}
		if re := r.GetRedirect(); re != nil {
			rd := &digitalocean.AppSpecIngressRuleRedirectArgs{
				Uri:       strPtr(re.GetUri()),
				Authority: strPtr(re.GetAuthority()),
				Scheme:    strPtr(re.GetScheme()),
			}
			if re.Port != nil {
				rd.Port = pulumi.Int(int(*re.Port))
			}
			if re.RedirectCode != nil {
				rd.RedirectCode = pulumi.Int(int(*re.RedirectCode))
			}
			rule.Redirect = rd
		}
		if cors := r.GetCors(); cors != nil {
			rule.Cors = buildCors(cors)
		}
		rules = append(rules, rule)
	}
	return &digitalocean.AppSpecIngressArgs{Rules: rules}
}

func buildCors(c *do.DigitalOceanAppCors) *digitalocean.AppSpecIngressRuleCorsArgs {
	args := &digitalocean.AppSpecIngressRuleCorsArgs{
		AllowMethods:     pulumi.ToStringArray(c.GetAllowMethods()),
		AllowHeaders:     pulumi.ToStringArray(c.GetAllowHeaders()),
		ExposeHeaders:    pulumi.ToStringArray(c.GetExposeHeaders()),
		MaxAge:           strPtr(c.GetMaxAge()),
		AllowCredentials: pulumi.Bool(c.GetAllowCredentials()),
	}
	if o := c.GetAllowOrigins(); o != nil {
		args.AllowOrigins = &digitalocean.AppSpecIngressRuleCorsAllowOriginsArgs{
			Exact: strPtr(o.GetExact()),
			Regex: strPtr(o.GetRegex()),
		}
	}
	return args
}
