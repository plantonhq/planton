package kuberneteskafkauiv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesKafkaUi(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesKafkaUi Suite")
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

// devCluster returns the minimal valid cluster entry every test
// mutates from.
func devCluster() *KubernetesKafkaUiCluster {
	return &KubernetesKafkaUiCluster{
		Name:             "dev",
		BootstrapServers: literal("my-kafka-kafka-bootstrap:9092"),
	}
}

// loginAuth returns a valid login_form auth block with one user.
func loginAuth() *KubernetesKafkaUiAuth {
	return &KubernetesKafkaUiAuth{
		Type: "login_form",
		User: &KubernetesKafkaUiUser{Username: "admin", Password: "s3cret"},
	}
}

var _ = ginkgo.Describe("KubernetesKafkaUi Validation Tests", func() {
	var input *KubernetesKafkaUi

	ginkgo.BeforeEach(func() {
		input = &KubernetesKafkaUi{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesKafkaUi",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-kafka-ui",
			},
			Spec: &KubernetesKafkaUiSpec{
				Namespace: literal("kafka-ui"),
				Clusters:  []*KubernetesKafkaUiCluster{devCluster()},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("bootstrap_servers as a reference should be valid", func() {
			input.Spec.Clusters[0].BootstrapServers = valueFrom(catalogkind.CatalogKind_KubernetesKafka, "my-kafka", "status.outputs.internal_bootstrap_endpoint")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("multiple clusters with distinct names should be valid", func() {
			second := devCluster()
			second.Name = "prod_us-east.1"
			second.ReadOnly = true
			input.Spec.Clusters = append(input.Spec.Clusters, second)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a cluster with TLS trust should be valid", func() {
			input.Spec.Clusters[0].Tls = &KubernetesKafkaUiClusterTls{
				CaSecretName:  literal("my-kafka-cluster-ca-cert"),
				CaCertificate: stringPtr("ca.crt"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("every SASL mechanism with credentials should be valid", func() {
			for _, mechanism := range []string{"PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"} {
				input.Spec.Clusters[0].Sasl = &KubernetesKafkaUiClusterSasl{
					Mechanism: mechanism,
					Username:  "console-user",
					PasswordSecret: &KubernetesKafkaUiPasswordSecret{
						SecretName: literal("console-user"),
						Key:        stringPtr("password"),
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("a schema registry with basic credentials should be valid", func() {
			input.Spec.Clusters[0].SchemaRegistry = &KubernetesKafkaUiSchemaRegistry{
				Url:      literal("http://karapace.kafka.svc.cluster.local:8081"),
				Username: "registry-user",
				PasswordSecret: &KubernetesKafkaUiBasicAuthPasswordSecret{
					SecretName: literal("registry-user"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("Connect clusters with distinct names should be valid", func() {
			input.Spec.Clusters[0].KafkaConnect = []*KubernetesKafkaUiConnectCluster{
				{Name: "cdc", Address: literal("http://my-connect-connect-api.kafka.svc:8083")},
				{Name: "sinks", Address: literal("http://sink-connect-connect-api.kafka.svc:8083")},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("login_form auth with the single user should be valid", func() {
			input.Spec.Auth = loginAuth()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("every service type should be valid", func() {
			for _, serviceType := range []string{"ClusterIP", "NodePort", "LoadBalancer"} {
				input.Spec.ServiceType = stringPtr(serviceType)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("full surface (replicas, port, registry override, properties, helm values) should be valid", func() {
			input.Spec.ChartVersion = "1.5.1"
			input.Spec.Replicas = int32Ptr(2)
			input.Spec.ServicePort = int32Ptr(8080)
			input.Spec.ImageRegistry = "mirror.example.com"
			input.Spec.HelmValues = "podAnnotations:\n  team: data\n"
			input.Spec.NodeSelector = map[string]string{"disktype": "ssd"}
			input.Spec.Clusters[0].Properties = map[string]string{"request.timeout.ms": "30000"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When top-level fields are invalid", func() {
		ginkgo.It("missing namespace should fail (required)", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an empty clusters list should fail (min_items)", func() {
			input.Spec.Clusters = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("duplicate cluster names should fail (spec.clusters.unique_names)", func() {
			input.Spec.Clusters = []*KubernetesKafkaUiCluster{devCluster(), devCluster()}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero replicas should fail (gte 1)", func() {
			input.Spec.Replicas = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown service type should fail (spec.service_type_enum)", func() {
			input.Spec.ServiceType = stringPtr("ExternalName")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("service port zero should fail (gte 1)", func() {
			input.Spec.ServicePort = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("service port above 65535 should fail (lte 65535)", func() {
			input.Spec.ServicePort = int32Ptr(70000)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When cluster entries are invalid", func() {
		ginkgo.It("a missing cluster name should fail (required)", func() {
			input.Spec.Clusters[0].Name = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a cluster name with a space should fail (spec.clusters.name.format)", func() {
			input.Spec.Clusters[0].Name = "prod cluster"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("missing bootstrap_servers should fail (required)", func() {
			input.Spec.Clusters[0].BootstrapServers = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a tls block without a CA secret name should fail (required)", func() {
			input.Spec.Clusters[0].Tls = &KubernetesKafkaUiClusterTls{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a schema registry without a url should fail (required)", func() {
			input.Spec.Clusters[0].SchemaRegistry = &KubernetesKafkaUiSchemaRegistry{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("duplicate Connect cluster names should fail (spec.clusters.kafka_connect.unique_names)", func() {
			input.Spec.Clusters[0].KafkaConnect = []*KubernetesKafkaUiConnectCluster{
				{Name: "cdc", Address: literal("http://a.kafka.svc:8083")},
				{Name: "cdc", Address: literal("http://b.kafka.svc:8083")},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a Connect cluster without a name should fail (required)", func() {
			input.Spec.Clusters[0].KafkaConnect = []*KubernetesKafkaUiConnectCluster{
				{Address: literal("http://a.kafka.svc:8083")},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a Connect cluster without an address should fail (required)", func() {
			input.Spec.Clusters[0].KafkaConnect = []*KubernetesKafkaUiConnectCluster{
				{Name: "cdc"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When SASL credentials are invalid", func() {
		ginkgo.It("a missing mechanism should fail (required)", func() {
			input.Spec.Clusters[0].Sasl = &KubernetesKafkaUiClusterSasl{
				Username: "console-user",
				PasswordSecret: &KubernetesKafkaUiPasswordSecret{
					SecretName: literal("console-user"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown mechanism should fail (spec.clusters.sasl.mechanism_enum)", func() {
			input.Spec.Clusters[0].Sasl = &KubernetesKafkaUiClusterSasl{
				Mechanism: "OAUTHBEARER",
				Username:  "console-user",
				PasswordSecret: &KubernetesKafkaUiPasswordSecret{
					SecretName: literal("console-user"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a missing username should fail (required)", func() {
			input.Spec.Clusters[0].Sasl = &KubernetesKafkaUiClusterSasl{
				Mechanism: "SCRAM-SHA-512",
				PasswordSecret: &KubernetesKafkaUiPasswordSecret{
					SecretName: literal("console-user"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a missing password secret should fail (required)", func() {
			input.Spec.Clusters[0].Sasl = &KubernetesKafkaUiClusterSasl{
				Mechanism: "SCRAM-SHA-512",
				Username:  "console-user",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a password secret without a secret name should fail (required)", func() {
			input.Spec.Clusters[0].Sasl = &KubernetesKafkaUiClusterSasl{
				Mechanism:      "SCRAM-SHA-512",
				Username:       "console-user",
				PasswordSecret: &KubernetesKafkaUiPasswordSecret{},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When auth is invalid", func() {
		ginkgo.It("a missing auth type should fail (required)", func() {
			input.Spec.Auth = loginAuth()
			input.Spec.Auth.Type = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown auth type should fail (spec.auth.type_enum)", func() {
			input.Spec.Auth = loginAuth()
			input.Spec.Auth.Type = "oauth2"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("login_form auth without a user should fail (required)", func() {
			input.Spec.Auth = loginAuth()
			input.Spec.Auth.User = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a user without a username should fail (required)", func() {
			input.Spec.Auth = loginAuth()
			input.Spec.Auth.User = &KubernetesKafkaUiUser{Password: "s3cret"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a user without a password should fail (required)", func() {
			input.Spec.Auth = loginAuth()
			input.Spec.Auth.User = &KubernetesKafkaUiUser{Username: "admin"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
