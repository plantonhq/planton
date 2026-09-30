package module

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/pkg/errors"
	cloudflareworkerv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareworker/v1alpha1"
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
		CompatibilityDate: pulumi.String(compatibilityDate),
		Bindings:          buildBindings(spec),
	}

	// Script source: inline content, else the R2 bundle body. A worker may instead
	// be a pure static site (assets only, no script) — main_module is then omitted.
	// body_part marks service-worker syntax and is mutually exclusive with main_module.
	// main_module defaults to "index.js" when unset (mirrors the tofu module):
	// an empty main_module makes Cloudflare treat the upload as a legacy
	// service-worker script, which rejects ES-module syntax at deploy with
	// "Uncaught SyntaxError: Unexpected token 'export'" (measured live).
	mainModule := spec.MainModule
	if mainModule == "" {
		mainModule = "index.js"
	}
	if spec.GetContent() != "" {
		scriptArgs.Content = pulumi.StringPtr(spec.GetContent())
		if spec.BodyPart != "" {
			scriptArgs.BodyPart = pulumi.String(spec.BodyPart)
		} else {
			scriptArgs.MainModule = pulumi.String(mainModule)
		}
	} else if bundle := spec.GetR2Bundle(); bundle != nil {
		scriptArgs.Content = bundleContent(ctx, r2Provider, fkValue(bundle.Bucket), bundle.Path)
		if spec.BodyPart != "" {
			scriptArgs.BodyPart = pulumi.String(spec.BodyPart)
		} else {
			scriptArgs.MainModule = pulumi.String(mainModule)
		}
	}

	if spec.ContentType != "" {
		scriptArgs.ContentType = pulumi.String(spec.ContentType)
	}

	// Workers Static Assets: upload a built site directory served from the edge.
	if a := spec.Assets; a != nil {
		assetsArgs := &cloudfl.WorkersScriptAssetsArgs{
			Directory: pulumi.String(a.Directory),
		}
		if c := a.Config; c != nil {
			cfgArgs := &cloudfl.WorkersScriptAssetsConfigArgs{}
			if c.HtmlHandling != "" {
				cfgArgs.HtmlHandling = pulumi.String(c.HtmlHandling)
			}
			if c.NotFoundHandling != "" {
				cfgArgs.NotFoundHandling = pulumi.String(c.NotFoundHandling)
			}
			if c.Headers != "" {
				cfgArgs.Headers = pulumi.String(c.Headers)
			}
			if c.Redirects != "" {
				cfgArgs.Redirects = pulumi.String(c.Redirects)
			}
			// run_worker_first is the provider's dynamic field: a list of path
			// rules when provided, else a bool toggle.
			if len(c.RunWorkerFirstRules) > 0 {
				cfgArgs.RunWorkerFirst = pulumi.ToStringArray(c.RunWorkerFirstRules)
			} else if c.RunWorkerFirst {
				cfgArgs.RunWorkerFirst = pulumi.Bool(true)
			}
			assetsArgs.Config = cfgArgs
		}
		scriptArgs.Assets = assetsArgs
	}

	if len(spec.CompatibilityFlags) > 0 {
		flags := make(pulumi.StringArray, 0, len(spec.CompatibilityFlags))
		for _, f := range spec.CompatibilityFlags {
			flags = append(flags, pulumi.String(f))
		}
		scriptArgs.CompatibilityFlags = flags
	}

	if spec.KeepAssets {
		scriptArgs.KeepAssets = pulumi.Bool(true)
	}
	if len(spec.KeepBindings) > 0 {
		scriptArgs.KeepBindings = pulumi.ToStringArray(spec.KeepBindings)
	}
	if spec.UsageModel != "" {
		scriptArgs.UsageModel = pulumi.String(spec.UsageModel)
	}

	if m := spec.Migrations; m != nil {
		scriptArgs.Migrations = buildMigrations(m)
	}

	if o := spec.Observability; o != nil {
		obsArgs := &cloudfl.WorkersScriptObservabilityArgs{Enabled: pulumi.Bool(o.Enabled)}
		if o.HeadSamplingRate > 0 {
			obsArgs.HeadSamplingRate = pulumi.Float64(o.HeadSamplingRate)
		}
		if l := o.Logs; l != nil {
			logsArgs := &cloudfl.WorkersScriptObservabilityLogsArgs{
				Enabled:        pulumi.Bool(l.Enabled),
				InvocationLogs: pulumi.Bool(l.InvocationLogs),
			}
			if len(l.Destinations) > 0 {
				logsArgs.Destinations = pulumi.ToStringArray(l.Destinations)
			}
			if l.HeadSamplingRate > 0 {
				logsArgs.HeadSamplingRate = pulumi.Float64(l.HeadSamplingRate)
			}
			if l.Persist {
				logsArgs.Persist = pulumi.Bool(true)
			}
			obsArgs.Logs = logsArgs
		}
		if t := o.Traces; t != nil {
			// PARITY-EXCEPTION: pulumi-cloudflare SDK v6.17.0 has no
			// PropagationPolicy on traces — tofu honors it, this engine skips it.
			tracesArgs := &cloudfl.WorkersScriptObservabilityTracesArgs{
				Enabled: pulumi.Bool(t.Enabled),
			}
			if len(t.Destinations) > 0 {
				tracesArgs.Destinations = pulumi.ToStringArray(t.Destinations)
			}
			if t.HeadSamplingRate > 0 {
				tracesArgs.HeadSamplingRate = pulumi.Float64(t.HeadSamplingRate)
			}
			if t.Persist {
				tracesArgs.Persist = pulumi.Bool(true)
			}
			obsArgs.Traces = tracesArgs
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
			a := cloudfl.WorkersScriptTailConsumerArgs{Service: pulumi.String(fkValue(t.Service))}
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

	if a := spec.Annotations; a != nil && (a.WorkersMessage != "" || a.WorkersTag != "") {
		ann := &cloudfl.WorkersScriptAnnotationsArgs{}
		if a.WorkersMessage != "" {
			ann.WorkersMessage = pulumi.String(a.WorkersMessage)
		}
		if a.WorkersTag != "" {
			ann.WorkersTag = pulumi.String(a.WorkersTag)
		}
		scriptArgs.Annotations = ann
	}

	// PARITY-EXCEPTION: pulumi-cloudflare SDK v6.17.0 has no CacheOptions,
	// Exports, or PackageDependencies inputs. tofu honors them; this engine
	// skips them so a set field is never silently dropped without a trail.
	if spec.CacheOptions != nil {
		_ = ctx.Log.Warn("PARITY-EXCEPTION: cache_options is absent from pulumi-cloudflare SDK v6.17.0; tofu honors it, this engine skips it", nil)
	}
	if len(spec.Exports) > 0 {
		_ = ctx.Log.Warn("PARITY-EXCEPTION: exports is absent from pulumi-cloudflare SDK v6.17.0; tofu honors it, this engine skips it", nil)
	}
	if len(spec.PackageDependencies) > 0 {
		_ = ctx.Log.Warn("PARITY-EXCEPTION: package_dependencies is absent from pulumi-cloudflare SDK v6.17.0; tofu honors it, this engine skips it", nil)
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

	// Managed custom domains. environment is deprecated on the provider and
	// is omitted — Cloudflare defaults the hostname to production.
	customDomainHostnames := make(pulumi.StringArray, 0, len(spec.CustomDomains))
	customDomainIds := pulumi.StringMap{}
	for i, cd := range spec.CustomDomains {
		cdArgs := &cloudfl.WorkersCustomDomainArgs{
			AccountId: pulumi.String(spec.AccountId),
			Hostname:  pulumi.String(cd.Hostname),
			Service:   createdScript.ScriptName,
		}
		// Zone is optional — Cloudflare infers it from the hostname when omitted.
		if v := fkValue(cd.ZoneId); v != "" {
			cdArgs.ZoneId = pulumi.String(v)
		}
		created, err := cloudfl.NewWorkersCustomDomain(ctx, fmt.Sprintf("custom-domain-%d", i), cdArgs, pulumi.Provider(cloudflareProvider))
		if err != nil {
			return errors.Wrap(err, "failed to create workers custom domain")
		}
		customDomainHostnames = append(customDomainHostnames, pulumi.String(cd.Hostname))
		customDomainIds[cd.Hostname] = created.ID()
	}

	// Pattern-based routes, keyed by list index so import can reassemble
	// {zone_id}/{route_id} from the two maps.
	routePatterns := make(pulumi.StringArray, 0, len(spec.Routes))
	routeIds := pulumi.StringMap{}
	routeZoneIds := pulumi.StringMap{}
	for i, r := range spec.Routes {
		zoneId := fkValue(r.ZoneId)
		created, err := cloudfl.NewWorkersRoute(ctx, fmt.Sprintf("workers-route-%d", i), &cloudfl.WorkersRouteArgs{
			ZoneId:  pulumi.String(zoneId),
			Pattern: pulumi.String(r.Pattern),
			Script:  createdScript.ScriptName,
		}, pulumi.Provider(cloudflareProvider))
		if err != nil {
			return errors.Wrap(err, "failed to create workers route")
		}
		key := fmt.Sprintf("%d", i)
		routePatterns = append(routePatterns, pulumi.String(r.Pattern))
		routeIds[key] = created.ID()
		routeZoneIds[key] = pulumi.String(zoneId)
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
	ctx.Export(OpCustomDomainIds, customDomainIds)
	ctx.Export(OpRouteIds, routeIds)
	ctx.Export(OpRouteZoneIds, routeZoneIds)

	return nil
}

// bundleContent reads the Worker's pre-built bundle from R2 and returns its text.
//
// DownloadBody returns the object's bytes as BodyBase64 whatever its Content-Type. The
// plain Body is filled only for a short list of "readable" types, which excludes
// application/javascript -- the type many upload tools give a .js file -- so a bundle
// read through it can come back empty. An empty bundle is refused here, naming the
// object, instead of reaching Cloudflare as a Worker with no code. Mirrors the tofu
// module's download_body read and its postcondition.
func bundleContent(ctx *pulumi.Context, r2Provider *aws.Provider, bucket, key string) pulumi.StringPtrOutput {
	obj := s3.GetObjectOutput(ctx, s3.GetObjectOutputArgs{
		Bucket:       pulumi.String(bucket),
		Key:          pulumi.String(key),
		DownloadBody: pulumi.String("true"),
	}, pulumi.Provider(r2Provider))

	return obj.BodyBase64().ApplyT(func(encoded string) (*string, error) {
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode the Worker bundle r2://%s/%s", bucket, key)
		}
		if len(decoded) == 0 {
			return nil, errors.Errorf("the Worker bundle r2://%s/%s is empty: upload the built script to that key, "+
				"or point spec.r2_bundle.path at the object that holds it", bucket, key)
		}
		content := string(decoded)
		return &content, nil
	}).(pulumi.StringPtrOutput)
}

func buildMigrations(m *cloudflareworkerv1alpha1.CloudflareWorkerMigrations) *cloudfl.WorkersScriptMigrationsArgs {
	args := &cloudfl.WorkersScriptMigrationsArgs{}
	if len(m.DeletedClasses) > 0 {
		args.DeletedClasses = pulumi.ToStringArray(m.DeletedClasses)
	}
	if len(m.NewClasses) > 0 {
		args.NewClasses = pulumi.ToStringArray(m.NewClasses)
	}
	if len(m.NewSqliteClasses) > 0 {
		args.NewSqliteClasses = pulumi.ToStringArray(m.NewSqliteClasses)
	}
	if m.NewTag != "" {
		args.NewTag = pulumi.String(m.NewTag)
	}
	if m.OldTag != "" {
		args.OldTag = pulumi.String(m.OldTag)
	}
	if len(m.RenamedClasses) > 0 {
		var renamed cloudfl.WorkersScriptMigrationsRenamedClassArray
		for _, r := range m.RenamedClasses {
			renamed = append(renamed, cloudfl.WorkersScriptMigrationsRenamedClassArgs{
				From: pulumi.String(r.From),
				To:   pulumi.String(r.To),
			})
		}
		args.RenamedClasses = renamed
	}
	if len(m.TransferredClasses) > 0 {
		args.TransferredClasses = buildTransferred(m.TransferredClasses)
	}
	if len(m.Steps) > 0 {
		var steps cloudfl.WorkersScriptMigrationsStepArray
		for _, s := range m.Steps {
			step := cloudfl.WorkersScriptMigrationsStepArgs{}
			if len(s.DeletedClasses) > 0 {
				step.DeletedClasses = pulumi.ToStringArray(s.DeletedClasses)
			}
			if len(s.NewClasses) > 0 {
				step.NewClasses = pulumi.ToStringArray(s.NewClasses)
			}
			if len(s.NewSqliteClasses) > 0 {
				step.NewSqliteClasses = pulumi.ToStringArray(s.NewSqliteClasses)
			}
			if len(s.RenamedClasses) > 0 {
				var renamed cloudfl.WorkersScriptMigrationsStepRenamedClassArray
				for _, r := range s.RenamedClasses {
					renamed = append(renamed, cloudfl.WorkersScriptMigrationsStepRenamedClassArgs{
						From: pulumi.String(r.From),
						To:   pulumi.String(r.To),
					})
				}
				step.RenamedClasses = renamed
			}
			if len(s.TransferredClasses) > 0 {
				var transferred cloudfl.WorkersScriptMigrationsStepTransferredClassArray
				for _, t := range s.TransferredClasses {
					ta := cloudfl.WorkersScriptMigrationsStepTransferredClassArgs{
						From: pulumi.String(t.From),
						To:   pulumi.String(t.To),
					}
					if v := fkValue(t.FromScript); v != "" {
						ta.FromScript = pulumi.String(v)
					}
					transferred = append(transferred, ta)
				}
				step.TransferredClasses = transferred
			}
			steps = append(steps, step)
		}
		args.Steps = steps
	}
	return args
}

func buildTransferred(in []*cloudflareworkerv1alpha1.CloudflareWorkerTransferredClass) cloudfl.WorkersScriptMigrationsTransferredClassArray {
	var out cloudfl.WorkersScriptMigrationsTransferredClassArray
	for _, t := range in {
		ta := cloudfl.WorkersScriptMigrationsTransferredClassArgs{
			From: pulumi.String(t.From),
			To:   pulumi.String(t.To),
		}
		if v := fkValue(t.FromScript); v != "" {
			ta.FromScript = pulumi.String(v)
		}
		out = append(out, ta)
	}
	return out
}
