package module

import (
	cloudflarepagesprojectv1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarepagesproject/v1"
	cloudfl "github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// deploymentConfigs builds the preview + production deployment configs. Preview
// and production share the same proto shape but are distinct Pulumi types, so we
// build each side explicitly.
func deploymentConfigs(dc *cloudflarepagesprojectv1.CloudflarePagesDeploymentConfigs) cloudfl.PagesProjectDeploymentConfigsPtrInput {
	if dc == nil {
		return nil
	}
	// Cloudflare treats preview and production as a paired configuration and
	// rejects inconsistent environments (e.g. fail_open must match). When only one
	// environment is supplied we mirror it to both so a single config "just works".
	previewSrc, productionSrc := dc.Preview, dc.Production
	if previewSrc == nil {
		previewSrc = productionSrc
	}
	if productionSrc == nil {
		productionSrc = previewSrc
	}
	if previewSrc == nil && productionSrc == nil {
		return nil
	}
	return &cloudfl.PagesProjectDeploymentConfigsArgs{
		Preview:    previewConfig(previewSrc),
		Production: productionConfig(productionSrc),
	}
}

func productionConfig(c *cloudflarepagesprojectv1.CloudflarePagesDeploymentConfig) cloudfl.PagesProjectDeploymentConfigsProductionPtrInput {
	args := &cloudfl.PagesProjectDeploymentConfigsProductionArgs{}
	if c.CompatibilityDate != "" {
		args.CompatibilityDate = pulumi.String(c.CompatibilityDate)
	}
	if len(c.CompatibilityFlags) > 0 {
		args.CompatibilityFlags = pulumi.ToStringArray(c.CompatibilityFlags)
	}
	if c.AlwaysUseLatestCompatibilityDate {
		args.AlwaysUseLatestCompatibilityDate = pulumi.Bool(true)
	}
	if c.BuildImageMajorVersion > 0 {
		args.BuildImageMajorVersion = pulumi.Int(int(c.BuildImageMajorVersion))
	}
	if c.FailOpen {
		args.FailOpen = pulumi.Bool(true)
	}
	if c.UsageModel != "" {
		args.UsageModel = pulumi.String(c.UsageModel)
	}
	if c.Limits != nil && c.Limits.CpuMs > 0 {
		args.Limits = &cloudfl.PagesProjectDeploymentConfigsProductionLimitsArgs{CpuMs: pulumi.Int(int(c.Limits.CpuMs))}
	}
	if c.Placement != nil && c.Placement.Mode != "" {
		args.Placement = &cloudfl.PagesProjectDeploymentConfigsProductionPlacementArgs{Mode: pulumi.String(c.Placement.Mode)}
	}
	if env := productionEnvVars(c); len(env) > 0 {
		args.EnvVars = env
	}
	if len(c.KvNamespaces) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsProductionKvNamespacesMap{}
		for _, b := range c.KvNamespaces {
			m[b.Name] = cloudfl.PagesProjectDeploymentConfigsProductionKvNamespacesArgs{NamespaceId: pulumi.String(b.NamespaceId.GetValue())}
		}
		args.KvNamespaces = m
	}
	if len(c.D1Databases) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsProductionD1DatabasesMap{}
		for _, b := range c.D1Databases {
			m[b.Name] = cloudfl.PagesProjectDeploymentConfigsProductionD1DatabasesArgs{Id: pulumi.String(b.DatabaseId.GetValue())}
		}
		args.D1Databases = m
	}
	if len(c.R2Buckets) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsProductionR2BucketsMap{}
		for _, b := range c.R2Buckets {
			r2 := cloudfl.PagesProjectDeploymentConfigsProductionR2BucketsArgs{Name: pulumi.String(b.BucketName.GetValue())}
			if b.Jurisdiction != "" {
				r2.Jurisdiction = pulumi.String(b.Jurisdiction)
			}
			m[b.Name] = r2
		}
		args.R2Buckets = m
	}
	if len(c.QueueProducers) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsProductionQueueProducersMap{}
		for _, b := range c.QueueProducers {
			m[b.Name] = cloudfl.PagesProjectDeploymentConfigsProductionQueueProducersArgs{Name: pulumi.String(b.QueueName.GetValue())}
		}
		args.QueueProducers = m
	}
	if len(c.HyperdriveBindings) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsProductionHyperdriveBindingsMap{}
		for _, b := range c.HyperdriveBindings {
			m[b.Name] = cloudfl.PagesProjectDeploymentConfigsProductionHyperdriveBindingsArgs{Id: pulumi.String(b.ConfigId.GetValue())}
		}
		args.HyperdriveBindings = m
	}
	if len(c.Services) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsProductionServicesMap{}
		for _, b := range c.Services {
			svc := cloudfl.PagesProjectDeploymentConfigsProductionServicesArgs{Service: pulumi.String(b.Service.GetValue())}
			if b.Entrypoint != "" {
				svc.Entrypoint = pulumi.String(b.Entrypoint)
			}
			if b.Environment != "" {
				svc.Environment = pulumi.String(b.Environment)
			}
			m[b.Name] = svc
		}
		args.Services = m
	}
	if len(c.DurableObjectNamespaces) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsProductionDurableObjectNamespacesMap{}
		for _, b := range c.DurableObjectNamespaces {
			m[b.Name] = cloudfl.PagesProjectDeploymentConfigsProductionDurableObjectNamespacesArgs{NamespaceId: pulumi.String(b.NamespaceId)}
		}
		args.DurableObjectNamespaces = m
	}
	if len(c.AnalyticsEngineDatasets) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsProductionAnalyticsEngineDatasetsMap{}
		for _, b := range c.AnalyticsEngineDatasets {
			m[b.Name] = cloudfl.PagesProjectDeploymentConfigsProductionAnalyticsEngineDatasetsArgs{Dataset: pulumi.String(b.Dataset)}
		}
		args.AnalyticsEngineDatasets = m
	}
	if len(c.VectorizeBindings) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsProductionVectorizeBindingsMap{}
		for _, b := range c.VectorizeBindings {
			m[b.Name] = cloudfl.PagesProjectDeploymentConfigsProductionVectorizeBindingsArgs{IndexName: pulumi.String(b.IndexName)}
		}
		args.VectorizeBindings = m
	}
	if len(c.AiBindings) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsProductionAiBindingsMap{}
		for _, b := range c.AiBindings {
			m[b.Name] = cloudfl.PagesProjectDeploymentConfigsProductionAiBindingsArgs{ProjectId: pulumi.String(b.ProjectId)}
		}
		args.AiBindings = m
	}
	if len(c.MtlsCertificates) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsProductionMtlsCertificatesMap{}
		for _, b := range c.MtlsCertificates {
			m[b.Name] = cloudfl.PagesProjectDeploymentConfigsProductionMtlsCertificatesArgs{CertificateId: pulumi.String(b.CertificateId)}
		}
		args.MtlsCertificates = m
	}
	if len(c.Browsers) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsProductionBrowsersMap{}
		for _, b := range c.Browsers {
			m[b.Name] = cloudfl.PagesProjectDeploymentConfigsProductionBrowsersArgs{}
		}
		args.Browsers = m
	}
	return args
}

func productionEnvVars(c *cloudflarepagesprojectv1.CloudflarePagesDeploymentConfig) cloudfl.PagesProjectDeploymentConfigsProductionEnvVarsMap {
	m := cloudfl.PagesProjectDeploymentConfigsProductionEnvVarsMap{}
	for k, v := range c.Vars {
		m[k] = cloudfl.PagesProjectDeploymentConfigsProductionEnvVarsArgs{Type: pulumi.String("plain_text"), Value: pulumi.String(v)}
	}
	for _, s := range c.Secrets {
		m[s.Name] = cloudfl.PagesProjectDeploymentConfigsProductionEnvVarsArgs{Type: pulumi.String("secret_text"), Value: pulumi.String(s.Value.GetValue())}
	}
	return m
}

func previewConfig(c *cloudflarepagesprojectv1.CloudflarePagesDeploymentConfig) cloudfl.PagesProjectDeploymentConfigsPreviewPtrInput {
	args := &cloudfl.PagesProjectDeploymentConfigsPreviewArgs{}
	if c.CompatibilityDate != "" {
		args.CompatibilityDate = pulumi.String(c.CompatibilityDate)
	}
	if len(c.CompatibilityFlags) > 0 {
		args.CompatibilityFlags = pulumi.ToStringArray(c.CompatibilityFlags)
	}
	if c.AlwaysUseLatestCompatibilityDate {
		args.AlwaysUseLatestCompatibilityDate = pulumi.Bool(true)
	}
	if c.BuildImageMajorVersion > 0 {
		args.BuildImageMajorVersion = pulumi.Int(int(c.BuildImageMajorVersion))
	}
	if c.FailOpen {
		args.FailOpen = pulumi.Bool(true)
	}
	if c.UsageModel != "" {
		args.UsageModel = pulumi.String(c.UsageModel)
	}
	if c.Limits != nil && c.Limits.CpuMs > 0 {
		args.Limits = &cloudfl.PagesProjectDeploymentConfigsPreviewLimitsArgs{CpuMs: pulumi.Int(int(c.Limits.CpuMs))}
	}
	if c.Placement != nil && c.Placement.Mode != "" {
		args.Placement = &cloudfl.PagesProjectDeploymentConfigsPreviewPlacementArgs{Mode: pulumi.String(c.Placement.Mode)}
	}
	if env := previewEnvVars(c); len(env) > 0 {
		args.EnvVars = env
	}
	if len(c.KvNamespaces) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsPreviewKvNamespacesMap{}
		for _, b := range c.KvNamespaces {
			m[b.Name] = cloudfl.PagesProjectDeploymentConfigsPreviewKvNamespacesArgs{NamespaceId: pulumi.String(b.NamespaceId.GetValue())}
		}
		args.KvNamespaces = m
	}
	if len(c.D1Databases) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsPreviewD1DatabasesMap{}
		for _, b := range c.D1Databases {
			m[b.Name] = cloudfl.PagesProjectDeploymentConfigsPreviewD1DatabasesArgs{Id: pulumi.String(b.DatabaseId.GetValue())}
		}
		args.D1Databases = m
	}
	if len(c.R2Buckets) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsPreviewR2BucketsMap{}
		for _, b := range c.R2Buckets {
			r2 := cloudfl.PagesProjectDeploymentConfigsPreviewR2BucketsArgs{Name: pulumi.String(b.BucketName.GetValue())}
			if b.Jurisdiction != "" {
				r2.Jurisdiction = pulumi.String(b.Jurisdiction)
			}
			m[b.Name] = r2
		}
		args.R2Buckets = m
	}
	if len(c.QueueProducers) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsPreviewQueueProducersMap{}
		for _, b := range c.QueueProducers {
			m[b.Name] = cloudfl.PagesProjectDeploymentConfigsPreviewQueueProducersArgs{Name: pulumi.String(b.QueueName.GetValue())}
		}
		args.QueueProducers = m
	}
	if len(c.HyperdriveBindings) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsPreviewHyperdriveBindingsMap{}
		for _, b := range c.HyperdriveBindings {
			m[b.Name] = cloudfl.PagesProjectDeploymentConfigsPreviewHyperdriveBindingsArgs{Id: pulumi.String(b.ConfigId.GetValue())}
		}
		args.HyperdriveBindings = m
	}
	if len(c.Services) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsPreviewServicesMap{}
		for _, b := range c.Services {
			svc := cloudfl.PagesProjectDeploymentConfigsPreviewServicesArgs{Service: pulumi.String(b.Service.GetValue())}
			if b.Entrypoint != "" {
				svc.Entrypoint = pulumi.String(b.Entrypoint)
			}
			if b.Environment != "" {
				svc.Environment = pulumi.String(b.Environment)
			}
			m[b.Name] = svc
		}
		args.Services = m
	}
	if len(c.DurableObjectNamespaces) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsPreviewDurableObjectNamespacesMap{}
		for _, b := range c.DurableObjectNamespaces {
			m[b.Name] = cloudfl.PagesProjectDeploymentConfigsPreviewDurableObjectNamespacesArgs{NamespaceId: pulumi.String(b.NamespaceId)}
		}
		args.DurableObjectNamespaces = m
	}
	if len(c.AnalyticsEngineDatasets) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsPreviewAnalyticsEngineDatasetsMap{}
		for _, b := range c.AnalyticsEngineDatasets {
			m[b.Name] = cloudfl.PagesProjectDeploymentConfigsPreviewAnalyticsEngineDatasetsArgs{Dataset: pulumi.String(b.Dataset)}
		}
		args.AnalyticsEngineDatasets = m
	}
	if len(c.VectorizeBindings) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsPreviewVectorizeBindingsMap{}
		for _, b := range c.VectorizeBindings {
			m[b.Name] = cloudfl.PagesProjectDeploymentConfigsPreviewVectorizeBindingsArgs{IndexName: pulumi.String(b.IndexName)}
		}
		args.VectorizeBindings = m
	}
	if len(c.AiBindings) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsPreviewAiBindingsMap{}
		for _, b := range c.AiBindings {
			m[b.Name] = cloudfl.PagesProjectDeploymentConfigsPreviewAiBindingsArgs{ProjectId: pulumi.String(b.ProjectId)}
		}
		args.AiBindings = m
	}
	if len(c.MtlsCertificates) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsPreviewMtlsCertificatesMap{}
		for _, b := range c.MtlsCertificates {
			m[b.Name] = cloudfl.PagesProjectDeploymentConfigsPreviewMtlsCertificatesArgs{CertificateId: pulumi.String(b.CertificateId)}
		}
		args.MtlsCertificates = m
	}
	if len(c.Browsers) > 0 {
		m := cloudfl.PagesProjectDeploymentConfigsPreviewBrowsersMap{}
		for _, b := range c.Browsers {
			m[b.Name] = cloudfl.PagesProjectDeploymentConfigsPreviewBrowsersArgs{}
		}
		args.Browsers = m
	}
	return args
}

func previewEnvVars(c *cloudflarepagesprojectv1.CloudflarePagesDeploymentConfig) cloudfl.PagesProjectDeploymentConfigsPreviewEnvVarsMap {
	m := cloudfl.PagesProjectDeploymentConfigsPreviewEnvVarsMap{}
	for k, v := range c.Vars {
		m[k] = cloudfl.PagesProjectDeploymentConfigsPreviewEnvVarsArgs{Type: pulumi.String("plain_text"), Value: pulumi.String(v)}
	}
	for _, s := range c.Secrets {
		m[s.Name] = cloudfl.PagesProjectDeploymentConfigsPreviewEnvVarsArgs{Type: pulumi.String("secret_text"), Value: pulumi.String(s.Value.GetValue())}
	}
	return m
}
