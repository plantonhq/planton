package module

import (
	"strings"

	"github.com/pkg/errors"
	do "github.com/plantonhq/planton/catalog/digitalocean"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func function(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.App, error) {
	spec := locals.DigitalOceanFunction.Spec

	for _, a := range spec.GetAlerts() {
		if a.GetDestinations() != nil &&
			(len(a.GetDestinations().GetEmails()) > 0 || len(a.GetDestinations().GetSlackWebhooks()) > 0) {
			return nil, errors.New("PARITY-EXCEPTION: alert destinations (emails / slack webhooks) are modeled and Terraform wires them; the Pulumi DigitalOcean SDK v4.49.0 has no destinations field on function alerts. Re-evaluate when the SDK exposes alert destinations.")
		}
	}

	fn := digitalocean.AppSpecFunctionArgs{
		Name:      pulumi.String(spec.GetFunctionName()),
		SourceDir: pulumi.String(spec.GetSourceDirectory()),
		Envs:      functionEnvs(spec.GetEnvs()),
		Alerts:    functionAlerts(spec.GetAlerts()),
	}

	if g := spec.GetGit(); g != nil {
		fn.Git = &digitalocean.AppSpecFunctionGitArgs{
			RepoCloneUrl: pulumi.String(g.GetRepoCloneUrl()),
			Branch:       pulumi.String(g.GetBranch()),
		}
	}
	if g := spec.GetGithub(); g != nil {
		fn.Github = &digitalocean.AppSpecFunctionGithubArgs{
			Repo:         pulumi.String(g.GetRepo()),
			Branch:       pulumi.String(g.GetBranch()),
			DeployOnPush: pulumi.Bool(g.GetDeployOnPush()),
		}
	}
	if g := spec.GetGitlab(); g != nil {
		fn.Gitlab = &digitalocean.AppSpecFunctionGitlabArgs{
			Repo:         pulumi.String(g.GetRepo()),
			Branch:       pulumi.String(g.GetBranch()),
			DeployOnPush: pulumi.Bool(g.GetDeployOnPush()),
		}
	}
	if g := spec.GetBitbucket(); g != nil {
		fn.Bitbucket = &digitalocean.AppSpecFunctionBitbucketArgs{
			Repo:         pulumi.String(g.GetRepo()),
			Branch:       pulumi.String(g.GetBranch()),
			DeployOnPush: pulumi.Bool(g.GetDeployOnPush()),
		}
	}
	fn.LogDestinations = functionLogs(spec.GetLogDestinations())

	appSpec := digitalocean.AppSpecArgs{
		Name:      pulumi.String(locals.DigitalOceanFunction.Metadata.Name),
		Region:    pulumi.String(spec.GetRegion().String()),
		Functions: digitalocean.AppSpecFunctionArray{fn},
	}

	args := &digitalocean.AppArgs{Spec: appSpec}
	if projectId := spec.GetProjectId().GetValue(); projectId != "" {
		args.ProjectId = pulumi.String(projectId)
	}

	created, err := digitalocean.NewApp(ctx, "function", args, pulumi.Provider(digitalOceanProvider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean app for function")
	}

	ctx.Export(OpFunctionId, created.ID())
	ctx.Export(OpHttpsEndpoint, created.LiveUrl)
	ctx.Export(OpDefaultHostname, created.DefaultIngress.ApplyT(func(u string) string {
		u = strings.TrimPrefix(u, "https://")
		return strings.TrimPrefix(u, "http://")
	}).(pulumi.StringOutput))

	return created, nil
}

func providerEnum(enumName string) string {
	if enumName == "" || strings.HasSuffix(enumName, "_unspecified") {
		return ""
	}
	return strings.ToUpper(enumName)
}

func strPtr(s string) pulumi.StringPtrInput {
	if s == "" {
		return nil
	}
	return pulumi.StringPtr(s)
}

func functionEnvs(envs []*do.DigitalOceanAppEnvVar) digitalocean.AppSpecFunctionEnvArray {
	out := digitalocean.AppSpecFunctionEnvArray{}
	for _, e := range envs {
		typ, value := "GENERAL", e.GetPlaintext()
		if e.GetSecret() != "" {
			typ, value = "SECRET", e.GetSecret()
		}
		scope := providerEnum(e.GetScope().String())
		if scope == "" {
			scope = "RUN_AND_BUILD_TIME"
		}
		out = append(out, digitalocean.AppSpecFunctionEnvArgs{
			Key:   pulumi.String(e.GetKey()),
			Value: pulumi.String(value),
			Type:  pulumi.String(typ),
			Scope: pulumi.String(scope),
		})
	}
	return out
}

func functionAlerts(in []*do.DigitalOceanAppComponentAlert) digitalocean.AppSpecFunctionAlertArray {
	out := digitalocean.AppSpecFunctionAlertArray{}
	for _, a := range in {
		out = append(out, digitalocean.AppSpecFunctionAlertArgs{
			Rule:     pulumi.String(providerEnum(a.GetRule().String())),
			Operator: pulumi.String(providerEnum(a.GetOperator().String())),
			Window:   pulumi.String(providerEnum(a.GetWindow().String())),
			Value:    pulumi.Float64(a.GetValue()),
			Disabled: pulumi.Bool(a.GetDisabled()),
		})
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
				Endpoint:    strPtr(o.GetEndpoint()),
				IndexName:   strPtr(o.GetIndexName()),
				ClusterName: strPtr(o.GetClusterName()),
				BasicAuth: digitalocean.AppSpecFunctionLogDestinationOpenSearchBasicAuthArgs{
					User: strPtr(user), Password: strPtr(pass),
				},
			}
		}
		out = append(out, args)
	}
	return out
}
