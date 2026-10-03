package module

import (
	"github.com/pkg/errors"
	awsredshiftserverlessworkgroupv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsredshiftserverlessworkgroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the Redshift Serverless workgroup -- the compute
// plane of the serverless warehouse. The workgroup composes onto its
// neighbors instead of embedding them: the namespace it serves, the
// subnets it places compute in, and the security groups on its endpoint
// all attach by reference, and warehouse ingress rules live on the
// referenced AwsSecurityGroup nodes -- this module never creates or
// mutates a resource that deserves to be its own node.
func Resources(ctx *pulumi.Context, iacInput *awsredshiftserverlessworkgroupv1alpha1.AwsRedshiftServerlessWorkgroupIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static keys, keyless
	// web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsRedshiftServerlessWorkgroup.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdWorkgroup, err := workgroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create Redshift Serverless workgroup")
	}

	createdCustomDomain, err := customDomain(ctx, locals, provider, createdWorkgroup)
	if err != nil {
		return errors.Wrap(err, "failed to associate custom domain")
	}

	// Satellite groups apply in a fixed serial order (custom domain ->
	// endpoint accesses -> usage limits): the endpoint-access create
	// conflicts unless the workgroup is idle, a usage-limit call keeps
	// the workgroup busy for ~15-30s after returning, and the provider
	// retries the ConflictException only on the workgroup's own
	// delete/update -- so the conflict-sensitive create runs first
	// (straight after the provider's own wait-for-available) and the
	// conflict-immune limit calls run last. On destroy the endpoints
	// ride the workgroup's cascading, conflict-retried delete via
	// DeletedWith (contract and parity exception in endpoint_access.go).
	// Terraform mirrors the order via depends_on and protects its
	// destroy crossing with a time_sleep settle.
	satelliteDeps := []pulumi.Resource{}
	if createdCustomDomain != nil {
		satelliteDeps = append(satelliteDeps, createdCustomDomain)
	}

	createdEndpointAddresses, createdEndpoints, err := endpointAccesses(ctx, locals, provider, createdWorkgroup, satelliteDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create endpoint accesses")
	}

	createdUsageLimitIds, err := usageLimits(ctx, locals, provider, createdWorkgroup,
		append(satelliteDeps, createdEndpoints...))
	if err != nil {
		return errors.Wrap(err, "failed to create usage limits")
	}

	ctx.Export(OpWorkgroupName, createdWorkgroup.WorkgroupName)
	ctx.Export(OpWorkgroupId, createdWorkgroup.WorkgroupId)
	ctx.Export(OpArn, createdWorkgroup.Arn)
	ctx.Export(OpPort, createdWorkgroup.Port)

	// The connection hostname lives on the workgroup's endpoint list
	// (exactly one endpoint once the workgroup is available). Index and
	// Elem both resolve to zero values when the endpoint is not yet
	// known, so the export shape is stable without an ApplyT applier.
	ctx.Export(OpEndpointAddress, createdWorkgroup.Endpoints.Index(pulumi.Int(0)).Address().Elem())

	// Per-satellite maps: endpoint addresses and AWS-generated usage-limit
	// IDs, keyed identically on both engines (imports and out-of-band CLI
	// operations address entries by these keys).
	ctx.Export(OpEndpointAccessAddresses, createdEndpointAddresses)
	ctx.Export(OpUsageLimitIds, createdUsageLimitIds)

	// The expiry keeps a stable string shape whether or not a custom
	// domain is configured.
	if createdCustomDomain != nil {
		ctx.Export(OpCustomDomainCertificateExpiryTime, createdCustomDomain.CustomDomainCertificateExpiryTime)
	} else {
		ctx.Export(OpCustomDomainCertificateExpiryTime, pulumi.String(""))
	}

	return nil
}
