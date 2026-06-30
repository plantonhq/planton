package module

import (
	"github.com/pkg/errors"
	awscodepipelinev1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awscodepipeline/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/codepipeline"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func pipeline(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
) (*codepipeline.Pipeline, error) {
	spec := locals.AwsCodePipeline.Spec

	// --- Artifact stores ---
	var artifactStores codepipeline.PipelineArtifactStoreArray
	for _, store := range spec.ArtifactStores {
		storeArgs := &codepipeline.PipelineArtifactStoreArgs{
			Location: pulumi.String(store.Location.GetValue()),
			Type:     pulumi.String("S3"),
		}
		if store.Region != "" {
			storeArgs.Region = pulumi.StringPtr(store.Region)
		}
		if store.EncryptionKeyId != nil && store.EncryptionKeyId.GetValue() != "" {
			storeArgs.EncryptionKey = &codepipeline.PipelineArtifactStoreEncryptionKeyArgs{
				Id:   pulumi.String(store.EncryptionKeyId.GetValue()),
				Type: pulumi.String("KMS"),
			}
		}
		artifactStores = append(artifactStores, storeArgs)
	}

	// --- Stages ---
	var stages codepipeline.PipelineStageArray
	for _, stage := range spec.Stages {
		var actions codepipeline.PipelineStageActionArray
		for _, action := range stage.Actions {
			actionArgs := &codepipeline.PipelineStageActionArgs{
				Name:     pulumi.String(action.Name),
				Category: pulumi.String(action.Category),
				Owner:    pulumi.String(action.Owner),
				Provider: pulumi.String(action.Provider),
				Version:  pulumi.String(action.Version),
			}
			if len(action.Configuration) > 0 {
				actionArgs.Configuration = pulumi.ToStringMap(action.Configuration)
			}
			if len(action.InputArtifacts) > 0 {
				actionArgs.InputArtifacts = pulumi.ToStringArray(action.InputArtifacts)
			}
			if len(action.OutputArtifacts) > 0 {
				actionArgs.OutputArtifacts = pulumi.ToStringArray(action.OutputArtifacts)
			}
			if action.Namespace != "" {
				actionArgs.Namespace = pulumi.StringPtr(action.Namespace)
			}
			if action.Region != "" {
				actionArgs.Region = pulumi.StringPtr(action.Region)
			}
			if action.RoleArn != nil && action.RoleArn.GetValue() != "" {
				actionArgs.RoleArn = pulumi.StringPtr(action.RoleArn.GetValue())
			}
			if action.RunOrder > 0 {
				actionArgs.RunOrder = pulumi.IntPtr(int(action.RunOrder))
			}
			if action.TimeoutInMinutes > 0 {
				actionArgs.TimeoutInMinutes = pulumi.IntPtr(int(action.TimeoutInMinutes))
			}
			actions = append(actions, actionArgs)
		}
		stages = append(stages, &codepipeline.PipelineStageArgs{
			Name:    pulumi.String(stage.Name),
			Actions: actions,
		})
	}

	// --- Pipeline args ---
	args := &codepipeline.PipelineArgs{
		Name:           pulumi.StringPtr(locals.AwsCodePipeline.Metadata.Id),
		RoleArn:        pulumi.String(spec.RoleArn.GetValue()),
		ArtifactStores: artifactStores,
		Stages:         stages,
		Tags:           pulumi.ToStringMap(locals.Labels),
	}

	if spec.GetPipelineType() != "" {
		args.PipelineType = pulumi.StringPtr(spec.GetPipelineType())
	}
	if spec.GetExecutionMode() != "" {
		args.ExecutionMode = pulumi.StringPtr(spec.GetExecutionMode())
	}

	// --- Triggers (V2 only) ---
	if len(spec.Triggers) > 0 {
		var triggers codepipeline.PipelineTriggerArray
		for _, trigger := range spec.Triggers {
			triggerArgs := &codepipeline.PipelineTriggerArgs{
				ProviderType: pulumi.String(trigger.ProviderType),
			}
			if trigger.GitConfiguration != nil {
				gitArgs := &codepipeline.PipelineTriggerGitConfigurationArgs{
					SourceActionName: pulumi.String(trigger.GitConfiguration.SourceActionName),
				}
				if len(trigger.GitConfiguration.Push) > 0 {
					gitArgs.Pushes = buildPushFilters(trigger.GitConfiguration.Push)
				}
				if len(trigger.GitConfiguration.PullRequest) > 0 {
					gitArgs.PullRequests = buildPullRequestFilters(trigger.GitConfiguration.PullRequest)
				}
				triggerArgs.GitConfiguration = gitArgs
			}
			triggers = append(triggers, triggerArgs)
		}
		args.Triggers = triggers
	}

	// --- Variables (V2 only) ---
	if len(spec.Variables) > 0 {
		var variables codepipeline.PipelineVariableArray
		for _, v := range spec.Variables {
			varArgs := &codepipeline.PipelineVariableArgs{
				Name: pulumi.String(v.Name),
			}
			if v.DefaultValue != "" {
				varArgs.DefaultValue = pulumi.StringPtr(v.DefaultValue)
			}
			if v.Description != "" {
				varArgs.Description = pulumi.StringPtr(v.Description)
			}
			variables = append(variables, varArgs)
		}
		args.Variables = variables
	}

	created, err := codepipeline.NewPipeline(ctx, "codepipeline", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create codepipeline")
	}

	return created, nil
}

func buildPushFilters(pushes []*awscodepipelinev1.AwsCodePipelineGitPush) codepipeline.PipelineTriggerGitConfigurationPushArray {
	var result codepipeline.PipelineTriggerGitConfigurationPushArray
	for _, push := range pushes {
		pushArgs := &codepipeline.PipelineTriggerGitConfigurationPushArgs{}
		if push.Branches != nil {
			pushArgs.Branches = &codepipeline.PipelineTriggerGitConfigurationPushBranchesArgs{
				Includes: pulumi.ToStringArray(push.Branches.Includes),
				Excludes: pulumi.ToStringArray(push.Branches.Excludes),
			}
		}
		if push.FilePaths != nil {
			pushArgs.FilePaths = &codepipeline.PipelineTriggerGitConfigurationPushFilePathsArgs{
				Includes: pulumi.ToStringArray(push.FilePaths.Includes),
				Excludes: pulumi.ToStringArray(push.FilePaths.Excludes),
			}
		}
		if push.Tags != nil {
			pushArgs.Tags = &codepipeline.PipelineTriggerGitConfigurationPushTagsArgs{
				Includes: pulumi.ToStringArray(push.Tags.Includes),
				Excludes: pulumi.ToStringArray(push.Tags.Excludes),
			}
		}
		result = append(result, pushArgs)
	}
	return result
}

func buildPullRequestFilters(prs []*awscodepipelinev1.AwsCodePipelineGitPullRequest) codepipeline.PipelineTriggerGitConfigurationPullRequestArray {
	var result codepipeline.PipelineTriggerGitConfigurationPullRequestArray
	for _, pr := range prs {
		prArgs := &codepipeline.PipelineTriggerGitConfigurationPullRequestArgs{}
		if pr.Branches != nil {
			prArgs.Branches = &codepipeline.PipelineTriggerGitConfigurationPullRequestBranchesArgs{
				Includes: pulumi.ToStringArray(pr.Branches.Includes),
				Excludes: pulumi.ToStringArray(pr.Branches.Excludes),
			}
		}
		if pr.FilePaths != nil {
			prArgs.FilePaths = &codepipeline.PipelineTriggerGitConfigurationPullRequestFilePathsArgs{
				Includes: pulumi.ToStringArray(pr.FilePaths.Includes),
				Excludes: pulumi.ToStringArray(pr.FilePaths.Excludes),
			}
		}
		if len(pr.Events) > 0 {
			prArgs.Events = pulumi.ToStringArray(pr.Events)
		}
		result = append(result, prArgs)
	}
	return result
}
