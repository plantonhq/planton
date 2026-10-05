package kuberneteskarapacev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesKarapace(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesKarapace Suite")
}

func int32Ptr(i int32) *int32    { return &i }
func stringPtr(s string) *string { return &s }

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valueFrom(kind catalogkind.CatalogKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Kind:      kind,
				Name:      name,
				FieldPath: fieldPath,
			},
		},
	}
}

// scramSasl returns a valid SASL block backed by a password Secret.
func scramSasl() *KubernetesKarapaceKafkaSasl {
	return &KubernetesKarapaceKafkaSasl{
		Mechanism: "SCRAM-SHA-512",
		Username:  "karapace-user",
		PasswordSecret: &KubernetesKarapacePasswordSecret{
			SecretName: literal("karapace-user"),
		},
	}
}

// caTls returns a valid Kafka-side TLS block trusting a cluster CA.
func caTls() *KubernetesKarapaceKafkaTls {
	return &KubernetesKarapaceKafkaTls{
		CaSecretName: literal("my-kafka-cluster-ca-cert"),
	}
}

var _ = ginkgo.Describe("KubernetesKarapace Validation Tests", func() {
	var input *KubernetesKarapace

	ginkgo.BeforeEach(func() {
		input = &KubernetesKarapace{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesKarapace",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-karapace",
			},
			Spec: &KubernetesKarapaceSpec{
				Namespace: literal("kafka"),
				Kafka: &KubernetesKarapaceKafka{
					BootstrapServers: literal("my-kafka-kafka-bootstrap:9092"),
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("bootstrap_servers as a reference should be valid", func() {
			input.Spec.Kafka.BootstrapServers = valueFrom(catalogkind.CatalogKind_KubernetesKafka, "my-kafka", "status.outputs.internal_bootstrap_endpoint")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("PLAINTEXT security protocol without tls or sasl should be valid", func() {
			input.Spec.Kafka.SecurityProtocol = stringPtr("PLAINTEXT")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("SSL security protocol with a tls block should be valid", func() {
			input.Spec.Kafka.SecurityProtocol = stringPtr("SSL")
			input.Spec.Kafka.Tls = caTls()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("SASL_PLAINTEXT with a sasl block should be valid", func() {
			input.Spec.Kafka.SecurityProtocol = stringPtr("SASL_PLAINTEXT")
			input.Spec.Kafka.Sasl = scramSasl()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("SASL_SSL with both tls and sasl blocks should be valid", func() {
			input.Spec.Kafka.SecurityProtocol = stringPtr("SASL_SSL")
			input.Spec.Kafka.Tls = caTls()
			input.Spec.Kafka.Sasl = scramSasl()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("mutual-TLS client identity within the tls block should be valid", func() {
			input.Spec.Kafka.SecurityProtocol = stringPtr("SSL")
			input.Spec.Kafka.Tls = caTls()
			input.Spec.Kafka.Tls.ClientCertSecretName = literal("karapace-tls-user")
			input.Spec.Kafka.Tls.ClientCertificate = stringPtr("user.crt")
			input.Spec.Kafka.Tls.ClientKey = stringPtr("user.key")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a sasl password declared as a value instead of a Secret should be valid", func() {
			input.Spec.Kafka.SecurityProtocol = stringPtr("SASL_PLAINTEXT")
			input.Spec.Kafka.Sasl = &KubernetesKarapaceKafkaSasl{
				Mechanism: "PLAIN",
				Username:  "api-key",
				Password:  "api-secret",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("every log level should be valid", func() {
			for _, level := range []string{"DEBUG", "INFO", "WARNING", "ERROR"} {
				input.Spec.LogLevel = stringPtr(level)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("every compatibility mode should be valid", func() {
			for _, mode := range []string{"BACKWARD", "BACKWARD_TRANSITIVE", "FORWARD", "FORWARD_TRANSITIVE", "FULL", "FULL_TRANSITIVE", "NONE"} {
				input.Spec.Registry = &KubernetesKarapaceRegistry{Compatibility: stringPtr(mode)}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("registry tuning (topic, replication factor, group, election) should be valid", func() {
			input.Spec.Registry = &KubernetesKarapaceRegistry{
				TopicName:              stringPtr("_schemas"),
				ReplicationFactor:      int32Ptr(3),
				GroupId:                "karapace-registry",
				MasterElectionStrategy: stringPtr("highest"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("basic HTTP authentication should be valid", func() {
			input.Spec.HttpAuthentication = &KubernetesKarapaceHttpAuthentication{
				Basic: &KubernetesKarapaceBasicAuth{
					SecretName: literal("karapace-authfile"),
					Key:        stringPtr("authfile.json"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("OIDC HTTP authentication with an https JWKS endpoint should be valid", func() {
			input.Spec.HttpAuthentication = &KubernetesKarapaceHttpAuthentication{
				Oidc: &KubernetesKarapaceOidc{
					JwksEndpointUrl:  "https://idp.example.com/.well-known/jwks.json",
					ExpectedIssuer:   "https://idp.example.com",
					ExpectedAudience: "karapace",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full surface (replicas, port, server TLS, rest proxy, image) should be valid", func() {
			input.Spec.Replicas = int32Ptr(2)
			input.Spec.Port = int32Ptr(8081)
			input.Spec.Image = "ghcr.io/aiven-open/karapace:4.1.0"
			input.Spec.ServerTls = &KubernetesKarapaceServerTls{
				SecretName:  literal("karapace-server-tls"),
				Certificate: stringPtr("tls.crt"),
				Key:         stringPtr("tls.key"),
			}
			input.Spec.RestProxy = &KubernetesKarapaceRestProxy{
				Enabled:  true,
				Replicas: int32Ptr(2),
				Port:     int32Ptr(8082),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When top-level fields are invalid", func() {
		ginkgo.It("missing namespace should fail (required)", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("missing kafka block should fail (required)", func() {
			input.Spec.Kafka = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero replicas should fail (gte 1)", func() {
			input.Spec.Replicas = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("port zero should fail (gte 1)", func() {
			input.Spec.Port = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("port above 65535 should fail (lte 65535)", func() {
			input.Spec.Port = int32Ptr(70000)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown log level should fail (spec.log_level_enum)", func() {
			input.Spec.LogLevel = stringPtr("TRACE")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When the Kafka connection is invalid", func() {
		ginkgo.It("missing bootstrap_servers should fail (required)", func() {
			input.Spec.Kafka.BootstrapServers = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown security protocol should fail (spec.kafka.security_protocol_enum)", func() {
			input.Spec.Kafka.SecurityProtocol = stringPtr("SASL_GSSAPI")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("SSL without a tls block should fail (spec.kafka.ssl_requires_tls)", func() {
			input.Spec.Kafka.SecurityProtocol = stringPtr("SSL")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("SASL_SSL without a tls block should fail (spec.kafka.ssl_requires_tls)", func() {
			input.Spec.Kafka.SecurityProtocol = stringPtr("SASL_SSL")
			input.Spec.Kafka.Sasl = scramSasl()
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("SASL_PLAINTEXT without a sasl block should fail (spec.kafka.sasl_requires_credentials)", func() {
			input.Spec.Kafka.SecurityProtocol = stringPtr("SASL_PLAINTEXT")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a sasl block on the PLAINTEXT protocol should fail (spec.kafka.sasl_only_with_sasl_protocol)", func() {
			input.Spec.Kafka.SecurityProtocol = stringPtr("PLAINTEXT")
			input.Spec.Kafka.Sasl = scramSasl()
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a sasl block without an explicit security protocol should fail (spec.kafka.sasl_only_with_sasl_protocol)", func() {
			input.Spec.Kafka.Sasl = scramSasl()
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a tls block without a CA secret name should fail (required)", func() {
			input.Spec.Kafka.SecurityProtocol = stringPtr("SSL")
			input.Spec.Kafka.Tls = &KubernetesKarapaceKafkaTls{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When SASL credentials are invalid", func() {
		ginkgo.BeforeEach(func() {
			input.Spec.Kafka.SecurityProtocol = stringPtr("SASL_PLAINTEXT")
		})

		ginkgo.It("a missing mechanism should fail (required)", func() {
			input.Spec.Kafka.Sasl = scramSasl()
			input.Spec.Kafka.Sasl.Mechanism = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown mechanism should fail (spec.kafka.sasl.mechanism_enum)", func() {
			input.Spec.Kafka.Sasl = scramSasl()
			input.Spec.Kafka.Sasl.Mechanism = "GSSAPI"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a missing username should fail (required)", func() {
			input.Spec.Kafka.Sasl = scramSasl()
			input.Spec.Kafka.Sasl.Username = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("both password_secret and password should fail (spec.kafka.sasl.password_exactly_one)", func() {
			input.Spec.Kafka.Sasl = scramSasl()
			input.Spec.Kafka.Sasl.Password = "also-a-literal"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("neither password_secret nor password should fail (spec.kafka.sasl.password_exactly_one)", func() {
			input.Spec.Kafka.Sasl = scramSasl()
			input.Spec.Kafka.Sasl.PasswordSecret = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a password secret without a secret name should fail (required)", func() {
			input.Spec.Kafka.Sasl = scramSasl()
			input.Spec.Kafka.Sasl.PasswordSecret = &KubernetesKarapacePasswordSecret{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When registry tuning is invalid", func() {
		ginkgo.It("zero replication factor should fail (gte 1)", func() {
			input.Spec.Registry = &KubernetesKarapaceRegistry{ReplicationFactor: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("replication factor above 32767 should fail (Kafka's int16 bound)", func() {
			input.Spec.Registry = &KubernetesKarapaceRegistry{ReplicationFactor: int32Ptr(40000)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown compatibility mode should fail (spec.registry.compatibility_enum)", func() {
			input.Spec.Registry = &KubernetesKarapaceRegistry{Compatibility: stringPtr("SIDEWAYS")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown election strategy should fail (spec.registry.election_enum)", func() {
			input.Spec.Registry = &KubernetesKarapaceRegistry{MasterElectionStrategy: stringPtr("random")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When the rest proxy is invalid", func() {
		ginkgo.It("zero proxy replicas should fail (gte 1)", func() {
			input.Spec.RestProxy = &KubernetesKarapaceRestProxy{Enabled: true, Replicas: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("proxy port above 65535 should fail (lte 65535)", func() {
			input.Spec.RestProxy = &KubernetesKarapaceRestProxy{Enabled: true, Port: int32Ptr(70000)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When server TLS or HTTP authentication is invalid", func() {
		ginkgo.It("server TLS without a secret name should fail (required)", func() {
			input.Spec.ServerTls = &KubernetesKarapaceServerTls{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("both basic and oidc should fail (spec.http_authentication.exactly_one)", func() {
			input.Spec.HttpAuthentication = &KubernetesKarapaceHttpAuthentication{
				Basic: &KubernetesKarapaceBasicAuth{SecretName: literal("karapace-authfile")},
				Oidc:  &KubernetesKarapaceOidc{JwksEndpointUrl: "https://idp.example.com/jwks.json"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("neither basic nor oidc should fail (spec.http_authentication.exactly_one)", func() {
			input.Spec.HttpAuthentication = &KubernetesKarapaceHttpAuthentication{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("basic authentication without a secret name should fail (required)", func() {
			input.Spec.HttpAuthentication = &KubernetesKarapaceHttpAuthentication{
				Basic: &KubernetesKarapaceBasicAuth{},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a missing JWKS endpoint should fail (required)", func() {
			input.Spec.HttpAuthentication = &KubernetesKarapaceHttpAuthentication{
				Oidc: &KubernetesKarapaceOidc{},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a plain-HTTP JWKS endpoint should fail (spec.oidc.jwks_https)", func() {
			input.Spec.HttpAuthentication = &KubernetesKarapaceHttpAuthentication{
				Oidc: &KubernetesKarapaceOidc{JwksEndpointUrl: "http://idp.example.com/jwks.json"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
