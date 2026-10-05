package module

import (
	"strconv"

	awsefsaccesspointv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsefsaccesspoint/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	AwsEfsAccessPoint *awsefsaccesspointv1alpha1.AwsEfsAccessPoint
	AwsTags           map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awsefsaccesspointv1alpha1.AwsEfsAccessPointIacInput) *Locals {
	locals := &Locals{}
	locals.AwsEfsAccessPoint = iacInput.Target

	// Resource-identity tags follow the catalog convention. An access point
	// has no name argument at all — the Name tag IS its console display name,
	// so metadata.name here is the resource's only human-readable identity.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsEfsAccessPoint.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsEfsAccessPoint.Metadata.Org,
		awstagkeys.Environment:  locals.AwsEfsAccessPoint.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsEfsAccessPoint.String(),
		awstagkeys.ResourceId:   locals.AwsEfsAccessPoint.Metadata.Id,
	}

	return locals
}
