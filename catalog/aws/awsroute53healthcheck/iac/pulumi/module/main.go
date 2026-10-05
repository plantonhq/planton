package module

import (
	"github.com/pkg/errors"
	awsroute53healthcheckv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsroute53healthcheck/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates one Route 53 health check. The spec's CEL rules guarantee
// the arguments already match the chosen check_type (endpoint probing,
// calculated aggregation, CloudWatch mirroring, or recovery-control
// mirroring), so this module maps fields 1:1 without re-validating.
func Resources(ctx *pulumi.Context, iacInput *awsroute53healthcheckv1alpha1.AwsRoute53HealthCheckIacInput) error {
	locals := initializeLocals(ctx, iacInput)
	spec := locals.AwsRoute53HealthCheck.Spec

	// Build the AWS provider from the IaC input via the shared builder, which
	// resolves the right credential mechanism (static keys, keyless web
	// identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// The monitoring model (ForceNew).
	args := &route53.HealthCheckArgs{
		Type: pulumi.String(spec.CheckType),
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}

	// The check type gates which argument families apply (CEL-enforced in
	// the spec); the probe-tuning family is sent only for endpoint checks so
	// both engines render identically.
	isEndpointCheck := map[string]bool{
		"HTTP": true, "HTTPS": true, "HTTP_STR_MATCH": true, "HTTPS_STR_MATCH": true, "TCP": true,
	}[spec.CheckType]

	// --- Endpoint checks (HTTP / HTTPS / *_STR_MATCH / TCP) ------------------
	// fqdn alone: Route 53 resolves and probes it (and sends it as the Host
	// header). ip_address alone or with fqdn: the probe goes to the IP.
	// Unset optionals stay nil so the provider applies its own defaults
	// (port 80/443, resource path "/", all checker regions).
	if spec.Fqdn != "" {
		args.Fqdn = pulumi.String(spec.Fqdn)
	}
	if spec.IpAddress != "" {
		args.IpAddress = pulumi.String(spec.IpAddress)
	}
	if spec.Port != 0 {
		args.Port = pulumi.Int(int(spec.Port))
	}
	if spec.ResourcePath != "" {
		args.ResourcePath = pulumi.String(spec.ResourcePath)
	}
	if spec.SearchString != "" {
		args.SearchString = pulumi.String(spec.SearchString)
	}
	// request_interval and measure_latency are create-time (ForceNew).
	// Probe tuning only exists where probing happens (endpoint checks) —
	// mirroring the Terraform module's type gate check-for-check.
	if isEndpointCheck {
		if spec.RequestInterval != nil {
			args.RequestInterval = pulumi.Int(int(spec.GetRequestInterval()))
		}
		if spec.FailureThreshold != nil {
			args.FailureThreshold = pulumi.Int(int(spec.GetFailureThreshold()))
		}
		// State-pinned like Terraform: an explicit false is sent, so a
		// true -> false edit is a real update on both engines.
		args.MeasureLatency = pulumi.Bool(spec.MeasureLatency)
		if spec.EnableSni != nil {
			args.EnableSni = pulumi.Bool(spec.GetEnableSni())
		}
	}
	if len(spec.Regions) > 0 {
		args.Regions = pulumi.ToStringArray(spec.Regions)
	}

	// --- State shaping (any type) ---------------------------------------------
	// State-pinned like Terraform (both booleans are sent unconditionally),
	// so switching either back to false is a real update on both engines.
	args.InvertHealthcheck = pulumi.Bool(spec.InvertHealthcheck)
	args.Disabled = pulumi.Bool(spec.Disabled)

	// --- CALCULATED: aggregate child checks -----------------------------------
	// Presence carries the AWS contract: an explicit 0 means "always
	// healthy", a value above the child count means "always unhealthy", and
	// omitting the field lets AWS apply its server-side default.
	if len(spec.ChildHealthChecks) > 0 {
		childIds := make([]string, 0, len(spec.ChildHealthChecks))
		for _, child := range spec.ChildHealthChecks {
			childIds = append(childIds, child.GetValue())
		}
		args.ChildHealthchecks = pulumi.ToStringArray(childIds)
	}
	if spec.ChildHealthThreshold != nil {
		args.ChildHealthThreshold = pulumi.Int(int(spec.GetChildHealthThreshold()))
	}

	// --- CLOUDWATCH_METRIC: mirror an alarm (the private-resource pattern) ----
	if spec.CloudwatchAlarmName != "" {
		args.CloudwatchAlarmName = pulumi.String(spec.CloudwatchAlarmName)
	}
	if spec.CloudwatchAlarmRegion != "" {
		args.CloudwatchAlarmRegion = pulumi.String(spec.CloudwatchAlarmRegion)
	}
	if spec.InsufficientDataHealthStatus != "" {
		args.InsufficientDataHealthStatus = pulumi.String(spec.InsufficientDataHealthStatus)
	}

	// --- RECOVERY_CONTROL: mirror an ARC routing control (ForceNew) -----------
	if spec.RoutingControlArn != "" {
		args.RoutingControlArn = pulumi.String(spec.RoutingControlArn)
	}

	createdHealthCheck, err := route53.NewHealthCheck(ctx,
		locals.AwsRoute53HealthCheck.Metadata.Name,
		args,
		pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create health check")
	}

	ctx.Export(OpHealthCheckId, createdHealthCheck.ID())
	ctx.Export(OpHealthCheckArn, createdHealthCheck.Arn)

	return nil
}
