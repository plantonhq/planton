package kubernetespodmonitorv1alpha1

import (
	"errors"
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

// The validation suite pins what the spec refuses before a PodMonitor reaches
// the operator, which otherwise skips the whole monitor in silence: two
// authentication methods on one endpoint, a client certificate without its
// key, the proxy combinations upstream rejects, a port number out of range, a
// target_port that is neither a port number nor a port name, and a relabeling
// step that breaks its action's rules. It also pins what differs from a
// ServiceMonitor: an empty endpoint list is accepted (upstream does not
// require one), and an endpoint's TLS material comes only from Secrets and
// ConfigMaps. The shared types' every rule is pinned once, by the
// KubernetesServiceMonitor suite.

func TestKubernetesPodMonitor(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesPodMonitor Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valueFrom(kind catalogkind.CatalogKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Kind: kind, Name: name, FieldPath: fieldPath},
		},
	}
}

// ptr returns a pointer to v, for setting proto3 `optional` scalar fields.
func ptr[T any](v T) *T { return &v }

func secretKey(name, key string) *kubernetes.KubernetesPrometheusOperatorApiSecretKeySelector {
	return &kubernetes.KubernetesPrometheusOperatorApiSecretKeySelector{Name: literal(name), Key: key}
}

func fromSecret(name, key string) *kubernetes.KubernetesPrometheusOperatorApiSecretOrConfigMap {
	return &kubernetes.KubernetesPrometheusOperatorApiSecretOrConfigMap{Secret: secretKey(name, key)}
}

// expectViolation asserts that validation failed on the rule with the given
// id, so a test proves which rule refused the input, not merely that one did.
func expectViolation(err error, ruleID string) {
	var verr *protovalidate.ValidationError
	gomega.Expect(errors.As(err, &verr)).To(gomega.BeTrue(), "expected a validation error, got: %v", err)
	ids := make([]string, 0, len(verr.Violations))
	for _, v := range verr.Violations {
		ids = append(ids, v.Proto.GetRuleId())
	}
	gomega.Expect(ids).To(gomega.ContainElement(ruleID), "expected rule %q to refuse the input, got: %v", ruleID, err)
}

var _ = ginkgo.Describe("KubernetesPodMonitor Validation Tests", func() {
	var input *KubernetesPodMonitor
	var endpoint *KubernetesPodMonitorPodMetricsEndpoint

	ginkgo.BeforeEach(func() {
		endpoint = &KubernetesPodMonitorPodMetricsEndpoint{Port: ptr("metrics")}
		input = &KubernetesPodMonitor{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesPodMonitor",
			Metadata:   &shared.CatalogObjectMetadata{Name: "orders-db"},
			Spec: &KubernetesPodMonitorSpec{
				Namespace: literal("orders"),
				Selector: &kubernetes.KubernetesPrometheusOperatorApiLabelSelector{
					MatchLabels: map[string]string{"cnpg.io/cluster": "orders-db"},
				},
				PodMetricsEndpoints: []*KubernetesPodMonitorPodMetricsEndpoint{endpoint},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts one named container port on labelled pods", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a port number, and the deprecated target_port as a number or a name", func() {
			endpoint.Port = nil
			endpoint.PortNumber = ptr(int32(9187))
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			endpoint.PortNumber = nil
			endpoint.TargetPort = ptr("9187")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			endpoint.TargetPort = ptr("metrics")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a monitor with no endpoint, which upstream allows", func() {
			input.Spec.PodMetricsEndpoints = nil
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts the namespace, the watched namespaces and the credentials as references", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "orders-ns", "spec.name")
			input.Spec.NamespaceSelector = &kubernetes.KubernetesPrometheusOperatorApiNamespaceSelector{
				MatchNames: []*foreignkeyv1.StringValueOrRef{valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "orders-ns", "spec.name")},
			}
			endpoint.BasicAuth = &kubernetes.KubernetesPrometheusOperatorApiBasicAuth{
				Username: &kubernetes.KubernetesPrometheusOperatorApiSecretKeySelector{
					Name: valueFrom(catalogkind.CatalogKind_KubernetesSecret, "exporter-auth", "status.outputs.secret_name"),
					Key:  "username",
				},
				Password: secretKey("exporter-auth", "password"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts every top-level setting at its upstream shape", func() {
			s := input.Spec
			s.Labels = map[string]string{"release": "management-hub-metrics"}
			s.Annotations = map[string]string{"planton.ai/owner": "platform"}
			s.JobLabel = "cnpg.io/cluster"
			s.PodTargetLabels = []string{"cnpg.io/instanceRole"}
			s.SelectorMechanism = ptr("RelabelConfig")
			s.NamespaceSelector = &kubernetes.KubernetesPrometheusOperatorApiNamespaceSelector{Any: true}
			s.SampleLimit = ptr(uint32(50000))
			s.TargetLimit = ptr(uint32(10))
			s.ScrapeProtocols = []string{"OpenMetricsText1.0.0", "PrometheusText0.0.4"}
			s.FallbackScrapeProtocol = ptr("PrometheusText0.0.4")
			s.LabelLimit = ptr(uint32(64))
			s.LabelNameLengthLimit = ptr(uint32(128))
			s.LabelValueLengthLimit = ptr(uint32(2048))
			s.ScrapeNativeHistograms = ptr(false)
			s.ScrapeClassicHistograms = ptr(true)
			s.NativeHistogramBucketLimit = ptr(uint32(160))
			s.NativeHistogramMinBucketFactor = ptr("2")
			s.ConvertClassicHistogramsToNhcb = ptr(false)
			s.KeepDroppedTargets = ptr(uint32(100))
			s.AttachMetadata = &kubernetes.KubernetesPrometheusOperatorApiAttachMetadata{Node: ptr(true)}
			s.ScrapeClass = ptr("default")
			s.BodySizeLimit = ptr("0")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts every endpoint setting at its upstream shape", func() {
			endpoint.Path = "/metrics"
			endpoint.Scheme = ptr("https")
			endpoint.Params = map[string]*kubernetes.KubernetesPrometheusOperatorApiStringList{"collect[]": {Values: []string{"pg_stat_database", "pg_replication"}}}
			endpoint.Interval = "1m"
			endpoint.ScrapeTimeout = "20s"
			endpoint.TlsConfig = &kubernetes.KubernetesPrometheusOperatorApiSafeTlsConfig{
				Ca:                 fromSecret("orders-db-ca", "ca.crt"),
				Cert:               fromSecret("orders-db-client", "tls.crt"),
				KeySecret:          secretKey("orders-db-client", "tls.key"),
				ServerName:         ptr("orders-db-rw"),
				InsecureSkipVerify: ptr(false),
				MinVersion:         ptr("TLS12"),
			}
			endpoint.Authorization = &kubernetes.KubernetesPrometheusOperatorApiSafeAuthorization{
				Type:        ptr("Bearer"),
				Credentials: secretKey("exporter-token", "token"),
			}
			endpoint.HonorLabels = true
			endpoint.HonorTimestamps = ptr(true)
			endpoint.TrackTimestampsStaleness = ptr(false)
			endpoint.MetricRelabelings = []*kubernetes.KubernetesPrometheusOperatorApiRelabelConfig{
				{SourceLabels: []string{"__name__"}, Regex: "pg_settings_.*", Action: "Drop"},
			}
			endpoint.Relabelings = []*kubernetes.KubernetesPrometheusOperatorApiRelabelConfig{
				{SourceLabels: []string{"__meta_kubernetes_pod_label_cnpg_io_instanceRole"}, TargetLabel: "role"},
			}
			endpoint.ProxyFromEnvironment = ptr(true)
			endpoint.FollowRedirects = ptr(true)
			endpoint.EnableHttp2 = ptr(true)
			endpoint.FilterRunning = ptr(true)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When the envelope or the monitor's shape is invalid", func() {
		ginkgo.It("refuses a missing namespace or selector", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			input.Spec.Namespace = literal("orders")
			input.Spec.Selector = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an object label value longer than 63 characters", func() {
			input.Spec.Labels = map[string]string{"release": strings.Repeat("a", 64)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When an endpoint is invalid", func() {
		ginkgo.It("refuses a port number out of range", func() {
			endpoint.PortNumber = ptr(int32(0))
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			endpoint.PortNumber = ptr(int32(65536))
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a target_port that is neither a port number nor a port name", func() {
			for _, bad := range []string{"0", "70000", "METRICS", "metrics_port"} {
				endpoint.TargetPort = ptr(bad)
				expectViolation(protovalidate.Validate(input), "pod_metrics_endpoint.target_port")
			}
		})

		ginkgo.It("refuses two authentication methods on one endpoint", func() {
			endpoint.BearerTokenSecret = secretKey("exporter-token", "token")
			endpoint.Authorization = &kubernetes.KubernetesPrometheusOperatorApiSafeAuthorization{Credentials: secretKey("exporter-token", "token")}
			expectViolation(protovalidate.Validate(input), "pod_metrics_endpoint.one_auth_method")
		})

		ginkgo.It("refuses a client certificate without its key", func() {
			endpoint.TlsConfig = &kubernetes.KubernetesPrometheusOperatorApiSafeTlsConfig{Cert: fromSecret("orders-db-client", "tls.crt")}
			expectViolation(protovalidate.Validate(input), "safe_tls_config.cert_with_key")
		})

		ginkgo.It("refuses the proxy combinations upstream rejects", func() {
			endpoint.ProxyFromEnvironment = ptr(true)
			endpoint.NoProxy = ptr(".svc")
			expectViolation(protovalidate.Validate(input), "pod_metrics_endpoint.proxy_from_environment_alone")
			endpoint.ProxyFromEnvironment = nil
			expectViolation(protovalidate.Validate(input), "pod_metrics_endpoint.no_proxy_needs_proxy_url")
			endpoint.NoProxy = nil
			endpoint.ProxyConnectHeader = map[string]*kubernetes.KubernetesPrometheusOperatorApiSecretKeySelectorList{
				"Proxy-Authorization": {Values: []*kubernetes.KubernetesPrometheusOperatorApiSecretKeySelector{secretKey("proxy", "header")}},
			}
			expectViolation(protovalidate.Validate(input), "pod_metrics_endpoint.proxy_connect_header_needs_proxy")
		})

		ginkgo.It("refuses a relabeling step that breaks its action's rules", func() {
			endpoint.Relabelings = []*kubernetes.KubernetesPrometheusOperatorApiRelabelConfig{
				{SourceLabels: []string{"__address__"}, TargetLabel: "__tmp_hash", Action: "HashMod"},
			}
			expectViolation(protovalidate.Validate(input), "relabel_config.hashmod_modulus")
		})
	})
})
