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

	// source_dir is omitted (nil) when the spec leaves source_directory unset
	// so App Platform reads project.yml from the repository root -- sending ""
	// would name a directory that does not exist.
	fn := digitalocean.AppSpecFunctionArgs{
		Name:      pulumi.String(spec.GetFunctionName()),
		SourceDir: strPtr(spec.GetSourceDirectory()),
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

	// The App Platform app is named from spec.app_name, never metadata.name:
	// the API caps app names at 32 characters and requires them to be unique
	// across the account, neither of which a Planton metadata name guarantees.
	appSpec := digitalocean.AppSpecArgs{
		Name:      pulumi.String(spec.GetAppName()),
		Region:    pulumi.String(spec.GetRegion().String()),
		Functions: digitalocean.AppSpecFunctionArray{fn},
	}

	args := &digitalocean.AppArgs{Spec: appSpec}
	if spec.GetProjectId() != "" {
		args.ProjectId = pulumi.String(spec.GetProjectId())
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

// functionAlerts wires component alerts including their email / Slack
// destinations (carried by the SDK since pulumi-digitalocean v4.53.0; the
// provider applies them through a side-channel call after the app exists and
// reads them back from ListAlerts). An empty destinations block is omitted so
// the provider never issues a clearing call. Webhook URLs are (sensitive) on
// the spec, so they are wrapped as Pulumi secrets -- the SDK does not flag
// them itself.
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
		if d := a.GetDestinations(); d != nil && (len(d.GetEmails()) > 0 || len(d.GetSlackWebhooks()) > 0) {
			hooks := digitalocean.AppSpecFunctionAlertDestinationsSlackWebhookArray{}
			for _, h := range d.GetSlackWebhooks() {
				hooks = append(hooks, digitalocean.AppSpecFunctionAlertDestinationsSlackWebhookArgs{
					Channel: pulumi.String(h.GetChannel()),
					Url:     pulumi.ToSecret(pulumi.String(h.GetUrl())).(pulumi.StringOutput),
				})
			}
			args.Destinations = &digitalocean.AppSpecFunctionAlertDestinationsArgs{
				Emails:        pulumi.ToStringArray(d.GetEmails()),
				SlackWebhooks: hooks,
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
