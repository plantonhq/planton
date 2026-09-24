package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/discoveryengine"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// chatEngine builds the CHAT arm (`google_discovery_engine_chat_engine`):
// a chat app that either creates a Dialogflow CX agent or links an
// existing one (exactly one, a spec rule). The whole chat_engine_config is
// immutable on Google's side. The created or linked agent's name is read
// back from chat_engine_metadata and exported.
func chatEngine(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider,
	dependsOn []pulumi.Resource) (*createdEngine, error) {
	spec := locals.GcpVertexAiSearchEngine.Spec
	cfg := spec.ChatEngineConfig

	chatConfig := &discoveryengine.ChatEngineChatEngineConfigArgs{
		AllowCrossRegion: pulumi.BoolPtr(cfg.AllowCrossRegion),
	}
	if acc := cfg.AgentCreationConfig; acc != nil {
		agentArgs := &discoveryengine.ChatEngineChatEngineConfigAgentCreationConfigArgs{
			DefaultLanguageCode: pulumi.String(acc.DefaultLanguageCode),
			TimeZone:            pulumi.String(acc.TimeZone),
		}
		if acc.Business != "" {
			agentArgs.Business = pulumi.String(acc.Business)
		}
		if acc.Location != "" {
			agentArgs.Location = pulumi.String(acc.Location)
		}
		chatConfig.AgentCreationConfig = agentArgs
	}
	if cfg.DialogflowAgentToLink.GetValue() != "" {
		chatConfig.DialogflowAgentToLink = pulumi.String(cfg.DialogflowAgentToLink.GetValue())
	}

	args := &discoveryengine.ChatEngineArgs{
		EngineId:         pulumi.String(locals.EngineId),
		DisplayName:      pulumi.String(locals.DisplayName),
		Location:         pulumi.String(spec.Location),
		CollectionId:     pulumi.String(locals.CollectionId),
		DataStoreIds:     dataStoreIds(spec),
		ChatEngineConfig: chatConfig,
	}
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.IndustryVertical != "" {
		args.IndustryVertical = pulumi.String(spec.IndustryVertical)
	}
	if spec.CommonConfig != nil && spec.CommonConfig.CompanyName != "" {
		args.CommonConfig = &discoveryengine.ChatEngineCommonConfigArgs{
			CompanyName: pulumi.String(spec.CommonConfig.CompanyName),
		}
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	created, err := discoveryengine.NewChatEngine(ctx,
		locals.GcpVertexAiSearchEngine.Metadata.Name, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn(dependsOn))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create discovery engine chat engine")
	}

	// Google reports the agent under chat_engine_metadata; the SDK types the
	// block as a list of one.
	dialogflowAgent := created.ChatEngineMetadatas.ApplyT(func(metadata []discoveryengine.ChatEngineChatEngineMetadata) string {
		if len(metadata) > 0 && metadata[0].DialogflowAgent != nil {
			return *metadata[0].DialogflowAgent
		}
		return ""
	}).(pulumi.StringOutput)

	return &createdEngine{
		Resource:        created,
		Name:            created.Name,
		EngineId:        created.EngineId,
		DialogflowAgent: dialogflowAgent,
	}, nil
}
