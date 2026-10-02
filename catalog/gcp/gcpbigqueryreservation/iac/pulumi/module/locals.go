package module

import (
	"fmt"
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpbigqueryreservationv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpbigqueryreservation/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Assignment is one folded assignment with its assignee composed into the
// form Google takes.
type Assignment struct {
	// Key is assignee|job_type|principal -- the uniqueness the spec
	// enforces and the Terraform module's for_each key.
	Key       string
	Assignee  string
	JobType   string
	Principal string
}

type Locals struct {
	GcpProviderConfig      *gcpprovider.GcpProviderConfig
	GcpBigQueryReservation *gcpbigqueryreservationv1alpha1.GcpBigQueryReservation
	GcpLabels              map[string]string

	// ReservationName is spec.reservation_name when set, otherwise
	// metadata.name -- identical to the Terraform module's
	// locals.reservation_name.
	ReservationName string

	// Assignments in declared order -- identical to the Terraform module's
	// locals.assignment_list and assignment_keys.
	Assignments []Assignment
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpbigqueryreservationv1alpha1.GcpBigQueryReservationStackInput) *Locals {
	locals := &Locals{}
	locals.GcpBigQueryReservation = stackInput.Target
	metadata := locals.GcpBigQueryReservation.Metadata
	spec := locals.GcpBigQueryReservation.Spec

	locals.ReservationName = spec.ReservationName
	if locals.ReservationName == "" {
		locals.ReservationName = metadata.Name
	}

	for _, a := range spec.Assignments {
		var assignee string
		switch {
		case a.Assignee.ProjectId.GetValue() != "":
			assignee = "projects/" + a.Assignee.ProjectId.GetValue()
		case a.Assignee.FolderId.GetValue() != "":
			assignee = "folders/" + a.Assignee.FolderId.GetValue()
		default:
			assignee = "organizations/" + a.Assignee.OrganizationId
		}
		locals.Assignments = append(locals.Assignments, Assignment{
			Key:       fmt.Sprintf("%s|%s|%s", assignee, a.JobType, a.Principal),
			Assignee:  assignee,
			JobType:   a.JobType,
			Principal: a.Principal,
		})
	}

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = metadata.Name
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpBigQueryReservation.String())

	if metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = metadata.Org
	}
	if metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = metadata.Env
	}
	if metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
