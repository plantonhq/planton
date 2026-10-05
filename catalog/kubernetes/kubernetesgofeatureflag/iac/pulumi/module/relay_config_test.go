package module

import (
	"encoding/json"
	"strings"
	"testing"

	kubernetesgofeatureflagv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgofeatureflag/v1alpha1"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func lit(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func cmRetriever(name, ns string) *kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRetriever {
	r := &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagConfigMapRetriever{ConfigMapName: lit(name), Key: lit("flags.goff.yaml")}
	if ns != "" {
		r.Namespace = lit(ns)
	}
	return &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRetriever{
		Kind: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRetriever_ConfigMap{ConfigMap: r},
	}
}

// parse strips the monitoring-port line and decodes the JSON body.
func parse(t *testing.T, doc string) map[string]interface{} {
	t.Helper()
	lines := strings.SplitN(doc, "\n", 2)
	if len(lines) != 2 || lines[0] != "# server.monitoringPort: 1032" {
		t.Fatalf("document must open with the monitoring-port line the chart scans for, got %q", lines[0])
	}
	body := strings.ReplaceAll(lines[1], `{{"{{"}}`, "{{")
	out := map[string]interface{}{}
	if err := json.Unmarshal([]byte(body), &out); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	return out
}

func flagSourceSpec(src *kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagFlagSource) *kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagSpec {
	return &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagSpec{
		Namespace: lit("flags"),
		Mode:      &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagSpec_FlagSource{FlagSource: src},
	}
}

func TestMinimalConfigServesHTTPWithPrefixedEnvAndOneGrant(t *testing.T) {
	cfg, err := buildRelayConfig(flagSourceSpec(&kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagFlagSource{
		Retrievers: []*kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRetriever{cmRetriever("release-flags", "")},
	}), "flags")
	if err != nil {
		t.Fatal(err)
	}
	doc := parse(t, cfg.Document)
	server := doc["server"].(map[string]interface{})
	if server["mode"] != "http" || server["port"].(float64) != 1031 || server["monitoringPort"].(float64) != 1032 {
		t.Fatalf("server block wrong: %v", server)
	}
	if doc["envVariablePrefix"] != "GOFFRELAY_" {
		t.Fatalf("the env prefix must default so service-link variables are never read as configuration, got %v", doc["envVariablePrefix"])
	}
	if doc["logLevel"] != "info" || doc["logFormat"] != "json" {
		t.Fatalf("log defaults wrong: %v %v", doc["logLevel"], doc["logFormat"])
	}
	ret := doc["retrievers"].([]interface{})[0].(map[string]interface{})
	if ret["kind"] != "configmap" || ret["namespace"] != "flags" || ret["configmap"] != "release-flags" || ret["key"] != "flags.goff.yaml" {
		t.Fatalf("configmap retriever wrong: %v", ret)
	}
	if len(cfg.SecretEnv) != 0 {
		t.Fatalf("no secret declared, yet env holds %v", cfg.SecretEnv)
	}
	if got := cfg.ConfigMapGrants["flags"]; len(got) != 1 || got[0] != "release-flags" {
		t.Fatalf("grants wrong: %v", cfg.ConfigMapGrants)
	}
}

func TestDefaultedKeysKeepTheirCamelCaseSoTheFileOverridesTheRelayDefaults(t *testing.T) {
	interval := int32(15000)
	start := true
	cfg, err := buildRelayConfig(flagSourceSpec(&kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagFlagSource{
		Retrievers:              []*kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRetriever{cmRetriever("f", "")},
		PollingIntervalMs:       &interval,
		FileFormat:              "json",
		StartWithRetrieverError: &start,
	}), "flags")
	if err != nil {
		t.Fatal(err)
	}
	doc := parse(t, cfg.Document)
	for _, k := range []string{"pollingInterval", "fileFormat", "logLevel", "startWithRetrieverError", "envVariablePrefix"} {
		if _, ok := doc[k]; !ok {
			t.Errorf("key %q must be spelled exactly so (the relay pre-loads or looks it up by that spelling); document keys: %v", k, keys(doc))
		}
	}
}

func TestEverySecretReachesItsIndexedVariableAndNeverTheFile(t *testing.T) {
	timeout := int32(5000)
	kafkaConfig, _ := structpb.NewStruct(map[string]interface{}{"Net": map[string]interface{}{"SASL": map[string]interface{}{"Enable": true, "User": "events"}}})
	src := &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagFlagSource{
		Retrievers: []*kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRetriever{
			cmRetriever("f", "other-ns"),
			{Kind: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRetriever_Http{Http: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagHttpRetriever{
				Url: "https://x.example.com/f.yaml", Headers: map[string]string{"Accept": "text/yaml"},
				SensitiveHeaders: map[string]string{"Authorization": "Bearer http-secret"}, TimeoutMs: &timeout,
			}}},
			{Kind: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRetriever_Github{Github: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagGitRetriever{
				RepositorySlug: "acme/flags", Path: "f.yaml", Token: "gh-secret",
			}}},
			{Kind: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRetriever_Redis{Redis: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRedisRetriever{
				Options: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRedisOptions{Addr: "redis:6379", Password: "redis-secret"},
			}}},
			{Kind: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRetriever_Postgresql{Postgresql: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagPostgresRetriever{
				Uri: "postgres://u:pg-secret@pg/flags", Table: "flags",
			}}},
		},
		Notifiers: []*kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagNotifier{
			{Kind: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagNotifier_Slack{Slack: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagWebhookUrlNotifier{WebhookUrl: "https://hooks.slack.com/slack-secret"}}},
			{Kind: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagNotifier_Webhook{Webhook: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagWebhookNotifier{
				EndpointUrl: "https://ops.example.com/hook", Secret: "hmac-secret", SensitiveHeaders: map[string]string{"X-Api-Key": "hdr-secret"},
			}}},
		},
		Exporters: []*kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagExporter{
			{Kind: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagExporter_Bigquery{Bigquery: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagBigQueryExporter{ProjectId: "p", DatasetId: "d", GoogleCredentials: "bq-secret"}}},
			{Kind: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagExporter_Kafka{Kafka: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagKafkaExporter{
				Topic: "t", Addresses: []string{"k:9092"}, Config: kafkaConfig, SaslPassword: "kafka-secret",
			}}},
			{Kind: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagExporter_Log{Log: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagLogExporter{LogFormat: "{{ .Key}}"}}},
		},
	}
	spec := flagSourceSpec(src)
	spec.AuthorizedKeys = &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagAuthorizedKeys{Admin: []string{"a1", "a2"}, Evaluation: []string{"e1"}}

	cfg, err := buildRelayConfig(spec, "flags")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"GOFFRELAY_AUTHORIZEDKEYS_ADMIN":                       "a1,a2",
		"GOFFRELAY_AUTHORIZEDKEYS_EVALUATION":                  "e1",
		"GOFFRELAY_RETRIEVERS_1_HEADERS_AUTHORIZATION":         "Bearer http-secret",
		"GOFFRELAY_RETRIEVERS_2_TOKEN":                         "gh-secret",
		"GOFFRELAY_RETRIEVERS_3_REDISOPTIONS_PASSWORD":         "redis-secret",
		"GOFFRELAY_RETRIEVERS_4_URI":                           "postgres://u:pg-secret@pg/flags",
		"GOFFRELAY_NOTIFIERS_0_WEBHOOKURL":                     "https://hooks.slack.com/slack-secret",
		"GOFFRELAY_NOTIFIERS_1_SECRET":                         "hmac-secret",
		"GOFFRELAY_NOTIFIERS_1_HEADERS_X-API-KEY":              "hdr-secret",
		"GOFFRELAY_EXPORTERS_0_GOOGLECREDENTIALS":              "bq-secret",
		"GOFFRELAY_EXPORTERS_1_KAFKA_CONFIG_NET_SASL_PASSWORD": "kafka-secret",
	}
	for k, v := range want {
		if cfg.SecretEnv[k] != v {
			t.Errorf("secret env %s = %q, want %q", k, cfg.SecretEnv[k], v)
		}
	}
	if len(cfg.SecretEnv) != len(want) {
		t.Errorf("secret env holds %d variables, want %d: %v", len(cfg.SecretEnv), len(want), cfg.SecretEnv)
	}
	for _, v := range want {
		if strings.Contains(cfg.Document, v) {
			t.Errorf("secret %q leaked into the configuration document", v)
		}
	}

	doc := parse(t, cfg.Document)
	retrievers := doc["retrievers"].([]interface{})
	if retrievers[0].(map[string]interface{})["namespace"] != "other-ns" {
		t.Errorf("a ConfigMap retriever keeps its own namespace")
	}
	if _, ok := retrievers[3].(map[string]interface{})["redisoptions"]; !ok {
		t.Errorf("redis options must be spelled `redisoptions`, the key the password variable writes into")
	}
	if h := retrievers[1].(map[string]interface{})["headers"].(map[string]interface{}); h["Accept"] == nil || h["Authorization"] != nil {
		t.Errorf("plain headers render, sensitive ones do not: %v", h)
	}
	kafka := doc["exporters"].([]interface{})[1].(map[string]interface{})["kafka"].(map[string]interface{})
	if _, ok := kafka["config"].(map[string]interface{})["net"].(map[string]interface{})["sasl"]; !ok {
		t.Errorf("Kafka config keys must be lowercased so the password variable merges into them: %v", kafka["config"])
	}
	if !strings.Contains(cfg.Document, `{{"{{"}} .Key}}`) {
		t.Errorf("a Go template in the configuration must be escaped for the chart's tpl pass")
	}
	if got := cfg.ConfigMapGrants["other-ns"]; len(got) != 1 || got[0] != "f" {
		t.Errorf("the grant follows the ConfigMap's namespace: %v", cfg.ConfigMapGrants)
	}
}

func TestFlagSetsIndexTheirSecretsAndRefuseSharedKeys(t *testing.T) {
	token := &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRetriever{Kind: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRetriever_Gitlab{
		Gitlab: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagGitRetriever{RepositorySlug: "a/b", Path: "f.yaml", Token: "gl-secret"},
	}}
	spec := &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagSpec{
		Namespace: lit("flags"),
		Mode: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagSpec_FlagSets{FlagSets: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagFlagSets{
			Items: []*kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagFlagSet{
				{Name: "web", ApiKeys: []string{"w1", "w2"}, Source: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagFlagSource{
					Retrievers: []*kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRetriever{cmRetriever("web-flags", "")},
				}},
				{Name: "mobile", ApiKeys: []string{"m1"}, Source: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagFlagSource{
					Retrievers: []*kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRetriever{cmRetriever("mobile-flags", ""), token},
				}},
			},
		}},
	}
	cfg, err := buildRelayConfig(spec, "flags")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SecretEnv["GOFFRELAY_FLAGSETS_0_APIKEYS"] != "w1,w2" || cfg.SecretEnv["GOFFRELAY_FLAGSETS_1_APIKEYS"] != "m1" ||
		cfg.SecretEnv["GOFFRELAY_FLAGSETS_1_RETRIEVERS_1_TOKEN"] != "gl-secret" {
		t.Fatalf("flag-set secrets wrong: %v", cfg.SecretEnv)
	}
	doc := parse(t, cfg.Document)
	if _, ok := doc["retrievers"]; ok {
		t.Fatalf("flag_sets mode must not render top-level retrievers (the relay ignores them beside flag sets)")
	}
	if got := cfg.ConfigMapGrants["flags"]; len(got) != 2 {
		t.Fatalf("both flag sets' ConfigMaps are granted: %v", cfg.ConfigMapGrants)
	}

	spec.GetFlagSets().GetItems()[1].ApiKeys = []string{"w2"}
	if _, err := buildRelayConfig(spec, "flags"); err == nil || !strings.Contains(err.Error(), "share an API key") {
		t.Fatalf("a key selecting two flag sets must be refused, got %v", err)
	}
}

func TestDeployTimeChecksRefuseWhatTheEnvironmentCannotCarry(t *testing.T) {
	base := func() *kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagSpec {
		return flagSourceSpec(&kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagFlagSource{
			Retrievers: []*kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRetriever{cmRetriever("f", "")},
		})
	}

	comma := base()
	comma.AuthorizedKeys = &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagAuthorizedKeys{Evaluation: []string{"a,b"}}
	if _, err := buildRelayConfig(comma, "flags"); err == nil || !strings.Contains(err.Error(), "comma") {
		t.Errorf("a key with a comma must be refused, got %v", err)
	}

	underscore := base()
	underscore.GetFlagSource().Notifiers = []*kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagNotifier{{Kind: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagNotifier_Webhook{
		Webhook: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagWebhookNotifier{EndpointUrl: "https://x.example.com", SensitiveHeaders: map[string]string{"X_Token": "t"}},
	}}}
	if _, err := buildRelayConfig(underscore, "flags"); err == nil || !strings.Contains(err.Error(), `"_"`) {
		t.Errorf("a sensitive header name with an underscore must be refused, got %v", err)
	}

	twice := base()
	twice.GetFlagSource().Notifiers = []*kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagNotifier{{Kind: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagNotifier_Webhook{
		Webhook: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagWebhookNotifier{EndpointUrl: "https://x.example.com",
			Headers: map[string]string{"authorization": "a"}, SensitiveHeaders: map[string]string{"Authorization": "b"}},
	}}}
	if _, err := buildRelayConfig(twice, "flags"); err == nil || !strings.Contains(err.Error(), "both plain and sensitive") {
		t.Errorf("a header declared plain and sensitive must be refused, got %v", err)
	}
}

func TestCustomPrefixAndPortsFlowThrough(t *testing.T) {
	prefix := "MYRELAY_"
	port, monitoring := int32(8080), int32(9090)
	spec := flagSourceSpec(&kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagFlagSource{
		Retrievers: []*kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRetriever{cmRetriever("f", "")},
	})
	spec.Runtime = &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRuntime{EnvVariablePrefix: &prefix}
	spec.Server = &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagServer{Port: &port, MonitoringPort: &monitoring}
	spec.AuthorizedKeys = &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagAuthorizedKeys{Evaluation: []string{"k"}}
	cfg, err := buildRelayConfig(spec, "flags")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(cfg.Document, "# server.monitoringPort: 9090\n") {
		t.Errorf("the monitoring-port line follows the spec: %q", strings.SplitN(cfg.Document, "\n", 2)[0])
	}
	if cfg.SecretEnv["MYRELAY_AUTHORIZEDKEYS_EVALUATION"] != "k" {
		t.Errorf("secret variables carry the configured prefix: %v", cfg.SecretEnv)
	}
}

func keys(m map[string]interface{}) []string {
	out := []string{}
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestAChangedSecretValueRollsThePods(t *testing.T) {
	values := func(evaluationKey string) map[string]interface{} {
		t.Helper()
		spec := flagSourceSpec(&kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagFlagSource{
			Retrievers: []*kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagRetriever{cmRetriever("release-flags", "")},
		})
		spec.PodAnnotations = map[string]string{"team": "platform"}
		spec.AuthorizedKeys = &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagAuthorizedKeys{Evaluation: []string{evaluationKey}}
		locals, err := initializeLocals(nil, &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlagIacInput{
			Target: &kubernetesgofeatureflagv1alpha1.KubernetesGoFeatureFlag{
				Metadata: &shared.CatalogObjectMetadata{Name: "flags"},
				Spec:     spec,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		v, err := buildHelmValues(locals)
		if err != nil {
			t.Fatal(err)
		}
		return v["podAnnotations"].(map[string]interface{})
	}
	before, after := values("old-key"), values("new-key")
	if before["team"] != "platform" {
		t.Fatalf("the spec's pod annotations must survive: %v", before)
	}
	sum, _ := before["checksum/env-secret"].(string)
	if len(sum) != 64 {
		t.Fatalf("pods must carry the env Secret's sha256, got %q", sum)
	}
	if after["checksum/env-secret"] == sum {
		t.Fatal("a rotated evaluation key must change the pod template, or the running relay keeps the old key")
	}
}
