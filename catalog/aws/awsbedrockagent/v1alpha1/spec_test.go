package awsbedrockagentv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsBedrockAgentSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsBedrockAgentSpec Validation Suite")
}

func svr(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

// minimalAgent is the smallest valid manifest: region, foundation model,
// and the service role.
func minimalAgent() *AwsBedrockAgentSpec {
	return &AwsBedrockAgentSpec{
		Region:               "us-west-2",
		FoundationModel:      "amazon.nova-micro-v1:0",
		AgentResourceRoleArn: svr("arn:aws:iam::123456789012:role/bedrock-agent"),
	}
}

func validActionGroup() *AwsBedrockAgentActionGroup {
	return &AwsBedrockAgentActionGroup{
		Name: "orders",
		Executor: &AwsBedrockAgentActionGroupExecutor{
			Lambda: svr("arn:aws:lambda:us-west-2:123456789012:function:orders"),
		},
		FunctionSchema: &AwsBedrockAgentFunctionSchema{
			Functions: []*AwsBedrockAgentFunction{
				{
					Name:        "get_order",
					Description: "Look up one order by its ID.",
					Parameters: []*AwsBedrockAgentFunctionParameter{
						{Name: "order_id", Type: "string", Required: true},
					},
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AwsBedrockAgentSpec validations", func() {

	// -----------------------------------------------------------------
	// Valid inputs
	// -----------------------------------------------------------------
	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.Context("with minimal required fields", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(minimalAgent())
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with the full surface configured", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := minimalAgent()
				spec.Description = "customer support agent"
				spec.Instruction = strings.Repeat("You are a helpful support agent. ", 3)
				spec.IdleSessionTtlSeconds = 900
				spec.CustomerEncryptionKeyArn = svr("arn:aws:kms:us-west-2:123456789012:key/abc")
				spec.AgentCollaboration = "SUPERVISOR"
				spec.Guardrail = &AwsBedrockAgentGuardrail{
					GuardrailId: svr("gr-abc123"),
					Version:     "1",
				}
				spec.Memory = &AwsBedrockAgentMemory{
					StorageDays:       30,
					MaxRecentSessions: 5,
				}
				spec.PromptOverride = &AwsBedrockAgentPromptOverride{
					OverrideLambda: svr("arn:aws:lambda:us-west-2:123456789012:function:parser"),
					PromptConfigurations: []*AwsBedrockAgentPromptConfiguration{
						{
							PromptType:         "ORCHESTRATION",
							BasePromptTemplate: "$instruction$ $question$",
							ParserMode:         "OVERRIDDEN",
							PromptState:        "ENABLED",
							InferenceConfiguration: &AwsBedrockAgentInferenceConfiguration{
								MaxLength:     int32Ptr(2048),
								StopSequences: []string{"</answer>"},
								Temperature:   float64Ptr(0),
								TopK:          int32Ptr(250),
								TopP:          float64Ptr(1),
							},
						},
					},
				}
				spec.ActionGroups = []*AwsBedrockAgentActionGroup{
					validActionGroup(),
					{
						Name:                       "user-input",
						ParentActionGroupSignature: "AMAZON.UserInput",
					},
				}
				spec.Aliases = []*AwsBedrockAgentAlias{
					{Name: "live", Description: "production endpoint"},
					{Name: "pinned", Routing: &AwsBedrockAgentAliasRouting{AgentVersion: "1"}},
				}
				spec.Collaborators = []*AwsBedrockAgentCollaborator{
					{
						Name:                     "billing",
						CollaborationInstruction: "Handle all billing and invoice questions.",
						CollaboratorAliasArn:     svr("arn:aws:bedrock:us-west-2:123456789012:agent-alias/AGENTID/ALIASID"),
						RelayConversationHistory: "TO_COLLABORATOR",
					},
				}
				spec.KnowledgeBaseAssociations = []*AwsBedrockAgentKnowledgeBaseAssociation{
					{
						Name:            "docs",
						KnowledgeBaseId: svr("EMDPPAYPZI"),
						Description:     "Product documentation for troubleshooting answers.",
						State:           "ENABLED",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with an api_schema action group using an S3 source", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := minimalAgent()
				spec.ActionGroups = []*AwsBedrockAgentActionGroup{{
					Name:     "api",
					Executor: &AwsBedrockAgentActionGroupExecutor{ReturnControl: true},
					ApiSchema: &AwsBedrockAgentApiSchema{
						S3: &AwsBedrockAgentApiSchemaS3{
							BucketName: svr("my-schemas"),
							ObjectKey:  "openapi.yaml",
						},
					},
				}}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	// -----------------------------------------------------------------
	// Field-level rules
	// -----------------------------------------------------------------
	ginkgo.Describe("Field-level rules", func() {

		ginkgo.It("should reject a missing role", func() {
			spec := minimalAgent()
			spec.AgentResourceRoleArn = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an empty foundation model", func() {
			spec := minimalAgent()
			spec.FoundationModel = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a too-short instruction", func() {
			spec := minimalAgent()
			spec.Instruction = "too short"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an idle TTL out of range", func() {
			spec := minimalAgent()
			spec.IdleSessionTtlSeconds = 30
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown collaboration mode", func() {
			spec := minimalAgent()
			spec.AgentCollaboration = "ROUTER"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a guardrail without a version", func() {
			spec := minimalAgent()
			spec.Guardrail = &AwsBedrockAgentGuardrail{GuardrailId: svr("gr-abc")}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept memory declared and switched off", func() {
			spec := minimalAgent()
			spec.Memory = &AwsBedrockAgentMemory{Enabled: boolPtr(false), StorageDays: 30}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("should reject memory storage days out of range", func() {
			spec := minimalAgent()
			spec.Memory = &AwsBedrockAgentMemory{StorageDays: 366}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -----------------------------------------------------------------
	// Satellite uniqueness and collaboration mode
	// -----------------------------------------------------------------
	ginkgo.Describe("Satellite collections", func() {

		ginkgo.It("should reject duplicate action group names", func() {
			spec := minimalAgent()
			g1 := validActionGroup()
			g2 := validActionGroup()
			spec.ActionGroups = []*AwsBedrockAgentActionGroup{g1, g2}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("unique"))
		})

		ginkgo.It("should reject duplicate alias names", func() {
			spec := minimalAgent()
			spec.Aliases = []*AwsBedrockAgentAlias{{Name: "live"}, {Name: "live"}}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject collaborators without a supervisor mode", func() {
			spec := minimalAgent()
			spec.Collaborators = []*AwsBedrockAgentCollaborator{{
				Name:                     "billing",
				CollaborationInstruction: "Handle billing.",
				CollaboratorAliasArn:     svr("arn:aws:bedrock:us-west-2:123456789012:agent-alias/A/B"),
			}}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("SUPERVISOR"))
		})

		ginkgo.It("should accept collaborators in SUPERVISOR_ROUTER mode", func() {
			spec := minimalAgent()
			spec.AgentCollaboration = "SUPERVISOR_ROUTER"
			spec.Collaborators = []*AwsBedrockAgentCollaborator{{
				Name:                     "billing",
				CollaborationInstruction: "Handle billing.",
				CollaboratorAliasArn:     svr("arn:aws:bedrock:us-west-2:123456789012:agent-alias/A/B"),
			}}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a knowledge base association without a description", func() {
			spec := minimalAgent()
			spec.KnowledgeBaseAssociations = []*AwsBedrockAgentKnowledgeBaseAssociation{{
				Name:            "docs",
				KnowledgeBaseId: svr("EMDPPAYPZI"),
			}}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -----------------------------------------------------------------
	// Action group shape
	// -----------------------------------------------------------------
	ginkgo.Describe("Action group shape", func() {

		ginkgo.It("should reject a custom group without an executor", func() {
			spec := minimalAgent()
			g := validActionGroup()
			g.Executor = nil
			spec.ActionGroups = []*AwsBedrockAgentActionGroup{g}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a custom group with both schemas", func() {
			spec := minimalAgent()
			g := validActionGroup()
			g.ApiSchema = &AwsBedrockAgentApiSchema{Payload: "{}"}
			spec.ActionGroups = []*AwsBedrockAgentActionGroup{g}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a reserved group carrying an executor", func() {
			spec := minimalAgent()
			spec.ActionGroups = []*AwsBedrockAgentActionGroup{{
				Name:                       "user-input",
				ParentActionGroupSignature: "AMAZON.UserInput",
				Executor:                   &AwsBedrockAgentActionGroupExecutor{ReturnControl: true},
			}}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a reserved group with a description", func() {
			spec := minimalAgent()
			spec.ActionGroups = []*AwsBedrockAgentActionGroup{{
				Name:                       "user-input",
				ParentActionGroupSignature: "AMAZON.UserInput",
				Description:                "reserved",
			}}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an executor with both arms", func() {
			spec := minimalAgent()
			g := validActionGroup()
			g.Executor = &AwsBedrockAgentActionGroupExecutor{
				Lambda:        svr("arn:aws:lambda:us-west-2:123456789012:function:orders"),
				ReturnControl: true,
			}
			spec.ActionGroups = []*AwsBedrockAgentActionGroup{g}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an api_schema with both payload and s3", func() {
			spec := minimalAgent()
			spec.ActionGroups = []*AwsBedrockAgentActionGroup{{
				Name:     "api",
				Executor: &AwsBedrockAgentActionGroupExecutor{ReturnControl: true},
				ApiSchema: &AwsBedrockAgentApiSchema{
					Payload: "{}",
					S3: &AwsBedrockAgentApiSchemaS3{
						BucketName: svr("b"),
						ObjectKey:  "k",
					},
				},
			}}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject duplicate function parameter names", func() {
			spec := minimalAgent()
			g := validActionGroup()
			g.FunctionSchema.Functions[0].Parameters = []*AwsBedrockAgentFunctionParameter{
				{Name: "id", Type: "string"},
				{Name: "id", Type: "number"},
			}
			spec.ActionGroups = []*AwsBedrockAgentActionGroup{g}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -----------------------------------------------------------------
	// Prompt override
	// -----------------------------------------------------------------
	ginkgo.Describe("Prompt override", func() {

		ginkgo.It("should reject an OVERRIDDEN parser without the lambda", func() {
			spec := minimalAgent()
			spec.PromptOverride = &AwsBedrockAgentPromptOverride{
				PromptConfigurations: []*AwsBedrockAgentPromptConfiguration{{
					PromptType:         "ORCHESTRATION",
					BasePromptTemplate: "$question$",
					ParserMode:         "OVERRIDDEN",
				}},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("override_lambda"))
		})

		ginkgo.It("should reject duplicate prompt types", func() {
			spec := minimalAgent()
			spec.PromptOverride = &AwsBedrockAgentPromptOverride{
				PromptConfigurations: []*AwsBedrockAgentPromptConfiguration{
					{PromptType: "ORCHESTRATION", BasePromptTemplate: "a"},
					{PromptType: "ORCHESTRATION", BasePromptTemplate: "b"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a temperature above 1", func() {
			spec := minimalAgent()
			spec.PromptOverride = &AwsBedrockAgentPromptOverride{
				PromptConfigurations: []*AwsBedrockAgentPromptConfiguration{{
					PromptType:         "ORCHESTRATION",
					BasePromptTemplate: "$question$",
					InferenceConfiguration: &AwsBedrockAgentInferenceConfiguration{
						Temperature: float64Ptr(1.5),
					},
				}},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -----------------------------------------------------------------
	// Alias routing
	// -----------------------------------------------------------------
	ginkgo.Describe("Alias routing", func() {

		ginkgo.It("should reject an empty routing entry", func() {
			spec := minimalAgent()
			spec.Aliases = []*AwsBedrockAgentAlias{{
				Name:    "live",
				Routing: &AwsBedrockAgentAliasRouting{},
			}}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept routing pinned to provisioned throughput", func() {
			spec := minimalAgent()
			spec.Aliases = []*AwsBedrockAgentAlias{{
				Name: "live",
				Routing: &AwsBedrockAgentAliasRouting{
					AgentVersion:          "2",
					ProvisionedThroughput: svr("arn:aws:bedrock:us-west-2:123456789012:provisioned-model/abc"),
				},
			}}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
})

func int32Ptr(v int32) *int32 {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}

func float64Ptr(v float64) *float64 {
	return &v
}
