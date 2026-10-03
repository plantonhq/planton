package module

import (
	"strconv"

	awseventbridgeschedulerv1alpha1 "github.com/plantonhq/planton/catalog/aws/awseventbridgescheduler/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awseventbridgeschedulerv1alpha1.AwsEventBridgeScheduler
	Spec   *awseventbridgeschedulerv1alpha1.AwsEventBridgeSchedulerSpec

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awseventbridgeschedulerv1alpha1.AwsEventBridgeSchedulerIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// Resource-identity tags match the Terraform module key-for-key.
	// They land on the OWNED GROUP only - the schedule itself is
	// untaggable at AWS (the deliberate tag-convention absence).
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsEventBridgeScheduler.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
