package kubernetesagentsandboxcontrollerv1alpha1

import (
	"errors"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

// The rules pinned here are the spec's whole validation surface: upstream's
// own startup checks (a controller or router that would exit at start is
// refused before deploy), and the settings that would otherwise be silently
// ignored. Every rejection names the rule it expects, so a case cannot pass
// because some other rule happened to fire.

func TestKubernetesAgentSandboxController(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesAgentSandboxController Suite")
}

func int32Ptr(i int32) *int32       { return &i }
func boolPtr(b bool) *bool          { return &b }
func stringPtr(s string) *string    { return &s }
func float64Ptr(f float64) *float64 { return &f }

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func secretKey(name, key string) *KubernetesAgentSandboxControllerSecretKey {
	return &KubernetesAgentSandboxControllerSecretKey{Secret: literal(name), Key: key}
}

func allowAll() *KubernetesAgentSandboxControllerRouterAuthorization {
	return &KubernetesAgentSandboxControllerRouterAuthorization{
		Mode: &KubernetesAgentSandboxControllerRouterAuthorization_AllowAll{
			AllowAll: &KubernetesAgentSandboxControllerRouterAllowAll{},
		},
	}
}

func tokenReview() *KubernetesAgentSandboxControllerRouterAuthorization {
	return &KubernetesAgentSandboxControllerRouterAuthorization{
		Mode: &KubernetesAgentSandboxControllerRouterAuthorization_TokenReview{
			TokenReview: &KubernetesAgentSandboxControllerRouterTokenReview{},
		},
	}
}

func scopedToken(st *KubernetesAgentSandboxControllerRouterScopedToken) *KubernetesAgentSandboxControllerRouterAuthorization {
	return &KubernetesAgentSandboxControllerRouterAuthorization{
		Mode: &KubernetesAgentSandboxControllerRouterAuthorization_ScopedToken{ScopedToken: st},
	}
}

var _ = ginkgo.Describe("KubernetesAgentSandboxController Validation Tests", func() {
	var input *KubernetesAgentSandboxController

	ginkgo.BeforeEach(func() {
		input = &KubernetesAgentSandboxController{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesAgentSandboxController",
			Metadata:   &shared.CatalogObjectMetadata{Name: "agent-sandbox"},
			Spec:       &KubernetesAgentSandboxControllerSpec{},
		}
	})

	// Cases that are about something else get upstream's namespace, the one
	// value it may take, filled in.
	withNamespace := func() {
		if input.Spec != nil && input.Spec.Namespace == nil {
			input.Spec.Namespace = literal("agent-sandbox-system")
		}
	}
	valid := func() {
		ginkgo.GinkgoHelper()
		withNamespace()
		gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
	}
	refusedBy := func(ruleID string) {
		ginkgo.GinkgoHelper()
		withNamespace()
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.HaveOccurred())
		var validationErr *protovalidate.ValidationError
		gomega.Expect(errors.As(err, &validationErr)).To(gomega.BeTrue(), "expected a validation error, got %v", err)
		ruleIDs := make([]string, 0, len(validationErr.Violations))
		for _, violation := range validationErr.Violations {
			ruleIDs = append(ruleIDs, violation.Proto.GetRuleId())
		}
		gomega.Expect(ruleIDs).To(gomega.ContainElement(ruleID), "refused, but not by %s: %v", ruleID, err)
	}

	ginkgo.Describe("the namespace", func() {
		ginkgo.It("requires it, so its owner is always a visible choice", func() {
			err := protovalidate.Validate(input)
			var validationErr *protovalidate.ValidationError
			gomega.Expect(errors.As(err, &validationErr)).To(gomega.BeTrue(), "expected a validation error, got %v", err)
			elements := validationErr.Violations[0].Proto.GetField().GetElements()
			gomega.Expect(elements[len(elements)-1].GetFieldName()).To(gomega.Equal("namespace"))
		})

		ginkgo.It("refuses any literal but upstream's", func() {
			input.Spec.Namespace = literal("sandboxes")
			refusedBy("namespace_fixed")
		})

		ginkgo.It("accepts a reference to the KubernetesNamespace that owns it", func() {
			input.Spec.Namespace = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{Name: "agent-sandbox-system"},
				},
			}
			valid()
		})
	})

	ginkgo.Describe("the minimal and complete shapes", func() {
		ginkgo.It("accepts a spec with only the namespace (core only, upstream's manifest as published)", func() {
			valid()
		})

		ginkgo.It("accepts every setting at once", func() {
			input.Spec = &KubernetesAgentSandboxControllerSpec{
				Namespace:        literal("agent-sandbox-system"),
				CreateNamespace:  true,
				Crds:             &KubernetesAgentSandboxControllerCrds{Install: boolPtr(true), KeepOnUninstall: boolPtr(true)},
				ClusterDomain:    stringPtr("cluster.local"),
				ImageRegistry:    "mirror.example.com/k8s",
				ImagePullSecrets: []string{"mirror-pull"},
				Extensions: &KubernetesAgentSandboxControllerExtensions{
					Enabled:             boolPtr(true),
					AllowedLabelDomains: []string{"sandbox.users.io", "example.com"},
					Claim: &KubernetesAgentSandboxControllerClaimSettings{
						ConcurrentWorkers:                int32Ptr(80),
						WarmCandidateGracePeriod:         stringPtr("500ms"),
						EventsDisabled:                   true,
						ObservabilityAnnotationsDisabled: true,
					},
					WarmPool: &KubernetesAgentSandboxControllerWarmPoolSettings{
						ConcurrentWorkers:            int32Ptr(4),
						MaxBatchSize:                 int32Ptr(100),
						ReplenishDelay:               stringPtr("10s"),
						MaxRefillRate:                float64Ptr(2.5),
						ReadinessGracePeriod:         stringPtr("10m"),
						UnschedulableRecheckInterval: stringPtr("30s"),
						EvictionEnabled:              boolPtr(false),
					},
					Template: &KubernetesAgentSandboxControllerTemplateSettings{ConcurrentWorkers: int32Ptr(2)},
				},
				Controller: &KubernetesAgentSandboxControllerManager{
					Image:    &kubernetes.ContainerImage{Repo: "mirror.example.com/agent-sandbox-controller", Tag: "v1.0.7"},
					Replicas: int32Ptr(2),
					Resources: &kubernetes.ContainerResources{
						Requests: &kubernetes.CpuMemory{Cpu: "100m", Memory: "128Mi"},
						Limits:   &kubernetes.CpuMemory{Cpu: "1", Memory: "512Mi"},
					},
					NodeSelector:   map[string]string{"role": "platform"},
					Tolerations:    []*kubernetes.WorkloadToleration{{Key: "platform", Operator: "Exists", Effect: "NoSchedule"}},
					LeaderElection: &KubernetesAgentSandboxControllerLeaderElection{Enabled: boolPtr(true), Namespace: "agent-sandbox-system"},
					Metrics: &KubernetesAgentSandboxControllerMetrics{
						Port:                     int32Ptr(8443),
						RestClientMetricsEnabled: boolPtr(false),
						Tls: &KubernetesAgentSandboxControllerMetricsTls{
							Enabled:           boolPtr(true),
							CertificateSecret: literal("agent-sandbox-metrics-tls"),
							MinVersion:        KubernetesAgentSandboxControllerTlsVersion_tls12,
							CipherSuites:      []string{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"},
						},
					},
					HealthProbePort:     int32Ptr(8081),
					Sandbox:             &KubernetesAgentSandboxControllerSandboxSettings{ConcurrentWorkers: int32Ptr(200), EventsDisabled: true, WriteBehindWindow: stringPtr("200ms")},
					CacheLabelSelectors: true,
					KubeApi:             &KubernetesAgentSandboxControllerKubeApi{Qps: float64Ptr(50), Burst: int32Ptr(100), Connections: int32Ptr(4), SeparateWatchConnection: true},
					Tracing:             &KubernetesAgentSandboxControllerTracing{OtlpEndpoint: "http://otel-collector.observability:4317", Insecure: true},
					Profiling:           &KubernetesAgentSandboxControllerProfiling{CpuEnabled: true, DebugEnabled: true, BlockProfileRate: int32Ptr(1), MutexProfileFraction: int32Ptr(100)},
					Logging: &KubernetesAgentSandboxControllerLogging{
						Development:     true,
						Encoder:         KubernetesAgentSandboxControllerLogging_console,
						Level:           "4",
						StacktraceLevel: KubernetesAgentSandboxControllerLogging_error,
						TimeEncoding:    KubernetesAgentSandboxControllerLogging_rfc3339nano,
					},
				},
				Router: &KubernetesAgentSandboxControllerRouter{
					Enabled:      boolPtr(true),
					Image:        &kubernetes.ContainerImage{Repo: "mirror.example.com/sandbox-router-go", Tag: "v1.0.7"},
					Replicas:     int32Ptr(3),
					MinAvailable: int32Ptr(2),
					Authorization: &KubernetesAgentSandboxControllerRouterAuthorization{
						Mode: &KubernetesAgentSandboxControllerRouterAuthorization_ScopedToken{
							ScopedToken: &KubernetesAgentSandboxControllerRouterScopedToken{
								V1Secret:         secretKey("router-hmac", "secret"),
								VerificationKeys: secretKey("router-keys", "keys.json"),
								V1AcceptUntil:    "2026-12-31T00:00:00Z",
							},
						},
						Cookie: &KubernetesAgentSandboxControllerRouterCookie{
							Name:                "sandbox-token",
							QueryParam:          "token",
							SameSite:            KubernetesAgentSandboxControllerRouterCookie_strict,
							AllowedOrigins:      []string{"https://ide.example.com"},
							TrustForwardedProto: true,
						},
					},
					Tls: &KubernetesAgentSandboxControllerRouterTls{
						Enabled:           boolPtr(true),
						CertificateSecret: literal("router-tls"),
						Port:              int32Ptr(8443),
						HttpEnabled:       boolPtr(false),
						MtlsMode:          KubernetesAgentSandboxControllerRouterTls_required,
						ClientCa:          secretKey("router-tls", "ca.crt"),
						MinVersion:        KubernetesAgentSandboxControllerTlsVersion_tls13,
					},
					Proxy: &KubernetesAgentSandboxControllerRouterProxy{
						Timeout:                   stringPtr("10m"),
						ResponseHeaderTimeout:     stringPtr("1m"),
						ShutdownTimeout:           stringPtr("0s"),
						MaxRequestBodyBytes:       10485760,
						AllowLoopbackPodIp:        true,
						PathRoutingPrefix:         "/sandboxes",
						UpstreamMaxRetries:        int32Ptr(0),
						UpstreamRetryInitialDelay: stringPtr("100ms"),
						UpstreamRetryMaxDelay:     stringPtr("2s"),
					},
					Cache:            &KubernetesAgentSandboxControllerRouterCache{Enabled: boolPtr(true), Namespace: "sandboxes"},
					Telemetry:        &KubernetesAgentSandboxControllerRouterTelemetry{OtlpEndpoint: "http://otel-collector.observability:4317", Insecure: true, TracesEnabled: boolPtr(true), MetricsEnabled: boolPtr(false)},
					AccessLogEnabled: boolPtr(false),
					Ports:            &KubernetesAgentSandboxControllerRouterPorts{Http: int32Ptr(8080), Metrics: int32Ptr(9090), HealthProbe: int32Ptr(8081)},
					NetworkPolicy: &KubernetesAgentSandboxControllerRouterNetworkPolicy{
						Enabled:               boolPtr(true),
						ClientNamespaceLabels: map[string]string{"kubernetes.io/metadata.name": "gateway"},
						MetricsNamespace:      stringPtr("observability"),
						SandboxPorts:          []int32{8888, 3000},
					},
				},
			}
			valid()
		})

		ginkgo.It("refuses a missing spec", func() {
			input.Spec = nil
			refusedBy("required")
		})

		ginkgo.It("refuses a wrong kind constant", func() {
			input.Kind = "AgentSandboxController"
			refusedBy("string.const")
		})
	})

	ginkgo.Describe("cluster-wide settings", func() {
		ginkgo.It("refuses a cluster domain that is not a DNS domain", func() {
			input.Spec.ClusterDomain = stringPtr("Cluster_Local")
			refusedBy("cluster_domain_format")
		})

		ginkgo.It("accepts a registry root with a path and refuses a scheme or trailing slash", func() {
			input.Spec.ImageRegistry = "asia-south1-docker.pkg.dev/example/k8s"
			valid()
			input.Spec.ImageRegistry = "https://mirror.example.com"
			refusedBy("image_registry_format")
			input.Spec.ImageRegistry = "mirror.example.com/"
			refusedBy("image_registry_format")
		})
	})

	ginkgo.Describe("extensions", func() {
		ginkgo.BeforeEach(func() {
			input.Spec.Extensions = &KubernetesAgentSandboxControllerExtensions{}
		})

		ginkgo.It("accepts an empty block (on, with defaults)", func() {
			valid()
		})

		ginkgo.It("refuses a label domain that is not a lowercase DNS domain, or a repeated one", func() {
			input.Spec.Extensions.AllowedLabelDomains = []string{"Example.com"}
			refusedBy("allowed_label_domains_format")
			input.Spec.Extensions.AllowedLabelDomains = []string{"example.com", "example.com"}
			refusedBy("repeated.unique")
		})

		ginkgo.It("refuses non-positive worker counts and batch sizes", func() {
			input.Spec.Extensions.Claim = &KubernetesAgentSandboxControllerClaimSettings{ConcurrentWorkers: int32Ptr(0)}
			refusedBy("int32.gt")
			input.Spec.Extensions.Claim = nil
			input.Spec.Extensions.WarmPool = &KubernetesAgentSandboxControllerWarmPoolSettings{MaxBatchSize: int32Ptr(0)}
			refusedBy("int32.gt")
			input.Spec.Extensions.WarmPool = nil
			input.Spec.Extensions.Template = &KubernetesAgentSandboxControllerTemplateSettings{ConcurrentWorkers: int32Ptr(0)}
			refusedBy("int32.gt")
		})

		ginkgo.It("refuses grace periods and intervals that are zero or not durations", func() {
			input.Spec.Extensions.Claim = &KubernetesAgentSandboxControllerClaimSettings{WarmCandidateGracePeriod: stringPtr("0s")}
			refusedBy("warm_candidate_grace_period_positive")
			input.Spec.Extensions.Claim.WarmCandidateGracePeriod = stringPtr("2 seconds")
			refusedBy("warm_candidate_grace_period_positive")
			input.Spec.Extensions.Claim = nil
			input.Spec.Extensions.WarmPool = &KubernetesAgentSandboxControllerWarmPoolSettings{ReadinessGracePeriod: stringPtr("0m")}
			refusedBy("readiness_grace_period_positive")
			input.Spec.Extensions.WarmPool = &KubernetesAgentSandboxControllerWarmPoolSettings{UnschedulableRecheckInterval: stringPtr("0s")}
			refusedBy("unschedulable_recheck_interval_positive")
		})

		ginkgo.It("accepts a zero replenish delay and refuses one that is not a duration", func() {
			input.Spec.Extensions.WarmPool = &KubernetesAgentSandboxControllerWarmPoolSettings{ReplenishDelay: stringPtr("0s")}
			valid()
			input.Spec.Extensions.WarmPool.ReplenishDelay = stringPtr("-5s")
			refusedBy("replenish_delay_format")
		})

		ginkgo.It("accepts a zero refill rate (unpaced) and refuses a negative one", func() {
			input.Spec.Extensions.WarmPool = &KubernetesAgentSandboxControllerWarmPoolSettings{MaxRefillRate: float64Ptr(0)}
			valid()
			input.Spec.Extensions.WarmPool.MaxRefillRate = float64Ptr(-1)
			refusedBy("max_refill_rate_range")
		})
	})

	ginkgo.Describe("the controller", func() {
		ginkgo.BeforeEach(func() {
			input.Spec.Controller = &KubernetesAgentSandboxControllerManager{}
		})

		ginkgo.It("refuses zero replicas", func() {
			input.Spec.Controller.Replicas = int32Ptr(0)
			refusedBy("int32.gte")
		})

		ginkgo.It("accepts several replicas with leader election and refuses them without", func() {
			input.Spec.Controller.Replicas = int32Ptr(2)
			valid()
			input.Spec.Controller.LeaderElection = &KubernetesAgentSandboxControllerLeaderElection{Enabled: boolPtr(false)}
			refusedBy("controller.replicas_need_leader_election")
			input.Spec.Controller.Replicas = int32Ptr(1)
			valid()
		})

		ginkgo.It("refuses a leader-election namespace that is not a namespace name", func() {
			input.Spec.Controller.LeaderElection = &KubernetesAgentSandboxControllerLeaderElection{Namespace: "Leases"}
			refusedBy("leader_election_namespace_format")
		})

		ginkgo.It("refuses a health port that collides with the default metrics port", func() {
			input.Spec.Controller.HealthProbePort = int32Ptr(8080)
			refusedBy("controller.distinct_ports")
		})

		ginkgo.It("refuses a port outside 1-65535", func() {
			input.Spec.Controller.Metrics = &KubernetesAgentSandboxControllerMetrics{Port: int32Ptr(70000)}
			refusedBy("int32.gte_lte")
		})

		ginkgo.It("refuses cipher suites alongside TLS 1.3, and a malformed suite name", func() {
			input.Spec.Controller.Metrics = &KubernetesAgentSandboxControllerMetrics{Tls: &KubernetesAgentSandboxControllerMetricsTls{
				MinVersion:   KubernetesAgentSandboxControllerTlsVersion_tls13,
				CipherSuites: []string{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"},
			}}
			refusedBy("metrics_tls.no_ciphers_with_tls13")
			input.Spec.Controller.Metrics.Tls.MinVersion = KubernetesAgentSandboxControllerTlsVersion_tls12
			input.Spec.Controller.Metrics.Tls.CipherSuites = []string{"aes128"}
			refusedBy("metrics_cipher_suites_format")
		})

		ginkgo.It("accepts metrics TLS with a self-signed certificate (no Secret)", func() {
			input.Spec.Controller.Metrics = &KubernetesAgentSandboxControllerMetrics{Tls: &KubernetesAgentSandboxControllerMetricsTls{}}
			valid()
		})

		ginkgo.It("refuses a write-behind window that is not a duration", func() {
			input.Spec.Controller.Sandbox = &KubernetesAgentSandboxControllerSandboxSettings{WriteBehindWindow: stringPtr("fast")}
			refusedBy("write_behind_window_format")
		})

		ginkgo.It("refuses a zero qps, burst or connection count", func() {
			input.Spec.Controller.KubeApi = &KubernetesAgentSandboxControllerKubeApi{Qps: float64Ptr(0)}
			refusedBy("kube_api_qps_positive")
			input.Spec.Controller.KubeApi = &KubernetesAgentSandboxControllerKubeApi{Burst: int32Ptr(0)}
			refusedBy("int32.gt")
			input.Spec.Controller.KubeApi = &KubernetesAgentSandboxControllerKubeApi{Connections: int32Ptr(0)}
			refusedBy("int32.gte")
		})

		ginkgo.It("refuses tracing without an endpoint, or with one that is not a URL", func() {
			input.Spec.Controller.Tracing = &KubernetesAgentSandboxControllerTracing{}
			refusedBy("required")
			input.Spec.Controller.Tracing.OtlpEndpoint = "otel-collector:4317"
			refusedBy("tracing_otlp_endpoint_format")
		})

		ginkgo.It("accepts default sampling without debug, and refuses changed sampling without it", func() {
			input.Spec.Controller.Profiling = &KubernetesAgentSandboxControllerProfiling{
				CpuEnabled: true, BlockProfileRate: int32Ptr(1000000), MutexProfileFraction: int32Ptr(10),
			}
			valid()
			input.Spec.Controller.Profiling.BlockProfileRate = int32Ptr(1)
			refusedBy("profiling.sampling_needs_debug")
			input.Spec.Controller.Profiling.DebugEnabled = true
			valid()
		})

		ginkgo.It("refuses a log level that is not a name or a positive number", func() {
			input.Spec.Controller.Logging = &KubernetesAgentSandboxControllerLogging{Level: "warn"}
			refusedBy("logging_level_format")
			input.Spec.Controller.Logging.Level = "0"
			refusedBy("logging_level_format")
		})
	})

	ginkgo.Describe("the router", func() {
		ginkgo.BeforeEach(func() {
			input.Spec.Router = &KubernetesAgentSandboxControllerRouter{Authorization: allowAll()}
		})

		ginkgo.It("accepts a router with an explicit allow-all", func() {
			valid()
		})

		ginkgo.It("refuses a router without an authorization mode", func() {
			input.Spec.Router.Authorization = nil
			refusedBy("required")
			input.Spec.Router.Authorization = &KubernetesAgentSandboxControllerRouterAuthorization{}
			refusedBy("required")
		})

		ginkgo.It("refuses a disruption budget that is not below the replicas, and accepts 0", func() {
			input.Spec.Router.MinAvailable = int32Ptr(2)
			refusedBy("router.budget_below_replicas")
			input.Spec.Router.Replicas = int32Ptr(1)
			input.Spec.Router.MinAvailable = int32Ptr(1)
			refusedBy("router.budget_below_replicas")
			input.Spec.Router.MinAvailable = int32Ptr(0)
			valid()
		})

		ginkgo.It("refuses colliding ports, counting the defaults", func() {
			input.Spec.Router.Ports = &KubernetesAgentSandboxControllerRouterPorts{Metrics: int32Ptr(8080)}
			refusedBy("router.distinct_ports")
			input.Spec.Router.Ports = nil
			input.Spec.Router.Tls = &KubernetesAgentSandboxControllerRouterTls{CertificateSecret: literal("tls"), Port: int32Ptr(9090)}
			refusedBy("router.distinct_ports")
		})

		ginkgo.It("refuses a router with every listener off", func() {
			input.Spec.Router.Tls = &KubernetesAgentSandboxControllerRouterTls{
				Enabled: boolPtr(false), HttpEnabled: boolPtr(false), CertificateSecret: literal("tls"),
			}
			refusedBy("router.listener_required")
			input.Spec.Router.Tls.Enabled = boolPtr(true)
			valid()
		})

		ginkgo.Describe("authorization", func() {
			ginkgo.It("refuses a cookie with allow-all", func() {
				input.Spec.Router.Proxy = &KubernetesAgentSandboxControllerRouterProxy{PathRoutingPrefix: "/s"}
				input.Spec.Router.Authorization.Cookie = &KubernetesAgentSandboxControllerRouterCookie{Name: "t"}
				refusedBy("authorization.cookie_needs_a_credential")
			})

			ginkgo.It("refuses a cookie without a path routing prefix", func() {
				input.Spec.Router.Authorization = tokenReview()
				input.Spec.Router.Authorization.Cookie = &KubernetesAgentSandboxControllerRouterCookie{Name: "t"}
				refusedBy("router.cookie_needs_path_routing")
				input.Spec.Router.Proxy = &KubernetesAgentSandboxControllerRouterProxy{PathRoutingPrefix: "/s"}
				valid()
			})

			ginkgo.It("refuses malformed cookie names, parameters and origins, and SameSite none without Secure", func() {
				input.Spec.Router.Authorization = tokenReview()
				input.Spec.Router.Proxy = &KubernetesAgentSandboxControllerRouterProxy{PathRoutingPrefix: "/s"}
				input.Spec.Router.Authorization.Cookie = &KubernetesAgentSandboxControllerRouterCookie{Name: "bad name"}
				refusedBy("cookie_name_format")
				input.Spec.Router.Authorization.Cookie = &KubernetesAgentSandboxControllerRouterCookie{Name: "t", QueryParam: "a=b"}
				refusedBy("cookie_query_param_format")
				input.Spec.Router.Authorization.Cookie = &KubernetesAgentSandboxControllerRouterCookie{Name: "t", AllowedOrigins: []string{"https://ide.example.com/app"}}
				refusedBy("cookie_allowed_origins_format")
				input.Spec.Router.Authorization.Cookie = &KubernetesAgentSandboxControllerRouterCookie{
					Name: "t", SameSite: KubernetesAgentSandboxControllerRouterCookie_none, Insecure: true,
				}
				refusedBy("cookie.same_site_none_needs_secure")
			})

			ginkgo.It("refuses a token review with a zero TTL or cache size", func() {
				input.Spec.Router.Authorization = &KubernetesAgentSandboxControllerRouterAuthorization{
					Mode: &KubernetesAgentSandboxControllerRouterAuthorization_TokenReview{
						TokenReview: &KubernetesAgentSandboxControllerRouterTokenReview{Ttl: stringPtr("0s")},
					},
				}
				refusedBy("token_review_ttl_positive")
				input.Spec.Router.Authorization.GetTokenReview().Ttl = nil
				input.Spec.Router.Authorization.GetTokenReview().CacheSize = int32Ptr(0)
				refusedBy("int32.gt")
			})

			ginkgo.It("refuses scoped tokens with no verifier", func() {
				input.Spec.Router.Authorization = scopedToken(&KubernetesAgentSandboxControllerRouterScopedToken{})
				refusedBy("scoped_token.verifier_required")
			})

			ginkgo.It("requires the v1 cutoff exactly while v1 and v2 both verify", func() {
				input.Spec.Router.Authorization = scopedToken(&KubernetesAgentSandboxControllerRouterScopedToken{
					V1Secret: secretKey("hmac", "secret"), VerificationKeys: secretKey("keys", "keys.json"),
				})
				refusedBy("scoped_token.cutoff_with_both")
				input.Spec.Router.Authorization.GetScopedToken().V1AcceptUntil = "31/12/2026"
				refusedBy("v1_accept_until_format")
				input.Spec.Router.Authorization.GetScopedToken().V1AcceptUntil = "2026-12-31T00:00:00+05:30"
				valid()
			})

			ginkgo.It("refuses a v1 cutoff without a v1 secret", func() {
				input.Spec.Router.Authorization = scopedToken(&KubernetesAgentSandboxControllerRouterScopedToken{
					VerificationKeys: secretKey("keys", "keys.json"), V1AcceptUntil: "2026-12-31T00:00:00Z",
				})
				refusedBy("scoped_token.cutoff_needs_v1")
			})

			ginkgo.It("refuses v2 verification with the pod-IP cache off", func() {
				input.Spec.Router.Authorization = scopedToken(&KubernetesAgentSandboxControllerRouterScopedToken{
					VerificationKeys: secretKey("keys", "keys.json"),
				})
				valid()
				input.Spec.Router.Cache = &KubernetesAgentSandboxControllerRouterCache{Enabled: boolPtr(false)}
				refusedBy("router.verification_keys_need_cache")
			})

			ginkgo.It("refuses a Secret key reference without a Secret or with a malformed key", func() {
				input.Spec.Router.Authorization = scopedToken(&KubernetesAgentSandboxControllerRouterScopedToken{
					V1Secret: &KubernetesAgentSandboxControllerSecretKey{Key: "secret"},
				})
				refusedBy("required")
				input.Spec.Router.Authorization.GetScopedToken().V1Secret = secretKey("hmac", "a/b")
				refusedBy("secret_key_format")
			})
		})

		ginkgo.Describe("TLS", func() {
			ginkgo.It("refuses TLS without a certificate Secret", func() {
				input.Spec.Router.Tls = &KubernetesAgentSandboxControllerRouterTls{}
				refusedBy("required")
			})

			ginkgo.It("refuses mutual TLS without a client CA", func() {
				input.Spec.Router.Tls = &KubernetesAgentSandboxControllerRouterTls{
					CertificateSecret: literal("tls"), MtlsMode: KubernetesAgentSandboxControllerRouterTls_optional,
				}
				refusedBy("router_tls.mtls_needs_client_ca")
				input.Spec.Router.Tls.ClientCa = secretKey("tls", "ca.crt")
				valid()
			})

			ginkgo.It("refuses cipher suites alongside TLS 1.3", func() {
				input.Spec.Router.Tls = &KubernetesAgentSandboxControllerRouterTls{
					CertificateSecret: literal("tls"),
					MinVersion:        KubernetesAgentSandboxControllerTlsVersion_tls13,
					CipherSuites:      []string{"TLS_AES_128_GCM_SHA256"},
				}
				refusedBy("router_tls.no_ciphers_with_tls13")
			})
		})

		ginkgo.Describe("the proxy", func() {
			ginkgo.It("refuses zero timeouts, and accepts a zero shutdown timeout", func() {
				input.Spec.Router.Proxy = &KubernetesAgentSandboxControllerRouterProxy{Timeout: stringPtr("0s")}
				refusedBy("proxy_timeout_positive")
				input.Spec.Router.Proxy = &KubernetesAgentSandboxControllerRouterProxy{ResponseHeaderTimeout: stringPtr("0s")}
				refusedBy("proxy_response_header_timeout_positive")
				input.Spec.Router.Proxy = &KubernetesAgentSandboxControllerRouterProxy{ShutdownTimeout: stringPtr("0s")}
				valid()
			})

			ginkgo.It("refuses a negative body limit or retry count", func() {
				input.Spec.Router.Proxy = &KubernetesAgentSandboxControllerRouterProxy{MaxRequestBodyBytes: -1}
				refusedBy("int64.gte")
				input.Spec.Router.Proxy = &KubernetesAgentSandboxControllerRouterProxy{UpstreamMaxRetries: int32Ptr(-1)}
				refusedBy("int32.gte")
			})

			ginkgo.It("refuses retry delays that are not durations", func() {
				input.Spec.Router.Proxy = &KubernetesAgentSandboxControllerRouterProxy{UpstreamRetryInitialDelay: stringPtr("soon")}
				refusedBy("upstream_retry_initial_delay_format")
				input.Spec.Router.Proxy = &KubernetesAgentSandboxControllerRouterProxy{UpstreamRetryMaxDelay: stringPtr("1 second")}
				refusedBy("upstream_retry_max_delay_format")
			})

			ginkgo.It("accepts a path prefix and refuses one without a leading slash, with a trailing slash, or with spaces", func() {
				input.Spec.Router.Proxy = &KubernetesAgentSandboxControllerRouterProxy{PathRoutingPrefix: "/sandboxes/v1"}
				valid()
				for _, bad := range []string{"sandboxes", "/sandboxes/", "/sand boxes", "/"} {
					input.Spec.Router.Proxy.PathRoutingPrefix = bad
					refusedBy("path_routing_prefix_format")
				}
			})
		})

		ginkgo.It("refuses a cache namespace that is not a namespace name", func() {
			input.Spec.Router.Cache = &KubernetesAgentSandboxControllerRouterCache{Namespace: "Sandboxes"}
			refusedBy("router_cache_namespace_format")
		})

		ginkgo.It("refuses telemetry without an endpoint", func() {
			input.Spec.Router.Telemetry = &KubernetesAgentSandboxControllerRouterTelemetry{}
			refusedBy("required")
		})

		ginkgo.It("refuses a malformed metrics namespace and sandbox ports outside 1-65535 or repeated", func() {
			input.Spec.Router.NetworkPolicy = &KubernetesAgentSandboxControllerRouterNetworkPolicy{MetricsNamespace: stringPtr("Monitoring")}
			refusedBy("router_metrics_namespace_format")
			input.Spec.Router.NetworkPolicy = &KubernetesAgentSandboxControllerRouterNetworkPolicy{SandboxPorts: []int32{0}}
			refusedBy("int32.gte_lte")
			input.Spec.Router.NetworkPolicy = &KubernetesAgentSandboxControllerRouterNetworkPolicy{SandboxPorts: []int32{8888, 8888}}
			refusedBy("repeated.unique")
		})
	})
})
