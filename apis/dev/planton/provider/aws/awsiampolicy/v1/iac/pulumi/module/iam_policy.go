package module

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps/convertstringmaps"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

// iamPolicy provisions the customer-managed policy. Name, path, and
// description are create-only in AWS (the provider replaces the policy when
// they change); only the document is updatable in place. Document updates
// create a new policy version and mark it default -- AWS keeps at most 5
// versions and the provider prunes the oldest non-default one before saving a
// new version, so repeated updates keep working without manual cleanup.
func iamPolicy(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource) error {
	policyName := locals.AwsIamPolicy.Metadata.Name
	spec := locals.AwsIamPolicy.Spec

	// policy_document is a free-form JSON object (google.protobuf.Struct);
	// iam.Policy wants the document as a JSON string, so encode it here.
	policyDocumentString, err := structToJSONString(spec.PolicyDocument)
	if err != nil {
		return errors.Wrap(err, "failed to marshal policy document JSON")
	}

	policyArgs := &iam.PolicyArgs{
		Name:   pulumi.String(policyName),
		Policy: pulumi.String(policyDocumentString),
		Tags: convertstringmaps.ConvertGoStringMapToPulumiStringMap(
			stringmaps.AddEntry(locals.AwsTags, "Name", policyName)),
	}

	if spec.Path != "" {
		policyArgs.Path = pulumi.StringPtr(spec.Path)
	}
	if spec.Description != "" {
		policyArgs.Description = pulumi.StringPtr(spec.Description)
	}

	createdPolicy, err := iam.NewPolicy(ctx, policyName, policyArgs, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create iam policy")
	}

	ctx.Export(OpPolicyArn, createdPolicy.Arn)
	ctx.Export(OpPolicyId, createdPolicy.PolicyId)
	ctx.Export(OpPolicyName, createdPolicy.Name)

	return nil
}

// structToJSONString converts a google.protobuf.Struct to a raw JSON string.
func structToJSONString(s *structpb.Struct) (string, error) {
	if s == nil {
		return "{}", nil
	}
	bytes, err := json.Marshal(s.AsMap())
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
