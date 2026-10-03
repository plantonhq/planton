package awsappsyncapiv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsAppSyncApi(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsAppSyncApi Suite")
}

func val(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func ref(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

var _ = ginkgo.Describe("AwsAppSyncApi", func() {

	var input *AwsAppSyncApi

	ginkgo.BeforeEach(func() {
		input = &AwsAppSyncApi{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsAppSyncApi",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "a-test-name",
			},
			Spec: &AwsAppSyncApiSpec{
				Region: "us-west-2",
				Graphql: &AwsAppSyncGraphqlApi{
					ApiName: "orders_api",
					Auth:    &AwsAppSyncGraphqlAuthProvider{Type: "API_KEY"},
				},
			},
		}
	})

	ginkgo.Context("when valid input is passed", func() {
		ginkgo.It("should accept a minimal GraphQL API", func() {
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a minimal Events API", func() {
			input.Spec.Graphql = nil
			input.Spec.Events = &AwsAppSyncEventsApi{
				AuthProviders:             []*AwsAppSyncEventsAuthProvider{{Type: "API_KEY"}},
				ConnectionAuthModes:       []string{"API_KEY"},
				DefaultPublishAuthModes:   []string{"API_KEY"},
				DefaultSubscribeAuthModes: []string{"API_KEY"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("The mode union (spec.exactly_one_api_arm)", func() {
		ginkgo.It("should reject a spec with neither arm", func() {
			input.Spec.Graphql = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a spec with both arms", func() {
			input.Spec.Events = &AwsAppSyncEventsApi{
				AuthProviders:             []*AwsAppSyncEventsAuthProvider{{Type: "API_KEY"}},
				ConnectionAuthModes:       []string{"API_KEY"},
				DefaultPublishAuthModes:   []string{"API_KEY"},
				DefaultSubscribeAuthModes: []string{"API_KEY"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("GraphQL naming (the explicit api_name convention)", func() {
		ginkgo.It("should reject a hyphenated api_name", func() {
			input.Spec.Graphql.ApiName = "orders-api"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an api_name starting with a digit", func() {
			input.Spec.Graphql.ApiName = "1orders"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("GraphQL auth pairing (graphql.auth_config_matches_type)", func() {
		ginkgo.It("should accept a Cognito primary with default_action", func() {
			input.Spec.Graphql.Auth = &AwsAppSyncGraphqlAuthProvider{
				Type: "AMAZON_COGNITO_USER_POOLS",
				UserPool: &AwsAppSyncCognitoUserPoolAuth{
					UserPoolId:    val("us-west-2_abc123"),
					DefaultAction: "ALLOW",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a Cognito primary missing its user_pool", func() {
			input.Spec.Graphql.Auth = &AwsAppSyncGraphqlAuthProvider{Type: "AMAZON_COGNITO_USER_POOLS"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an API_KEY primary carrying a user_pool", func() {
			input.Spec.Graphql.Auth = &AwsAppSyncGraphqlAuthProvider{
				Type: "API_KEY",
				UserPool: &AwsAppSyncCognitoUserPoolAuth{
					UserPoolId:    val("us-west-2_abc123"),
					DefaultAction: "ALLOW",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a Cognito primary without default_action", func() {
			input.Spec.Graphql.Auth = &AwsAppSyncGraphqlAuthProvider{
				Type:     "AMAZON_COGNITO_USER_POOLS",
				UserPool: &AwsAppSyncCognitoUserPoolAuth{UserPoolId: val("us-west-2_abc123")},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an additional Cognito provider WITH default_action", func() {
			input.Spec.Graphql.AdditionalAuthProviders = []*AwsAppSyncGraphqlAuthProvider{{
				Type: "AMAZON_COGNITO_USER_POOLS",
				UserPool: &AwsAppSyncCognitoUserPoolAuth{
					UserPoolId:    val("us-west-2_abc123"),
					DefaultAction: "DENY",
				},
			}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept an additional Cognito provider without default_action", func() {
			input.Spec.Graphql.AdditionalAuthProviders = []*AwsAppSyncGraphqlAuthProvider{{
				Type:     "AMAZON_COGNITO_USER_POOLS",
				UserPool: &AwsAppSyncCognitoUserPoolAuth{UserPoolId: val("us-west-2_abc123")},
			}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a duplicate auth type across primary and additionals", func() {
			input.Spec.Graphql.AdditionalAuthProviders = []*AwsAppSyncGraphqlAuthProvider{{Type: "API_KEY"}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept a Lambda authorizer with its function reference", func() {
			input.Spec.Graphql.Auth = &AwsAppSyncGraphqlAuthProvider{
				Type:   "AWS_LAMBDA",
				Lambda: &AwsAppSyncLambdaAuth{AuthorizerUri: ref("auth-fn")},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a Lambda authorizer TTL above 3600", func() {
			input.Spec.Graphql.Auth = &AwsAppSyncGraphqlAuthProvider{
				Type: "AWS_LAMBDA",
				Lambda: &AwsAppSyncLambdaAuth{
					AuthorizerUri:                ref("auth-fn"),
					AuthorizerResultTtlInSeconds: 7200,
					IdentityValidationExpression: "",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("The MERGED arm (graphql.merged_owns_no_schema_or_resolvers)", func() {
		ginkgo.BeforeEach(func() {
			input.Spec.Graphql.Merged = &AwsAppSyncGraphqlMerged{
				ExecutionRoleArn: ref("merge-role"),
				SourceApis: []*AwsAppSyncSourceApi{{
					Name:        "orders",
					SourceApiId: ref("orders-api"),
				}},
			}
		})

		ginkgo.It("should accept a merged API with sources", func() {
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a merged API carrying a schema", func() {
			input.Spec.Graphql.Schema = "type Query { ping: String }"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a merged API carrying resolvers", func() {
			input.Spec.Graphql.Resolvers = []*AwsAppSyncGraphqlResolver{{
				Type: "Query", Field: "ping", DataSourceName: "none",
			}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a merged block missing its execution role", func() {
			input.Spec.Graphql.Merged.ExecutionRoleArn = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject duplicate source API entry names", func() {
			input.Spec.Graphql.Merged.SourceApis = append(input.Spec.Graphql.Merged.SourceApis,
				&AwsAppSyncSourceApi{Name: "orders", SourceApiId: ref("other-api")})
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Resolvers (resolver.unit_xor_pipeline)", func() {
		ginkgo.It("should accept a UNIT resolver", func() {
			input.Spec.Graphql.Resolvers = []*AwsAppSyncGraphqlResolver{{
				Type: "Query", Field: "getOrder", DataSourceName: "orders_table",
			}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a PIPELINE resolver with code", func() {
			input.Spec.Graphql.Resolvers = []*AwsAppSyncGraphqlResolver{{
				Type: "Mutation", Field: "putOrder",
				PipelineFunctions: []string{"validate", "persist"},
				Code:              "export function request(ctx) { return {}; }",
			}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a resolver with both a data source and a pipeline", func() {
			input.Spec.Graphql.Resolvers = []*AwsAppSyncGraphqlResolver{{
				Type: "Query", Field: "getOrder",
				DataSourceName:    "orders_table",
				PipelineFunctions: []string{"validate"},
			}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a resolver with neither a data source nor a pipeline", func() {
			input.Spec.Graphql.Resolvers = []*AwsAppSyncGraphqlResolver{{
				Type: "Query", Field: "getOrder",
			}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a resolver mixing code and VTL templates", func() {
			input.Spec.Graphql.Resolvers = []*AwsAppSyncGraphqlResolver{{
				Type: "Query", Field: "getOrder", DataSourceName: "orders_table",
				Code:            "export function request(ctx) { return {}; }",
				RequestTemplate: "{}",
			}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject duplicate type.field positions", func() {
			input.Spec.Graphql.Resolvers = []*AwsAppSyncGraphqlResolver{
				{Type: "Query", Field: "getOrder", DataSourceName: "a"},
				{Type: "Query", Field: "getOrder", DataSourceName: "b"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a hyphenated resolver type", func() {
			input.Spec.Graphql.Resolvers = []*AwsAppSyncGraphqlResolver{{
				Type: "My-Type", Field: "getOrder", DataSourceName: "a",
			}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a LAMBDA conflict handler without its function", func() {
			input.Spec.Graphql.Resolvers = []*AwsAppSyncGraphqlResolver{{
				Type: "Query", Field: "getOrder", DataSourceName: "a",
				SyncConfig: &AwsAppSyncSyncConfig{
					ConflictDetection: "VERSION",
					ConflictHandler:   "LAMBDA",
				},
			}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Functions (function.code_or_templates)", func() {
		ginkgo.It("should reject a function mixing code and templates", func() {
			input.Spec.Graphql.Functions = []*AwsAppSyncGraphqlFunction{{
				Name: "persist", DataSourceName: "orders_table",
				Code:                   "export function request(ctx) { return {}; }",
				RequestMappingTemplate: "{}",
			}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject duplicate function names", func() {
			input.Spec.Graphql.Functions = []*AwsAppSyncGraphqlFunction{
				{Name: "persist", DataSourceName: "a"},
				{Name: "persist", DataSourceName: "b"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a hyphenated function name", func() {
			input.Spec.Graphql.Functions = []*AwsAppSyncGraphqlFunction{
				{Name: "persist-order", DataSourceName: "a"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a max_batch_size above 2000", func() {
			input.Spec.Graphql.Functions = []*AwsAppSyncGraphqlFunction{
				{Name: "persist", DataSourceName: "a", MaxBatchSize: 4000},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("The cache singleton", func() {
		ginkgo.It("should accept a full-request cache", func() {
			input.Spec.Graphql.Cache = &AwsAppSyncGraphqlCache{
				ApiCachingBehavior: "FULL_REQUEST_CACHING",
				Ttl:                300,
				Type:               "SMALL",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a cache ttl above 3600", func() {
			input.Spec.Graphql.Cache = &AwsAppSyncGraphqlCache{
				ApiCachingBehavior: "FULL_REQUEST_CACHING",
				Ttl:                7200,
				Type:               "SMALL",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown cache instance type", func() {
			input.Spec.Graphql.Cache = &AwsAppSyncGraphqlCache{
				ApiCachingBehavior: "FULL_REQUEST_CACHING",
				Ttl:                300,
				Type:               "HUGE",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Limits", func() {
		ginkgo.It("should reject query_depth_limit above 75", func() {
			input.Spec.Graphql.QueryDepthLimit = 76
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject resolver_count_limit above 10000", func() {
			input.Spec.Graphql.ResolverCountLimit = 10001
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Data sources (datasource.config_matches_type)", func() {
		ginkgo.It("should accept a NONE data source with no block", func() {
			input.Spec.Datasources = []*AwsAppSyncDatasource{{Name: "local", Type: "NONE"}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a DynamoDB data source with its block", func() {
			input.Spec.Datasources = []*AwsAppSyncDatasource{{
				Name:           "orders_table",
				Type:           "AMAZON_DYNAMODB",
				ServiceRoleArn: ref("appsync-ddb-role"),
				Dynamodb:       &AwsAppSyncDatasourceDynamodb{TableName: ref("orders")},
			}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a DynamoDB data source missing its block", func() {
			input.Spec.Datasources = []*AwsAppSyncDatasource{{
				Name: "orders_table", Type: "AMAZON_DYNAMODB",
			}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a NONE data source carrying a lambda block", func() {
			input.Spec.Datasources = []*AwsAppSyncDatasource{{
				Name:   "local",
				Type:   "NONE",
				Lambda: &AwsAppSyncDatasourceLambda{FunctionArn: ref("fn")},
			}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an EventBridge block on an HTTP-typed data source", func() {
			input.Spec.Datasources = []*AwsAppSyncDatasource{{
				Name:        "bus",
				Type:        "HTTP",
				Http:        &AwsAppSyncDatasourceHttp{Endpoint: "https://api.example.com"},
				Eventbridge: &AwsAppSyncDatasourceEventbridge{EventBusArn: ref("bus")},
			}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject duplicate data source names", func() {
			input.Spec.Datasources = []*AwsAppSyncDatasource{
				{Name: "local", Type: "NONE"},
				{Name: "local", Type: "NONE"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a hyphenated data source name", func() {
			input.Spec.Datasources = []*AwsAppSyncDatasource{{Name: "orders-table", Type: "NONE"}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("The Events arm", func() {
		ginkgo.BeforeEach(func() {
			input.Spec.Graphql = nil
			input.Spec.Events = &AwsAppSyncEventsApi{
				AuthProviders:             []*AwsAppSyncEventsAuthProvider{{Type: "API_KEY"}},
				ConnectionAuthModes:       []string{"API_KEY"},
				DefaultPublishAuthModes:   []string{"API_KEY"},
				DefaultSubscribeAuthModes: []string{"API_KEY"},
			}
		})

		ginkgo.It("should reject a publish mode naming an undeclared provider", func() {
			input.Spec.Events.DefaultPublishAuthModes = []string{"AWS_IAM"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept multiple providers with phase-scoped modes", func() {
			input.Spec.Events.AuthProviders = []*AwsAppSyncEventsAuthProvider{
				{Type: "API_KEY"},
				{Type: "AWS_IAM"},
			}
			input.Spec.Events.DefaultPublishAuthModes = []string{"AWS_IAM"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject an Events Cognito provider without aws_region", func() {
			input.Spec.Events.AuthProviders = []*AwsAppSyncEventsAuthProvider{{
				Type:    "AMAZON_COGNITO_USER_POOLS",
				Cognito: &AwsAppSyncEventsCognitoAuth{UserPoolId: val("us-west-2_abc123")},
			}}
			input.Spec.Events.ConnectionAuthModes = []string{"AMAZON_COGNITO_USER_POOLS"}
			input.Spec.Events.DefaultPublishAuthModes = []string{"AMAZON_COGNITO_USER_POOLS"}
			input.Spec.Events.DefaultSubscribeAuthModes = []string{"AMAZON_COGNITO_USER_POOLS"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject duplicate provider types", func() {
			input.Spec.Events.AuthProviders = []*AwsAppSyncEventsAuthProvider{
				{Type: "API_KEY"},
				{Type: "API_KEY"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept a channel namespace with handlers", func() {
			input.Spec.Datasources = []*AwsAppSyncDatasource{{Name: "handler_fn", Type: "AWS_LAMBDA",
				ServiceRoleArn: ref("appsync-lambda-role"),
				Lambda:         &AwsAppSyncDatasourceLambda{FunctionArn: ref("fn")}}}
			input.Spec.Events.ChannelNamespaces = []*AwsAppSyncChannelNamespace{{
				Name: "chat",
				HandlerConfigs: &AwsAppSyncChannelHandlerConfigs{
					OnPublish: &AwsAppSyncChannelHandler{
						Behavior:       "DIRECT",
						DataSourceName: "handler_fn",
					},
				},
			}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a channel namespace name with a leading hyphen", func() {
			input.Spec.Events.ChannelNamespaces = []*AwsAppSyncChannelNamespace{{Name: "-chat"}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject duplicate channel namespace names", func() {
			input.Spec.Events.ChannelNamespaces = []*AwsAppSyncChannelNamespace{
				{Name: "chat"},
				{Name: "chat"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("API keys", func() {
		ginkgo.It("should accept a key with an RFC 3339 expiry", func() {
			input.Spec.ApiKeys = []*AwsAppSyncApiKey{{Name: "web_client", Expires: "2027-02-01T00:00:00Z"}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed expiry", func() {
			input.Spec.ApiKeys = []*AwsAppSyncApiKey{{Name: "web_client", Expires: "next year"}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject duplicate key names", func() {
			input.Spec.ApiKeys = []*AwsAppSyncApiKey{{Name: "web_client"}, {Name: "web_client"}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("The custom domain (custom_domain.certificate_in_us_east_1)", func() {
		ginkgo.It("should accept a us-east-1 certificate literal", func() {
			input.Spec.CustomDomain = &AwsAppSyncCustomDomain{
				DomainName:     "api.example.com",
				CertificateArn: val("arn:aws:acm:us-east-1:123456789012:certificate/abc-123"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a certificate literal outside us-east-1", func() {
			input.Spec.CustomDomain = &AwsAppSyncCustomDomain{
				DomainName:     "api.example.com",
				CertificateArn: val("arn:aws:acm:us-west-2:123456789012:certificate/abc-123"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept a reference-shaped certificate (region checked at deploy)", func() {
			input.Spec.CustomDomain = &AwsAppSyncCustomDomain{
				DomainName:     "api.example.com",
				CertificateArn: ref("api-cert"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid domain name", func() {
			input.Spec.CustomDomain = &AwsAppSyncCustomDomain{
				DomainName:     "not a domain",
				CertificateArn: val("arn:aws:acm:us-east-1:123456789012:certificate/abc-123"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a domain missing its certificate", func() {
			input.Spec.CustomDomain = &AwsAppSyncCustomDomain{DomainName: "api.example.com"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
