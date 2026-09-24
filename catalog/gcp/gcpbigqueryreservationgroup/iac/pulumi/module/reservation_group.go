package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/bigquery"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// reservationGroup creates the reservation group reservations join from
// their own reservation_group field.
func reservationGroup(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpBigQueryReservationGroup.Spec
	resourceName := locals.GcpBigQueryReservationGroup.Metadata.Name

	// Enable the BigQuery Reservation API first so a fresh project works on
	// the first deploy. DisableOnDestroy stays false: tearing down one group
	// must never disable the API its admin project runs on.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("bigqueryreservation.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		apiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdApi, err := projects.NewService(ctx,
		"gcpbqrg-bigqueryreservation.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable bigqueryreservation.googleapis.com api")
	}

	args := &bigquery.ReservationGroupArgs{
		Name: pulumi.String(locals.ReservationGroupName),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors; an empty location
	// leaves Google's US default.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.Location != "" {
		args.Location = pulumi.String(spec.Location)
	}

	// Engine-side destroy stance: PREVENT fails destroys, ABANDON removes
	// from management without deleting. Sent only when set so the provider
	// default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	created, err := bigquery.NewReservationGroup(ctx, resourceName, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create bigquery reservation group")
	}

	ctx.Export(OpName, created.ID().ToStringOutput())
	ctx.Export(OpReservationGroupName, created.Name)
	ctx.Export(OpLocation, created.Location)
	return nil
}
