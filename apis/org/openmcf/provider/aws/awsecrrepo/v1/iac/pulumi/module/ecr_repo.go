package module

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ecr"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ecrRepo creates an AWS ECR repository and optional lifecycle and repository policies
// based on the fields in AwsEcrRepoSpec.
func ecrRepo(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.AwsEcrRepo.Spec

	imageTagMutability := "MUTABLE"

	if spec.ImageImmutable {
		imageTagMutability = "IMMUTABLE"
	}

	// Configure image scanning (defaults to true for security)
	imageScanningConfiguration := &ecr.RepositoryImageScanningConfigurationArgs{
		ScanOnPush: pulumi.Bool(spec.GetScanOnPush()),
	}

	repo, err := ecr.NewRepository(ctx, locals.AwsEcrRepo.Metadata.Name, &ecr.RepositoryArgs{
		Name:                       pulumi.String(spec.RepositoryName),
		ImageTagMutability:         pulumi.String(imageTagMutability),
		ImageScanningConfiguration: imageScanningConfiguration,
		ForceDelete:                pulumi.Bool(spec.ForceDelete),
		Tags:                       pulumi.ToStringMap(locals.AwsTags),
		EncryptionConfigurations: ecr.RepositoryEncryptionConfigurationArray{
			&ecr.RepositoryEncryptionConfigurationArgs{
				EncryptionType: pulumi.String(spec.GetEncryptionType()),
				KmsKey: func() pulumi.StringInput {
					if spec.KmsKeyId != nil && spec.KmsKeyId.GetValue() != "" {
						return pulumi.String(spec.KmsKeyId.GetValue())
					}
					return pulumi.String("")
				}(),
			},
		},
	}, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "unable to create AWS ECR Repository")
	}

	// Create lifecycle policy if specified
	if spec.LifecyclePolicy != nil {
		if err := createLifecyclePolicy(ctx, locals, repo, provider); err != nil {
			return errors.Wrap(err, "unable to create lifecycle policy")
		}
	}

	ctx.Export(OpEcrRepoName, repo.Name)
	ctx.Export(OpEcrRepoUrl, repo.RepositoryUrl)
	ctx.Export(OpEcrRepoArn, repo.Arn)
	ctx.Export(OpEcrRepoRegistryId, repo.RegistryId)

	return nil
}

// createLifecyclePolicy generates and applies an ECR lifecycle policy
// based on the simplified configuration in the spec.
func createLifecyclePolicy(ctx *pulumi.Context, locals *Locals, repo *ecr.Repository, provider *aws.Provider) error {
	spec := locals.AwsEcrRepo.Spec
	lifecyclePolicy := spec.LifecyclePolicy

	// Build lifecycle policy rules
	rules := []map[string]interface{}{}

	// Rule 1: Expire untagged images after specified days
	if lifecyclePolicy.GetExpireUntaggedAfterDays() > 0 {
		rules = append(rules, map[string]interface{}{
			"rulePriority": 1,
			"description":  fmt.Sprintf("Expire untagged images after %d days", lifecyclePolicy.GetExpireUntaggedAfterDays()),
			"selection": map[string]interface{}{
				"tagStatus":   "untagged",
				"countType":   "sinceImagePushed",
				"countUnit":   "days",
				"countNumber": lifecyclePolicy.GetExpireUntaggedAfterDays(),
			},
			"action": map[string]interface{}{
				"type": "expire",
			},
		})
	}

	// Rule 2: Keep only the most recent N images
	if lifecyclePolicy.GetMaxImageCount() > 0 {
		rules = append(rules, map[string]interface{}{
			"rulePriority": 2,
			"description":  fmt.Sprintf("Keep only the last %d images", lifecyclePolicy.GetMaxImageCount()),
			"selection": map[string]interface{}{
				"tagStatus":   "any",
				"countType":   "imageCountMoreThan",
				"countNumber": lifecyclePolicy.GetMaxImageCount(),
			},
			"action": map[string]interface{}{
				"type": "expire",
			},
		})
	}

	// Only create the lifecycle policy if there are rules
	if len(rules) == 0 {
		return nil
	}

	policyDocument := map[string]interface{}{
		"rules": rules,
	}

	policyJSON, err := json.Marshal(policyDocument)
	if err != nil {
		return errors.Wrap(err, "failed to marshal lifecycle policy JSON")
	}

	_, err = ecr.NewLifecyclePolicy(ctx, fmt.Sprintf("%s-lifecycle", locals.AwsEcrRepo.Metadata.Name), &ecr.LifecyclePolicyArgs{
		Repository: repo.Name,
		Policy:     pulumi.String(string(policyJSON)),
	}, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create lifecycle policy")
	}

	return nil
}
