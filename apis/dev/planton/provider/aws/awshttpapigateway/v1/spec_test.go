package awshttpapigatewayv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAwsHttpApiGatewaySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsHttpApiGatewaySpec Validation Suite")
}

// helper to create a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

// helper to create a minimal valid spec with a single $default route.
func minimalValidSpec() *AwsHttpApiGatewaySpec {
	return &AwsHttpApiGatewaySpec{
		Region: "us-west-2",
		Routes: []*AwsHttpApiGatewayRoute{
			{
				RouteKey: "$default",
				Integration: &AwsHttpApiGatewayIntegration{
					IntegrationType: "AWS_PROXY",
					IntegrationUri:  strRef("arn:aws:lambda:us-east-1:123456789012:function:my-func"),
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AwsHttpApiGatewaySpec validations", func() {
	var spec *AwsHttpApiGatewaySpec

	ginkgo.BeforeEach(func() {
		spec = minimalValidSpec()
	})

	// -------------------------------------------------------------------------
	// Happy paths
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal spec with single $default route", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts multiple routes to different Lambda functions", func() {
		spec.Routes = []*AwsHttpApiGatewayRoute{
			{
				RouteKey: "GET /users",
				Integration: &AwsHttpApiGatewayIntegration{
					IntegrationType: "AWS_PROXY",
					IntegrationUri:  strRef("arn:aws:lambda:us-east-1:123456789012:function:users"),
				},
			},
			{
				RouteKey: "POST /orders",
				Integration: &AwsHttpApiGatewayIntegration{
					IntegrationType: "AWS_PROXY",
					IntegrationUri:  strRef("arn:aws:lambda:us-east-1:123456789012:function:orders"),
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with HTTP_PROXY integration", func() {
		spec.Routes[0].Integration.IntegrationType = "HTTP_PROXY"
		spec.Routes[0].Integration.IntegrationUri = strRef("https://api.example.com")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with CORS configuration", func() {
		spec.CorsConfiguration = &AwsHttpApiGatewayCorsConfig{
			AllowOrigins:     []string{"https://example.com"},
			AllowMethods:     []string{"GET", "POST", "OPTIONS"},
			AllowHeaders:     []string{"Content-Type", "Authorization"},
			MaxAgeSeconds:    3600,
			AllowCredentials: true,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with stage configuration", func() {
		spec.Stage = &AwsHttpApiGatewayStageConfig{
			Name:       "prod",
			AutoDeploy: true,
			StageVariables: map[string]string{
				"env": "production",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with access logging", func() {
		spec.Stage = &AwsHttpApiGatewayStageConfig{
			AccessLog: &AwsHttpApiGatewayAccessLogConfig{
				DestinationArn: strRef("arn:aws:logs:us-east-1:123456789012:log-group:/aws/apigateway/my-api"),
				Format:         `{"requestId":"$context.requestId","ip":"$context.identity.sourceIp"}`,
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with throttling", func() {
		spec.Stage = &AwsHttpApiGatewayStageConfig{
			DefaultThrottle: &AwsHttpApiGatewayThrottleConfig{
				BurstLimit: 100,
				RateLimit:  50.0,
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with JWT authorizer", func() {
		spec.Authorizers = []*AwsHttpApiGatewayAuthorizer{
			{
				Name:           "cognito",
				AuthorizerType: "JWT",
				JwtConfiguration: &AwsHttpApiGatewayJwtConfig{
					Issuer:    "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_abc123",
					Audiences: []string{"my-app-client-id"},
				},
				IdentitySources: []string{"$request.header.Authorization"},
			},
		}
		spec.Routes[0].AuthorizationType = "JWT"
		spec.Routes[0].AuthorizerName = "cognito"
		spec.Routes[0].AuthorizationScopes = []string{"read:users"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with REQUEST authorizer", func() {
		spec.Authorizers = []*AwsHttpApiGatewayAuthorizer{
			{
				Name:                           "custom-lambda",
				AuthorizerType:                 "REQUEST",
				AuthorizerUri:                  strRef("arn:aws:lambda:us-east-1:123456789012:function:auth"),
				AuthorizerCredentialsArn:       strRef("arn:aws:iam::123456789012:role/api-auth-role"),
				IdentitySources:                []string{"$request.header.Authorization"},
				ResultTtlSeconds:               300,
				EnableSimpleResponses:          true,
				AuthorizerPayloadFormatVersion: "2.0",
			},
		}
		spec.Routes[0].AuthorizationType = "JWT"
		spec.Routes[0].AuthorizerName = "custom-lambda"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with NONE authorization (explicit)", func() {
		spec.Routes[0].AuthorizationType = "NONE"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with AWS_IAM authorization", func() {
		spec.Routes[0].AuthorizationType = "AWS_IAM"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts integration with explicit payload format version 1.0", func() {
		spec.Routes[0].Integration.PayloadFormatVersion = "1.0"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts integration with explicit payload format version 2.0", func() {
		spec.Routes[0].Integration.PayloadFormatVersion = "2.0"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts integration with timeout", func() {
		spec.Routes[0].Integration.TimeoutMilliseconds = 5000
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a fully-configured production spec", func() {
		spec.Description = "Production API for order management"
		spec.CorsConfiguration = &AwsHttpApiGatewayCorsConfig{
			AllowOrigins:     []string{"https://app.example.com"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Content-Type", "Authorization"},
			ExposeHeaders:    []string{"X-Request-Id"},
			MaxAgeSeconds:    7200,
			AllowCredentials: true,
		}
		spec.Stage = &AwsHttpApiGatewayStageConfig{
			AutoDeploy: true,
			AccessLog: &AwsHttpApiGatewayAccessLogConfig{
				DestinationArn: strRef("arn:aws:logs:us-east-1:123456789012:log-group:/aws/apigateway/orders"),
				Format:         `{"requestId":"$context.requestId"}`,
			},
			DefaultThrottle: &AwsHttpApiGatewayThrottleConfig{
				BurstLimit: 500,
				RateLimit:  100.0,
			},
		}
		spec.Authorizers = []*AwsHttpApiGatewayAuthorizer{
			{
				Name:           "cognito",
				AuthorizerType: "JWT",
				JwtConfiguration: &AwsHttpApiGatewayJwtConfig{
					Issuer:    "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_abc123",
					Audiences: []string{"orders-client"},
				},
				IdentitySources: []string{"$request.header.Authorization"},
			},
		}
		spec.Routes = []*AwsHttpApiGatewayRoute{
			{
				RouteKey: "GET /orders",
				Integration: &AwsHttpApiGatewayIntegration{
					IntegrationType:      "AWS_PROXY",
					IntegrationUri:       strRef("arn:aws:lambda:us-east-1:123456789012:function:get-orders"),
					PayloadFormatVersion: "2.0",
				},
				AuthorizationType:   "JWT",
				AuthorizerName:      "cognito",
				AuthorizationScopes: []string{"orders:read"},
			},
			{
				RouteKey: "POST /orders",
				Integration: &AwsHttpApiGatewayIntegration{
					IntegrationType:      "AWS_PROXY",
					IntegrationUri:       strRef("arn:aws:lambda:us-east-1:123456789012:function:create-order"),
					PayloadFormatVersion: "2.0",
				},
				AuthorizationType:   "JWT",
				AuthorizerName:      "cognito",
				AuthorizationScopes: []string{"orders:write"},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Route validation failures
	// -------------------------------------------------------------------------

	ginkgo.It("fails when no routes are provided", func() {
		spec.Routes = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when routes is empty", func() {
		spec.Routes = []*AwsHttpApiGatewayRoute{}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when route_key is empty", func() {
		spec.Routes[0].RouteKey = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when route integration is nil", func() {
		spec.Routes[0].Integration = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when integration_type is empty", func() {
		spec.Routes[0].Integration.IntegrationType = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when integration_uri is nil", func() {
		spec.Routes[0].Integration.IntegrationUri = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL validation failures: integration_type
	// -------------------------------------------------------------------------

	ginkgo.It("fails when integration_type is invalid", func() {
		spec.Routes[0].Integration.IntegrationType = "MOCK"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL validation failures: authorization
	// -------------------------------------------------------------------------

	ginkgo.It("fails when authorization_type is invalid", func() {
		spec.Routes[0].AuthorizationType = "CUSTOM"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when JWT authorization has no authorizer_name", func() {
		spec.Routes[0].AuthorizationType = "JWT"
		spec.Routes[0].AuthorizerName = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when authorizer_name references non-existent authorizer", func() {
		spec.Routes[0].AuthorizationType = "JWT"
		spec.Routes[0].AuthorizerName = "does-not-exist"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL validation failures: authorizer
	// -------------------------------------------------------------------------

	ginkgo.It("fails when authorizer_type is invalid", func() {
		spec.Authorizers = []*AwsHttpApiGatewayAuthorizer{
			{
				Name:           "bad",
				AuthorizerType: "INVALID",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when JWT authorizer has no jwt_configuration", func() {
		spec.Authorizers = []*AwsHttpApiGatewayAuthorizer{
			{
				Name:           "jwt-missing-config",
				AuthorizerType: "JWT",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when JWT authorizer has empty issuer", func() {
		spec.Authorizers = []*AwsHttpApiGatewayAuthorizer{
			{
				Name:           "jwt-empty-issuer",
				AuthorizerType: "JWT",
				JwtConfiguration: &AwsHttpApiGatewayJwtConfig{
					Issuer: "",
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when REQUEST authorizer has no authorizer_uri", func() {
		spec.Authorizers = []*AwsHttpApiGatewayAuthorizer{
			{
				Name:           "lambda-missing-uri",
				AuthorizerType: "REQUEST",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL validation failures: payload format version
	// -------------------------------------------------------------------------

	ginkgo.It("fails when payload_format_version is invalid", func() {
		spec.Routes[0].Integration.PayloadFormatVersion = "3.0"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when authorizer_payload_format_version is invalid", func() {
		spec.Authorizers = []*AwsHttpApiGatewayAuthorizer{
			{
				Name:                           "bad-payload-ver",
				AuthorizerType:                 "REQUEST",
				AuthorizerUri:                  strRef("arn:aws:lambda:us-east-1:123456789012:function:auth"),
				AuthorizerPayloadFormatVersion: "3.0",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL validation failures: range validations
	// -------------------------------------------------------------------------

	ginkgo.It("fails when integration timeout is below minimum", func() {
		spec.Routes[0].Integration.TimeoutMilliseconds = 10
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when integration timeout exceeds maximum", func() {
		spec.Routes[0].Integration.TimeoutMilliseconds = 31000
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when authorizer TTL exceeds 3600", func() {
		spec.Authorizers = []*AwsHttpApiGatewayAuthorizer{
			{
				Name:             "bad-ttl",
				AuthorizerType:   "REQUEST",
				AuthorizerUri:    strRef("arn:aws:lambda:us-east-1:123456789012:function:auth"),
				ResultTtlSeconds: 3601,
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Field-level validation failures
	// -------------------------------------------------------------------------

	ginkgo.It("fails when CORS max_age_seconds is negative", func() {
		spec.CorsConfiguration = &AwsHttpApiGatewayCorsConfig{
			MaxAgeSeconds: -1,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when CORS max_age_seconds exceeds 86400", func() {
		spec.CorsConfiguration = &AwsHttpApiGatewayCorsConfig{
			MaxAgeSeconds: 86401,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when authorizer name is empty", func() {
		spec.Authorizers = []*AwsHttpApiGatewayAuthorizer{
			{
				Name:           "",
				AuthorizerType: "JWT",
				JwtConfiguration: &AwsHttpApiGatewayJwtConfig{
					Issuer: "https://example.com",
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when authorizer name exceeds 128 characters", func() {
		longName := ""
		for i := 0; i < 130; i++ {
			longName += "a"
		}
		spec.Authorizers = []*AwsHttpApiGatewayAuthorizer{
			{
				Name:           longName,
				AuthorizerType: "JWT",
				JwtConfiguration: &AwsHttpApiGatewayJwtConfig{
					Issuer: "https://example.com",
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when access log format is empty", func() {
		spec.Stage = &AwsHttpApiGatewayStageConfig{
			AccessLog: &AwsHttpApiGatewayAccessLogConfig{
				DestinationArn: strRef("arn:aws:logs:us-east-1:123456789012:log-group:test"),
				Format:         "",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when access log destination_arn is missing", func() {
		spec.Stage = &AwsHttpApiGatewayStageConfig{
			AccessLog: &AwsHttpApiGatewayAccessLogConfig{
				Format: `{"requestId":"$context.requestId"}`,
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when description exceeds 1024 characters", func() {
		longDesc := ""
		for i := 0; i < 1025; i++ {
			longDesc += "a"
		}
		spec.Description = longDesc
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
