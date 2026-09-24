package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/bigquery"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// urnSafe turns an assignment key (projects/x|QUERY|principal://...) into a
// Pulumi resource name segment.
var urnSafe = strings.NewReplacer("/", "-", "|", "--", ":", "-")

// reservation creates the slot reservation and its folded assignments.
func reservation(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpBigQueryReservation.Spec
	resourceName := locals.GcpBigQueryReservation.Metadata.Name

	// Enable the BigQuery Reservation API first so a fresh project works on
	// the first deploy. DisableOnDestroy stays false: tearing down one
	// reservation must never disable the API for the admin project.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("bigqueryreservation.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		apiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdApi, err := projects.NewService(ctx,
		"gcpbqrs-bigqueryreservation.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable bigqueryreservation.googleapis.com api")
	}

	args := &bigquery.ReservationArgs{
		Name:         pulumi.String(locals.ReservationName),
		SlotCapacity: pulumi.Int(int(spec.SlotCapacity)),
		Labels:       pulumi.ToStringMap(locals.GcpLabels),
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
	if spec.Edition != "" {
		args.Edition = pulumi.String(spec.Edition)
	}
	if spec.IgnoreIdleSlots {
		args.IgnoreIdleSlots = pulumi.Bool(true)
	}
	if spec.Concurrency > 0 {
		args.Concurrency = pulumi.Int(int(spec.Concurrency))
	}
	if spec.ReservationGroup.GetValue() != "" {
		args.ReservationGroup = pulumi.String(spec.ReservationGroup.GetValue())
	}
	if spec.SecondaryLocation != "" {
		args.SecondaryLocation = pulumi.String(spec.SecondaryLocation)
	}
	// The spec lifts the block's one input; 0 means no autoscaling.
	if spec.AutoscaleMaxSlots > 0 {
		args.Autoscale = &bigquery.ReservationAutoscaleArgs{
			MaxSlots: pulumi.Int(int(spec.AutoscaleMaxSlots)),
		}
	}

	// Engine-side destroy stance: PREVENT fails destroys, ABANDON removes
	// from management without deleting. Sent only when set so the provider
	// default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	created, err := bigquery.NewReservation(ctx, resourceName, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create bigquery reservation")
	}

	// The folded assignments share the reservation's project, location, and
	// destroy stance. Every argument is immutable, so a change replaces the
	// assignment (its key changes with it).
	assignmentNames := pulumi.StringArray{}
	for _, assignment := range locals.Assignments {
		assignmentArgs := &bigquery.ReservationAssignmentArgs{
			Location:    created.Location,
			Reservation: created.Name,
			Assignee:    pulumi.String(assignment.Assignee),
			JobType:     pulumi.String(assignment.JobType),
		}
		if spec.ProjectId.GetValue() != "" {
			assignmentArgs.Project = pulumi.String(spec.ProjectId.GetValue())
		}
		if assignment.Principal != "" {
			assignmentArgs.Principal = pulumi.String(assignment.Principal)
		}
		if spec.DeletionPolicy != "" {
			assignmentArgs.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
		}
		createdAssignment, err := bigquery.NewReservationAssignment(ctx,
			resourceName+"-"+urnSafe.Replace(assignment.Key), assignmentArgs, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create reservation assignment %s", assignment.Key)
		}
		assignmentNames = append(assignmentNames, createdAssignment.ID().ToStringOutput())
	}

	ctx.Export(OpName, created.ID().ToStringOutput())
	ctx.Export(OpReservationName, created.Name)
	ctx.Export(OpLocation, created.Location)
	ctx.Export(OpAssignmentNames, assignmentNames)
	return nil
}
