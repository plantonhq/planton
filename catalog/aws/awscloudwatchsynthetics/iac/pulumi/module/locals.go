package module

import (
	"strconv"
	"strings"

	awscloudwatchsyntheticsv1alpha1 "github.com/plantonhq/planton/catalog/aws/awscloudwatchsynthetics/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awscloudwatchsyntheticsv1alpha1.AwsCloudwatchSynthetics
	Spec   *awscloudwatchsyntheticsv1alpha1.AwsCloudwatchSyntheticsSpec

	AwsTags map[string]string

	// The provider wants one "s3://bucket/prefix" artifact location
	// string; the spec models bucket and prefix separately for chart
	// wiring.
	ArtifactS3Location string
}

func initializeLocals(_ *pulumi.Context, in *awscloudwatchsyntheticsv1alpha1.AwsCloudwatchSyntheticsIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// Resource-identity tags match the Terraform module key-for-key
	// (applied to the canary and every owned group; the association
	// join is untaggable at AWS).
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsCloudwatchSynthetics.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	if canary := locals.Spec.Canary; canary != nil {
		locals.ArtifactS3Location = "s3://" + canary.ArtifactBucket.GetValue()
		if canary.ArtifactPrefix != "" {
			locals.ArtifactS3Location += "/" + strings.Trim(canary.ArtifactPrefix, "/")
		}
	}

	return locals
}
