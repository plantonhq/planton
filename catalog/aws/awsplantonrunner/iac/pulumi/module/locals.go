package module

import (
	"strconv"

	"github.com/plantonhq/planton/shared/catalogkind"

	awsplantonrunnerv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsplantonrunner/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals defines local variables used throughout the Pulumi module: the
// target AwsPlantonRunner resource, the resource-identity tags applied to
// everything the module creates, and the runner's registration name.
type Locals struct {
	AwsPlantonRunner *awsplantonrunnerv1alpha1.AwsPlantonRunner
	AwsTags          map[string]string

	// RegistrationName is the name the runner registers itself under when
	// it joins the control plane: "<env>-<metadata.name>" (metadata.name
	// outside an environment) -- the SAME derivation the platform uses for
	// records that reference this runner (its minted token, its managed
	// destroy); changing this formula breaks arrival attribution and
	// managed teardown.
	RegistrationName string
}

// initializeLocals pulls values from the IaC input and populates the
// Locals struct. Similar to Terraform's "locals" concept.
func initializeLocals(ctx *pulumi.Context, iacInput *awsplantonrunnerv1alpha1.AwsPlantonRunnerIacInput) *Locals {
	locals := &Locals{
		AwsPlantonRunner: iacInput.Target,
	}

	locals.RegistrationName = locals.AwsPlantonRunner.Metadata.Name
	if locals.AwsPlantonRunner.Metadata.Env != "" {
		locals.RegistrationName = locals.AwsPlantonRunner.Metadata.Env + "-" + locals.AwsPlantonRunner.Metadata.Name
	}

	// Resource-identity tags match the Terraform module key-for-key.
	// DELIBERATELY five keys, no Name: the appliance's resources (cluster,
	// service, roles, secret, log group) each carry their own explicit
	// resource name -- a shared Name tag would mislabel ten distinct
	// resources with one value (the same recorded convention as AwsEcrRepo
	// and AwsGlobalAccelerator).
	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsPlantonRunner.Metadata.Org,
		awstagkeys.Environment:  locals.AwsPlantonRunner.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsPlantonRunner.String(),
		awstagkeys.ResourceId:   locals.AwsPlantonRunner.Metadata.Id,
	}

	return locals
}
