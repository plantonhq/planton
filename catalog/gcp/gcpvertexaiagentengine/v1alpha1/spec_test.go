package gcpvertexaiagentenginev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpVertexAiAgentEngineSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func nameRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: v}},
	}
}

var _ = ginkgo.Describe("GcpVertexAiAgentEngineSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	inlinePython := func() *GcpVertexAiAgentEngineSourceCodeSpec {
		return &GcpVertexAiAgentEngineSourceCodeSpec{
			InlineSource: &GcpVertexAiAgentEngineInlineSource{SourceArchive: "H4sIAAAAAAAAA+3BMQEAAADCoPVPbQwfoAAAAAAAAAAAAAAAAAAAAIC3AYbSVKsAKAAA"},
			PythonSpec: &GcpVertexAiAgentEnginePythonSpec{
				EntrypointModule: "agent",
				EntrypointObject: "root_agent",
			},
		}
	}

	// An ADK agent built from an inline archive.
	minimal := func() *GcpVertexAiAgentEngine {
		return &GcpVertexAiAgentEngine{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpVertexAiAgentEngine",
			Metadata:   &shared.CloudResourceMetadata{Name: "support-agent"},
			Spec: &GcpVertexAiAgentEngineSpec{
				Location: "us-central1",
				Spec: &GcpVertexAiAgentEngineSpecConfig{
					AgentFramework: "google-adk",
					SourceCodeSpec: inlinePython(),
				},
			},
		}
	}

	textPart := func(text string) *GcpVertexAiAgentEngineContentPart {
		return &GcpVertexAiAgentEngineContentPart{Text: text}
	}

	ginkgo.It("should accept an ADK agent from an inline archive", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a bare agent with no spec block", func() {
		msg := minimal()
		msg.Spec.Spec = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a container agent with identity, deployment, networking, and a memory bank", func() {
		msg := minimal()
		msg.Spec.ProjectId = nameRef("ai-project")
		msg.Spec.DisplayName = "Support agent"
		msg.Spec.Description = "Answers support tickets"
		msg.Spec.Labels = map[string]string{"team": "support"}
		msg.Spec.KmsKeyName = litRef("projects/ai-project/locations/us-central1/keyRings/ai/cryptoKeys/agents")
		msg.Spec.Spec = &GcpVertexAiAgentEngineSpecConfig{
			AgentFramework: "custom",
			ClassMethods:   `[{"name":"query","api_mode":"","parameters":{"type":"object","properties":{"input":{"type":"string"}}}}]`,
			IdentityType:   "SERVICE_ACCOUNT",
			ServiceAccount: nameRef("support-agent-sa"),
			ContainerSpec: &GcpVertexAiAgentEngineContainerSpec{
				ImageUri: "us-central1-docker.pkg.dev/ai-project/agents/support:1.4.0",
				Port:     proto.Int32(8080),
			},
			BuildSpec: &GcpVertexAiAgentEngineBuildSpec{WorkerPool: "projects/ai-project/locations/us-central1/workerPools/private"},
			DeploymentSpec: &GcpVertexAiAgentEngineDeploymentSpec{
				Env:                  []*GcpVertexAiAgentEngineEnvVar{{Name: "LOG_LEVEL", Value: "info"}},
				SecretEnv:            []*GcpVertexAiAgentEngineSecretEnvVar{{Name: "OPENAI_KEY", SecretRef: &GcpVertexAiAgentEngineSecretRef{Secret: nameRef("openai-key"), Version: "latest"}}},
				MinInstances:         proto.Int32(1),
				MaxInstances:         proto.Int32(20),
				ContainerConcurrency: proto.Int32(9),
				ResourceLimits:       map[string]string{"cpu": "4", "memory": "8Gi"},
				PscInterfaceConfig: &GcpVertexAiAgentEnginePscInterfaceConfig{
					NetworkAttachment: "agents-attachment",
					DnsPeeringConfigs: []*GcpVertexAiAgentEngineDnsPeeringConfig{{
						Domain:        "internal.corp.",
						TargetProject: litRef("network-project"),
						TargetNetwork: nameRef("shared-vpc"),
					}},
				},
				AgentGatewayConfig: &GcpVertexAiAgentEngineAgentGatewayConfig{
					ClientToAgentConfig:   &GcpVertexAiAgentEngineGatewayTarget{AgentGateway: "projects/ai-project/locations/us-central1/agentGateways/front"},
					AgentToAnywhereConfig: &GcpVertexAiAgentEngineGatewayTarget{AgentGateway: "projects/ai-project/locations/us-central1/agentGateways/egress"},
				},
			},
		}
		msg.Spec.ContextSpec = &GcpVertexAiAgentEngineContextSpec{
			MemoryBankConfig: &GcpVertexAiAgentEngineMemoryBankConfig{
				GenerationConfig: &GcpVertexAiAgentEngineGenerationConfig{
					Model: "projects/ai-project/locations/us-central1/publishers/google/models/gemini-2.5-flash",
					GenerationTriggerConfig: &GcpVertexAiAgentEngineGenerationTriggerConfig{
						GenerationRule: &GcpVertexAiAgentEngineGenerationRule{EventCount: proto.Int32(5), OverlapEventCount: proto.Int32(1)},
					},
				},
				SimilaritySearchConfig: &GcpVertexAiAgentEngineSimilaritySearchConfig{
					EmbeddingModel: "projects/ai-project/locations/us-central1/publishers/google/models/text-embedding-005",
				},
				TtlConfig: &GcpVertexAiAgentEngineTtlConfig{
					GranularTtlConfig:        &GcpVertexAiAgentEngineGranularTtlConfig{CreateTtl: "7776000s", GenerateCreatedTtl: "2592000s", GenerateUpdatedTtl: "2592000s"},
					MemoryRevisionDefaultTtl: "604800s",
				},
				DisableMemoryRevisions: false,
				StructuredMemoryConfigs: []*GcpVertexAiAgentEngineStructuredMemoryConfig{{
					ScopeKeys:     []string{"user_id"},
					SchemaConfigs: []*GcpVertexAiAgentEngineSchemaConfig{{Id: "profile", MemorySchema: `{"type":"object","properties":{"name":{"type":"string"}}}`}},
				}},
				CustomizationConfigs: []*GcpVertexAiAgentEngineCustomizationConfig{{
					ScopeKeys: []string{"user_id"},
					MemoryTopics: []*GcpVertexAiAgentEngineMemoryTopic{
						{ManagedMemoryTopic: &GcpVertexAiAgentEngineManagedMemoryTopic{ManagedTopicEnum: "USER_PREFERENCES"}},
						{CustomMemoryTopic: &GcpVertexAiAgentEngineCustomMemoryTopic{Label: "billing", Description: "Anything about invoices"}},
					},
					GenerateMemoriesExamples: []*GcpVertexAiAgentEngineGenerateMemoriesExample{{
						ConversationSource: &GcpVertexAiAgentEngineConversationSource{Events: []*GcpVertexAiAgentEngineConversationEvent{
							{Content: &GcpVertexAiAgentEngineContent{Role: "user", Parts: []*GcpVertexAiAgentEngineContentPart{textPart("I prefer email over phone calls.")}}},
							{Content: &GcpVertexAiAgentEngineContent{Role: "model", Parts: []*GcpVertexAiAgentEngineContentPart{
								{Thought: true, Text: "Noting a preference."},
								{FunctionCall: &GcpVertexAiAgentEngineFunctionCall{Id: "c1", Name: "save_pref", Args: `{"channel":"email"}`}},
								{FunctionResponse: &GcpVertexAiAgentEngineFunctionResponse{Id: "c1", Name: "save_pref", Response: `{"ok":true}`}},
								{ExecutableCode: &GcpVertexAiAgentEngineExecutableCode{Id: "e1", Language: "PYTHON", Code: "print(1)"}},
								{CodeExecutionResult: &GcpVertexAiAgentEngineCodeExecutionResult{Id: "e1", Outcome: "OUTCOME_OK", Output: "1"}},
								{InlineData: &GcpVertexAiAgentEngineInlineData{MimeType: "image/png", Data: "iVBORw0KGgo="}},
								{FileData: &GcpVertexAiAgentEngineFileData{MimeType: "video/mp4", FileUri: "gs://bucket/clip.mp4"}, VideoMetadata: &GcpVertexAiAgentEngineVideoMetadata{StartOffset: "0s", EndOffset: "3.5s"}},
							}}},
						}},
						GeneratedMemories: []*GcpVertexAiAgentEngineGeneratedMemory{{
							Fact: "The user prefers email.",
							Topics: []*GcpVertexAiAgentEngineGeneratedMemoryTopic{
								{ManagedMemoryTopic: "USER_PREFERENCES"},
								{CustomMemoryTopicLabel: "billing"},
							},
						}},
					}},
					ConsolidationConfig:            &GcpVertexAiAgentEngineConsolidationConfig{RevisionsPerCandidateCount: proto.Int32(3)},
					DisableNaturalLanguageMemories: false,
					EnableThirdPersonMemories:      true,
				}},
			},
		}
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept the legacy pickled package and an ADK config source", func() {
		msg := minimal()
		msg.Spec.Spec = &GcpVertexAiAgentEngineSpecConfig{
			PackageSpec: &GcpVertexAiAgentEnginePackageSpec{
				PickleObjectGcsUri:    "gs://agents/support.pkl",
				DependencyFilesGcsUri: "gs://agents/deps.tar.gz",
				RequirementsGcsUri:    "gs://agents/requirements.txt",
				PythonVersion:         "3.11",
			},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.Spec = &GcpVertexAiAgentEngineSpecConfig{
			SourceCodeSpec: &GcpVertexAiAgentEngineSourceCodeSpec{
				AgentConfigSource: &GcpVertexAiAgentEngineAgentConfigSource{AdkConfig: &GcpVertexAiAgentEngineAdkConfig{JsonConfig: `{"name":"root_agent","model":"gemini-2.5-flash"}`}},
				ImageSpec:         &GcpVertexAiAgentEngineImageSpec{BuildArgs: map[string]string{"PY": "3.12"}},
			},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.Spec.SourceCodeSpec = &GcpVertexAiAgentEngineSourceCodeSpec{
			DeveloperConnectSource: &GcpVertexAiAgentEngineDeveloperConnectSource{Config: &GcpVertexAiAgentEngineDeveloperConnectSourceConfig{
				GitRepositoryLink: "projects/ai-project/locations/us-central1/connections/github/gitRepositoryLinks/agents",
				Dir:               "support",
				Revision:          "main",
			}},
			PythonSpec: &GcpVertexAiAgentEnginePythonSpec{Version: "3.12", RequirementsFile: "requirements.txt"},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require location", func() {
		msg := minimal()
		msg.Spec.Location = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a container spec beside a source spec", func() {
		msg := minimal()
		msg.Spec.Spec.ContainerSpec = &GcpVertexAiAgentEngineContainerSpec{ImageUri: "gcr.io/x/y"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require exactly one source and exactly one build recipe", func() {
		msg := minimal()
		msg.Spec.Spec.SourceCodeSpec.AgentConfigSource = &GcpVertexAiAgentEngineAgentConfigSource{AdkConfig: &GcpVertexAiAgentEngineAdkConfig{JsonConfig: "{}"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Spec.SourceCodeSpec.InlineSource = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Spec.SourceCodeSpec.ImageSpec = &GcpVertexAiAgentEngineImageSpec{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Spec.SourceCodeSpec.PythonSpec = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should forbid a service account under AGENT_IDENTITY and reject an unknown identity type", func() {
		msg := minimal()
		msg.Spec.Spec.IdentityType = "AGENT_IDENTITY"
		msg.Spec.Spec.ServiceAccount = nameRef("sa")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Spec.ServiceAccount = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.Spec.IdentityType = "WORKLOAD"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should bound instances and restrict resource_limits keys", func() {
		msg := minimal()
		msg.Spec.Spec.DeploymentSpec = &GcpVertexAiAgentEngineDeploymentSpec{MinInstances: proto.Int32(5), MaxInstances: proto.Int32(2)}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Spec.DeploymentSpec = &GcpVertexAiAgentEngineDeploymentSpec{MinInstances: proto.Int32(11)}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Spec.DeploymentSpec = &GcpVertexAiAgentEngineDeploymentSpec{ResourceLimits: map[string]string{"gpu": "1"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a secret on every secret env var and a dot-terminated DNS domain", func() {
		msg := minimal()
		msg.Spec.Spec.DeploymentSpec = &GcpVertexAiAgentEngineDeploymentSpec{
			SecretEnv: []*GcpVertexAiAgentEngineSecretEnvVar{{Name: "KEY", SecretRef: &GcpVertexAiAgentEngineSecretRef{}}},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Spec.DeploymentSpec = &GcpVertexAiAgentEngineDeploymentSpec{
			PscInterfaceConfig: &GcpVertexAiAgentEnginePscInterfaceConfig{DnsPeeringConfigs: []*GcpVertexAiAgentEngineDnsPeeringConfig{{
				Domain: "internal.corp", TargetProject: litRef("p"), TargetNetwork: litRef("n"),
			}}},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require exactly one payload per conversation part", func() {
		example := func(part *GcpVertexAiAgentEngineContentPart) *GcpVertexAiAgentEngine {
			msg := minimal()
			msg.Spec.ContextSpec = &GcpVertexAiAgentEngineContextSpec{MemoryBankConfig: &GcpVertexAiAgentEngineMemoryBankConfig{
				CustomizationConfigs: []*GcpVertexAiAgentEngineCustomizationConfig{{
					GenerateMemoriesExamples: []*GcpVertexAiAgentEngineGenerateMemoriesExample{{
						ConversationSource: &GcpVertexAiAgentEngineConversationSource{Events: []*GcpVertexAiAgentEngineConversationEvent{
							{Content: &GcpVertexAiAgentEngineContent{Parts: []*GcpVertexAiAgentEngineContentPart{part}}},
						}},
					}},
				}},
			}}
			return msg
		}
		gomega.Expect(validator.Validate(example(&GcpVertexAiAgentEngineContentPart{Thought: true}))).ToNot(gomega.Succeed())
		gomega.Expect(validator.Validate(example(&GcpVertexAiAgentEngineContentPart{
			Text:     "hi",
			FileData: &GcpVertexAiAgentEngineFileData{MimeType: "image/png", FileUri: "gs://b/o"},
		}))).ToNot(gomega.Succeed())
		gomega.Expect(validator.Validate(example(textPart("hi")))).To(gomega.Succeed())
	})

	ginkgo.It("should require exactly one topic arm and exactly one TTL shape", func() {
		msg := minimal()
		msg.Spec.ContextSpec = &GcpVertexAiAgentEngineContextSpec{MemoryBankConfig: &GcpVertexAiAgentEngineMemoryBankConfig{
			CustomizationConfigs: []*GcpVertexAiAgentEngineCustomizationConfig{{MemoryTopics: []*GcpVertexAiAgentEngineMemoryTopic{{}}}},
		}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.ContextSpec = &GcpVertexAiAgentEngineContextSpec{MemoryBankConfig: &GcpVertexAiAgentEngineMemoryBankConfig{
			TtlConfig: &GcpVertexAiAgentEngineTtlConfig{DefaultTtl: "3600s", GranularTtlConfig: &GcpVertexAiAgentEngineGranularTtlConfig{CreateTtl: "3600s"}},
		}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.ContextSpec.MemoryBankConfig.TtlConfig = &GcpVertexAiAgentEngineTtlConfig{DefaultTtl: "3600s"}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should allow at most one generation trigger", func() {
		msg := minimal()
		msg.Spec.ContextSpec = &GcpVertexAiAgentEngineContextSpec{MemoryBankConfig: &GcpVertexAiAgentEngineMemoryBankConfig{
			GenerationConfig: &GcpVertexAiAgentEngineGenerationConfig{
				Model: "projects/p/locations/us-central1/publishers/google/models/gemini-2.5-flash",
				GenerationTriggerConfig: &GcpVertexAiAgentEngineGenerationTriggerConfig{GenerationRule: &GcpVertexAiAgentEngineGenerationRule{
					EventCount: proto.Int32(3), IdleDuration: "300s",
				}},
			},
		}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.ContextSpec.MemoryBankConfig.GenerationConfig.Model = "gemini-2.5-flash"
		msg.Spec.ContextSpec.MemoryBankConfig.GenerationConfig.GenerationTriggerConfig = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
