package module

import (
	"sort"

	"github.com/pkg/errors"
	awsbedrockagentv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbedrockagent/v1alpha1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/bedrock"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// agent creates the Bedrock agent, its folded satellites (action groups,
// collaborators, knowledge-base associations, aliases), and exports
// outputs.
//
// Lifecycle facts the renders below depend on:
//   - every satellite change re-"prepares" the agent (provider-managed;
//     the provider retries the OptLock/preparing conflicts itself);
//   - creating an alias SNAPSHOTS the draft into a new numbered version,
//     so aliases depend on every other satellite -- the snapshot must
//     capture the fully-assembled draft;
//   - prepare_agent and skip_resource_in_use_check are apply-behavior
//     knobs, not desired state: both engines keep the provider defaults
//     (prepare always, in-use check on).
func agent(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec

	args := &bedrock.AgentAgentArgs{
		// Create-time naming basis; doubles as the Name tag. metadata.name
		// on both engines.
		AgentName: pulumi.String(locals.AgentName),
		// Required by AWS: the model that powers the agent and the role
		// the Bedrock service assumes to operate it.
		FoundationModel:      pulumi.String(spec.FoundationModel),
		AgentResourceRoleArn: pulumi.String(spec.AgentResourceRoleArn.GetValue()),
		Tags:                 pulumi.ToStringMap(locals.AwsTags),
	}

	// Optional+Computed at the provider: sent only when set so the module
	// never fights AWS's normalization.
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}
	if spec.Instruction != "" {
		args.Instruction = pulumi.String(spec.Instruction)
	}
	if spec.IdleSessionTtlSeconds != 0 {
		args.IdleSessionTtlInSeconds = pulumi.Int(int(spec.IdleSessionTtlSeconds))
	}
	if spec.CustomerEncryptionKeyArn.GetValue() != "" {
		args.CustomerEncryptionKeyArn = pulumi.String(spec.CustomerEncryptionKeyArn.GetValue())
	}
	if spec.AgentCollaboration != "" {
		args.AgentCollaboration = pulumi.String(spec.AgentCollaboration)
	}

	// Guardrail attachment.
	if spec.Guardrail != nil {
		args.GuardrailConfigurations = bedrock.AgentAgentGuardrailConfigurationArray{
			&bedrock.AgentAgentGuardrailConfigurationArgs{
				GuardrailIdentifier: pulumi.String(spec.Guardrail.GuardrailId.GetValue()),
				GuardrailVersion:    pulumi.String(spec.Guardrail.Version),
			},
		}
	}

	// Session-summary memory -- SESSION_SUMMARY is the only memory type
	// AWS defines; the block's switch (on by default once the block is
	// declared, filled by the platform before this runs) enables it and the
	// module owns
	// the constant.
	if spec.Memory != nil && spec.Memory.GetEnabled() {
		memory := &bedrock.AgentAgentMemoryConfigurationArgs{
			EnabledMemoryTypes: pulumi.StringArray{pulumi.String("SESSION_SUMMARY")},
		}
		if spec.Memory.StorageDays != 0 {
			memory.StorageDays = pulumi.Int(int(spec.Memory.StorageDays))
		}
		if spec.Memory.MaxRecentSessions != 0 {
			memory.SessionSummaryConfigurations = bedrock.AgentAgentMemoryConfigurationSessionSummaryConfigurationArray{
				&bedrock.AgentAgentMemoryConfigurationSessionSummaryConfigurationArgs{
					MaxRecentSessions: pulumi.Int(int(spec.Memory.MaxRecentSessions)),
				},
			}
		}
		args.MemoryConfigurations = bedrock.AgentAgentMemoryConfigurationArray{memory}
	}

	// Prompt-template overrides. Authoring an entry IS the override, so
	// the module marks every entry OVERRIDDEN (the provider strips non-
	// overridden AWS defaults from state -- a DEFAULT creation mode here
	// would vanish on read and drift forever).
	if spec.PromptOverride != nil {
		override := &bedrock.AgentAgentPromptOverrideConfigurationArgs{}
		if spec.PromptOverride.OverrideLambda.GetValue() != "" {
			override.OverrideLambda = pulumi.String(spec.PromptOverride.OverrideLambda.GetValue())
		}
		var configurations bedrock.AgentAgentPromptOverrideConfigurationPromptConfigurationArray
		for _, p := range spec.PromptOverride.PromptConfigurations {
			configuration := &bedrock.AgentAgentPromptOverrideConfigurationPromptConfigurationArgs{
				PromptType:         pulumi.String(p.PromptType),
				BasePromptTemplate: pulumi.String(p.BasePromptTemplate),
				PromptCreationMode: pulumi.String("OVERRIDDEN"),
			}
			if p.ParserMode != "" {
				configuration.ParserMode = pulumi.String(p.ParserMode)
			}
			if p.PromptState != "" {
				configuration.PromptState = pulumi.String(p.PromptState)
			}
			if p.InferenceConfiguration != nil {
				inference := &bedrock.AgentAgentPromptOverrideConfigurationPromptConfigurationInferenceConfigurationArgs{}
				if p.InferenceConfiguration.MaxLength != nil {
					inference.MaxLength = pulumi.Int(int(*p.InferenceConfiguration.MaxLength))
				}
				if len(p.InferenceConfiguration.StopSequences) > 0 {
					inference.StopSequences = pulumi.ToStringArray(p.InferenceConfiguration.StopSequences)
				}
				if p.InferenceConfiguration.Temperature != nil {
					inference.Temperature = pulumi.Float64(*p.InferenceConfiguration.Temperature)
				}
				if p.InferenceConfiguration.TopK != nil {
					inference.TopK = pulumi.Int(int(*p.InferenceConfiguration.TopK))
				}
				if p.InferenceConfiguration.TopP != nil {
					inference.TopP = pulumi.Float64(*p.InferenceConfiguration.TopP)
				}
				configuration.InferenceConfigurations = bedrock.AgentAgentPromptOverrideConfigurationPromptConfigurationInferenceConfigurationArray{inference}
			}
			configurations = append(configurations, configuration)
		}
		override.PromptConfigurations = configurations
		args.PromptOverrideConfigurations = bedrock.AgentAgentPromptOverrideConfigurationArray{override}
	}

	createdAgent, err := bedrock.NewAgentAgent(ctx, locals.AgentName, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "create agent")
	}

	ctx.Export(OpAgentId, createdAgent.AgentId)
	ctx.Export(OpAgentArn, createdAgent.AgentArn)
	ctx.Export(OpDraftVersion, pulumi.String("DRAFT"))

	// Satellites attach to the DRAFT version, keyed by their stable entry
	// names. Iteration is name-sorted for deterministic previews.
	draftSatellites := []pulumi.Resource{}

	actionGroupIds := pulumi.StringMap{}
	for _, g := range sortedActionGroups(spec.ActionGroups) {
		groupArgs := &bedrock.AgentAgentActionGroupArgs{
			ActionGroupName: pulumi.String(g.Name),
			AgentId:         createdAgent.AgentId,
			AgentVersion:    pulumi.String("DRAFT"),
		}
		if g.Description != "" {
			groupArgs.Description = pulumi.String(g.Description)
		}
		if g.State != "" {
			groupArgs.ActionGroupState = pulumi.String(g.State)
		}
		if g.ParentActionGroupSignature != "" {
			groupArgs.ParentActionGroupSignature = pulumi.String(g.ParentActionGroupSignature)
		}
		if g.Executor != nil {
			executor := &bedrock.AgentAgentActionGroupActionGroupExecutorArgs{}
			if g.Executor.Lambda.GetValue() != "" {
				executor.Lambda = pulumi.String(g.Executor.Lambda.GetValue())
			}
			// RETURN_CONTROL is the only custom-control method AWS defines
			// -- the spec models it as a bool and the module owns the
			// constant.
			if g.Executor.ReturnControl {
				executor.CustomControl = pulumi.String("RETURN_CONTROL")
			}
			groupArgs.ActionGroupExecutor = executor
		}
		if g.ApiSchema != nil {
			apiSchema := &bedrock.AgentAgentActionGroupApiSchemaArgs{}
			if g.ApiSchema.Payload != "" {
				apiSchema.Payload = pulumi.String(g.ApiSchema.Payload)
			}
			if g.ApiSchema.S3 != nil {
				apiSchema.S3 = &bedrock.AgentAgentActionGroupApiSchemaS3Args{
					S3BucketName: pulumi.String(g.ApiSchema.S3.BucketName.GetValue()),
					S3ObjectKey:  pulumi.String(g.ApiSchema.S3.ObjectKey),
				}
			}
			groupArgs.ApiSchema = apiSchema
		}
		if g.FunctionSchema != nil {
			var functions bedrock.AgentAgentActionGroupFunctionSchemaMemberFunctionsFunctionArray
			for _, f := range g.FunctionSchema.Functions {
				function := &bedrock.AgentAgentActionGroupFunctionSchemaMemberFunctionsFunctionArgs{
					Name: pulumi.String(f.Name),
				}
				if f.Description != "" {
					function.Description = pulumi.String(f.Description)
				}
				var parameters bedrock.AgentAgentActionGroupFunctionSchemaMemberFunctionsFunctionParameterArray
				for _, p := range f.Parameters {
					parameter := &bedrock.AgentAgentActionGroupFunctionSchemaMemberFunctionsFunctionParameterArgs{
						// The provider calls the parameter name
						// `map_block_key` for its own compatibility
						// reasons; the spec calls it what it is.
						MapBlockKey: pulumi.String(p.Name),
						Type:        pulumi.String(p.Type),
						Required:    pulumi.Bool(p.Required),
					}
					if p.Description != "" {
						parameter.Description = pulumi.String(p.Description)
					}
					parameters = append(parameters, parameter)
				}
				function.Parameters = parameters
				functions = append(functions, function)
			}
			groupArgs.FunctionSchema = &bedrock.AgentAgentActionGroupFunctionSchemaArgs{
				MemberFunctions: &bedrock.AgentAgentActionGroupFunctionSchemaMemberFunctionsArgs{
					Functions: functions,
				},
			}
		}
		createdGroup, err := bedrock.NewAgentAgentActionGroup(ctx, "action-group-"+g.Name, groupArgs,
			pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{createdAgent}))
		if err != nil {
			return errors.Wrapf(err, "create action group %q", g.Name)
		}
		actionGroupIds[g.Name] = createdGroup.ActionGroupId
		draftSatellites = append(draftSatellites, createdGroup)
	}
	ctx.Export(OpActionGroupIds, actionGroupIds)

	collaboratorIds := pulumi.StringMap{}
	for _, c := range sortedCollaborators(spec.Collaborators) {
		collaboratorArgs := &bedrock.AgentAgentCollaboratorArgs{
			AgentId:                  createdAgent.AgentId,
			CollaboratorName:         pulumi.String(c.Name),
			CollaborationInstruction: pulumi.String(c.CollaborationInstruction),
			AgentDescriptor: &bedrock.AgentAgentCollaboratorAgentDescriptorArgs{
				AliasArn: pulumi.String(c.CollaboratorAliasArn.GetValue()),
			},
		}
		if c.RelayConversationHistory != "" {
			collaboratorArgs.RelayConversationHistory = pulumi.String(c.RelayConversationHistory)
		}
		createdCollaborator, err := bedrock.NewAgentAgentCollaborator(ctx, "collaborator-"+c.Name, collaboratorArgs,
			pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{createdAgent}))
		if err != nil {
			return errors.Wrapf(err, "create collaborator %q", c.Name)
		}
		collaboratorIds[c.Name] = createdCollaborator.CollaboratorId
		draftSatellites = append(draftSatellites, createdCollaborator)
	}
	ctx.Export(OpCollaboratorIds, collaboratorIds)

	associatedKnowledgeBaseIds := pulumi.StringMap{}
	for _, k := range sortedKbAssociations(spec.KnowledgeBaseAssociations) {
		// knowledge_base_state is required at the provider; the spec's
		// omitted default is ENABLED (the AWS default for new
		// associations).
		state := k.State
		if state == "" {
			state = "ENABLED"
		}
		createdAssociation, err := bedrock.NewAgentAgentKnowledgeBaseAssociation(ctx, "kb-association-"+k.Name,
			&bedrock.AgentAgentKnowledgeBaseAssociationArgs{
				AgentId:         createdAgent.AgentId,
				KnowledgeBaseId: pulumi.String(k.KnowledgeBaseId.GetValue()),
				// Required by AWS -- the model reads it to decide when to
				// retrieve.
				Description:        pulumi.String(k.Description),
				KnowledgeBaseState: pulumi.String(state),
			},
			pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{createdAgent}))
		if err != nil {
			return errors.Wrapf(err, "create knowledge base association %q", k.Name)
		}
		associatedKnowledgeBaseIds[k.Name] = createdAssociation.KnowledgeBaseId
		draftSatellites = append(draftSatellites, createdAssociation)
	}
	ctx.Export(OpAssociatedKnowledgeBaseIds, associatedKnowledgeBaseIds)

	// Immutable serving endpoints. Each alias without an explicit routing
	// snapshots the CURRENT draft into a new numbered version, so aliases
	// are created only after every other satellite has landed on the
	// draft -- otherwise the snapshot captures a half-assembled agent.
	aliasIds := pulumi.StringMap{}
	aliasArns := pulumi.StringMap{}
	for _, a := range sortedAliases(spec.Aliases) {
		aliasArgs := &bedrock.AgentAgentAliasArgs{
			AgentAliasName: pulumi.String(a.Name),
			AgentId:        createdAgent.AgentId,
			Tags:           pulumi.ToStringMap(locals.AwsTags),
		}
		if a.Description != "" {
			aliasArgs.Description = pulumi.String(a.Description)
		}
		if a.Routing != nil {
			routing := &bedrock.AgentAgentAliasRoutingConfigurationArgs{}
			if a.Routing.AgentVersion != "" {
				routing.AgentVersion = pulumi.String(a.Routing.AgentVersion)
			}
			if a.Routing.ProvisionedThroughput.GetValue() != "" {
				routing.ProvisionedThroughput = pulumi.String(a.Routing.ProvisionedThroughput.GetValue())
			}
			aliasArgs.RoutingConfigurations = bedrock.AgentAgentAliasRoutingConfigurationArray{routing}
		}
		createdAlias, err := bedrock.NewAgentAgentAlias(ctx, "alias-"+a.Name, aliasArgs,
			pulumi.Provider(provider), pulumi.DependsOn(append([]pulumi.Resource{createdAgent}, draftSatellites...)))
		if err != nil {
			return errors.Wrapf(err, "create alias %q", a.Name)
		}
		aliasIds[a.Name] = createdAlias.AgentAliasId
		aliasArns[a.Name] = createdAlias.AgentAliasArn
	}
	ctx.Export(OpAliasIds, aliasIds)
	ctx.Export(OpAliasArns, aliasArns)

	return nil
}

func sortedActionGroups(in []*awsbedrockagentv1alpha1.AwsBedrockAgentActionGroup) []*awsbedrockagentv1alpha1.AwsBedrockAgentActionGroup {
	out := append([]*awsbedrockagentv1alpha1.AwsBedrockAgentActionGroup{}, in...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func sortedAliases(in []*awsbedrockagentv1alpha1.AwsBedrockAgentAlias) []*awsbedrockagentv1alpha1.AwsBedrockAgentAlias {
	out := append([]*awsbedrockagentv1alpha1.AwsBedrockAgentAlias{}, in...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func sortedCollaborators(in []*awsbedrockagentv1alpha1.AwsBedrockAgentCollaborator) []*awsbedrockagentv1alpha1.AwsBedrockAgentCollaborator {
	out := append([]*awsbedrockagentv1alpha1.AwsBedrockAgentCollaborator{}, in...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func sortedKbAssociations(in []*awsbedrockagentv1alpha1.AwsBedrockAgentKnowledgeBaseAssociation) []*awsbedrockagentv1alpha1.AwsBedrockAgentKnowledgeBaseAssociation {
	out := append([]*awsbedrockagentv1alpha1.AwsBedrockAgentKnowledgeBaseAssociation{}, in...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
