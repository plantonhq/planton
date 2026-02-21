package module

import (
	"github.com/pkg/errors"
	ociapigatewayv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ociapigateway/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/oci/pulumiociprovider"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var endpointTypeMap = map[ociapigatewayv1.OciApiGatewaySpec_EndpointType]string{
	ociapigatewayv1.OciApiGatewaySpec_endpoint_type_public:  "PUBLIC",
	ociapigatewayv1.OciApiGatewaySpec_endpoint_type_private: "PRIVATE",
}

var logLevelMap = map[ociapigatewayv1.OciApiGatewaySpec_LogLevel]string{
	ociapigatewayv1.OciApiGatewaySpec_info:  "INFO",
	ociapigatewayv1.OciApiGatewaySpec_warn:  "WARN",
	ociapigatewayv1.OciApiGatewaySpec_error: "ERROR",
}

var backendTypeMap = map[ociapigatewayv1.OciApiGatewaySpec_BackendType]string{
	ociapigatewayv1.OciApiGatewaySpec_http:             "HTTP_BACKEND",
	ociapigatewayv1.OciApiGatewaySpec_oracle_functions: "ORACLE_FUNCTIONS_BACKEND",
	ociapigatewayv1.OciApiGatewaySpec_stock_response:   "STOCK_RESPONSE_BACKEND",
}

var publicKeyTypeMap = map[ociapigatewayv1.OciApiGatewaySpec_PublicKeyType]string{
	ociapigatewayv1.OciApiGatewaySpec_remote_jwks: "REMOTE_JWKS",
	ociapigatewayv1.OciApiGatewaySpec_static_keys: "STATIC_KEYS",
}

var keyFormatMap = map[ociapigatewayv1.OciApiGatewaySpec_KeyFormat]string{
	ociapigatewayv1.OciApiGatewaySpec_pem:          "PEM",
	ociapigatewayv1.OciApiGatewaySpec_json_web_key: "JSON_WEB_KEY",
}

var rateKeyMap = map[ociapigatewayv1.OciApiGatewaySpec_RateKey]string{
	ociapigatewayv1.OciApiGatewaySpec_client_ip: "CLIENT_IP",
	ociapigatewayv1.OciApiGatewaySpec_total:     "TOTAL",
}

var authorizationTypeMap = map[ociapigatewayv1.OciApiGatewaySpec_AuthorizationType]string{
	ociapigatewayv1.OciApiGatewaySpec_anonymous:           "ANONYMOUS",
	ociapigatewayv1.OciApiGatewaySpec_any_of:              "ANY_OF",
	ociapigatewayv1.OciApiGatewaySpec_authentication_only: "AUTHENTICATION_ONLY",
}

func Resources(ctx *pulumi.Context, stackInput *ociapigatewayv1.OciApiGatewayStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	ociProvider, err := pulumiociprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup oci provider")
	}

	if err := gatewayResource(ctx, locals, ociProvider); err != nil {
		return errors.Wrap(err, "failed to create api gateway")
	}

	return nil
}

func pulumiOciOpt(provider *oci.Provider) pulumi.ResourceOption {
	return pulumi.Provider(provider)
}
