package kuberneteskafkaconnectv1alpha1

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

func TestKubernetesKafkaConnect(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesKafkaConnect Suite")
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

// caTrust returns the minimal valid TLS trust block — one trusted
// certificate naming a single file within the Secret.
func caTrust() *kubernetes.StrimziKafkaClientTls {
	return &kubernetes.StrimziKafkaClientTls{
		TrustedCertificates: []*kubernetes.StrimziKafkaClientTrustedCertificate{
			{SecretName: literal("my-kafka-cluster-ca-cert"), Certificate: "ca.crt"},
		},
	}
}

// scramAuth returns a valid SASL SCRAM authentication block.
func scramAuth() *kubernetes.StrimziKafkaClientAuthentication {
	return &kubernetes.StrimziKafkaClientAuthentication{
		Type:     "scram-sha-512",
		Username: "connect-user",
		PasswordSecret: &kubernetes.StrimziKafkaClientPasswordSecret{
			SecretName: literal("connect-user"),
		},
	}
}

// jarBuild returns the minimal valid build block — a docker output and
// one plugin made of a single jar artifact.
func jarBuild() *KubernetesKafkaConnectBuild {
	return &KubernetesKafkaConnectBuild{
		Output: &KubernetesKafkaConnectBuildOutput{
			Image: "registry.example.com/team/my-connect:latest",
		},
		Plugins: []*KubernetesKafkaConnectBuildPlugin{
			{
				Name: "debezium-postgres",
				Artifacts: []*KubernetesKafkaConnectBuildArtifact{
					{Type: "jar", Url: "https://repo1.maven.org/debezium-connector-postgres-3.1.0.jar"},
				},
			},
		},
	}
}

var _ = ginkgo.Describe("KubernetesKafkaConnect Validation Tests", func() {
	var input *KubernetesKafkaConnect

	ginkgo.BeforeEach(func() {
		input = &KubernetesKafkaConnect{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesKafkaConnect",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-connect",
			},
			Spec: &KubernetesKafkaConnectSpec{
				Namespace:        literal("kafka"),
				BootstrapServers: literal("my-kafka-kafka-bootstrap:9092"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "kafka", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("bootstrap_servers as a reference should be valid", func() {
			input.Spec.BootstrapServers = valueFrom(catalogkind.CatalogKind_KubernetesKafka, "my-kafka", "status.outputs.internal_bootstrap_endpoint")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("three replicas should be valid", func() {
			input.Spec.Replicas = int32Ptr(3)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("worker config without operator-owned keys should be valid", func() {
			input.Spec.Config = map[string]string{
				"key.converter":                     "org.apache.kafka.connect.json.JsonConverter",
				"config.storage.replication.factor": "3",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("TLS trust with a single certificate file should be valid", func() {
			input.Spec.Tls = caTrust()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("TLS trust with a glob pattern should be valid", func() {
			input.Spec.Tls = &kubernetes.StrimziKafkaClientTls{
				TrustedCertificates: []*kubernetes.StrimziKafkaClientTrustedCertificate{
					{SecretName: literal("external-ca"), Pattern: "*.crt"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("tls authentication with a certificate and key should be valid", func() {
			input.Spec.Tls = caTrust()
			input.Spec.Authentication = &kubernetes.StrimziKafkaClientAuthentication{
				Type: "tls",
				CertificateAndKey: &kubernetes.StrimziKafkaClientCertificateAndKey{
					SecretName: literal("connect-user"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("every SASL authentication type with credentials should be valid", func() {
			for _, authType := range []string{"scram-sha-512", "scram-sha-256", "plain"} {
				input.Spec.Authentication = &kubernetes.StrimziKafkaClientAuthentication{
					Type:     authType,
					Username: "connect-user",
					PasswordSecret: &kubernetes.StrimziKafkaClientPasswordSecret{
						SecretName: literal("connect-user"),
						Password:   stringPtr("password"),
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("custom authentication with sasl and config should be valid", func() {
			input.Spec.Authentication = &kubernetes.StrimziKafkaClientAuthentication{
				Type: "custom",
				Sasl: true,
				Config: map[string]string{
					"sasl.mechanism": "AWS_MSK_IAM",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("OCI plugins with unique names and pull policies should be valid", func() {
			input.Spec.Plugins = []*KubernetesKafkaConnectOciPlugin{
				{
					Name: "debezium-postgres",
					Artifacts: []*KubernetesKafkaConnectOciArtifact{
						{Reference: "quay.io/example/debezium-postgres:3.1.0", PullPolicy: stringPtr("IfNotPresent")},
					},
				},
				{
					Name: "s3_sink",
					Artifacts: []*KubernetesKafkaConnectOciArtifact{
						{Reference: "quay.io/example/s3-sink:1.0.0"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a build with a jar artifact should be valid", func() {
			input.Spec.Build = jarBuild()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a build with a maven artifact carrying full coordinates should be valid", func() {
			input.Spec.Build = jarBuild()
			input.Spec.Build.Plugins[0].Artifacts = []*KubernetesKafkaConnectBuildArtifact{
				{Type: "maven", Group: "io.debezium", Artifact: "debezium-connector-postgres", Version: "3.1.0.Final"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("every URL artifact type with a url should be valid", func() {
			for _, artifactType := range []string{"jar", "tgz", "zip", "other"} {
				input.Spec.Build = jarBuild()
				input.Spec.Build.Plugins[0].Artifacts = []*KubernetesKafkaConnectBuildArtifact{
					{Type: artifactType, Url: "https://example.com/plugin-artifact", FileName: "plugin.bin"},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("both build output types should be valid", func() {
			for _, outputType := range []string{"docker", "imagestream"} {
				input.Spec.Build = jarBuild()
				input.Spec.Build.Output.Type = stringPtr(outputType)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("full surface (group identity, image, jvm, rack, metrics, selectors) should be valid", func() {
			input.Spec.Version = "4.3.0"
			input.Spec.GroupId = "my-connect-group"
			input.Spec.ConfigStorageTopic = "my-connect-configs"
			input.Spec.StatusStorageTopic = "my-connect-status"
			input.Spec.OffsetStorageTopic = "my-connect-offsets"
			input.Spec.Image = "quay.io/example/my-connect:latest"
			input.Spec.Jvm = &KubernetesKafkaConnectJvm{Xms: "1g", Xmx: "1g"}
			input.Spec.Rack = &KubernetesKafkaConnectRack{TopologyKey: "topology.kubernetes.io/zone"}
			input.Spec.Metrics = &KubernetesKafkaConnectMetrics{Enabled: true}
			input.Spec.NodeSelector = map[string]string{"disktype": "ssd"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When top-level fields are invalid", func() {
		ginkgo.It("missing namespace should fail (required)", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("missing bootstrap_servers should fail (required)", func() {
			input.Spec.BootstrapServers = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero replicas should fail (gte 1)", func() {
			input.Spec.Replicas = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("connector.plugin.version in worker config should fail (spec.config.no_connector_plugin_version)", func() {
			input.Spec.Config = map[string]string{"connector.plugin.version": "3.1.0"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a rack block without a topology key should fail (required)", func() {
			input.Spec.Rack = &KubernetesKafkaConnectRack{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("image together with build should fail (spec.image_xor_build)", func() {
			input.Spec.Image = "registry.example.com/team/prebuilt-connect:1.0"
			input.Spec.Build = &KubernetesKafkaConnectBuild{
				Output: &KubernetesKafkaConnectBuildOutput{
					Image: "registry.example.com/team/built-connect:1.0",
				},
				Plugins: []*KubernetesKafkaConnectBuildPlugin{{
					Name: "debezium-postgres",
					Artifacts: []*KubernetesKafkaConnectBuildArtifact{{
						Type: "maven", Group: "io.debezium", Artifact: "debezium-connector-postgres", Version: "3.1.0.Final",
					}},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When TLS trust is invalid", func() {
		ginkgo.It("an empty trusted-certificates list should fail (min_items)", func() {
			input.Spec.Tls = &kubernetes.StrimziKafkaClientTls{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a trusted certificate without a secret name should fail (required)", func() {
			input.Spec.Tls = &kubernetes.StrimziKafkaClientTls{
				TrustedCertificates: []*kubernetes.StrimziKafkaClientTrustedCertificate{
					{Certificate: "ca.crt"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("both certificate and pattern should fail (strimzi.client.trusted_cert.certificate_xor_pattern)", func() {
			input.Spec.Tls = &kubernetes.StrimziKafkaClientTls{
				TrustedCertificates: []*kubernetes.StrimziKafkaClientTrustedCertificate{
					{SecretName: literal("ca-secret"), Certificate: "ca.crt", Pattern: "*.crt"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("neither certificate nor pattern should fail (strimzi.client.trusted_cert.certificate_xor_pattern)", func() {
			input.Spec.Tls = &kubernetes.StrimziKafkaClientTls{
				TrustedCertificates: []*kubernetes.StrimziKafkaClientTrustedCertificate{
					{SecretName: literal("ca-secret")},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When authentication is invalid", func() {
		ginkgo.It("a missing authentication type should fail (required)", func() {
			input.Spec.Authentication = &kubernetes.StrimziKafkaClientAuthentication{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown authentication type should fail (strimzi.client.auth.type_enum)", func() {
			input.Spec.Authentication = &kubernetes.StrimziKafkaClientAuthentication{Type: "oauth"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("tls authentication without certificate_and_key should fail (strimzi.client.auth.tls_requires_cert)", func() {
			input.Spec.Authentication = &kubernetes.StrimziKafkaClientAuthentication{Type: "tls"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("scram authentication without a username should fail (strimzi.client.auth.sasl_requires_credentials)", func() {
			input.Spec.Authentication = &kubernetes.StrimziKafkaClientAuthentication{
				Type: "scram-sha-512",
				PasswordSecret: &kubernetes.StrimziKafkaClientPasswordSecret{
					SecretName: literal("connect-user"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("scram authentication without a password secret should fail (strimzi.client.auth.sasl_requires_credentials)", func() {
			input.Spec.Authentication = &kubernetes.StrimziKafkaClientAuthentication{
				Type:     "scram-sha-512",
				Username: "connect-user",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("certificate_and_key on scram authentication should fail (strimzi.client.auth.cert_only_for_tls)", func() {
			input.Spec.Authentication = scramAuth()
			input.Spec.Authentication.CertificateAndKey = &kubernetes.StrimziKafkaClientCertificateAndKey{
				SecretName: literal("connect-user"),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("certificate_and_key without a secret name should fail (required)", func() {
			input.Spec.Authentication = &kubernetes.StrimziKafkaClientAuthentication{
				Type:              "tls",
				CertificateAndKey: &kubernetes.StrimziKafkaClientCertificateAndKey{},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a password secret without a secret name should fail (required)", func() {
			input.Spec.Authentication = scramAuth()
			input.Spec.Authentication.PasswordSecret = &kubernetes.StrimziKafkaClientPasswordSecret{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When OCI plugins are invalid", func() {
		ginkgo.It("duplicate plugin names should fail (spec.plugins.unique_names)", func() {
			input.Spec.Plugins = []*KubernetesKafkaConnectOciPlugin{
				{Name: "debezium", Artifacts: []*KubernetesKafkaConnectOciArtifact{{Reference: "quay.io/a:1"}}},
				{Name: "debezium", Artifacts: []*KubernetesKafkaConnectOciArtifact{{Reference: "quay.io/b:1"}}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an uppercase plugin name should fail (spec.plugins.name.format)", func() {
			input.Spec.Plugins = []*KubernetesKafkaConnectOciPlugin{
				{Name: "Debezium", Artifacts: []*KubernetesKafkaConnectOciArtifact{{Reference: "quay.io/a:1"}}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a plugin without artifacts should fail (min_items)", func() {
			input.Spec.Plugins = []*KubernetesKafkaConnectOciPlugin{{Name: "debezium"}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an artifact without a reference should fail (required)", func() {
			input.Spec.Plugins = []*KubernetesKafkaConnectOciPlugin{
				{Name: "debezium", Artifacts: []*KubernetesKafkaConnectOciArtifact{{}}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown pull policy should fail (spec.plugins.artifacts.pull_policy_enum)", func() {
			input.Spec.Plugins = []*KubernetesKafkaConnectOciPlugin{
				{Name: "debezium", Artifacts: []*KubernetesKafkaConnectOciArtifact{
					{Reference: "quay.io/a:1", PullPolicy: stringPtr("Sometimes")},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When the build block is invalid", func() {
		ginkgo.It("a build without an output should fail (required)", func() {
			input.Spec.Build = jarBuild()
			input.Spec.Build.Output = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a build without plugins should fail (min_items)", func() {
			input.Spec.Build = jarBuild()
			input.Spec.Build.Plugins = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("duplicate build plugin names should fail (spec.build.plugins.unique_names)", func() {
			input.Spec.Build = jarBuild()
			input.Spec.Build.Plugins = append(input.Spec.Build.Plugins, jarBuild().Plugins[0])
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown build output type should fail (spec.build.output.type_enum)", func() {
			input.Spec.Build = jarBuild()
			input.Spec.Build.Output.Type = stringPtr("s3")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a build output without an image should fail (required)", func() {
			input.Spec.Build = jarBuild()
			input.Spec.Build.Output.Image = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an uppercase build plugin name should fail (spec.build.plugins.name.format)", func() {
			input.Spec.Build = jarBuild()
			input.Spec.Build.Plugins[0].Name = "Debezium"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a build plugin without artifacts should fail (min_items)", func() {
			input.Spec.Build = jarBuild()
			input.Spec.Build.Plugins[0].Artifacts = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a missing artifact type should fail (required)", func() {
			input.Spec.Build = jarBuild()
			input.Spec.Build.Plugins[0].Artifacts[0].Type = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown artifact type should fail (spec.build.artifacts.type_enum)", func() {
			input.Spec.Build = jarBuild()
			input.Spec.Build.Plugins[0].Artifacts[0].Type = "rpm"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a jar artifact without a url should fail (spec.build.artifacts.url_required)", func() {
			input.Spec.Build = jarBuild()
			input.Spec.Build.Plugins[0].Artifacts[0].Url = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a maven artifact missing coordinates should fail (spec.build.artifacts.maven_coordinates)", func() {
			input.Spec.Build = jarBuild()
			input.Spec.Build.Plugins[0].Artifacts = []*KubernetesKafkaConnectBuildArtifact{
				{Type: "maven", Group: "io.debezium"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a maven artifact carrying a url should fail (spec.build.artifacts.maven_no_url)", func() {
			input.Spec.Build = jarBuild()
			input.Spec.Build.Plugins[0].Artifacts = []*KubernetesKafkaConnectBuildArtifact{
				{Type: "maven", Group: "io.debezium", Artifact: "debezium-connector-postgres", Version: "3.1.0.Final", Url: "https://example.com/x.jar"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
