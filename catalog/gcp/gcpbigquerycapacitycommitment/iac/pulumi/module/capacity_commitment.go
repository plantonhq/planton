package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/bigquery"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// capacityCommitment buys the slot commitment -- a purchase. Creating it
// starts a billed term; Google refuses the delete before the term ends.
func capacityCommitment(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpBigQueryCapacityCommitment.Spec
	resourceName := locals.GcpBigQueryCapacityCommitment.Metadata.Name

	// Enable the BigQuery Reservation API first so a fresh project works on
	// the first deploy. DisableOnDestroy stays false: a commitment's
	// teardown must never disable the API its admin project runs on.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("bigqueryreservation.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		apiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdApi, err := projects.NewService(ctx,
		"gcpbqcc-bigqueryreservation.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable bigqueryreservation.googleapis.com api")
	}

	args := &bigquery.CapacityCommitmentArgs{
		CapacityCommitmentId: pulumi.String(locals.CapacityCommitmentId),
		SlotCount:            pulumi.Int(int(spec.SlotCount)),
		Plan:                 pulumi.String(spec.Plan),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors. Unset optionals stay
	// out of the payload so Google's defaults apply.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.Location != "" {
		args.Location = pulumi.String(spec.Location)
	}
	if spec.RenewalPlan != "" {
		args.RenewalPlan = pulumi.String(spec.RenewalPlan)
	}
	if spec.Edition != "" {
		args.Edition = pulumi.String(spec.Edition)
	}
	// The provider types this boolean as a string; the spec's bool is sent
	// as "true" when set and omitted otherwise -- the Terraform module's
	// mapping.
	if spec.EnforceSingleAdminProjectPerOrg {
		args.EnforceSingleAdminProjectPerOrg = pulumi.String("true")
	}

	// Engine-side destroy stance: PREVENT fails destroys, ABANDON removes
	// from management without deleting. Sent only when set so the provider
	// default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	created, err := bigquery.NewCapacityCommitment(ctx, resourceName, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create bigquery capacity commitment")
	}

	ctx.Export(OpName, created.Name)
	ctx.Export(OpState, created.State)
	ctx.Export(OpCommitmentStartTime, created.CommitmentStartTime)
	ctx.Export(OpCommitmentEndTime, created.CommitmentEndTime)
	return nil
}
