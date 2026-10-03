package module

import (
	"strconv"

	awsconfigrecorderv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsconfigrecorder/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsconfigrecorderv1alpha1.AwsConfigRecorder
	Spec   *awsconfigrecorderv1alpha1.AwsConfigRecorderSpec

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsconfigrecorderv1alpha1.AwsConfigRecorderIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// Resource-identity tags match the Terraform module key-for-key.
	// None of the four resources this module manages carries tags in
	// the provider (the recorder family is untagged AWS surface) --
	// the map exists for convention and future taggable satellites.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsConfigRecorder.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
