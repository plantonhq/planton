package module

import (
	gcpvertexaiagentenginev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaiagentengine/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/vertex"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildMemoryBank maps the Memory Bank configuration: generation, lookup,
// expiry, structured schemas, and per-scope customization with its worked
// examples. Example conversations carry Gemini's Content.Part union; the
// audio_transcription payload is not offered because the pinned SDK lacks
// it (an argument one engine cannot send is never a one-engine field).
func buildMemoryBank(mb *gcpvertexaiagentenginev1alpha1.GcpVertexAiAgentEngineMemoryBankConfig) *vertex.AiReasoningEngineContextSpecMemoryBankConfigArgs {
	args := &vertex.AiReasoningEngineContextSpecMemoryBankConfigArgs{}

	if g := mb.GenerationConfig; g != nil {
		gen := &vertex.AiReasoningEngineContextSpecMemoryBankConfigGenerationConfigArgs{
			Model: pulumi.String(g.Model),
		}
		if t := g.GenerationTriggerConfig; t != nil {
			trigger := &vertex.AiReasoningEngineContextSpecMemoryBankConfigGenerationConfigGenerationTriggerConfigArgs{}
			if r := t.GenerationRule; r != nil {
				rule := &vertex.AiReasoningEngineContextSpecMemoryBankConfigGenerationConfigGenerationTriggerConfigGenerationRuleArgs{}
				if r.EventCount != nil {
					rule.EventCount = pulumi.Int(int(r.GetEventCount()))
				}
				if r.FixedInterval != "" {
					rule.FixedInterval = pulumi.String(r.FixedInterval)
				}
				if r.IdleDuration != "" {
					rule.IdleDuration = pulumi.String(r.IdleDuration)
				}
				if r.OverlapEventCount != nil {
					rule.OverlapEventCount = pulumi.Int(int(r.GetOverlapEventCount()))
				}
				trigger.GenerationRule = rule
			}
			gen.GenerationTriggerConfig = trigger
		}
		args.GenerationConfig = gen
	}

	if s := mb.SimilaritySearchConfig; s != nil {
		args.SimilaritySearchConfig = &vertex.AiReasoningEngineContextSpecMemoryBankConfigSimilaritySearchConfigArgs{
			EmbeddingModel: pulumi.String(s.EmbeddingModel),
		}
	}

	if t := mb.TtlConfig; t != nil {
		ttl := &vertex.AiReasoningEngineContextSpecMemoryBankConfigTtlConfigArgs{}
		if t.DefaultTtl != "" {
			ttl.DefaultTtl = pulumi.String(t.DefaultTtl)
		}
		if t.MemoryRevisionDefaultTtl != "" {
			ttl.MemoryRevisionDefaultTtl = pulumi.String(t.MemoryRevisionDefaultTtl)
		}
		if g := t.GranularTtlConfig; g != nil {
			granular := &vertex.AiReasoningEngineContextSpecMemoryBankConfigTtlConfigGranularTtlConfigArgs{}
			if g.CreateTtl != "" {
				granular.CreateTtl = pulumi.String(g.CreateTtl)
			}
			if g.GenerateCreatedTtl != "" {
				granular.GenerateCreatedTtl = pulumi.String(g.GenerateCreatedTtl)
			}
			if g.GenerateUpdatedTtl != "" {
				granular.GenerateUpdatedTtl = pulumi.String(g.GenerateUpdatedTtl)
			}
			ttl.GranularTtlConfig = granular
		}
		args.TtlConfig = ttl
	}

	if mb.DisableMemoryRevisions {
		args.DisableMemoryRevisions = pulumi.Bool(true)
	}

	if len(mb.StructuredMemoryConfigs) > 0 {
		configs := vertex.AiReasoningEngineContextSpecMemoryBankConfigStructuredMemoryConfigArray{}
		for _, c := range mb.StructuredMemoryConfigs {
			cfg := &vertex.AiReasoningEngineContextSpecMemoryBankConfigStructuredMemoryConfigArgs{}
			if len(c.ScopeKeys) > 0 {
				cfg.ScopeKeys = pulumi.ToStringArray(c.ScopeKeys)
			}
			if len(c.SchemaConfigs) > 0 {
				schemas := vertex.AiReasoningEngineContextSpecMemoryBankConfigStructuredMemoryConfigSchemaConfigArray{}
				for _, s := range c.SchemaConfigs {
					schema := &vertex.AiReasoningEngineContextSpecMemoryBankConfigStructuredMemoryConfigSchemaConfigArgs{
						Id: pulumi.String(s.Id),
					}
					if s.MemorySchema != "" {
						schema.MemorySchema = pulumi.String(s.MemorySchema)
					}
					schemas = append(schemas, schema)
				}
				cfg.SchemaConfigs = schemas
			}
			configs = append(configs, cfg)
		}
		args.StructuredMemoryConfigs = configs
	}

	if len(mb.CustomizationConfigs) > 0 {
		configs := vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigArray{}
		for _, c := range mb.CustomizationConfigs {
			configs = append(configs, buildCustomizationConfig(c))
		}
		args.CustomizationConfigs = configs
	}

	return args
}

func buildCustomizationConfig(c *gcpvertexaiagentenginev1alpha1.GcpVertexAiAgentEngineCustomizationConfig) *vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigArgs {
	args := &vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigArgs{}
	if len(c.ScopeKeys) > 0 {
		args.ScopeKeys = pulumi.ToStringArray(c.ScopeKeys)
	}
	if len(c.MemoryTopics) > 0 {
		topics := vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigMemoryTopicArray{}
		for _, t := range c.MemoryTopics {
			topic := &vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigMemoryTopicArgs{}
			if t.CustomMemoryTopic != nil {
				custom := &vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigMemoryTopicCustomMemoryTopicArgs{
					Label: pulumi.String(t.CustomMemoryTopic.Label),
				}
				if t.CustomMemoryTopic.Description != "" {
					custom.Description = pulumi.String(t.CustomMemoryTopic.Description)
				}
				topic.CustomMemoryTopic = custom
			}
			if t.ManagedMemoryTopic != nil {
				topic.ManagedMemoryTopic = &vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigMemoryTopicManagedMemoryTopicArgs{
					ManagedTopicEnum: pulumi.String(t.ManagedMemoryTopic.ManagedTopicEnum),
				}
			}
			topics = append(topics, topic)
		}
		args.MemoryTopics = topics
	}
	if len(c.GenerateMemoriesExamples) > 0 {
		examples := vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigGenerateMemoriesExampleArray{}
		for _, e := range c.GenerateMemoriesExamples {
			examples = append(examples, buildGenerateMemoriesExample(e))
		}
		args.GenerateMemoriesExamples = examples
	}
	if cc := c.ConsolidationConfig; cc != nil {
		consolidation := &vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigConsolidationConfigArgs{}
		if cc.RevisionsPerCandidateCount != nil {
			consolidation.RevisionsPerCandidateCount = pulumi.Int(int(cc.GetRevisionsPerCandidateCount()))
		}
		args.ConsolidationConfig = consolidation
	}
	if c.DisableNaturalLanguageMemories {
		args.DisableNaturalLanguageMemories = pulumi.Bool(true)
	}
	if c.EnableThirdPersonMemories {
		args.EnableThirdPersonMemories = pulumi.Bool(true)
	}
	return args
}

func buildGenerateMemoriesExample(e *gcpvertexaiagentenginev1alpha1.GcpVertexAiAgentEngineGenerateMemoriesExample) *vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigGenerateMemoriesExampleArgs {
	args := &vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigGenerateMemoriesExampleArgs{}
	if cs := e.ConversationSource; cs != nil {
		events := vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigGenerateMemoriesExampleConversationSourceEventArray{}
		for _, ev := range cs.Events {
			content := &vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigGenerateMemoriesExampleConversationSourceEventContentArgs{}
			if ev.Content != nil {
				if ev.Content.Role != "" {
					content.Role = pulumi.String(ev.Content.Role)
				}
				parts := vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigGenerateMemoriesExampleConversationSourceEventContentPartArray{}
				for _, p := range ev.Content.Parts {
					parts = append(parts, buildContentPart(p))
				}
				content.Parts = parts
			}
			events = append(events, &vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigGenerateMemoriesExampleConversationSourceEventArgs{
				Content: content,
			})
		}
		args.ConversationSource = &vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigGenerateMemoriesExampleConversationSourceArgs{
			Events: events,
		}
	}
	if len(e.GeneratedMemories) > 0 {
		memories := vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigGenerateMemoriesExampleGeneratedMemoryArray{}
		for _, m := range e.GeneratedMemories {
			memory := &vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigGenerateMemoriesExampleGeneratedMemoryArgs{
				Fact: pulumi.String(m.Fact),
			}
			if len(m.Topics) > 0 {
				topics := vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigGenerateMemoriesExampleGeneratedMemoryTopicArray{}
				for _, t := range m.Topics {
					topic := &vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigGenerateMemoriesExampleGeneratedMemoryTopicArgs{}
					if t.CustomMemoryTopicLabel != "" {
						topic.CustomMemoryTopicLabel = pulumi.String(t.CustomMemoryTopicLabel)
					}
					if t.ManagedMemoryTopic != "" {
						topic.ManagedMemoryTopic = pulumi.String(t.ManagedMemoryTopic)
					}
					topics = append(topics, topic)
				}
				memory.Topics = topics
			}
			memories = append(memories, memory)
		}
		args.GeneratedMemories = memories
	}
	return args
}

// buildContentPart maps one Content.Part: exactly one payload
// (proto-enforced) plus the thought flag and video clip bounds.
func buildContentPart(p *gcpvertexaiagentenginev1alpha1.GcpVertexAiAgentEngineContentPart) *vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigGenerateMemoriesExampleConversationSourceEventContentPartArgs {
	args := &vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigGenerateMemoriesExampleConversationSourceEventContentPartArgs{}
	if p.Text != "" {
		args.Text = pulumi.String(p.Text)
	}
	if p.Thought {
		args.Thought = pulumi.Bool(true)
	}
	if d := p.InlineData; d != nil {
		args.InlineData = &vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigGenerateMemoriesExampleConversationSourceEventContentPartInlineDataArgs{
			MimeType: pulumi.String(d.MimeType),
			Data:     pulumi.String(d.Data),
		}
	}
	if f := p.FileData; f != nil {
		args.FileData = &vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigGenerateMemoriesExampleConversationSourceEventContentPartFileDataArgs{
			MimeType: pulumi.String(f.MimeType),
			FileUri:  pulumi.String(f.FileUri),
		}
	}
	if fc := p.FunctionCall; fc != nil {
		call := &vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigGenerateMemoriesExampleConversationSourceEventContentPartFunctionCallArgs{}
		if fc.Id != "" {
			call.Id = pulumi.String(fc.Id)
		}
		if fc.Name != "" {
			call.Name = pulumi.String(fc.Name)
		}
		if fc.Args != "" {
			call.Args = pulumi.String(fc.Args)
		}
		args.FunctionCall = call
	}
	if fr := p.FunctionResponse; fr != nil {
		response := &vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigGenerateMemoriesExampleConversationSourceEventContentPartFunctionResponseArgs{
			Name: pulumi.String(fr.Name),
		}
		if fr.Id != "" {
			response.Id = pulumi.String(fr.Id)
		}
		if fr.Response != "" {
			response.Response = pulumi.String(fr.Response)
		}
		args.FunctionResponse = response
	}
	if ec := p.ExecutableCode; ec != nil {
		code := &vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigGenerateMemoriesExampleConversationSourceEventContentPartExecutableCodeArgs{
			Language: pulumi.String(ec.Language),
			Code:     pulumi.String(ec.Code),
		}
		if ec.Id != "" {
			code.Id = pulumi.String(ec.Id)
		}
		args.ExecutableCode = code
	}
	if cr := p.CodeExecutionResult; cr != nil {
		result := &vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigGenerateMemoriesExampleConversationSourceEventContentPartCodeExecutionResultArgs{
			Outcome: pulumi.String(cr.Outcome),
		}
		if cr.Id != "" {
			result.Id = pulumi.String(cr.Id)
		}
		if cr.Output != "" {
			result.Output = pulumi.String(cr.Output)
		}
		args.CodeExecutionResult = result
	}
	if vm := p.VideoMetadata; vm != nil {
		video := &vertex.AiReasoningEngineContextSpecMemoryBankConfigCustomizationConfigGenerateMemoriesExampleConversationSourceEventContentPartVideoMetadataArgs{}
		if vm.StartOffset != "" {
			video.StartOffset = pulumi.String(vm.StartOffset)
		}
		if vm.EndOffset != "" {
			video.EndOffset = pulumi.String(vm.EndOffset)
		}
		args.VideoMetadata = video
	}
	return args
}
