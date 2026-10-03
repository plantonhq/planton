package module

import (
	"strconv"

	awssagemakermlflowserverv1alpha1 "github.com/plantonhq/planton/catalog/aws/awssagemakermlflowserver/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awssagemakermlflowserverv1alpha1.AwsSagemakerMlflowServer
	Spec   *awssagemakermlflowserverv1alpha1.AwsSagemakerMlflowServerSpec

	TrackingServerName string
	AwsTags            map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awssagemakermlflowserverv1alpha1.AwsSagemakerMlflowServerIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// The component's name IS the tracking server name.
	locals.TrackingServerName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsSagemakerMlflowServer.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
