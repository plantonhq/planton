package module

import (
	"strconv"

	awseventbridgeapidestinationv1alpha1 "github.com/plantonhq/planton/catalog/aws/awseventbridgeapidestination/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awseventbridgeapidestinationv1alpha1.AwsEventBridgeApiDestination
	Spec   *awseventbridgeapidestinationv1alpha1.AwsEventBridgeApiDestinationSpec

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awseventbridgeapidestinationv1alpha1.AwsEventBridgeApiDestinationIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// Resource-identity tags match the Terraform module key-for-key.
	// NOTE: neither the connection nor the destination is taggable at
	// AWS - the deliberate tag-convention absence (the
	// AwsCloudwatchDashboard precedent). Kept for the day AWS adds
	// tagging.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsEventBridgeApiDestination.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
