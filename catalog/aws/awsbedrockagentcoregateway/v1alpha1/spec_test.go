package awsbedrockagentcoregatewayv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAwsBedrockAgentCoreGatewaySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsBedrockAgentCoreGatewaySpec Validation Suite")
}

func svr(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

// minimalGateway is the smallest valid manifest: region, role, and IAM
// inbound auth.
func minimalGateway() *AwsBedrockAgentCoreGatewaySpec {
	return &AwsBedrockAgentCoreGatewaySpec{
		Region:         "us-west-2",
		RoleArn:        svr("arn:aws:iam::123456789012:role/agentcore-gateway"),
		AuthorizerType: "AWS_IAM",
	}
}

func validLambdaTarget() *AwsBedrockAgentCoreGatewayTarget {
	return &AwsBedrockAgentCoreGatewayTarget{
		Name: "orders",
		Backend: &AwsBedrockAgentCoreGatewayTargetBackend{
			Lambda: &AwsBedrockAgentCoreGatewayLambdaTarget{
				LambdaArn: svr("arn:aws:lambda:us-west-2:123456789012:function:orders"),
				Tools: []*AwsBedrockAgentCoreGatewayToolDefinition{
					{
						Name:        "get_order",
						Description: "Look up one order by its ID.",
						InputSchema: &AwsBedrockAgentCoreGatewaySchemaDefinition{
							Type: "object",
							Properties: []*AwsBedrockAgentCoreGatewaySchemaProperty{
								{Name: "order_id", Type: "string", Required: true},
							},
						},
					},
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AwsBedrockAgentCoreGatewaySpec validations", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.Context("with minimal required fields", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(minimalGateway())
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with the full surface configured", func() {
			ginkgo.It("should not return a validation error", func() {
				deepItems, perr := structpb.NewStruct(map[string]any{"type": "string"})
				gomega.Expect(perr).To(gomega.BeNil())
				spec := minimalGateway()
				spec.Description = "front door for order tools"
				spec.AuthorizerType = "CUSTOM_JWT"
				spec.CustomJwtAuthorizer = &AwsBedrockAgentCoreGatewayJwtAuthorizer{
					DiscoveryUrl:    "https://issuer.example.com/.well-known/openid-configuration",
					AllowedAudience: []string{"agents"},
					AllowedClients:  []string{"client-1"},
				}
				spec.KmsKeyArn = svr("arn:aws:kms:us-west-2:123456789012:key/abc")
				spec.ExposeDebugExceptions = true
				spec.Mcp = &AwsBedrockAgentCoreGatewayMcp{
					Instructions:            "Tools for querying and managing orders.",
					EnableSemanticSearch:    true,
					SupportedVersions:       []string{"2025-03-26"},
					SessionTimeoutSeconds:   900,
					EnableResponseStreaming: true,
				}
				spec.Interceptors = []*AwsBedrockAgentCoreGatewayInterceptor{
					{
						InterceptionPoints: []string{"REQUEST"},
						LambdaArn:          svr("arn:aws:lambda:us-west-2:123456789012:function:rewrite"),
						PassRequestHeaders: boolPtr(true),
					},
				}
				spec.PolicyEngine = &AwsBedrockAgentCoreGatewayPolicyEngine{
					PolicyEngineArn: svr("arn:aws:bedrock-agentcore:us-west-2:123456789012:policy-engine/pe-abc"),
					Mode:            "LOG_ONLY",
				}
				lambdaTarget := validLambdaTarget()
				lambdaTarget.Backend.Lambda.Tools[0].OutputSchema = &AwsBedrockAgentCoreGatewaySchemaDefinition{
					Type: "array",
					Items: &AwsBedrockAgentCoreGatewaySchemaItems{
						Type: "object",
						Properties: []*AwsBedrockAgentCoreGatewaySchemaPropertyLeaf{
							{Name: "line_items", Type: "array", ItemsJson: deepItems},
						},
					},
				}
				lambdaTarget.Credentials = &AwsBedrockAgentCoreGatewayTargetCredentials{
					Provider: &AwsBedrockAgentCoreGatewayTargetCredentials_GatewayIamRole{GatewayIamRole: &AwsBedrockAgentCoreGatewaySigv4Credentials{
						Service: "lambda",
						Region:  "us-west-2",
					}},
				}
				spec.Targets = []*AwsBedrockAgentCoreGatewayTarget{
					lambdaTarget,
					{
						Name:        "docs-api",
						Description: "existing OpenAPI backend",
						Backend: &AwsBedrockAgentCoreGatewayTargetBackend{
							OpenApiSchema: &AwsBedrockAgentCoreGatewaySchemaTarget{
								InlinePayload: `{"openapi":"3.0.0"}`,
							},
						},
						Credentials: &AwsBedrockAgentCoreGatewayTargetCredentials{
							Provider: &AwsBedrockAgentCoreGatewayTargetCredentials_ApiKey{ApiKey: &AwsBedrockAgentCoreGatewayApiKeyCredentials{
								ProviderArn:             svr("arn:aws:bedrock-agentcore:us-west-2:123456789012:token-vault/default/apikeycredentialprovider/docs"),
								CredentialLocation:      "HEADER",
								CredentialParameterName: "X-Api-Key",
							}},
						},
						Metadata: &AwsBedrockAgentCoreGatewayTargetMetadata{
							AllowedRequestHeaders: []string{"X-Trace-Id"},
						},
					},
					{
						Name: "runtime",
						Backend: &AwsBedrockAgentCoreGatewayTargetBackend{
							AgentcoreRuntime: &AwsBedrockAgentCoreGatewayRuntimeTarget{
								AgentRuntimeArn: svr("arn:aws:bedrock-agentcore:us-west-2:123456789012:runtime/rt-abc"),
								Qualifier:       "DEFAULT",
							},
						},
					},
					{
						Name: "remote-mcp",
						Backend: &AwsBedrockAgentCoreGatewayTargetBackend{
							McpServer: &AwsBedrockAgentCoreGatewayMcpServerTarget{
								Endpoint:    "https://mcp.example.com/mcp",
								ListingMode: "DYNAMIC",
							},
						},
						Credentials: &AwsBedrockAgentCoreGatewayTargetCredentials{
							Provider: &AwsBedrockAgentCoreGatewayTargetCredentials_Oauth{Oauth: &AwsBedrockAgentCoreGatewayOauthCredentials{
								ProviderArn: svr("arn:aws:bedrock-agentcore:us-west-2:123456789012:token-vault/default/oauth2credentialprovider/gh"),
								Scopes:      []string{"repo"},
								GrantType:   "CLIENT_CREDENTIALS",
							}},
						},
					},
					{
						Name: "rest-api",
						Backend: &AwsBedrockAgentCoreGatewayTargetBackend{
							ApiGateway: &AwsBedrockAgentCoreGatewayApiGatewayTarget{
								RestApiId: "a1b2c3",
								Stage:     "prod",
								ToolFilters: []*AwsBedrockAgentCoreGatewayApiGatewayToolFilter{
									{FilterPath: "/orders/*", Methods: []string{"GET", "POST"}},
								},
								ToolOverrides: []*AwsBedrockAgentCoreGatewayApiGatewayToolOverride{
									{Path: "/orders/{id}", Method: "GET", Name: "get_order", Description: "Fetch one order."},
								},
							},
						},
						Credentials: &AwsBedrockAgentCoreGatewayTargetCredentials{
							Provider: &AwsBedrockAgentCoreGatewayTargetCredentials_CallerIamCredentials{CallerIamCredentials: &AwsBedrockAgentCoreGatewaySigv4Credentials{
								Service: "execute-api",
							}},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.Context("with an unknown authorizer type", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGateway()
				spec.AuthorizerType = "API_KEY"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with CUSTOM_JWT and no authorizer rules", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGateway()
				spec.AuthorizerType = "CUSTOM_JWT"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a session timeout below AWS's floor", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGateway()
				spec.Mcp = &AwsBedrockAgentCoreGatewayMcp{SessionTimeoutSeconds: 60}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with three interceptors", func() {
			ginkgo.It("should return a validation error", func() {
				i := &AwsBedrockAgentCoreGatewayInterceptor{
					InterceptionPoints: []string{"REQUEST"},
					LambdaArn:          svr("arn:aws:lambda:us-west-2:123456789012:function:f"),
				}
				spec := minimalGateway()
				spec.Interceptors = []*AwsBedrockAgentCoreGatewayInterceptor{i, i, i}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a target name carrying an underscore", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGateway()
				target := validLambdaTarget()
				target.Name = "orders_v1"
				spec.Targets = []*AwsBedrockAgentCoreGatewayTarget{target}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with duplicate target names", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGateway()
				spec.Targets = []*AwsBedrockAgentCoreGatewayTarget{validLambdaTarget(), validLambdaTarget()}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a target setting two backend arms", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGateway()
				target := validLambdaTarget()
				target.Backend.McpServer = &AwsBedrockAgentCoreGatewayMcpServerTarget{
					Endpoint: "https://mcp.example.com/mcp",
				}
				spec.Targets = []*AwsBedrockAgentCoreGatewayTarget{target}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a lambda target carrying both tool sources", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGateway()
				target := validLambdaTarget()
				target.Backend.Lambda.ToolsS3 = &AwsBedrockAgentCoreGatewaySchemaS3{
					Uri: "s3://bucket/tools.json",
				}
				spec.Targets = []*AwsBedrockAgentCoreGatewayTarget{target}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with duplicate tool names on a lambda target", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGateway()
				target := validLambdaTarget()
				target.Backend.Lambda.Tools = append(target.Backend.Lambda.Tools, target.Backend.Lambda.Tools[0])
				spec.Targets = []*AwsBedrockAgentCoreGatewayTarget{target}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a schema node carrying both properties and items", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGateway()
				target := validLambdaTarget()
				target.Backend.Lambda.Tools[0].InputSchema.Items = &AwsBedrockAgentCoreGatewaySchemaItems{Type: "string"}
				spec.Targets = []*AwsBedrockAgentCoreGatewayTarget{target}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a leaf node carrying both raw-JSON shapes", func() {
			ginkgo.It("should return a validation error", func() {
				j, perr := structpb.NewStruct(map[string]any{"type": "string"})
				gomega.Expect(perr).To(gomega.BeNil())
				spec := minimalGateway()
				target := validLambdaTarget()
				target.Backend.Lambda.Tools[0].InputSchema.Properties[0].Properties = []*AwsBedrockAgentCoreGatewaySchemaPropertyLeaf{
					{Name: "deep", Type: "object", ItemsJson: j, PropertiesJson: j},
				}
				spec.Targets = []*AwsBedrockAgentCoreGatewayTarget{target}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with an mcp_server endpoint that is not https", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGateway()
				spec.Targets = []*AwsBedrockAgentCoreGatewayTarget{
					{
						Name: "remote",
						Backend: &AwsBedrockAgentCoreGatewayTargetBackend{
							McpServer: &AwsBedrockAgentCoreGatewayMcpServerTarget{
								Endpoint: "http://mcp.example.com/mcp",
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with the credential provider as a oneof", func() {
			ginkgo.It("carries at most one arm by construction and accepts the JWT pass-through arm", func() {
				spec := minimalGateway()
				target := validLambdaTarget()
				target.Credentials = &AwsBedrockAgentCoreGatewayTargetCredentials{
					Provider: &AwsBedrockAgentCoreGatewayTargetCredentials_GatewayIamRole{GatewayIamRole: &AwsBedrockAgentCoreGatewaySigv4Credentials{Service: "lambda"}},
				}
				target.Credentials.Provider = &AwsBedrockAgentCoreGatewayTargetCredentials_JwtPassthrough{JwtPassthrough: &AwsBedrockAgentCoreGatewayJwtPassthroughCredentials{}}
				gomega.Expect(target.Credentials.GetGatewayIamRole()).To(gomega.BeNil())
				spec.Targets = []*AwsBedrockAgentCoreGatewayTarget{target}
				gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
			})

			ginkgo.It("accepts metadata propagation declared and switched off", func() {
				spec := minimalGateway()
				target := validLambdaTarget()
				target.Metadata = &AwsBedrockAgentCoreGatewayTargetMetadata{Enabled: proto.Bool(false), AllowedRequestHeaders: []string{"X-Trace-Id"}}
				spec.Targets = []*AwsBedrockAgentCoreGatewayTarget{target}
				gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a SigV4 region and no service", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGateway()
				target := validLambdaTarget()
				target.Credentials = &AwsBedrockAgentCoreGatewayTargetCredentials{
					Provider: &AwsBedrockAgentCoreGatewayTargetCredentials_GatewayIamRole{GatewayIamRole: &AwsBedrockAgentCoreGatewaySigv4Credentials{Region: "us-west-2"}},
				}
				spec.Targets = []*AwsBedrockAgentCoreGatewayTarget{target}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with an AUTHORIZATION_CODE oauth credential missing its return URL", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGateway()
				target := validLambdaTarget()
				target.Credentials = &AwsBedrockAgentCoreGatewayTargetCredentials{
					Provider: &AwsBedrockAgentCoreGatewayTargetCredentials_Oauth{Oauth: &AwsBedrockAgentCoreGatewayOauthCredentials{
						ProviderArn: svr("arn:aws:bedrock-agentcore:us-west-2:123456789012:token-vault/default/oauth2credentialprovider/gh"),
						Scopes:      []string{"repo"},
						GrantType:   "AUTHORIZATION_CODE",
					}},
				}
				spec.Targets = []*AwsBedrockAgentCoreGatewayTarget{target}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a schema-target setting both sources", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGateway()
				spec.Targets = []*AwsBedrockAgentCoreGatewayTarget{
					{
						Name: "docs",
						Backend: &AwsBedrockAgentCoreGatewayTargetBackend{
							SmithyModel: &AwsBedrockAgentCoreGatewaySchemaTarget{
								InlinePayload: "namespace demo",
								S3:            &AwsBedrockAgentCoreGatewaySchemaS3{Uri: "s3://bucket/model.smithy"},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with eleven allowed request headers", func() {
			ginkgo.It("should return a validation error", func() {
				headers := make([]string, 11)
				for i := range headers {
					headers[i] = "X-H"
				}
				spec := minimalGateway()
				target := validLambdaTarget()
				target.Metadata = &AwsBedrockAgentCoreGatewayTargetMetadata{AllowedRequestHeaders: headers}
				spec.Targets = []*AwsBedrockAgentCoreGatewayTarget{target}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})

func boolPtr(v bool) *bool {
	return &v
}
