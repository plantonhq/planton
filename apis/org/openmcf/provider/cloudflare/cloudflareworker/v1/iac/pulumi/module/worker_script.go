package module

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
	cloudfl "github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// worker provisions the Worker script and its routing, schedules, and settings.
func worker(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudfl.Provider,
	r2Provider *aws.Provider,
) error {
	spec := locals.CloudflareWorker.Spec

	// Compatibility date defaults to today when unset.
	compatibilityDate := spec.CompatibilityDate
	if compatibilityDate == "" {
		compatibilityDate = time.Now().UTC().Format("2006-01-02")
	}

	scriptArgs := &cloudfl.WorkersScriptArgs{
		AccountId:         pulumi.String(spec.AccountId),
		ScriptName:        pulumi.String(spec.WorkerName),
		MainModule:        pulumi.String(spec.MainModule),
		CompatibilityDate: pulumi.String(compatibilityDate),
		Bindings:          buildBindings(spec),
	}

	// Script source: inline content, else the R2 bundle body.
	if spec.GetContent() != "" {
		scriptArgs.Content = pulumi.StringPtr(spec.GetContent())
	} else if bundle := spec.GetR2Bundle(); bundle != nil {
		obj := s3.GetObjectOutput(ctx, s3.GetObjectOutputArgs{
			Bucket: pulumi.String(bundle.Bucket),
			Key:    pulumi.String(bundle.Path),
		}, pulumi.Provider(r2Provider))
		scriptArgs.Content = obj.Body().ApplyT(func(s string) *string { return &s }).(pulumi.StringPtrOutput)
	}

	if len(spec.CompatibilityFlags) > 0 {
		flags := make(pulumi.StringArray, 0, len(spec.CompatibilityFlags))
		for _, f := range spec.CompatibilityFlags {
			flags = append(flags, pulumi.String(f))
		}
		scriptArgs.CompatibilityFlags = flags
	}

	if o := spec.Observability; o != nil {
		obsArgs := &cloudfl.WorkersScriptObservabilityArgs{Enabled: pulumi.Bool(o.Enabled)}
		if o.HeadSamplingRate > 0 {
			obsArgs.HeadSamplingRate = pulumi.Float64(o.HeadSamplingRate)
		}
		scriptArgs.Observability = obsArgs
	}

	if p := spec.Placement; p != nil && p.Mode != "" {
		scriptArgs.Placement = &cloudfl.WorkersScriptPlacementArgs{Mode: pulumi.String(p.Mode)}
	}

	if l := spec.Limits; l != nil && (l.CpuMs > 0 || l.Subrequests > 0) {
		limitsArgs := &cloudfl.WorkersScriptLimitsArgs{}
		if l.CpuMs > 0 {
			limitsArgs.CpuMs = pulumi.Int(int(l.CpuMs))
		}
		if l.Subrequests > 0 {
			limitsArgs.Subrequests = pulumi.Int(int(l.Subrequests))
		}
		scriptArgs.Limits = limitsArgs
	}

	if spec.Logpush {
		scriptArgs.Logpush = pulumi.Bool(true)
	}

	if len(spec.TailConsumers) > 0 {
		var tc cloudfl.WorkersScriptTailConsumerArray
		for _, t := range spec.TailConsumers {
			a := cloudfl.WorkersScriptTailConsumerArgs{Service: pulumi.String(t.Service)}
			if t.Environment != "" {
				a.Environment = pulumi.String(t.Environment)
			}
			if t.Namespace != "" {
				a.Namespace = pulumi.String(t.Namespace)
			}
			tc = append(tc, a)
		}
		scriptArgs.TailConsumers = tc
	}

	createdScript, err := cloudfl.NewWorkersScript(ctx, "workers-script", scriptArgs, pulumi.Provider(cloudflareProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create workers script")
	}

	// workers.dev subdomain.
	if wd := spec.WorkersDev; wd != nil && wd.Enabled {
		if _, err := cloudfl.NewWorkersScriptSubdomain(ctx, "workers-dev", &cloudfl.WorkersScriptSubdomainArgs{
			AccountId:       pulumi.String(spec.AccountId),
			ScriptName:      createdScript.ScriptName,
			Enabled:         pulumi.Bool(true),
			PreviewsEnabled: pulumi.Bool(wd.PreviewsEnabled),
		}, pulumi.Provider(cloudflareProvider)); err != nil {
			return errors.Wrap(err, "failed to create workers.dev subdomain")
		}
	}

	// Managed custom domains.
	customDomainHostnames := make(pulumi.StringArray, 0, len(spec.CustomDomains))
	for i, cd := range spec.CustomDomains {
		cdArgs := &cloudfl.WorkersCustomDomainArgs{
			AccountId:   pulumi.String(spec.AccountId),
			Environment: pulumi.String("production"),
			Hostname:    pulumi.String(cd.Hostname),
			Service:     createdScript.ScriptName,
		}
		// Zone is optional — Cloudflare infers it from the hostname when omitted.
		if cd.ZoneId != nil && cd.ZoneId.GetValue() != "" {
			cdArgs.ZoneId = pulumi.String(cd.ZoneId.GetValue())
		}
		if _, err := cloudfl.NewWorkersCustomDomain(ctx, fmt.Sprintf("custom-domain-%d", i), cdArgs, pulumi.Provider(cloudflareProvider)); err != nil {
			return errors.Wrap(err, "failed to create workers custom domain")
		}
		customDomainHostnames = append(customDomainHostnames, pulumi.String(cd.Hostname))
	}

	// Pattern-based routes.
	routePatterns := make(pulumi.StringArray, 0, len(spec.Routes))
	for i, r := range spec.Routes {
		zoneId := ""
		if r.ZoneId != nil {
			zoneId = r.ZoneId.GetValue()
		}
		if _, err := cloudfl.NewWorkersRoute(ctx, fmt.Sprintf("workers-route-%d", i), &cloudfl.WorkersRouteArgs{
			ZoneId:  pulumi.String(zoneId),
			Pattern: pulumi.String(r.Pattern),
			Script:  createdScript.ScriptName,
		}, pulumi.Provider(cloudflareProvider)); err != nil {
			return errors.Wrap(err, "failed to create workers route")
		}
		routePatterns = append(routePatterns, pulumi.String(r.Pattern))
	}

	// Cron-triggered invocations.
	if len(spec.Schedules) > 0 {
		var schedules cloudfl.WorkersCronTriggerScheduleArray
		for _, s := range spec.Schedules {
			schedules = append(schedules, cloudfl.WorkersCronTriggerScheduleArgs{Cron: pulumi.String(s)})
		}
		if _, err := cloudfl.NewWorkersCronTrigger(ctx, "cron-trigger", &cloudfl.WorkersCronTriggerArgs{
			AccountId:  pulumi.String(spec.AccountId),
			ScriptName: createdScript.ScriptName,
			Schedules:  schedules,
		}, pulumi.Provider(cloudflareProvider)); err != nil {
			return errors.Wrap(err, "failed to create workers cron trigger")
		}
	}

	ctx.Export(OpScriptId, createdScript.ID())
	ctx.Export(OpScriptName, createdScript.ScriptName)
	ctx.Export(OpCustomDomainHostnames, customDomainHostnames)
	ctx.Export(OpRoutePatterns, routePatterns)

	return nil
}
