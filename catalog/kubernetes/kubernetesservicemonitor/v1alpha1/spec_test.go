package kubernetesservicemonitorv1alpha1

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

// The validation suite pins what the spec refuses before a ServiceMonitor
// reaches the operator, which otherwise skips the whole monitor in silence:
// two authentication methods on one endpoint, a client certificate without
// its key (or a part taken from both a Secret and a file), the proxy
// combinations upstream rejects, OAuth2 without a client id, an authorization
// of type Basic, relabeling steps that break their action's rules, a
// target_port that is neither a port number nor a port name, label selector
// requirements whose values do not fit their operator, and a monitor with no
// endpoint. It also pins the Kubernetes label grammar on the object's own
// labels and the reference forms of every Secret, ConfigMap and namespace.

func TestKubernetesServiceMonitor(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesServiceMonitor Suite")
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

func fromConfigMap(name, key string) *kubernetes.KubernetesPrometheusOperatorApiSecretOrConfigMap {
	return &kubernetes.KubernetesPrometheusOperatorApiSecretOrConfigMap{
		ConfigMap: &kubernetes.KubernetesPrometheusOperatorApiConfigMapKeySelector{Name: literal(name), Key: key},
	}
}

func bearer() *kubernetes.KubernetesPrometheusOperatorApiSafeAuthorization {
	return &kubernetes.KubernetesPrometheusOperatorApiSafeAuthorization{Credentials: secretKey("api-scrape-token", "token")}
}

func oauth2() *kubernetes.KubernetesPrometheusOperatorApiOAuth2 {
	return &kubernetes.KubernetesPrometheusOperatorApiOAuth2{
		ClientId:     fromConfigMap("oauth-client", "id"),
		ClientSecret: secretKey("oauth-client", "secret"),
		TokenUrl:     "https://issuer.example.com/oauth2/token",
	}
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

var _ = ginkgo.Describe("KubernetesServiceMonitor Validation Tests", func() {
	var input *KubernetesServiceMonitor
	var endpoint *KubernetesServiceMonitorEndpoint

	ginkgo.BeforeEach(func() {
		endpoint = &KubernetesServiceMonitorEndpoint{
			Port:     "http-metrics",
			Interval: "30s",
		}
		input = &KubernetesServiceMonitor{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesServiceMonitor",
			Metadata:   &shared.CatalogObjectMetadata{Name: "api"},
			Spec: &KubernetesServiceMonitorSpec{
				Namespace: literal("api"),
				Selector: &kubernetes.KubernetesPrometheusOperatorApiLabelSelector{
					MatchLabels: map[string]string{"app.kubernetes.io/name": "api"},
				},
				Endpoints: []*KubernetesServiceMonitorEndpoint{endpoint},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts one named port on a labelled Service", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an empty selector, which selects every Service", func() {
			input.Spec.Selector = &kubernetes.KubernetesPrometheusOperatorApiLabelSelector{}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts the namespace, the watched namespaces and every credential as references", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "api-ns", "spec.name")
			input.Spec.NamespaceSelector = &kubernetes.KubernetesPrometheusOperatorApiNamespaceSelector{
				MatchNames: []*foreignkeyv1.StringValueOrRef{
					valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "api-ns", "spec.name"),
					literal("workers"),
				},
			}
			endpoint.Authorization = &kubernetes.KubernetesPrometheusOperatorApiSafeAuthorization{
				Credentials: &kubernetes.KubernetesPrometheusOperatorApiSecretKeySelector{
					Name: valueFrom(catalogkind.CatalogKind_KubernetesSecret, "api-scrape-token", "status.outputs.secret_name"),
					Key:  "token",
				},
			}
			endpoint.TlsConfig = &kubernetes.KubernetesPrometheusOperatorApiTlsConfig{
				Ca: &kubernetes.KubernetesPrometheusOperatorApiSecretOrConfigMap{
					ConfigMap: &kubernetes.KubernetesPrometheusOperatorApiConfigMapKeySelector{
						Name: valueFrom(catalogkind.CatalogKind_KubernetesConfigMap, "internal-ca", "status.outputs.configmap_name"),
						Key:  "ca.crt",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts every top-level setting at its upstream shape", func() {
			s := input.Spec
			s.Labels = map[string]string{"release": "management-hub-metrics"}
			s.Annotations = map[string]string{"planton.ai/owner": "platform"}
			s.JobLabel = "app.kubernetes.io/name"
			s.TargetLabels = []string{"team"}
			s.PodTargetLabels = []string{"app.kubernetes.io/version"}
			s.SelectorMechanism = ptr("RoleSelector")
			s.SampleLimit = ptr(uint32(0))
			s.TargetLimit = ptr(uint32(50))
			s.ScrapeProtocols = []string{"PrometheusProto", "OpenMetricsText1.0.0", "PrometheusText0.0.4"}
			s.FallbackScrapeProtocol = ptr("PrometheusText0.0.4")
			s.LabelLimit = ptr(uint32(64))
			s.LabelNameLengthLimit = ptr(uint32(128))
			s.LabelValueLengthLimit = ptr(uint32(2048))
			s.ScrapeNativeHistograms = ptr(true)
			s.ScrapeClassicHistograms = ptr(false)
			s.NativeHistogramBucketLimit = ptr(uint32(160))
			s.NativeHistogramMinBucketFactor = ptr("1.1")
			s.ConvertClassicHistogramsToNhcb = ptr(true)
			s.KeepDroppedTargets = ptr(uint32(0))
			s.AttachMetadata = &kubernetes.KubernetesPrometheusOperatorApiAttachMetadata{Node: ptr(true)}
			s.ScrapeClass = ptr("tls")
			s.BodySizeLimit = ptr("10MiB")
			s.ServiceDiscoveryRole = ptr("EndpointSlice")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts every endpoint setting at its upstream shape", func() {
			endpoint.Port = ""
			endpoint.TargetPort = ptr("9090")
			endpoint.Path = "/stats/prometheus"
			endpoint.Scheme = ptr("HTTPS")
			endpoint.Params = map[string]*kubernetes.KubernetesPrometheusOperatorApiStringList{"format": {Values: []string{"prometheus"}}}
			endpoint.ScrapeTimeout = "10s"
			endpoint.TlsConfig = &kubernetes.KubernetesPrometheusOperatorApiTlsConfig{
				CaFile:     "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
				ServerName: ptr("api.api.svc"),
				MinVersion: ptr("TLS12"),
				MaxVersion: ptr("TLS13"),
			}
			endpoint.Authorization = bearer()
			endpoint.HonorLabels = true
			endpoint.HonorTimestamps = ptr(false)
			endpoint.TrackTimestampsStaleness = ptr(true)
			endpoint.MetricRelabelings = []*kubernetes.KubernetesPrometheusOperatorApiRelabelConfig{
				{SourceLabels: []string{"__name__"}, Regex: "go_gc_.*", Action: "drop"},
			}
			endpoint.Relabelings = []*kubernetes.KubernetesPrometheusOperatorApiRelabelConfig{
				{SourceLabels: []string{"__meta_kubernetes_pod_node_name"}, TargetLabel: "node"},
				{SourceLabels: []string{"__address__"}, TargetLabel: "__tmp_hash", Modulus: ptr(uint32(4)), Action: "HashMod"},
				{SourceLabels: []string{"__tmp_hash"}, Regex: "0", Action: "keep"},
				{Regex: "__meta_kubernetes_service_label_(.+)", Action: "labelmap", Replacement: ptr("svc_$1")},
				{Regex: "pod_template_hash", Action: "labeldrop"},
				{SourceLabels: []string{"instance"}, TargetLabel: "node", Action: "keepequal"},
				{SourceLabels: []string{"__name__"}, TargetLabel: "__name__", Action: "lowercase"},
			}
			endpoint.ProxyUrl = ptr("http://proxy.internal:3128")
			endpoint.NoProxy = ptr("10.0.0.0/8,.svc")
			endpoint.ProxyConnectHeader = map[string]*kubernetes.KubernetesPrometheusOperatorApiSecretKeySelectorList{
				"Proxy-Authorization": {Values: []*kubernetes.KubernetesPrometheusOperatorApiSecretKeySelector{secretKey("proxy", "header")}},
			}
			endpoint.FollowRedirects = ptr(false)
			endpoint.EnableHttp2 = ptr(false)
			endpoint.FilterRunning = ptr(false)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a port name as target_port", func() {
			endpoint.TargetPort = ptr("metrics")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts basic auth, OAuth2 and the deprecated bearer forms one at a time", func() {
			endpoint.BasicAuth = &kubernetes.KubernetesPrometheusOperatorApiBasicAuth{
				Username: secretKey("api-basic", "username"),
				Password: secretKey("api-basic", "password"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())

			endpoint.BasicAuth = nil
			endpoint.Oauth2 = oauth2()
			endpoint.Oauth2.Scopes = []string{"metrics.read"}
			endpoint.Oauth2.EndpointParams = map[string]string{"audience": "api"}
			endpoint.Oauth2.TlsConfig = &kubernetes.KubernetesPrometheusOperatorApiSafeTlsConfig{Ca: fromConfigMap("internal-ca", "ca.crt")}
			endpoint.Oauth2.ProxyFromEnvironment = ptr(true)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())

			endpoint.Oauth2 = nil
			endpoint.BearerTokenSecret = secretKey("api-scrape-token", "token")
			endpoint.BearerTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts mutual TLS from Secrets and from files", func() {
			endpoint.TlsConfig = &kubernetes.KubernetesPrometheusOperatorApiTlsConfig{
				Ca:        fromSecret("api-mtls", "ca.crt"),
				Cert:      fromSecret("api-mtls", "tls.crt"),
				KeySecret: secretKey("api-mtls", "tls.key"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())

			endpoint.TlsConfig = &kubernetes.KubernetesPrometheusOperatorApiTlsConfig{
				CaFile:   "/etc/prometheus/certs/ca.crt",
				CertFile: "/etc/prometheus/certs/tls.crt",
				KeyFile:  "/etc/prometheus/certs/tls.key",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts set-based selector requirements", func() {
			input.Spec.Selector.MatchExpressions = []*kubernetes.KubernetesPrometheusOperatorApiLabelSelectorRequirement{
				{Key: "tier", Operator: "In", Values: []string{"api", "worker"}},
				{Key: "canary", Operator: "DoesNotExist"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When the envelope or the monitor's shape is invalid", func() {
		ginkgo.It("refuses a missing namespace", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a missing selector", func() {
			input.Spec.Selector = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a monitor with no endpoint", func() {
			input.Spec.Endpoints = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an object label key that is not a Kubernetes qualified name", func() {
			input.Spec.Labels = map[string]string{"not a key": "x"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an object label value longer than 63 characters", func() {
			input.Spec.Labels = map[string]string{"release": strings.Repeat("a", 64)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an unknown scrape protocol, a repeated one, and an unknown fallback", func() {
			input.Spec.ScrapeProtocols = []string{"PrometheusProto", "Json"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			input.Spec.ScrapeProtocols = []string{"PrometheusProto", "PrometheusProto"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			input.Spec.ScrapeProtocols = nil
			input.Spec.FallbackScrapeProtocol = ptr("Json")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an unknown selector mechanism or discovery role", func() {
			input.Spec.SelectorMechanism = ptr("Labels")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			input.Spec.SelectorMechanism = nil
			input.Spec.ServiceDiscoveryRole = ptr("Pod")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a malformed body size limit and bucket factor, and an empty scrape class", func() {
			input.Spec.BodySizeLimit = ptr("10MB/s")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			input.Spec.BodySizeLimit = nil
			input.Spec.NativeHistogramMinBucketFactor = ptr("fast")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			input.Spec.NativeHistogramMinBucketFactor = nil
			input.Spec.ScrapeClass = ptr("")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses selector requirements whose values do not fit the operator", func() {
			input.Spec.Selector.MatchExpressions = []*kubernetes.KubernetesPrometheusOperatorApiLabelSelectorRequirement{
				{Key: "tier", Operator: "In"},
			}
			expectViolation(protovalidate.Validate(input), "label_selector_requirement.values_by_operator")
			input.Spec.Selector.MatchExpressions[0] = &kubernetes.KubernetesPrometheusOperatorApiLabelSelectorRequirement{
				Key: "canary", Operator: "Exists", Values: []string{"true"},
			}
			expectViolation(protovalidate.Validate(input), "label_selector_requirement.values_by_operator")
			input.Spec.Selector.MatchExpressions[0] = &kubernetes.KubernetesPrometheusOperatorApiLabelSelectorRequirement{
				Key: "tier", Operator: "Equals", Values: []string{"api"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When an endpoint's address or timing is invalid", func() {
		ginkgo.It("refuses a target_port out of range, a name without a letter, and a name too long", func() {
			for _, bad := range []string{"0", "65536", "123456", "Metrics", "-metrics", "metrics-port-name-long"} {
				endpoint.TargetPort = ptr(bad)
				expectViolation(protovalidate.Validate(input), "endpoint.target_port")
			}
		})

		ginkgo.It("refuses an unknown scheme and malformed durations", func() {
			endpoint.Scheme = ptr("grpc")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			endpoint.Scheme = nil
			endpoint.Interval = "30 seconds"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			endpoint.Interval = "30s"
			endpoint.ScrapeTimeout = "10sec"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When an endpoint's authentication is invalid", func() {
		ginkgo.It("refuses two authentication methods on one endpoint", func() {
			endpoint.Authorization = bearer()
			endpoint.BasicAuth = &kubernetes.KubernetesPrometheusOperatorApiBasicAuth{Password: secretKey("api-basic", "password")}
			expectViolation(protovalidate.Validate(input), "endpoint.one_auth_method")

			endpoint.BasicAuth = nil
			endpoint.Oauth2 = oauth2()
			expectViolation(protovalidate.Validate(input), "endpoint.one_auth_method")

			endpoint.Oauth2 = nil
			endpoint.BearerTokenSecret = secretKey("api-scrape-token", "token")
			expectViolation(protovalidate.Validate(input), "endpoint.one_auth_method")
		})

		ginkgo.It("refuses an authorization without credentials, and of type Basic", func() {
			endpoint.Authorization = &kubernetes.KubernetesPrometheusOperatorApiSafeAuthorization{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			endpoint.Authorization = bearer()
			endpoint.Authorization.Type = ptr("basic")
			expectViolation(protovalidate.Validate(input), "safe_authorization.type_not_basic")
		})

		ginkgo.It("refuses a Secret selector without a name or with a malformed key", func() {
			endpoint.Authorization = bearer()
			endpoint.Authorization.Credentials.Name = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			endpoint.Authorization = bearer()
			endpoint.Authorization.Credentials.Key = "a key"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses OAuth2 without a client id source, a client secret or a token URL", func() {
			endpoint.Oauth2 = oauth2()
			endpoint.Oauth2.ClientId = &kubernetes.KubernetesPrometheusOperatorApiSecretOrConfigMap{}
			expectViolation(protovalidate.Validate(input), "oauth2.client_id_source")
			endpoint.Oauth2 = oauth2()
			endpoint.Oauth2.ClientSecret = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			endpoint.Oauth2 = oauth2()
			endpoint.Oauth2.TokenUrl = "issuer.example.com/token"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a value taken from both a Secret and a ConfigMap", func() {
			endpoint.Oauth2 = oauth2()
			endpoint.Oauth2.ClientId.Secret = secretKey("oauth-client", "id")
			expectViolation(protovalidate.Validate(input), "secret_or_config_map.one_source")
		})
	})

	ginkgo.Describe("When an endpoint's TLS is invalid", func() {
		ginkgo.It("refuses a client certificate without its key, and a key without a certificate", func() {
			endpoint.TlsConfig = &kubernetes.KubernetesPrometheusOperatorApiTlsConfig{Cert: fromSecret("api-mtls", "tls.crt")}
			expectViolation(protovalidate.Validate(input), "tls_config.cert_with_key")
			endpoint.TlsConfig = &kubernetes.KubernetesPrometheusOperatorApiTlsConfig{KeyFile: "/etc/prometheus/certs/tls.key"}
			expectViolation(protovalidate.Validate(input), "tls_config.cert_with_key")
		})

		ginkgo.It("refuses a part taken from both a Secret and a file", func() {
			endpoint.TlsConfig = &kubernetes.KubernetesPrometheusOperatorApiTlsConfig{Ca: fromSecret("api-mtls", "ca.crt"), CaFile: "/etc/ca.crt"}
			expectViolation(protovalidate.Validate(input), "tls_config.one_ca_source")
			endpoint.TlsConfig = &kubernetes.KubernetesPrometheusOperatorApiTlsConfig{
				Cert: fromSecret("api-mtls", "tls.crt"), CertFile: "/etc/tls.crt", KeyFile: "/etc/tls.key",
			}
			expectViolation(protovalidate.Validate(input), "tls_config.one_cert_source")
			endpoint.TlsConfig = &kubernetes.KubernetesPrometheusOperatorApiTlsConfig{
				CertFile: "/etc/tls.crt", KeySecret: secretKey("api-mtls", "tls.key"), KeyFile: "/etc/tls.key",
			}
			expectViolation(protovalidate.Validate(input), "tls_config.one_key_source")
		})

		ginkgo.It("refuses a maximum TLS version below the minimum, and an unknown version", func() {
			endpoint.TlsConfig = &kubernetes.KubernetesPrometheusOperatorApiTlsConfig{MinVersion: ptr("TLS13"), MaxVersion: ptr("TLS12")}
			expectViolation(protovalidate.Validate(input), "tls_config.version_order")
			endpoint.TlsConfig = &kubernetes.KubernetesPrometheusOperatorApiTlsConfig{MinVersion: ptr("TLS1.2")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an OAuth2 token request's client certificate without its key", func() {
			endpoint.Oauth2 = oauth2()
			endpoint.Oauth2.TlsConfig = &kubernetes.KubernetesPrometheusOperatorApiSafeTlsConfig{Cert: fromSecret("oauth-mtls", "tls.crt")}
			expectViolation(protovalidate.Validate(input), "safe_tls_config.cert_with_key")
		})
	})

	ginkgo.Describe("When an endpoint's proxy is invalid", func() {
		ginkgo.It("refuses CONNECT headers without a proxy", func() {
			endpoint.ProxyConnectHeader = map[string]*kubernetes.KubernetesPrometheusOperatorApiSecretKeySelectorList{
				"Proxy-Authorization": {Values: []*kubernetes.KubernetesPrometheusOperatorApiSecretKeySelector{secretKey("proxy", "header")}},
			}
			expectViolation(protovalidate.Validate(input), "endpoint.proxy_connect_header_needs_proxy")
		})

		ginkgo.It("refuses a CONNECT header with no value", func() {
			endpoint.ProxyUrl = ptr("http://proxy.internal:3128")
			endpoint.ProxyConnectHeader = map[string]*kubernetes.KubernetesPrometheusOperatorApiSecretKeySelectorList{
				"Proxy-Authorization": {},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses the environment's proxy beside an explicit proxy or bypass list", func() {
			endpoint.ProxyFromEnvironment = ptr(true)
			endpoint.ProxyUrl = ptr("http://proxy.internal:3128")
			expectViolation(protovalidate.Validate(input), "endpoint.proxy_from_environment_alone")
		})

		ginkgo.It("refuses a bypass list without a proxy, and a proxy URL without a scheme", func() {
			endpoint.NoProxy = ptr(".svc")
			expectViolation(protovalidate.Validate(input), "endpoint.no_proxy_needs_proxy_url")
			endpoint.NoProxy = nil
			endpoint.ProxyUrl = ptr("proxy.internal:3128")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("holds an OAuth2 token request's proxy to the same rules", func() {
			endpoint.Oauth2 = oauth2()
			endpoint.Oauth2.NoProxy = ptr(".svc")
			expectViolation(protovalidate.Validate(input), "oauth2.no_proxy_needs_proxy_url")
		})
	})

	ginkgo.Describe("When a relabeling step breaks its action's rules", func() {
		relabel := func(rc *kubernetes.KubernetesPrometheusOperatorApiRelabelConfig) {
			endpoint.Relabelings = []*kubernetes.KubernetesPrometheusOperatorApiRelabelConfig{rc}
		}

		ginkgo.It("refuses replace (also as the default action) without target_label", func() {
			relabel(&kubernetes.KubernetesPrometheusOperatorApiRelabelConfig{SourceLabels: []string{"__address__"}})
			expectViolation(protovalidate.Validate(input), "relabel_config.target_label_required")
			relabel(&kubernetes.KubernetesPrometheusOperatorApiRelabelConfig{SourceLabels: []string{"__address__"}, Action: "Replace"})
			expectViolation(protovalidate.Validate(input), "relabel_config.target_label_required")
		})

		ginkgo.It("refuses hashmod without a modulus", func() {
			relabel(&kubernetes.KubernetesPrometheusOperatorApiRelabelConfig{SourceLabels: []string{"__address__"}, TargetLabel: "__tmp_hash", Action: "hashmod"})
			expectViolation(protovalidate.Validate(input), "relabel_config.hashmod_modulus")
		})

		ginkgo.It("refuses a replacement on lowercase", func() {
			relabel(&kubernetes.KubernetesPrometheusOperatorApiRelabelConfig{SourceLabels: []string{"job"}, TargetLabel: "job", Action: "lowercase", Replacement: ptr("x")})
			expectViolation(protovalidate.Validate(input), "relabel_config.no_replacement")
		})

		ginkgo.It("refuses a regex on keepequal", func() {
			relabel(&kubernetes.KubernetesPrometheusOperatorApiRelabelConfig{SourceLabels: []string{"instance"}, TargetLabel: "node", Regex: "a.*", Action: "keepequal"})
			expectViolation(protovalidate.Validate(input), "relabel_config.equal_actions_fields")
		})

		ginkgo.It("refuses source labels on labeldrop", func() {
			relabel(&kubernetes.KubernetesPrometheusOperatorApiRelabelConfig{SourceLabels: []string{"job"}, Regex: "tmp_.*", Action: "labeldrop"})
			expectViolation(protovalidate.Validate(input), "relabel_config.label_drop_keep_fields")
		})

		ginkgo.It("refuses an action upstream does not know", func() {
			relabel(&kubernetes.KubernetesPrometheusOperatorApiRelabelConfig{SourceLabels: []string{"job"}, TargetLabel: "job", Action: "rename"})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
