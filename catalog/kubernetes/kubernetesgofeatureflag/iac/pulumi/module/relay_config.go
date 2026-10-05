package module

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
	kubernetesgofeatureflagv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgofeatureflag/v1alpha1"
)

// RelayConfig is everything the relay's configuration renders into: the
// configuration document (secret-free), the secret values that reach the
// relay as environment variables, and the ConfigMaps its retrievers read.
type RelayConfig struct {
	// Document is the relay configuration as the chart's
	// `relayproxy.config` string: a `# server.monitoringPort: <port>` line
	// (the chart discovers the monitoring port by scanning the string for
	// exactly that text) followed by the configuration as one JSON object
	// (JSON is YAML; both engines produce it byte for byte - Go
	// json.Marshal and OpenTofu jsonencode sort keys and escape alike), with
	// every "{{" escaped because the chart runs the string through Helm's
	// tpl.
	Document string

	// SecretEnv maps each generated environment variable name to its secret
	// value. The relay reads list entries from indexed variables
	// (RETRIEVERS_<i>_TOKEN, FLAGSETS_<i>_RETRIEVERS_<j>_TOKEN, ...) and key
	// lists from comma-separated ones (AUTHORIZEDKEYS_ADMIN,
	// FLAGSETS_<i>_APIKEYS); every name carries the env-variable prefix.
	SecretEnv map[string]string

	// ConfigMapGrants lists, per namespace, the ConfigMap names the
	// retrievers read - the module grants `get` on exactly these.
	ConfigMapGrants map[string][]string
}

// KEY SPELLING (verified against the relay's loader at v1.56.0). The relay
// loads defaults, then this file, then its environment, into one key tree,
// and decodes it case-insensitively - but a file key and an environment key
// that differ only in case are TWO keys, and which one wins is undefined.
// So:
//   - keys the relay pre-loads with defaults stay in their documented
//     camelCase (fileFormat, pollingInterval, logLevel) so this file
//     overrides the same key, and envVariablePrefix stays camelCase because
//     the relay looks it up by that exact spelling;
//   - every map that a secret environment variable also writes into is
//     spelled in the lowercase form the environment produces: `headers`,
//     `redisoptions`, `kafka.config` and every key inside a Kafka config;
//   - secret fields never appear in the file, only in the environment.

// buildRelayConfig renders the relay configuration from the spec.
func buildRelayConfig(spec *kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagSpec, relayNamespace string) (*RelayConfig, error) {
	r := &renderer{
		prefix:         envPrefix(spec),
		relayNamespace: relayNamespace,
		secretEnv:      map[string]string{},
		grants:         map[string]map[string]bool{},
	}

	port := vars.DefaultPort
	if spec.GetServer() != nil && spec.GetServer().Port != nil {
		port = int(spec.GetServer().GetPort())
	}
	monitoringPort := vars.DefaultMonitoringPort
	if spec.GetServer() != nil && spec.GetServer().MonitoringPort != nil {
		monitoringPort = int(spec.GetServer().GetMonitoringPort())
	}

	doc := map[string]interface{}{
		"server": map[string]interface{}{
			"mode":           "http",
			"port":           port,
			"monitoringPort": monitoringPort,
		},
		"envVariablePrefix": r.prefix,
	}

	logLevel := "info"
	if spec.GetLog().GetLevel() != "" {
		logLevel = spec.GetLog().GetLevel()
	}
	doc["logLevel"] = logLevel
	logFormat := "json"
	if spec.GetLog().GetFormat() != "" {
		logFormat = spec.GetLog().GetFormat()
	}
	doc["logFormat"] = logFormat

	if keys := spec.GetAuthorizedKeys(); keys != nil {
		if err := r.keyList("AUTHORIZEDKEYS_ADMIN", "authorized_keys.admin", keys.GetAdmin()); err != nil {
			return nil, err
		}
		if err := r.keyList("AUTHORIZEDKEYS_EVALUATION", "authorized_keys.evaluation", keys.GetEvaluation()); err != nil {
			return nil, err
		}
	}

	switch {
	case spec.GetFlagSource() != nil:
		source, err := r.source(spec.GetFlagSource(), "")
		if err != nil {
			return nil, err
		}
		for k, v := range source {
			doc[k] = v
		}
	case spec.GetFlagSets() != nil:
		seenKeys := map[string]int{}
		flagsets := []interface{}{}
		for i, fs := range spec.GetFlagSets().GetItems() {
			envPath := fmt.Sprintf("FLAGSETS_%d_", i)
			rendered, err := r.source(fs.GetSource(), envPath)
			if err != nil {
				return nil, err
			}
			if fs.GetName() != "" {
				rendered["name"] = fs.GetName()
			}
			for _, k := range fs.GetApiKeys() {
				if j, dup := seenKeys[k]; dup {
					return nil, errors.Errorf("flag_sets.items[%d] and flag_sets.items[%d] share an API key; every key must select exactly one flag set", j, i)
				}
				seenKeys[k] = i
			}
			if err := r.keyList(envPath+"APIKEYS", fmt.Sprintf("flag_sets.items[%d].api_keys", i), fs.GetApiKeys()); err != nil {
				return nil, err
			}
			flagsets = append(flagsets, rendered)
		}
		doc["flagsets"] = flagsets
	}

	if es := spec.GetOfrepEventStream(); es != nil && (es.GetBaseUrl() != "" || es.GetInactivityDelaySec() != 0) {
		stream := map[string]interface{}{}
		if es.GetBaseUrl() != "" {
			stream["baseUrl"] = es.GetBaseUrl()
		}
		if es.GetInactivityDelaySec() != 0 {
			stream["inactivityDelaySec"] = int(es.GetInactivityDelaySec())
		}
		doc["ofrepEventStream"] = stream
	}

	if t := spec.GetTelemetry(); t != nil {
		otel := map[string]interface{}{}
		if t.GetOtlpEndpoint() != "" {
			otel["exporter"] = map[string]interface{}{"otlp": map[string]interface{}{"endpoint": t.GetOtlpEndpoint()}}
		}
		if t.GetSdkDisabled() {
			otel["sdk"] = map[string]interface{}{"disabled": true}
		}
		if t.GetServiceName() != "" {
			otel["service"] = map[string]interface{}{"name": t.GetServiceName()}
		}
		if t.GetTracesSampler() != "" {
			otel["traces"] = map[string]interface{}{"sampler": t.GetTracesSampler()}
		}
		if len(t.GetResourceAttributes()) > 0 {
			otel["resource"] = map[string]interface{}{"attributes": stringMap(t.GetResourceAttributes())}
		}
		if len(otel) > 0 {
			doc["otel"] = otel
		}
		if js := t.GetJaegerSampler(); js != nil {
			sampler := map[string]interface{}{}
			if js.GetManagerHostPort() != "" {
				sampler["manager"] = map[string]interface{}{"host": map[string]interface{}{"port": js.GetManagerHostPort()}}
			}
			if js.GetRefreshInterval() != "" {
				sampler["refresh"] = map[string]interface{}{"interval": js.GetRefreshInterval()}
			}
			if js.GetMaxOperations() != 0 {
				sampler["max"] = map[string]interface{}{"operations": int(js.GetMaxOperations())}
			}
			if len(sampler) > 0 {
				doc["jaeger"] = map[string]interface{}{"sampler": sampler}
			}
		}
	}

	if sw := spec.GetSwagger(); sw != nil && (sw.GetEnabled() || sw.GetHost() != "") {
		swagger := map[string]interface{}{"enabled": sw.GetEnabled()}
		if sw.GetHost() != "" {
			swagger["host"] = sw.GetHost()
		}
		doc["swagger"] = swagger
	}

	if rt := spec.GetRuntime(); rt != nil {
		setTrue(doc, "hideBanner", rt.GetHideBanner())
		setTrue(doc, "enablePprof", rt.GetEnablePprof())
		setTrue(doc, "disableVersionHeader", rt.GetDisableVersionHeader())
		setTrue(doc, "enableBulkMetricFlagNames", rt.GetEnableBulkMetricFlagNames())
		setTrue(doc, "disableFlagDetailsInStream", rt.GetDisableFlagDetailsInStream())
		if rt.GetExporterCleanQueueInterval() != "" {
			doc["exporterCleanQueueInterval"] = rt.GetExporterCleanQueueInterval()
		}
	}

	body, err := json.Marshal(doc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode the relay configuration")
	}

	grants := map[string][]string{}
	for ns, names := range r.grants {
		list := make([]string, 0, len(names))
		for n := range names {
			list = append(list, n)
		}
		sort.Strings(list)
		grants[ns] = list
	}

	return &RelayConfig{
		Document:        fmt.Sprintf("# server.monitoringPort: %d\n%s", monitoringPort, escapeHelmTemplate(string(body))),
		SecretEnv:       r.secretEnv,
		ConfigMapGrants: grants,
	}, nil
}

// envPrefix resolves the env-variable prefix (the spec default when unset).
func envPrefix(spec *kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagSpec) string {
	if p := spec.GetRuntime().GetEnvVariablePrefix(); p != "" {
		return p
	}
	return vars.DefaultEnvVariablePrefix
}

// escapeHelmTemplate keeps "{{" literal through the chart's tpl pass.
func escapeHelmTemplate(s string) string {
	return strings.ReplaceAll(s, "{{", `{{"{{"}}`)
}

type renderer struct {
	prefix         string
	relayNamespace string
	secretEnv      map[string]string
	grants         map[string]map[string]bool
}

// keyList records a comma-separated key list variable, refusing a key with
// a comma (the relay splits the variable on commas).
func (r *renderer) keyList(name, field string, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	for _, k := range keys {
		if strings.Contains(k, ",") {
			return errors.Errorf("%s holds a key containing a comma; the relay reads key lists comma-separated, so a key may not contain one", field)
		}
	}
	r.secretEnv[r.prefix+name] = strings.Join(keys, ",")
	return nil
}

// secret records one secret value under its indexed variable name.
func (r *renderer) secret(name, value string) {
	if value == "" {
		return
	}
	r.secretEnv[r.prefix+name] = value
}

// headers renders plain headers into the file and sensitive headers into
// the environment (both land in the same lowercase `headers` map).
func (r *renderer) headers(element map[string]interface{}, envPath, field string, plain, sensitive map[string]string) error {
	lowerPlain := map[string]bool{}
	for k := range plain {
		lowerPlain[strings.ToLower(k)] = true
	}
	for k, v := range sensitive {
		if strings.Contains(k, "_") {
			return errors.Errorf("%s.sensitive_headers[%q]: a sensitive header name may not contain \"_\" (the relay reads it from an environment variable whose name it splits on \"_\")", field, k)
		}
		if lowerPlain[strings.ToLower(k)] {
			return errors.Errorf("%s: header %q is declared both plain and sensitive; declare it once", field, k)
		}
		r.secret(envPath+"HEADERS_"+strings.ToUpper(k), v)
	}
	if len(plain) > 0 {
		h := map[string]interface{}{}
		for k, v := range plain {
			h[k] = []interface{}{v}
		}
		element["headers"] = h
	}
	return nil
}

// source renders one flag source (retrievers, notifiers, exporters and the
// common settings); envPath is "" in flag_source mode and "FLAGSETS_<i>_"
// in flag_sets mode.
func (r *renderer) source(src *kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagFlagSource, envPath string) (map[string]interface{}, error) {
	out := map[string]interface{}{}

	retrievers := []interface{}{}
	for j, ret := range src.GetRetrievers() {
		element, err := r.retriever(ret, fmt.Sprintf("%sRETRIEVERS_%d_", envPath, j), fmt.Sprintf("retrievers[%d]", j))
		if err != nil {
			return nil, err
		}
		retrievers = append(retrievers, element)
	}
	out["retrievers"] = retrievers

	if len(src.GetNotifiers()) > 0 {
		notifiers := []interface{}{}
		for j, n := range src.GetNotifiers() {
			element, err := r.notifier(n, fmt.Sprintf("%sNOTIFIERS_%d_", envPath, j), fmt.Sprintf("notifiers[%d]", j))
			if err != nil {
				return nil, err
			}
			notifiers = append(notifiers, element)
		}
		out["notifiers"] = notifiers
	}

	if len(src.GetExporters()) > 0 {
		exporters := []interface{}{}
		for j, e := range src.GetExporters() {
			element, err := r.exporter(e, fmt.Sprintf("%sEXPORTERS_%d_", envPath, j), fmt.Sprintf("exporters[%d]", j))
			if err != nil {
				return nil, err
			}
			exporters = append(exporters, element)
		}
		out["exporters"] = exporters
	}

	if src.GetFileFormat() != "" {
		out["fileFormat"] = src.GetFileFormat()
	}
	if src.PollingIntervalMs != nil {
		out["pollingInterval"] = int(src.GetPollingIntervalMs())
	}
	if src.StartWithRetrieverError != nil {
		out["startWithRetrieverError"] = src.GetStartWithRetrieverError()
	}
	setTrue(out, "enablePollingJitter", src.GetEnablePollingJitter())
	setTrue(out, "disableNotifierOnInit", src.GetDisableNotifierOnInit())
	if enrichment := src.GetEvaluationContextEnrichment(); len(enrichment.GetFields()) > 0 {
		out["evaluationContextEnrichment"] = enrichment.AsMap()
	}
	return out, nil
}

func (r *renderer) retriever(ret *kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRetriever, envPath, field string) (map[string]interface{}, error) {
	element := map[string]interface{}{}
	switch {
	case ret.GetConfigMap() != nil:
		cm := ret.GetConfigMap()
		ns := cm.GetNamespace().GetValue()
		if ns == "" {
			ns = r.relayNamespace
		}
		element["kind"] = "configmap"
		element["namespace"] = ns
		element["configmap"] = cm.GetConfigMapName().GetValue()
		element["key"] = cm.GetKey().GetValue()
		if r.grants[ns] == nil {
			r.grants[ns] = map[string]bool{}
		}
		r.grants[ns][cm.GetConfigMapName().GetValue()] = true
	case ret.GetHttp() != nil:
		h := ret.GetHttp()
		element["kind"] = "http"
		element["url"] = h.GetUrl()
		setString(element, "method", h.GetMethod())
		setString(element, "body", h.GetBody())
		if h.TimeoutMs != nil {
			element["timeout"] = int(h.GetTimeoutMs())
		}
		if err := r.headers(element, envPath, field+".http", h.GetHeaders(), h.GetSensitiveHeaders()); err != nil {
			return nil, err
		}
	case ret.GetGithub() != nil:
		r.git(element, "github", ret.GetGithub(), envPath)
	case ret.GetGitlab() != nil:
		r.git(element, "gitlab", ret.GetGitlab(), envPath)
	case ret.GetBitbucket() != nil:
		r.git(element, "bitbucket", ret.GetBitbucket(), envPath)
	case ret.GetS3() != nil:
		element["kind"] = "s3"
		element["bucket"] = ret.GetS3().GetBucket()
		element["item"] = ret.GetS3().GetItem()
	case ret.GetGoogleStorage() != nil:
		element["kind"] = "googleStorage"
		element["bucket"] = ret.GetGoogleStorage().GetBucket()
		element["object"] = ret.GetGoogleStorage().GetObject()
	case ret.GetAzureBlobStorage() != nil:
		a := ret.GetAzureBlobStorage()
		element["kind"] = "azureBlobStorage"
		element["accountName"] = a.GetAccountName()
		element["container"] = a.GetContainer()
		element["object"] = a.GetObject()
		r.secret(envPath+"ACCOUNTKEY", a.GetAccountKey())
	case ret.GetMongodb() != nil:
		m := ret.GetMongodb()
		element["kind"] = "mongodb"
		element["database"] = m.GetDatabase()
		element["collection"] = m.GetCollection()
		r.secret(envPath+"URI", m.GetUri())
	case ret.GetRedis() != nil:
		element["kind"] = "redis"
		element["redisoptions"] = redisOptions(ret.GetRedis().GetOptions())
		setString(element, "redisPrefix", ret.GetRedis().GetPrefix())
		r.secret(envPath+"REDISOPTIONS_PASSWORD", ret.GetRedis().GetOptions().GetPassword())
	case ret.GetPostgresql() != nil:
		p := ret.GetPostgresql()
		element["kind"] = "postgresql"
		element["table"] = p.GetTable()
		if len(p.GetColumns()) > 0 {
			element["columns"] = stringMap(p.GetColumns())
		}
		r.secret(envPath+"URI", p.GetUri())
	}
	return element, nil
}

func (r *renderer) git(element map[string]interface{}, kind string, g *kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagGitRetriever, envPath string) {
	element["kind"] = kind
	element["repositorySlug"] = g.GetRepositorySlug()
	element["path"] = g.GetPath()
	setString(element, "branch", g.GetBranch())
	setString(element, "baseUrl", g.GetBaseUrl())
	if g.TimeoutMs != nil {
		element["timeout"] = int(g.GetTimeoutMs())
	}
	r.secret(envPath+"TOKEN", g.GetToken())
}

func redisOptions(o *kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRedisOptions) map[string]interface{} {
	out := map[string]interface{}{"addr": o.GetAddr()}
	setString(out, "network", o.GetNetwork())
	setString(out, "username", o.GetUsername())
	if o.GetDb() != 0 {
		out["db"] = int(o.GetDb())
	}
	setTrue(out, "tlsEnabled", o.GetTlsEnabled())
	if o.Protocol != nil {
		out["protocol"] = int(o.GetProtocol())
	}
	setString(out, "clientName", o.GetClientName())
	setString(out, "identitySuffix", o.GetIdentitySuffix())
	setTrue(out, "disableIdentity", o.GetDisableIdentity())
	if o.MaxRetries != nil {
		out["maxRetries"] = int(o.GetMaxRetries())
	}
	setInt64(out, "minRetryBackoff", o.MinRetryBackoffMs)
	setInt64(out, "maxRetryBackoff", o.MaxRetryBackoffMs)
	setInt64(out, "dialTimeout", o.DialTimeoutMs)
	setInt64(out, "readTimeout", o.ReadTimeoutMs)
	setInt64(out, "writeTimeout", o.WriteTimeoutMs)
	setTrue(out, "contextTimeoutEnabled", o.GetContextTimeoutEnabled())
	setTrue(out, "poolFIFO", o.GetPoolFifo())
	if o.PoolSize != nil {
		out["poolSize"] = int(o.GetPoolSize())
	}
	setInt64(out, "poolTimeout", o.PoolTimeoutMs)
	if o.GetMinIdleConns() != 0 {
		out["minIdleConns"] = int(o.GetMinIdleConns())
	}
	if o.GetMaxIdleConns() != 0 {
		out["maxIdleConns"] = int(o.GetMaxIdleConns())
	}
	setInt64(out, "connMaxIdleTime", o.ConnMaxIdleTimeMs)
	setInt64(out, "connMaxLifetime", o.ConnMaxLifetimeMs)
	return out
}

func (r *renderer) notifier(n *kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagNotifier, envPath, field string) (map[string]interface{}, error) {
	element := map[string]interface{}{}
	switch {
	case n.GetSlack() != nil:
		element["kind"] = "slack"
		r.secret(envPath+"WEBHOOKURL", n.GetSlack().GetWebhookUrl())
	case n.GetMicrosoftTeams() != nil:
		element["kind"] = "microsoftteams"
		r.secret(envPath+"WEBHOOKURL", n.GetMicrosoftTeams().GetWebhookUrl())
	case n.GetDiscord() != nil:
		element["kind"] = "discord"
		r.secret(envPath+"WEBHOOKURL", n.GetDiscord().GetWebhookUrl())
	case n.GetWebhook() != nil:
		w := n.GetWebhook()
		element["kind"] = "webhook"
		element["endpointUrl"] = w.GetEndpointUrl()
		if len(w.GetMeta()) > 0 {
			element["meta"] = stringMap(w.GetMeta())
		}
		r.secret(envPath+"SECRET", w.GetSecret())
		if err := r.headers(element, envPath, field+".webhook", w.GetHeaders(), w.GetSensitiveHeaders()); err != nil {
			return nil, err
		}
	}
	return element, nil
}

func (r *renderer) exporter(e *kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagExporter, envPath, field string) (map[string]interface{}, error) {
	element := map[string]interface{}{}
	switch {
	case e.GetWebhook() != nil:
		w := e.GetWebhook()
		element["kind"] = "webhook"
		element["endpointUrl"] = w.GetEndpointUrl()
		if len(w.GetMeta()) > 0 {
			element["meta"] = stringMap(w.GetMeta())
		}
		r.secret(envPath+"SECRET", w.GetSecret())
		if err := r.headers(element, envPath, field+".webhook", w.GetHeaders(), w.GetSensitiveHeaders()); err != nil {
			return nil, err
		}
	case e.GetLog() != nil:
		element["kind"] = "log"
		setString(element, "logFormat", e.GetLog().GetLogFormat())
	case e.GetS3() != nil:
		element["kind"] = "s3"
		element["bucket"] = e.GetS3().GetBucket()
		setString(element, "path", e.GetS3().GetPath())
		fileFormat(element, e.GetS3().GetFile())
	case e.GetGoogleStorage() != nil:
		element["kind"] = "googleStorage"
		element["bucket"] = e.GetGoogleStorage().GetBucket()
		setString(element, "path", e.GetGoogleStorage().GetPath())
		fileFormat(element, e.GetGoogleStorage().GetFile())
	case e.GetAzureBlobStorage() != nil:
		a := e.GetAzureBlobStorage()
		element["kind"] = "azureBlobStorage"
		element["accountName"] = a.GetAccountName()
		element["container"] = a.GetContainer()
		setString(element, "path", a.GetPath())
		fileFormat(element, a.GetFile())
		r.secret(envPath+"ACCOUNTKEY", a.GetAccountKey())
	case e.GetSqs() != nil:
		element["kind"] = "sqs"
		element["queueUrl"] = e.GetSqs().GetQueueUrl()
	case e.GetKinesis() != nil:
		k := e.GetKinesis()
		element["kind"] = "kinesis"
		setString(element, "streamArn", k.GetStreamArn())
		setString(element, "streamName", k.GetStreamName())
		setString(element, "format", k.GetFormat())
	case e.GetPubsub() != nil:
		element["kind"] = "pubsub"
		element["projectID"] = e.GetPubsub().GetProjectId()
		element["topic"] = e.GetPubsub().GetTopic()
	case e.GetBigquery() != nil:
		b := e.GetBigquery()
		element["kind"] = "bigquery"
		element["projectID"] = b.GetProjectId()
		element["datasetID"] = b.GetDatasetId()
		setString(element, "tableName", b.GetTableName())
		setTrue(element, "autoMigrate", b.GetAutoMigrate())
		r.secret(envPath+"GOOGLECREDENTIALS", b.GetGoogleCredentials())
	case e.GetKafka() != nil:
		k := e.GetKafka()
		element["kind"] = "kafka"
		kafka := map[string]interface{}{
			"topic":     k.GetTopic(),
			"addresses": stringSlice(k.GetAddresses()),
		}
		if cfg := k.GetConfig(); len(cfg.GetFields()) > 0 {
			kafka["config"] = lowerKeys(cfg.AsMap())
		}
		element["kafka"] = kafka
		r.secret(envPath+"KAFKA_CONFIG_NET_SASL_PASSWORD", k.GetSaslPassword())
	case e.GetOpentelemetry() != nil:
		element["kind"] = "opentelemetry"
		setString(element, "tracerName", e.GetOpentelemetry().GetTracerName())
	}
	if e.FlushIntervalMs != nil {
		element["flushInterval"] = e.GetFlushIntervalMs()
	}
	if e.MaxEventInMemory != nil {
		element["maxEventInMemory"] = e.GetMaxEventInMemory()
	}
	setString(element, "eventType", e.GetEventType())
	return element, nil
}

func fileFormat(element map[string]interface{}, f *kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagEventFileFormat) {
	if f == nil {
		return
	}
	setString(element, "format", f.GetFormat())
	setString(element, "filename", f.GetFilename())
	setString(element, "csvTemplate", f.GetCsvTemplate())
	setString(element, "parquetCompressionCodec", f.GetParquetCompressionCodec())
}

// lowerKeys lowercases every map key of a free-form object, recursively
// (the relay matches them case-insensitively; lowercase lets a secret
// environment variable merge into the same keys).
//
// PARITY-EXCEPTION: the Terraform module lowercases to a depth of six and
// leaves keys inside lists as written (HCL has no recursion); a Sarama
// configuration nests four levels deep and holds no lists of objects, so
// both engines render every real configuration identically.
func lowerKeys(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(t))
		for k, val := range t {
			out[strings.ToLower(k)] = lowerKeys(val)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(t))
		for i, val := range t {
			out[i] = lowerKeys(val)
		}
		return out
	default:
		return v
	}
}

func setString(m map[string]interface{}, key, value string) {
	if value != "" {
		m[key] = value
	}
}

func setTrue(m map[string]interface{}, key string, value bool) {
	if value {
		m[key] = true
	}
}

func setInt64(m map[string]interface{}, key string, value *int64) {
	if value != nil {
		m[key] = *value
	}
}

func stringMap(m map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func stringSlice(s []string) []interface{} {
	out := make([]interface{}, len(s))
	for i, v := range s {
		out[i] = v
	}
	return out
}
