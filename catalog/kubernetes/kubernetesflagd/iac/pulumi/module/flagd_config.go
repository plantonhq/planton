package module

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	kubernetesflagdv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesflagd/v1alpha1"
)

// FlagdConfig is everything flagd's own configuration renders into.
type FlagdConfig struct {
	// Args are the `flagd start` arguments (every setting except the
	// sources).
	Args []string

	// SourcesJSON is the SourceConfig array flagd reads from FLAGD_SOURCES.
	// It carries authorization headers, so it lives in the module-owned
	// Secret; both engines produce it byte for byte (Go json.Marshal and
	// OpenTofu jsonencode sort keys and escape alike).
	SourcesJSON string

	// SourcesChecksum is the SHA-256 of SourcesJSON, stamped on the pod
	// template so a source change rolls the pods (a variable read from a
	// Secret is fixed at container start).
	SourcesChecksum string

	// ConfigMapMounts are the ConfigMap sources, mounted as directories.
	ConfigMapMounts []ConfigMapMount

	// GrpcCaMounts are the CA certificates gRPC sources verify against.
	GrpcCaMounts []SecretKeyMount

	// FeatureFlagNamespaces are the namespaces FeatureFlag resources are
	// read from (the module grants get/list/watch there).
	FeatureFlagNamespaces []string
}

// ConfigMapMount is one ConfigMap source volume.
type ConfigMapMount struct {
	Volume    string
	ConfigMap string
	MountPath string
}

// SecretKeyMount is one Secret key mounted as a file.
type SecretKeyMount struct {
	Volume    string
	Secret    string
	Key       string
	MountPath string
}

// buildFlagdConfig renders flagd's arguments and sources from the spec.
// Terraform twin: locals.tf (args, sources_json).
func buildFlagdConfig(spec *kubernetesflagdv1alpha1.KubernetesFlagdSpec, namespace string) (*FlagdConfig, error) {
	cfg := &FlagdConfig{}

	port, managementPort, syncPort, ofrepPort := ports(spec)
	args := []string{
		"start",
		"--port=" + strconv.Itoa(port),
		"--management-port=" + strconv.Itoa(managementPort),
		"--sync-port=" + strconv.Itoa(syncPort),
		"--ofrep-port=" + strconv.Itoa(ofrepPort),
	}

	logFormat := "json"
	if spec.GetLog().GetFormat() != "" {
		logFormat = spec.GetLog().GetFormat()
	}
	args = append(args, "--log-format="+logFormat)
	if spec.GetLog().GetDebug() {
		args = append(args, "--debug")
	}

	if spec.GetServer().GetTlsSecretName() != "" {
		args = append(args,
			"--server-cert-path="+vars.ServerTlsMountPath+"/tls.crt",
			"--server-key-path="+vars.ServerTlsMountPath+"/tls.key")
	}

	ev := spec.GetEvaluation()
	for _, k := range sortedKeys(ev.GetContextValues()) {
		args = append(args, "--context-value="+k+"="+ev.GetContextValues()[k])
	}
	for _, k := range sortedKeys(ev.GetContextFromHeader()) {
		args = append(args, "--context-from-header="+k+"="+ev.GetContextFromHeader()[k])
	}
	if len(ev.GetCorsOrigins()) > 0 {
		args = append(args, "--cors-origin="+strings.Join(ev.GetCorsOrigins(), ","))
	}

	if sse := spec.GetOfrepSse(); sse != nil {
		if sse.Enabled != nil {
			args = append(args, "--ofrep-sse-enabled="+strconv.FormatBool(sse.GetEnabled()))
		}
		if sse.InactivityDelaySeconds != nil {
			args = append(args, "--ofrep-sse-inactivity-delay="+strconv.Itoa(int(sse.GetInactivityDelaySeconds())))
		}
		if sse.GetPublicUrl() != "" {
			args = append(args, "--ofrep-sse-public-url="+sse.GetPublicUrl())
		}
	}

	if sync := spec.GetSync(); sync != nil {
		if sync.HttpEnabled != nil {
			args = append(args, "--sync-http-enabled="+strconv.FormatBool(sync.GetHttpEnabled()))
		}
		if sync.GetDisableMetadata() {
			args = append(args, "--disable-sync-metadata")
		}
		if sync.GetStreamDeadline() != "" {
			args = append(args, "--stream-deadline="+sync.GetStreamDeadline())
		}
	}

	if t := spec.GetTelemetry(); t != nil {
		if t.GetMetricsExporter() == "otel" {
			args = append(args, "--metrics-exporter=otel")
		}
		if t.GetOtelCollectorUri() != "" {
			args = append(args, "--otel-collector-uri="+t.GetOtelCollectorUri())
		}
		if ca := t.GetOtelCaCertSecret(); ca != nil {
			args = append(args, "--otel-ca-path="+vars.OtelCaMountPath+"/"+ca.GetKey())
		}
		if t.GetOtelClientTlsSecretName() != "" {
			args = append(args,
				"--otel-cert-path="+vars.OtelTlsMountPath+"/tls.crt",
				"--otel-key-path="+vars.OtelTlsMountPath+"/tls.key")
		}
		if t.GetOtelReloadInterval() != "" {
			args = append(args, "--otel-reload-interval="+t.GetOtelReloadInterval())
		}
	}

	if l := spec.GetLimits(); l != nil {
		if l.MaxRequestBodyBytes != nil {
			args = append(args, "--max-request-body="+strconv.FormatInt(l.GetMaxRequestBodyBytes(), 10))
		}
		if l.MaxRequestHeaderBytes != nil {
			args = append(args, "--max-request-header="+strconv.FormatInt(l.GetMaxRequestHeaderBytes(), 10))
		}
	}
	cfg.Args = args

	sources := []interface{}{}
	ffNamespaces := map[string]bool{}
	for i, s := range spec.GetSources() {
		src := map[string]interface{}{}
		switch {
		case s.GetConfigMap() != nil:
			cm := s.GetConfigMap()
			volume := fmt.Sprintf("source-%d", i)
			mountPath := fmt.Sprintf("%s/%d", vars.SourcesMountRoot, i)
			cfg.ConfigMapMounts = append(cfg.ConfigMapMounts, ConfigMapMount{
				Volume: volume, ConfigMap: cm.GetConfigMapName().GetValue(), MountPath: mountPath,
			})
			provider := "file"
			if cm.GetWatcher() != "" {
				provider = cm.GetWatcher()
			}
			src["uri"] = mountPath + "/" + cm.GetKey().GetValue()
			src["provider"] = provider
		case s.GetHttp() != nil:
			h := s.GetHttp()
			src["uri"] = h.GetUrl()
			src["provider"] = "http"
			if h.GetAuthHeader() != "" {
				src["authHeader"] = h.GetAuthHeader()
			}
			headers, err := mergeHeaders(fmt.Sprintf("sources[%d].http", i), h.GetHeaders(), h.GetSensitiveHeaders())
			if err != nil {
				return nil, err
			}
			if headers != nil {
				src["headers"] = headers
			}
			if h.IntervalSeconds != nil {
				src["interval"] = int(h.GetIntervalSeconds())
			}
			if h.GetIntervalSeed() != "" {
				src["intervalSeed"] = h.GetIntervalSeed()
			}
			if h.TimeoutSeconds != nil {
				src["timeoutS"] = int(h.GetTimeoutSeconds())
			}
			if o := h.GetOauth(); o != nil {
				src["oauth"] = map[string]interface{}{
					"clientID":     o.GetClientId(),
					"clientSecret": o.GetClientSecret(),
					"tokenUrl":     o.GetTokenUrl(),
				}
			}
		case s.GetGrpc() != nil:
			g := s.GetGrpc()
			src["uri"] = g.GetTarget()
			src["provider"] = "grpc"
			if g.GetTls() {
				src["tls"] = true
			}
			if ca := g.GetCaCertSecret(); ca != nil {
				mountPath := fmt.Sprintf("%s/%d", vars.GrpcCaMountRoot, i)
				cfg.GrpcCaMounts = append(cfg.GrpcCaMounts, SecretKeyMount{
					Volume: fmt.Sprintf("grpc-ca-%d", i), Secret: ca.GetName(), Key: ca.GetKey(), MountPath: mountPath,
				})
				src["certPath"] = mountPath + "/" + ca.GetKey()
			}
			if g.GetProviderId() != "" {
				src["providerID"] = g.GetProviderId()
			}
			if g.MaxMsgSize != nil {
				src["maxMsgSize"] = int(g.GetMaxMsgSize())
			}
			if g.GetIncrementalUpdates() {
				src["incrementalUpdates"] = true
			}
			headers, err := mergeHeaders(fmt.Sprintf("sources[%d].grpc", i), g.GetHeaders(), g.GetSensitiveHeaders())
			if err != nil {
				return nil, err
			}
			if headers != nil {
				src["headers"] = headers
			}
			if g.GetSelector() != "" {
				src["selector"] = g.GetSelector()
			}
		case s.GetFeatureFlag() != nil:
			ff := s.GetFeatureFlag()
			ns := ff.GetNamespace()
			if ns == "" {
				ns = namespace
			}
			ffNamespaces[ns] = true
			src["uri"] = ns + "/" + ff.GetName()
			src["provider"] = "kubernetes"
		case s.GetGoogleStorage() != nil:
			objectSource(src, "gs", "gcs", s.GetGoogleStorage())
		case s.GetAzureBlob() != nil:
			objectSource(src, "azblob", "azblob", s.GetAzureBlob())
		case s.GetS3() != nil:
			objectSource(src, "s3", "s3", s.GetS3())
		}
		sources = append(sources, src)
	}

	body, err := json.Marshal(sources)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode flagd sources")
	}
	cfg.SourcesJSON = string(body)
	sum := sha256.Sum256(body)
	cfg.SourcesChecksum = hex.EncodeToString(sum[:])

	for ns := range ffNamespaces {
		cfg.FeatureFlagNamespaces = append(cfg.FeatureFlagNamespaces, ns)
	}
	sort.Strings(cfg.FeatureFlagNamespaces)
	return cfg, nil
}

func objectSource(src map[string]interface{}, scheme, provider string, o *kubernetesflagdv1alpha1.KubernetesFlagdObjectSource) {
	src["uri"] = scheme + "://" + o.GetBucket() + "/" + o.GetObject()
	src["provider"] = provider
	if o.IntervalSeconds != nil {
		src["interval"] = int(o.GetIntervalSeconds())
	}
	if o.GetIntervalSeed() != "" {
		src["intervalSeed"] = o.GetIntervalSeed()
	}
}

// mergeHeaders merges plain and credential headers into one map (both land
// in the sources Secret), refusing a header declared both ways.
func mergeHeaders(field string, plain, sensitive map[string]string) (map[string]interface{}, error) {
	if len(plain)+len(sensitive) == 0 {
		return nil, nil
	}
	out := map[string]interface{}{}
	lower := map[string]bool{}
	for k, v := range plain {
		out[k] = v
		lower[strings.ToLower(k)] = true
	}
	for k, v := range sensitive {
		if lower[strings.ToLower(k)] {
			return nil, errors.Errorf("%s: header %q is declared both plain and sensitive; declare it once", field, k)
		}
		out[k] = v
	}
	return out, nil
}

func ports(spec *kubernetesflagdv1alpha1.KubernetesFlagdSpec) (int, int, int, int) {
	s := spec.GetServer()
	pick := func(set bool, v int32, def int) int {
		if set {
			return int(v)
		}
		return def
	}
	return pick(s != nil && s.Port != nil, s.GetPort(), vars.DefaultPort),
		pick(s != nil && s.ManagementPort != nil, s.GetManagementPort(), vars.DefaultManagementPort),
		pick(s != nil && s.SyncPort != nil, s.GetSyncPort(), vars.DefaultSyncPort),
		pick(s != nil && s.OfrepPort != nil, s.GetOfrepPort(), vars.DefaultOfrepPort)
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
