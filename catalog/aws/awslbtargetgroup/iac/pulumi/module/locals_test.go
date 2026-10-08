package module

// The AWS name is spec.target_group_name when set, else metadata.name
// truncated to 32 characters; the Name tag carries the AWS name.

import (
	"testing"

	awslbtargetgroupv1alpha1 "github.com/plantonhq/planton/catalog/aws/awslbtargetgroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared"
)

func localsFor(name, targetGroupName string) *Locals {
	return initializeLocals(nil, &awslbtargetgroupv1alpha1.AwsLbTargetGroupIacInput{
		Target: &awslbtargetgroupv1alpha1.AwsLbTargetGroup{
			Metadata: &shared.CatalogObjectMetadata{Name: name},
			Spec:     &awslbtargetgroupv1alpha1.AwsLbTargetGroupSpec{TargetGroupName: targetGroupName},
		},
	})
}

func TestTheAwsNameIsTargetGroupNameWhenSet(t *testing.T) {
	locals := localsFor("orders-api-production", "tg-0a1b2c3")
	if locals.TargetGroupName != "tg-0a1b2c3" {
		t.Fatalf("AWS name is %q, want %q", locals.TargetGroupName, "tg-0a1b2c3")
	}
	if locals.AwsTags[awstagkeys.Name] != "tg-0a1b2c3" {
		t.Fatalf("Name tag is %q, want %q", locals.AwsTags[awstagkeys.Name], "tg-0a1b2c3")
	}
}

func TestTheAwsNameIsMetadataNameTruncatedTo32WhenUnset(t *testing.T) {
	locals := localsFor("orders-api-production-blue-green-canary", "")
	const want = "orders-api-production-blue-green"
	if locals.TargetGroupName != want {
		t.Fatalf("AWS name is %q, want %q", locals.TargetGroupName, want)
	}
	if locals.AwsTags[awstagkeys.Name] != want {
		t.Fatalf("Name tag is %q, want %q", locals.AwsTags[awstagkeys.Name], want)
	}
}
