variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsHttpApiGateway specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Human-readable description of the API (max 1024 characters).
    description = optional(string, "")

    # A version identifier for the API (max 64 characters). Purely informational
    # metadata surfaced in the AWS console and exports -- it does not create
    # stages or affect routing.
    api_version = optional(string, "")

    # CORS configuration for cross-origin requests. When not set, no CORS
    # headers are returned by the API.
    cors_configuration = optional(object({
      # Origins allowed to make cross-origin requests (e.g., "https://example.com", "*").
      allow_origins = optional(list(string), [])

      # HTTP methods allowed for cross-origin requests (e.g., "GET", "POST", "OPTIONS").
      allow_methods = optional(list(string), [])

      # Request headers allowed in cross-origin requests (e.g., "Content-Type", "Authorization").
      allow_headers = optional(list(string), [])

      # Response headers exposed to the browser in cross-origin responses.
      expose_headers = optional(list(string), [])

      # Maximum time in seconds that browsers can cache CORS preflight results.
      # Reduces the number of preflight OPTIONS requests. Range: 0-86400.
      max_age_seconds = optional(number, 0)

      # Whether the API supports credentials (cookies, authorization headers)
      # in cross-origin requests.
      allow_credentials = optional(bool, false)
    }))

    # Disable the default execute-api endpoint. Set to true when a custom domain
    # (AwsHttpApiDomain) fronts this API to prevent callers from bypassing the
    # domain (and its TLS policy / WAF) via the default endpoint.
    disable_execute_api_endpoint = optional(bool, false)

    # IP address type for the API's default endpoint.
    # - "ipv4": Resolve the endpoint to IPv4 addresses only.
    # - "dualstack": Resolve to both IPv4 and IPv6.
    # When omitted, AWS defaults new APIs to dualstack. Changing the value
    # updates the endpoint in place.
    ip_address_type = optional(string, "")

    # Stage configuration for the deployed API. When not set, a "$default" stage
    # with auto_deploy=true is created automatically, which is the recommended
    # configuration for most HTTP APIs.
    stage = optional(object({
      # Stage name. Defaults to "$default" when empty, which is the recommended
      # configuration for HTTP APIs. Named stages (e.g., "prod", "dev") append
      # the stage name to the invoke URL path.
      name = optional(string, "")

      # Enable automatic deployment when routes, integrations, or authorizers
      # change. When omitted, the modules default to true -- the configuration
      # that makes a declarative spec self-applying. Set explicitly to false to
      # require deployments to be created outside this resource (an advanced
      # pattern; changes to routes then have no effect until a deployment is
      # published).
      auto_deploy = optional(bool)

      # Human-readable description of the stage (max 1024 characters).
      description = optional(string, "")

      # Access logging configuration. When set, API Gateway streams access logs
      # to the specified CloudWatch Log Group.
      access_log = optional(object({
        # CloudWatch Log Group ARN for access log delivery. Accepts a direct ARN
        # or a reference to an AwsCloudwatchLogGroup resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        destination_arn = string

        # Log format template. Uses API Gateway access log variables (e.g.,
        # $context.requestId, $context.identity.sourceIp, $context.httpMethod).
        #
        # Common JSON format:
        #   {"requestId":"$context.requestId","ip":"$context.identity.sourceIp",
        #    "method":"$context.httpMethod","path":"$context.routeKey",
        #    "status":"$context.status","latency":"$context.responseLatency"}
        format = string
      }))

      # Default throttling settings applied to all routes unless overridden
      # per route in route_settings.
      default_throttle = optional(object({
        # Maximum number of concurrent requests allowed (burst). API Gateway uses
        # the token bucket algorithm where burst_limit is the bucket size.
        burst_limit = optional(number, 0)

        # Steady-state request rate limit (requests per second). API Gateway uses
        # this as the token refill rate.
        rate_limit = optional(number, 0)
      }))

      # Emit detailed CloudWatch metrics for all routes (per-route dimensions for
      # count, latency, and errors). Applies as the stage default; can be
      # overridden per route in route_settings. Detailed metrics carry additional
      # CloudWatch cost.
      detailed_metrics_enabled = optional(bool, false)

      # Per-route overrides of throttling and detailed metrics. Each entry targets
      # one route (by its route_key) and overrides the stage defaults for that
      # route only -- e.g. a lower rate limit on an expensive search route, or
      # detailed metrics on just the checkout path.
      route_settings = optional(list(object({
        # The route to override, addressed by its route_key exactly as defined in
        # routes (e.g. "GET /search", "$default").
        route_key = string

        # Maximum number of concurrent requests for this route (token bucket size).
        # Zero inherits the stage default.
        throttling_burst_limit = optional(number, 0)

        # Steady-state request rate limit for this route (requests per second).
        # Zero inherits the stage default.
        throttling_rate_limit = optional(number, 0)

        # Emit detailed CloudWatch metrics for this route regardless of the stage
        # default.
        detailed_metrics_enabled = optional(bool, false)
      })), [])

      # Stage variables passed to integrations. These act as environment-specific
      # configuration values accessible in integration request parameters.
      stage_variables = optional(map(string), {})
    }))

    # API routes mapping request patterns to backend integrations. Each route
    # specifies a route key (e.g., "GET /users", "$default") and an inline
    # integration that defines the backend target.
    #
    # When multiple routes share the same integration configuration (same type,
    # URI, and payload format), the IaC modules automatically deduplicate and
    # create a single integration resource.
    #
    # At least one route is required -- an API without routes has no function.
    routes = list(object({
      # Route key defining the request pattern to match. Format: "{METHOD} {PATH}"
      # for specific routes (e.g., "GET /users", "POST /orders/{id}") or "$default"
      # for a catch-all route that handles unmatched requests.
      route_key = string

      # Backend integration that processes requests matching this route.
      integration = object({
        # Integration type. Valid values:
        # - "AWS_PROXY": Lambda proxy integration or (with integration_subtype)
        #   a first-class AWS service integration.
        # - "HTTP_PROXY": HTTP proxy integration -- API Gateway forwards the
        #   request to an upstream HTTP endpoint.
        integration_type = string

        # Integration URI for proxy integrations. For AWS_PROXY (Lambda) this is
        # the Lambda function ARN; for HTTP_PROXY this is the upstream URL (for
        # private integrations through a VPC link, the ALB/NLB listener ARN or
        # Cloud Map service ARN). Accepts a direct value or a reference to another
        # resource's output. Must be omitted for AWS service integrations
        # (integration_subtype set) -- their target is expressed in
        # request_parameters instead.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        integration_uri = optional(string, "")

        # AWS service integration subtype. Selects a first-class service action
        # that API Gateway invokes directly -- no Lambda glue required. Known
        # subtypes include: "EventBridge-PutEvents", "SQS-SendMessage",
        # "SQS-ReceiveMessage", "SQS-DeleteMessage", "SQS-PurgeQueue",
        # "Kinesis-PutRecord", "StepFunctions-StartExecution",
        # "StepFunctions-StartSyncExecution", "StepFunctions-StopExecution",
        # "AppConfig-GetConfiguration". The action's parameters (e.g. QueueUrl and
        # MessageBody for SQS-SendMessage, StateMachineArn and Input for
        # StepFunctions-StartExecution) are supplied via request_parameters, and
        # credentials_arn must grant API Gateway permission to call the action.
        integration_subtype = optional(string, "")

        # Payload format version for Lambda integrations. Controls the format of the
        # event sent to the Lambda function.
        # - "2.0" (recommended): Simplified event structure with direct body access.
        # - "1.0": Legacy format with multi-value headers and base64 encoding.
        # Defaults to "2.0" when empty. Only applicable to AWS_PROXY integrations.
        # AWS service integrations (integration_subtype) always use "1.0".
        payload_format_version = optional(string, "")

        # HTTP method used for the integration request. Valid values: "ANY",
        # "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT" (uppercase).
        # Defaults to the route's HTTP method for HTTP_PROXY integrations. For
        # AWS_PROXY (Lambda) integrations, this is always POST regardless of the
        # value set here.
        integration_method = optional(string, "")

        # Integration timeout in milliseconds. If the backend does not respond within
        # this duration, API Gateway returns a 504 Gateway Timeout.
        # Range: 50-30000 (50ms to 30s). AWS default: 30000 (30s).
        # Leave at 0 to use the AWS default.
        timeout_milliseconds = optional(number, 0)

        # Connection type for the integration.
        # - "INTERNET" (default): Route to the target over the public internet.
        # - "VPC_LINK": Route through a VPC link to a private ALB, NLB, or Cloud
        #   Map service inside a VPC. Requires connection_id and integration_type
        #   HTTP_PROXY.
        connection_type = optional(string, "")

        # The VPC link to route through for private integrations. Accepts a direct
        # VPC link ID or a reference to an AwsHttpApiVpcLink resource. Required
        # when connection_type is "VPC_LINK"; must be omitted otherwise.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        connection_id = optional(string, "")

        # IAM role that API Gateway assumes to invoke the integration target.
        # Required for AWS service integrations (integration_subtype set) -- the
        # role must trust apigateway.amazonaws.com and grant the service action
        # (e.g. sqs:SendMessage on the target queue). Optional for Lambda proxy
        # integrations, which normally authorize through the function's resource
        # policy instead.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        credentials_arn = optional(string, "")

        # Parameter mappings applied to the integration request.
        #
        # For proxy integrations (Lambda / HTTP), keys are mapping instructions of
        # the form "append:header.<name>", "overwrite:header.<name>",
        # "remove:querystring.<name>", "overwrite:path", etc., and values are
        # static strings or context expressions (e.g. "$context.requestId",
        # "$request.header.Authorization").
        #
        # For AWS service integrations (integration_subtype set), keys are the
        # service action's parameter names (e.g. "QueueUrl", "MessageBody" for
        # SQS-SendMessage; "StateMachineArn", "Input" for
        # StepFunctions-StartExecution) and values are static strings or request
        # expressions like "$request.body".
        request_parameters = optional(map(string), {})

        # Response parameter mappings, keyed by the backend status code they apply
        # to. Each entry transforms the response API Gateway returns to the caller
        # for that status code -- e.g. overwriting the status code or injecting a
        # header. Supported by proxy and service integrations on HTTP APIs.
        response_parameters = optional(list(object({
          # The backend response status code these mappings apply to (e.g. "403",
          # "500"). API Gateway accepts 200-599 -- informational (1xx) codes cannot
          # carry response overrides.
          status_code = optional(string, "")

          # Mapping instructions applied to the response. Keys are instructions such
          # as "overwrite:statuscode", "append:header.<name>",
          # "overwrite:header.<name>", "remove:header.<name>"; values are static
          # strings or context expressions (e.g. "$context.requestId").
          mappings = optional(map(string), {})
        })), [])

        # Server name TLS verification for private HTTP_PROXY integrations. When
        # set, API Gateway verifies the target's certificate against this server
        # name (SNI) instead of the resolved address -- required when a private
        # ALB terminates TLS with a certificate issued for the public domain name.
        # Maps to the integration's tls_config block.
        tls_server_name_to_verify = optional(string, "")

        # Human-readable description of the integration (max 1024 characters).
        description = optional(string, "")
      })

      # Authorization type for this route. Valid values:
      # - "NONE" (default): No authorization required.
      # - "JWT": JSON Web Token authorization using a JWT authorizer.
      # - "AWS_IAM": AWS IAM authorization using SigV4 signatures.
      # - "CUSTOM": Lambda authorization using a REQUEST authorizer.
      authorization_type = optional(string, "")

      # Name of the authorizer to use for this route. Must match the name of an
      # authorizer defined in the `authorizers` field. Required when
      # `authorization_type` is "JWT" (bind a JWT authorizer) or "CUSTOM" (bind
      # a REQUEST authorizer).
      authorizer_name = optional(string, "")

      # OAuth 2.0 scopes required for JWT authorization. The request must include
      # all specified scopes to be authorized. Only applicable when
      # `authorization_type` is "JWT".
      authorization_scopes = optional(list(string), [])

      # Operation name for this route (max 64 characters). Surfaced in OpenAPI
      # exports as the operationId -- useful when generated client SDKs need
      # stable method names.
      operation_name = optional(string, "")
    }))

    # Named authorizers that can be referenced by routes. Define JWT authorizers
    # for Cognito/Auth0/OIDC integration, or Lambda (REQUEST) authorizers for
    # custom authorization logic. Routes reference authorizers by name.
    authorizers = optional(list(object({
      # Unique name for this authorizer. Routes reference authorizers by this name.
      # Must be unique across all authorizers in the spec.
      name = string

      # Authorizer type. Valid values:
      # - "JWT": Validates a JSON Web Token using the configured issuer and audiences.
      # - "REQUEST": Invokes a Lambda function that returns an authorization decision.
      authorizer_type = string

      # JWT configuration. Required when authorizer_type is "JWT".
      # Configures the token issuer and expected audiences for JWT validation.
      jwt_configuration = optional(object({
        # Token issuer URL. API Gateway validates that the JWT's "iss" claim matches
        # this value. Accepts a direct issuer URL -- for Cognito,
        # "https://cognito-idp.{region}.amazonaws.com/{userPoolId}"; for Auth0,
        # "https://{domain}/" -- or a reference to an AwsCognitoUserPool resource
        # (its issuer output carries exactly this URL).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        issuer = string

        # Expected audiences. API Gateway validates that the JWT's "aud" claim matches
        # one of these values. Each entry accepts a direct value -- for Cognito, an
        # app client ID -- or a reference to an AwsCognitoUserPoolClient resource;
        # literals and references can be mixed.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        audiences = optional(list(string), [])
      }))

      # Lambda function URI for REQUEST authorizers. This is the Lambda function
      # invoke ARN. Required when authorizer_type is "REQUEST".
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      authorizer_uri = optional(string, "")

      # IAM role ARN that API Gateway assumes to invoke the Lambda authorizer.
      # Only applicable to REQUEST authorizers.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      authorizer_credentials_arn = optional(string, "")

      # Identity sources used to extract the authorization token or context.
      # For JWT authorizers: "$request.header.Authorization" (typical).
      # For REQUEST authorizers: varies by implementation.
      identity_sources = optional(list(string), [])

      # Time in seconds that API Gateway caches the authorizer result.
      # Range: 0-3600. Set to 0 to disable caching.
      # AWS default: 300 for REQUEST authorizers with identity sources.
      result_ttl_seconds = optional(number, 0)

      # Enable simple boolean responses from Lambda authorizers.
      # When true, the Lambda returns {"isAuthorized": true/false} instead of an
      # IAM policy document. Simpler to implement for most use cases.
      # Only applicable to REQUEST authorizers.
      enable_simple_responses = optional(bool, false)

      # Payload format version for the Lambda authorizer event.
      # - "2.0" (recommended): Simplified event with direct access to headers/query params.
      # - "1.0": Legacy format.
      # Only applicable to REQUEST authorizers.
      authorizer_payload_format_version = optional(string, "")
    })), [])
  })
}
