package module

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	kuberneteslangfusev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteslangfuse/v1alpha1"
)

// envVar is one environment variable the module adds through the chart's
// additionalEnv channels: a plain value, a key of the instance Secret, or a
// key of a Secret the spec references.
type envVar struct {
	Name        string
	Value       string
	InstanceKey string
	SecretRef   *kubernetesprovider.KubernetesSecretKey
}

// envSet collects variables by name; empty plain values are dropped, so a
// spec field left empty sets nothing and Langfuse's own default applies.
type envSet map[string]envVar

func (e envSet) plain(name, value string) {
	if value != "" {
		e[name] = envVar{Name: name, Value: value}
	}
}

func (e envSet) instance(name, key string) {
	e[name] = envVar{Name: name, InstanceKey: key}
}

func (e envSet) secret(name string, ref *kubernetesprovider.KubernetesSecretKey) {
	e[name] = envVar{Name: name, SecretRef: ref}
}

func (e envSet) sorted() []envVar {
	names := make([]string, 0, len(e))
	for name := range e {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]envVar, 0, len(names))
	for _, name := range names {
		out = append(out, e[name])
	}
	return out
}

// body renders the variable as a Kubernetes EnvVar map.
func (v envVar) body(instanceSecretName string) map[string]interface{} {
	switch {
	case v.InstanceKey != "":
		return map[string]interface{}{"name": v.Name, "valueFrom": map[string]interface{}{
			"secretKeyRef": map[string]interface{}{"name": instanceSecretName, "key": v.InstanceKey},
		}}
	case v.SecretRef != nil:
		return map[string]interface{}{"name": v.Name, "valueFrom": map[string]interface{}{
			"secretKeyRef": map[string]interface{}{"name": v.SecretRef.GetName(), "key": v.SecretRef.GetKey()},
		}}
	default:
		return map[string]interface{}{"name": v.Name, "value": v.Value}
	}
}

// buildEnvironment renders every typed setting the chart has no value for
// into the three additionalEnv channels. Variables the chart renders itself
// (the stores, the salt, encryption key and NextAuth secret, telemetry,
// sign-up, the license key, logging) are never repeated here: the chart
// refuses a variable set both ways. Terraform twin: local.common_env,
// local.web_env and local.worker_env.
func (l *Locals) buildEnvironment() {
	spec := l.Spec
	common, web, worker := envSet{}, envSet{}, envSet{}

	// ---- postgres --------------------------------------------------------
	if l.DeclaredSecrets[keyPostgresDirectUrl] != "" {
		common.instance("DIRECT_URL", keyPostgresDirectUrl)
	}
	if l.DeclaredSecrets[keyPostgresShadowUrl] != "" {
		common.instance("SHADOW_DATABASE_URL", keyPostgresShadowUrl)
	}

	// ---- clickhouse ------------------------------------------------------
	ch := spec.GetClickhouse()
	common.plain("CLICKHOUSE_CLUSTER_NAME", ch.GetCluster().GetName().GetValue())
	common.plain("CLICKHOUSE_READ_ONLY_URL", ch.GetReadOnlyUrl())
	if m := ch.GetDeleteMode(); m != kuberneteslangfusev1alpha1.KubernetesLangfuseClickHouse_unspecified {
		common.plain("CLICKHOUSE_LIGHTWEIGHT_DELETE_MODE", m.String())
	}
	common.plain("CLICKHOUSE_USE_LIGHTWEIGHT_UPDATE", trueOrEmpty(ch.GetUseLightweightUpdate()))
	common.plain("LANGFUSE_CLICKHOUSE_DELETION_TIMEOUT_MS", optionalInt(ch.DeletionTimeoutMs))

	// ---- redis -----------------------------------------------------------
	redis := spec.GetRedis()
	common.plain("REDIS_KEY_PREFIX", redis.GetKeyPrefix())
	common.plain("REDIS_SOCKET_TIMEOUT_MS", optionalInt(redis.SocketTimeoutMs))
	if cluster := redis.GetCluster(); cluster != nil {
		common.plain("REDIS_CLUSTER_SLOTS_REFRESH_TIMEOUT", optionalInt(cluster.SlotsRefreshTimeoutMs))
	}
	common.plain("REDIS_SENTINEL_TLS_ENABLED", trueOrEmpty(redis.GetSentinel().GetTlsEnabled()))
	tls := redis.GetTls()
	if tls == nil {
		tls = &kuberneteslangfusev1alpha1.KubernetesLangfuseRedisTls{}
	}
	common.plain("REDIS_TLS_SERVERNAME", tls.GetServerName())
	common.plain("REDIS_TLS_REJECT_UNAUTHORIZED", optionalBool(tls.RejectUnauthorized))
	common.plain("REDIS_TLS_CHECK_SERVER_IDENTITY", optionalBool(tls.CheckServerIdentity))
	common.plain("REDIS_TLS_CIPHERS", tls.GetCiphers())
	common.plain("REDIS_TLS_SECURE_PROTOCOL", tls.GetSecureProtocol())
	common.plain("REDIS_TLS_HONOR_CIPHER_ORDER", optionalBool(tls.HonorCipherOrder))
	if l.DeclaredSecrets[keyRedisTlsKeyPassphrase] != "" {
		common.instance("REDIS_TLS_KEY_PASSPHRASE", keyRedisTlsKeyPassphrase)
	}

	// ---- blob storage ----------------------------------------------------
	storage := spec.GetBlobStorage()
	sse(common, "EVENT_UPLOAD", storage.GetEventUpload().GetSse())
	sse(common, "MEDIA_UPLOAD", storage.GetMediaUpload().GetSse())
	sse(common, "BATCH_EXPORT", storage.GetBatchExport().GetSse())
	if storage.GetMediaUpload().GetPublicEndpoint() != "" {
		common.plain("LANGFUSE_S3_MEDIA_UPLOAD_INTERNAL_ENDPOINT", l.MediaTarget.Endpoint)
	}
	common.plain("LANGFUSE_S3_BATCH_EXPORT_EXTERNAL_ENDPOINT", storage.GetBatchExport().GetPublicEndpoint())
	if a := storage.GetDeleteObjectsChecksumAlgorithm(); a != kuberneteslangfusev1alpha1.KubernetesLangfuseBlobStorage_unspecified {
		common.plain("LANGFUSE_S3_DELETE_OBJECTS_CHECKSUM_ALGORITHM", strings.ToUpper(a.String()))
	}
	if batch := storage.GetBatchExport(); batch != nil {
		worker.plain("BATCH_EXPORT_PAGE_SIZE", optionalInt(batch.PageSize))
		worker.plain("BATCH_EXPORT_ROW_LIMIT", optionalInt(batch.RowLimit))
		worker.plain("BATCH_EXPORT_DOWNLOAD_LINK_EXPIRATION_HOURS", optionalInt(batch.DownloadLinkExpirationHours))
	}

	// ---- seeding (web only: the web server runs headless initialization) -
	if l.InitEnabled {
		organization := spec.GetInit().GetOrganization()
		if organization == nil {
			organization = &kuberneteslangfusev1alpha1.KubernetesLangfuseInitOrganization{}
		}
		project := spec.GetInit().GetProject()
		if project == nil {
			project = &kuberneteslangfusev1alpha1.KubernetesLangfuseInitProject{}
		}
		web.plain("LANGFUSE_INIT_ORG_ID", l.OrganizationId)
		web.plain("LANGFUSE_INIT_ORG_NAME", stringOr(organization.Name, vars.DefaultOrganizationName))
		web.plain("LANGFUSE_INIT_PROJECT_ID", l.ProjectId)
		web.plain("LANGFUSE_INIT_PROJECT_NAME", stringOr(project.Name, vars.DefaultProjectName))
		web.plain("LANGFUSE_INIT_PROJECT_RETENTION", optionalInt(project.RetentionDays))
		web.instance("LANGFUSE_INIT_PROJECT_PUBLIC_KEY", keyInitProjectPublicKey)
		web.instance("LANGFUSE_INIT_PROJECT_SECRET_KEY", keyInitProjectSecretKey)
		if l.AdminSeeded {
			web.plain("LANGFUSE_INIT_USER_EMAIL", l.AdminEmail)
			web.plain("LANGFUSE_INIT_USER_NAME", stringOr(spec.GetInit().GetAdmin().Name, vars.DefaultAdminName))
			web.instance("LANGFUSE_INIT_USER_PASSWORD", keyInitUserPassword)
		}
	}

	// ---- auth (web only) -------------------------------------------------
	auth := spec.GetAuth()
	if auth == nil {
		auth = &kuberneteslangfusev1alpha1.KubernetesLangfuseAuth{}
	}
	web.plain("AUTH_EMAIL_VERIFICATION_REQUIRED", trueOrEmpty(auth.GetEmailVerificationRequired()))
	web.plain("AUTH_SESSION_MAX_AGE", optionalInt(auth.SessionMaxAgeMinutes))
	web.plain("AUTH_DOMAINS_WITH_SSO_ENFORCEMENT", csv(auth.GetSsoEnforcedDomains()))
	web.plain("AUTH_IGNORE_ACCOUNT_FIELDS", csv(auth.GetIgnoreAccountFields()))
	web.plain("AUTH_HTTPS_PROXY", auth.GetHttpsProxy())
	web.plain("AUTH_HTTP_PROXY", auth.GetHttpProxy())
	web.plain("AUTH_SSO_TIMEOUT", optionalInt(auth.SsoTimeoutMs))
	access := auth.GetNewUserAccess()
	web.plain("LANGFUSE_DEFAULT_ORG_ID", access.GetOrganizationId())
	if r := access.GetOrganizationRole(); r != kuberneteslangfusev1alpha1.KubernetesLangfuseNewUserAccess_unspecified {
		web.plain("LANGFUSE_DEFAULT_ORG_ROLE", r.String())
	}
	web.plain("LANGFUSE_DEFAULT_PROJECT_ID", csv(access.GetProjectIds()))
	if r := access.GetProjectRole(); r != kuberneteslangfusev1alpha1.KubernetesLangfuseNewUserAccess_unspecified {
		web.plain("LANGFUSE_DEFAULT_PROJECT_ROLE", r.String())
	}
	for _, p := range ssoProviders(auth.GetSso()) {
		p.env(web, fmt.Sprintf(keySsoClientSecretTmpl, p.Id))
	}
	if claims := auth.GetSso().GetCustom().GetClaims(); claims != nil {
		web.plain("LANGFUSE_CUSTOM_SSO_EMAIL_CLAIM", claims.GetEmail())
		web.plain("LANGFUSE_CUSTOM_SSO_NAME_CLAIM", claims.GetName())
		web.plain("LANGFUSE_CUSTOM_SSO_SUB_CLAIM", claims.GetSubject())
		web.plain("LANGFUSE_CUSTOM_SSO_IMAGE_CLAIM", claims.GetImage())
	}

	// ---- smtp ------------------------------------------------------------
	if smtp := spec.GetSmtp(); smtp != nil {
		common.instance("SMTP_CONNECTION_URL", keySmtpConnectionUrl)
		common.plain("EMAIL_FROM_ADDRESS", smtp.GetFromAddress())
	}

	// ---- features --------------------------------------------------------
	// The chart's own experimental-features value sets a variable Langfuse
	// never reads (ENABLE_EXPERIMENTAL_FEATURES); the module sets the one it
	// does, on the web server that serves the features.
	features := spec.GetFeatures()
	if features == nil {
		features = &kuberneteslangfusev1alpha1.KubernetesLangfuseFeatures{}
	}
	web.plain("LANGFUSE_ENABLE_EXPERIMENTAL_FEATURES", trueOrEmpty(features.GetExperimentalFeaturesEnabled()))
	web.plain("LANGFUSE_CSP_ENFORCE_HTTPS", trueOrEmpty(features.GetCspEnforceHttps()))
	web.plain("LANGFUSE_MARKDOWN_RENDER_CHARACTER_LIMIT", optionalInt(features.MarkdownRenderCharacterLimit))

	// ---- caching ---------------------------------------------------------
	caching := spec.GetCaching()
	if caching == nil {
		caching = &kuberneteslangfusev1alpha1.KubernetesLangfuseCaching{}
	}
	web.plain("LANGFUSE_CACHE_API_KEY_ENABLED", optionalBool(caching.ApiKeyCacheEnabled))
	web.plain("LANGFUSE_CACHE_API_KEY_TTL_SECONDS", optionalInt(caching.ApiKeyCacheTtlSeconds))
	common.plain("LANGFUSE_CACHE_PROMPT_ENABLED", optionalBool(caching.PromptCacheEnabled))
	common.plain("LANGFUSE_CACHE_PROMPT_TTL_SECONDS", optionalInt(caching.PromptCacheTtlSeconds))

	// ---- ingestion -------------------------------------------------------
	// Shard counts reach both workloads: the web server enqueues to the
	// shards the worker consumes.
	ingestion := spec.GetIngestion()
	if ingestion == nil {
		ingestion = &kuberneteslangfusev1alpha1.KubernetesLangfuseIngestion{}
	}
	common.plain("LANGFUSE_INGESTION_QUEUE_SHARD_COUNT", optionalInt(ingestion.QueueShardCount))
	common.plain("LANGFUSE_TRACE_UPSERT_QUEUE_SHARD_COUNT", optionalInt(ingestion.TraceUpsertQueueShardCount))
	worker.plain("LANGFUSE_INGESTION_QUEUE_PROCESSING_CONCURRENCY", optionalInt(ingestion.QueueProcessingConcurrency))
	worker.plain("LANGFUSE_TRACE_UPSERT_WORKER_CONCURRENCY", optionalInt(ingestion.TraceUpsertWorkerConcurrency))
	worker.plain("LANGFUSE_SKIP_INGESTION_CLICKHOUSE_READ_MIN_PROJECT_CREATE_DATE", ingestion.GetSkipClickhouseReadMinProjectCreateDate())

	// ---- langfuse ai -----------------------------------------------------
	// Rendered here rather than through the chart's aiFeatures values: the
	// chart's provider list predates Vertex.
	ai := spec.GetAiFeatures()
	if p := ai.GetProvider(); p != kuberneteslangfusev1alpha1.KubernetesLangfuseAiFeatures_unspecified {
		common.plain("LANGFUSE_AI_PROVIDER", p.String())
	}
	common.plain("LANGFUSE_AI_MODEL", ai.GetModel())
	common.plain("LANGFUSE_AI_SMALL_MODEL", ai.GetSmallModel())
	if l.DeclaredSecrets[keyAiApiKey] != "" {
		common.instance("LANGFUSE_AI_API_KEY", keyAiApiKey)
	}
	common.plain("LANGFUSE_AI_BASE_URL", ai.GetBaseUrl())
	if len(ai.GetExtraHeaders()) > 0 {
		headers, _ := json.Marshal(ai.GetExtraHeaders())
		common.plain("LANGFUSE_AI_EXTRA_HEADERS", string(headers))
	}
	common.plain("LANGFUSE_AI_USE_RESPONSES_API", trueOrEmpty(ai.GetUseResponsesApi()))
	common.plain("LANGFUSE_AI_AWS_BEDROCK_REGION", ai.GetBedrockRegion())
	common.plain("LANGFUSE_AI_VERTEX_LOCATION", ai.GetVertexLocation())
	common.plain("LANGFUSE_AI_FEATURES_PROJECT_ID", ai.GetTraceProjectId())
	assistant := ai.GetAssistant()
	if assistant == nil {
		assistant = &kuberneteslangfusev1alpha1.KubernetesLangfuseAssistant{}
	}
	common.plain("LANGFUSE_IN_APP_AGENT_ENABLED", trueOrEmpty(assistant.GetEnabled()))
	web.plain("LANGFUSE_IN_APP_AGENT_MAX_ACTIVE_RUNS_PER_USER", optionalInt(assistant.MaxActiveRunsPerUser))
	web.plain("LANGFUSE_IN_APP_AGENT_MAX_ACTIVE_RUNS_PER_ORG", optionalInt(assistant.MaxActiveRunsPerOrg))
	worker.plain("LANGFUSE_IN_APP_AGENT_RUN_QUEUE_PROCESSING_CONCURRENCY", optionalInt(assistant.RunConcurrency))
	if assistant.GetUseInternalWebUrl() {
		web.plain("LANGFUSE_MCP_ALLOWED_HOSTS", l.WebServiceName)
		worker.plain("LANGFUSE_MCP_BASE_URL", fmt.Sprintf("http://%s:%d", l.WebServiceName, l.WebServicePort))
	}
	if sandbox := assistant.GetSandbox(); sandbox != nil {
		worker.plain("LANGFUSE_IN_APP_AGENT_SANDBOX_PROVIDER", "lambda-microvm")
		worker.plain("LANGFUSE_IN_APP_AGENT_SANDBOX_AWS_LAMBDA_MICROVM_IMAGE_IDENTIFIER", sandbox.GetImageIdentifier())
		worker.plain("LANGFUSE_IN_APP_AGENT_SANDBOX_AWS_LAMBDA_MICROVM_EXECUTION_ROLE_ARN", sandbox.GetExecutionRoleArn().GetValue())
		worker.plain("LANGFUSE_IN_APP_AGENT_SANDBOX_AWS_LAMBDA_MICROVM_REGION", sandbox.GetRegion())
		worker.plain("LANGFUSE_IN_APP_AGENT_SANDBOX_AWS_LAMBDA_MICROVM_EGRESS_NETWORK_CONNECTOR_ARN", sandbox.GetEgressNetworkConnectorArn())
	}

	// ---- evaluation ------------------------------------------------------
	evaluation := spec.GetEvaluation()
	if evaluation == nil {
		evaluation = &kuberneteslangfusev1alpha1.KubernetesLangfuseEvaluation{}
	}
	if t := evaluation.GetMediaTransport(); t != kuberneteslangfusev1alpha1.KubernetesLangfuseEvaluation_unspecified {
		common.plain("LANGFUSE_EVALUATOR_MEDIA_TRANSPORT", t.String())
	}
	common.plain("LANGFUSE_EVALUATOR_MEDIA_INLINE_MAX_BYTES", optionalInt(evaluation.MediaInlineMaxBytes))
	if code := evaluation.GetCodeEvaluators(); code != nil {
		if lambda := code.GetAwsLambda(); lambda != nil {
			common.plain("LANGFUSE_CODE_EVAL_DISPATCHER", "aws-lambda")
			common.plain("LANGFUSE_CODE_EVAL_AWS_LAMBDA_NODE_FUNCTION_NAME", lambda.GetNodeFunctionName())
			common.plain("LANGFUSE_CODE_EVAL_AWS_LAMBDA_PYTHON_FUNCTION_NAME", lambda.GetPythonFunctionName())
			common.plain("LANGFUSE_CODE_EVAL_AWS_LAMBDA_ENDPOINT", lambda.GetEndpoint())
		} else {
			common.plain("LANGFUSE_CODE_EVAL_DISPATCHER", "insecure-local")
			common.plain("LANGFUSE_CODE_EVAL_LOCAL_TIMEOUT_MS", optionalInt(code.GetInsecureLocal().TimeoutMs))
		}
		common.plain("LANGFUSE_CODE_EVAL_EXECUTION_QUEUE_SHARD_COUNT", optionalInt(code.QueueShardCount))
		worker.plain("LANGFUSE_CODE_EVAL_EXECUTION_WORKER_CONCURRENCY", optionalInt(code.Concurrency))
		worker.plain("QUEUE_CONSUMER_CODE_EVAL_EXECUTION_QUEUE_IS_ENABLED", "true")
	}

	// ---- slack -----------------------------------------------------------
	if slack := spec.GetSlack(); slack != nil {
		common.plain("SLACK_CLIENT_ID", slack.GetClientId())
		common.instance("SLACK_CLIENT_SECRET", keySlackClientSecret)
		common.instance("SLACK_STATE_SECRET", keySlackStateSecret)
	}

	// ---- enterprise ------------------------------------------------------
	enterprise := spec.GetEnterprise()
	if l.DeclaredSecrets[keyAdminApiKey] != "" {
		web.instance("ADMIN_API_KEY", keyAdminApiKey)
	}
	ui := enterprise.GetUi()
	web.plain("LANGFUSE_UI_DOCUMENTATION_HREF", ui.GetDocumentationHref())
	web.plain("LANGFUSE_UI_SUPPORT_HREF", ui.GetSupportHref())
	web.plain("LANGFUSE_UI_FEEDBACK_HREF", ui.GetFeedbackHref())
	web.plain("LANGFUSE_UI_API_HOST", ui.GetApiHost())
	web.plain("LANGFUSE_UI_LOGO_LIGHT_MODE_HREF", ui.GetLogoLightModeHref())
	web.plain("LANGFUSE_UI_LOGO_DARK_MODE_HREF", ui.GetLogoDarkModeHref())
	if a := ui.GetDefaultModelAdapter(); a != kuberneteslangfusev1alpha1.KubernetesLangfuseUiCustomization_unspecified {
		web.plain("LANGFUSE_UI_DEFAULT_MODEL_ADAPTER", a.String())
	}
	web.plain("LANGFUSE_UI_DEFAULT_BASE_URL_OPENAI", ui.GetDefaultBaseUrlOpenai())
	web.plain("LANGFUSE_UI_DEFAULT_BASE_URL_ANTHROPIC", ui.GetDefaultBaseUrlAnthropic())
	web.plain("LANGFUSE_UI_DEFAULT_BASE_URL_AZURE", ui.GetDefaultBaseUrlAzure())
	web.plain("LANGFUSE_UI_VISIBLE_PRODUCT_MODULES", csv(ui.GetVisibleProductModules()))
	web.plain("LANGFUSE_UI_HIDDEN_PRODUCT_MODULES", csv(ui.GetHiddenProductModules()))
	if links := ui.GetInstanceLinks(); len(links) > 0 {
		list := make([]map[string]string, 0, len(links))
		for _, link := range links {
			list = append(list, map[string]string{"name": link.GetName(), "url": link.GetUrl()})
		}
		encoded, _ := json.Marshal(list)
		web.plain("LANGFUSE_UI_INSTANCE_LINKS", string(encoded))
	}
	if masking := enterprise.GetIngestionMasking(); masking != nil {
		common.plain("LANGFUSE_INGESTION_MASKING_CALLBACK_URL", masking.GetCallbackUrl())
		common.plain("LANGFUSE_INGESTION_MASKING_CALLBACK_TIMEOUT_MS", optionalInt(masking.TimeoutMs))
		common.plain("LANGFUSE_INGESTION_MASKING_CALLBACK_FAIL_CLOSED", trueOrEmpty(masking.GetFailClosed()))
		common.plain("LANGFUSE_INGESTION_MASKING_MAX_RETRIES", optionalInt(masking.MaxRetries))
		common.plain("LANGFUSE_INGESTION_MASKING_PROPAGATED_HEADERS", csv(masking.GetPropagatedHeaders()))
	}

	// ---- observability ---------------------------------------------------
	observability := spec.GetObservability()
	if observability == nil {
		observability = &kuberneteslangfusev1alpha1.KubernetesLangfuseObservability{}
	}
	common.plain("LANGFUSE_LOG_PROPAGATED_HEADERS", csv(observability.GetLogPropagatedHeaders()))
	common.plain("OTEL_EXPORTER_OTLP_ENDPOINT", observability.GetOtlpEndpoint().GetValue())
	if observability.TraceSamplingRatio != nil {
		web.plain("OTEL_TRACE_SAMPLING_RATIO", strconv.FormatFloat(observability.GetTraceSamplingRatio(), 'f', -1, 64))
	}

	// ---- network ---------------------------------------------------------
	network := spec.GetNetwork()
	common.plain("HTTPS_PROXY", network.GetHttpsProxy())
	common.plain("NO_PROXY", csv(network.GetNoProxy()))
	common.plain("NODE_EXTRA_CA_CERTS", l.TrustedCaPath)
	allowlists := network.GetAllowlists()
	allowlist(common, "LANGFUSE_WEBHOOK", allowlists.GetWebhooks())
	allowlist(common, "LANGFUSE_LLM_CONNECTION", allowlists.GetLlmConnections())
	allowlist(common, "LANGFUSE_BLOB_STORAGE_ENDPOINT", allowlists.GetBlobStorageEndpoints())
	allowlist(common, "LANGFUSE_EXTERNAL_MEDIA_STORAGE_ENDPOINT", allowlists.GetExternalMediaStorageEndpoints())
	allowlist(common, "LANGFUSE_SSO_DISCOVERY", allowlists.GetSsoDiscovery())

	// ---- worker health ---------------------------------------------------
	if w := spec.GetWorker(); w != nil {
		worker.plain("LANGFUSE_QUEUE_CONSUMPTION_STUCK_THRESHOLD_MINUTES", optionalInt(w.QueueConsumptionStuckThresholdMinutes))
	}

	// ---- extra environment (last: the spec's own entries win) ------------
	extraEnv(common, spec.GetExtraEnv(), spec.GetExtraEnvFromSecret())
	extraEnv(web, spec.GetWeb().GetExtraEnv(), spec.GetWeb().GetExtraEnvFromSecret())
	extraEnv(worker, spec.GetWorker().GetExtraEnv(), spec.GetWorker().GetExtraEnvFromSecret())

	l.CommonEnv, l.WebEnv, l.WorkerEnv = common.sorted(), web.sorted(), worker.sorted()
}

func sse(env envSet, uploadType string, s *kuberneteslangfusev1alpha1.KubernetesLangfuseServerSideEncryption) {
	if s == nil {
		return
	}
	algorithm := "AES256"
	if s.GetAlgorithm() == kuberneteslangfusev1alpha1.KubernetesLangfuseServerSideEncryption_aws_kms {
		algorithm = "aws:kms"
	}
	env.plain("LANGFUSE_S3_"+uploadType+"_SSE", algorithm)
	env.plain("LANGFUSE_S3_"+uploadType+"_SSE_KMS_KEY_ID", s.GetKmsKeyId().GetValue())
}

func allowlist(env envSet, prefix string, a *kuberneteslangfusev1alpha1.KubernetesLangfuseAllowlist) {
	env.plain(prefix+"_WHITELISTED_HOST", csv(a.GetHosts()))
	env.plain(prefix+"_WHITELISTED_IPS", csv(a.GetIps()))
	env.plain(prefix+"_WHITELISTED_IP_SEGMENTS", csv(a.GetIpSegments()))
}

func extraEnv(env envSet, plain map[string]string, fromSecret map[string]*kubernetesprovider.KubernetesSecretKey) {
	for name, value := range plain {
		env.plain(name, value)
	}
	for name, ref := range fromSecret {
		env.secret(name, ref)
	}
}

func optionalInt(value *int32) string {
	if value == nil {
		return ""
	}
	return strconv.Itoa(int(*value))
}

func optionalBool(value *bool) string {
	if value == nil {
		return ""
	}
	return strconv.FormatBool(*value)
}

func trueOrEmpty(value bool) string {
	if value {
		return "true"
	}
	return ""
}
