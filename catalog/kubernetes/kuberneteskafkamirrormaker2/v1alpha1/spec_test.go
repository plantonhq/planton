package kuberneteskafkamirrormaker2v1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesKafkaMirrorMaker2(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesKafkaMirrorMaker2 Suite")
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

// mskMirror returns the minimal valid mirror — one source cluster with
// an alias and a bootstrap address (the migrate-from shape every test
// mutates from).
func mskMirror() *KubernetesKafkaMirrorMaker2Mirror {
	return &KubernetesKafkaMirrorMaker2Mirror{
		Source: &KubernetesKafkaMirrorMaker2Source{
			Alias:            "prod-msk",
			BootstrapServers: literal("b-1.prod.kafka.us-east-1.amazonaws.com:9096"),
		},
	}
}

var _ = ginkgo.Describe("KubernetesKafkaMirrorMaker2 Validation Tests", func() {
	var input *KubernetesKafkaMirrorMaker2

	ginkgo.BeforeEach(func() {
		input = &KubernetesKafkaMirrorMaker2{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesKafkaMirrorMaker2",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "msk-migration",
			},
			Spec: &KubernetesKafkaMirrorMaker2Spec{
				Namespace: literal("kafka"),
				Target: &KubernetesKafkaMirrorMaker2Target{
					BootstrapServers: literal("my-kafka-kafka-bootstrap:9092"),
				},
				Mirrors: []*KubernetesKafkaMirrorMaker2Mirror{mskMirror()},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("target bootstrap_servers as a reference should be valid", func() {
			input.Spec.Target.BootstrapServers = valueFrom(catalogkind.CatalogKind_KubernetesKafka, "my-kafka", "status.outputs.internal_bootstrap_endpoint")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a target alias with dots, underscores and dashes should be valid", func() {
			input.Spec.Target.Alias = stringPtr("onprem_dc-1.east")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a target with TLS trust and scram authentication should be valid", func() {
			input.Spec.Target.Tls = &kubernetes.StrimziKafkaClientTls{
				TrustedCertificates: []*kubernetes.StrimziKafkaClientTrustedCertificate{
					{SecretName: literal("my-kafka-cluster-ca-cert"), Certificate: "ca.crt"},
				},
			}
			input.Spec.Target.Authentication = &kubernetes.StrimziKafkaClientAuthentication{
				Type:     "scram-sha-512",
				Username: "mm2-user",
				PasswordSecret: &kubernetes.StrimziKafkaClientPasswordSecret{
					SecretName: literal("mm2-user"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a mirror source with TLS pattern trust and plain authentication should be valid", func() {
			input.Spec.Mirrors[0].Source.Tls = &kubernetes.StrimziKafkaClientTls{
				TrustedCertificates: []*kubernetes.StrimziKafkaClientTrustedCertificate{
					{SecretName: literal("msk-ca"), Pattern: "*.crt"},
				},
			}
			input.Spec.Mirrors[0].Source.Authentication = &kubernetes.StrimziKafkaClientAuthentication{
				Type:     "plain",
				Username: "api-key",
				PasswordSecret: &kubernetes.StrimziKafkaClientPasswordSecret{
					SecretName: literal("confluent-api-secret"),
					Password:   stringPtr("password"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("multiple mirrors with distinct source aliases should be valid", func() {
			second := mskMirror()
			second.Source.Alias = "confluent"
			input.Spec.Mirrors = append(input.Spec.Mirrors, second)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("mirror connectors with tasks and identity replication config should be valid", func() {
			input.Spec.Mirrors[0].SourceConnector = &KubernetesKafkaMirrorMaker2Connector{
				TasksMax: int32Ptr(8),
				Config: map[string]string{
					"replication.policy.class": "org.apache.kafka.connect.mirror.IdentityReplicationPolicy",
				},
				AutoRestart: &KubernetesKafkaMirrorMaker2AutoRestart{Enabled: true, MaxRestarts: int32Ptr(10)},
			}
			input.Spec.Mirrors[0].CheckpointConnector = &KubernetesKafkaMirrorMaker2Connector{
				Config: map[string]string{
					"replication.policy.class": "org.apache.kafka.connect.mirror.IdentityReplicationPolicy",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full surface (patterns, identity, replicas, jvm, rack, metrics) should be valid", func() {
			input.Spec.Replicas = int32Ptr(3)
			input.Spec.Version = "4.3.0"
			input.Spec.Target.GroupId = "mm2-group"
			input.Spec.Target.ConfigStorageTopic = "mm2-configs"
			input.Spec.Target.StatusStorageTopic = "mm2-status"
			input.Spec.Target.OffsetStorageTopic = "mm2-offsets"
			input.Spec.Target.Config = map[string]string{"producer.compression.type": "lz4"}
			input.Spec.Mirrors[0].TopicsPattern = stringPtr("orders.*")
			input.Spec.Mirrors[0].TopicsExcludePattern = "orders.internal.*"
			input.Spec.Mirrors[0].GroupsPattern = stringPtr(".*")
			input.Spec.Mirrors[0].GroupsExcludePattern = "console-consumer-.*"
			input.Spec.Mirrors[0].Source.Config = map[string]string{"fetch.max.bytes": "52428800"}
			input.Spec.Jvm = &KubernetesKafkaMirrorMaker2Jvm{Xms: "1g", Xmx: "1g"}
			input.Spec.Rack = &KubernetesKafkaMirrorMaker2Rack{TopologyKey: "topology.kubernetes.io/zone"}
			input.Spec.Metrics = &KubernetesKafkaMirrorMaker2Metrics{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When top-level fields are invalid", func() {
		ginkgo.It("missing namespace should fail (required)", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("missing target should fail (required)", func() {
			input.Spec.Target = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero replicas should fail (gte 1)", func() {
			input.Spec.Replicas = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a rack block without a topology key should fail (required)", func() {
			input.Spec.Rack = &KubernetesKafkaMirrorMaker2Rack{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When the target is invalid", func() {
		ginkgo.It("missing target bootstrap_servers should fail (required)", func() {
			input.Spec.Target.BootstrapServers = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a target alias with a slash should fail (spec.target.alias.format)", func() {
			input.Spec.Target.Alias = stringPtr("target/1")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("tls authentication without certificate_and_key on the target should fail (strimzi.client.auth.tls_requires_cert)", func() {
			input.Spec.Target.Authentication = &kubernetes.StrimziKafkaClientAuthentication{Type: "tls"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("certificate_and_key on scram target authentication should fail (strimzi.client.auth.cert_only_for_tls)", func() {
			input.Spec.Target.Authentication = &kubernetes.StrimziKafkaClientAuthentication{
				Type:     "scram-sha-512",
				Username: "mm2-user",
				PasswordSecret: &kubernetes.StrimziKafkaClientPasswordSecret{
					SecretName: literal("mm2-user"),
				},
				CertificateAndKey: &kubernetes.StrimziKafkaClientCertificateAndKey{
					SecretName: literal("mm2-user"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When mirrors are invalid", func() {
		ginkgo.It("an empty mirrors list should fail (min_items)", func() {
			input.Spec.Mirrors = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("duplicate source aliases should fail (spec.mirrors.unique_aliases)", func() {
			input.Spec.Mirrors = []*KubernetesKafkaMirrorMaker2Mirror{mskMirror(), mskMirror()}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a source alias equal to the declared target alias should fail (spec.target_alias_distinct_from_sources)", func() {
			input.Spec.Target.Alias = stringPtr("onprem")
			input.Spec.Mirrors[0].Source.Alias = "onprem"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a source alias colliding with the DEFAULT target alias should fail (spec.target_alias_distinct_from_sources)", func() {
			input.Spec.Target.Alias = nil
			input.Spec.Mirrors[0].Source.Alias = "target"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a mirror without a source should fail (required)", func() {
			input.Spec.Mirrors[0].Source = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a missing source alias should fail (required)", func() {
			input.Spec.Mirrors[0].Source.Alias = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a source alias with a slash should fail (spec.mirrors.source.alias.format)", func() {
			input.Spec.Mirrors[0].Source.Alias = "prod/msk"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("missing source bootstrap_servers should fail (required)", func() {
			input.Spec.Mirrors[0].Source.BootstrapServers = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("both certificate and pattern on a source trusted certificate should fail (strimzi.client.trusted_cert.certificate_xor_pattern)", func() {
			input.Spec.Mirrors[0].Source.Tls = &kubernetes.StrimziKafkaClientTls{
				TrustedCertificates: []*kubernetes.StrimziKafkaClientTrustedCertificate{
					{SecretName: literal("msk-ca"), Certificate: "ca.crt", Pattern: "*.crt"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an empty source trusted-certificates list should fail (min_items)", func() {
			input.Spec.Mirrors[0].Source.Tls = &kubernetes.StrimziKafkaClientTls{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown source authentication type should fail (strimzi.client.auth.type_enum)", func() {
			input.Spec.Mirrors[0].Source.Authentication = &kubernetes.StrimziKafkaClientAuthentication{Type: "iam"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("plain source authentication without credentials should fail (strimzi.client.auth.sasl_requires_credentials)", func() {
			input.Spec.Mirrors[0].Source.Authentication = &kubernetes.StrimziKafkaClientAuthentication{Type: "plain"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When mirror connectors are invalid", func() {
		ginkgo.It("zero tasks_max should fail (gte 1)", func() {
			input.Spec.Mirrors[0].SourceConnector = &KubernetesKafkaMirrorMaker2Connector{TasksMax: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("connector.plugin.version in connector config should fail (spec.mirrors.connector.config.no_plugin_version)", func() {
			input.Spec.Mirrors[0].SourceConnector = &KubernetesKafkaMirrorMaker2Connector{
				Config: map[string]string{"connector.plugin.version": "3.1.0"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("connector.plugin.version in the checkpoint connector's config should fail (spec.mirrors.connector.config.no_plugin_version)", func() {
			input.Spec.Mirrors[0].CheckpointConnector = &KubernetesKafkaMirrorMaker2Connector{
				Config: map[string]string{"connector.plugin.version": "3.1.0"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero max_restarts should fail (gte 1)", func() {
			input.Spec.Mirrors[0].SourceConnector = &KubernetesKafkaMirrorMaker2Connector{
				AutoRestart: &KubernetesKafkaMirrorMaker2AutoRestart{Enabled: true, MaxRestarts: int32Ptr(0)},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
