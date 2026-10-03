package kuberneteskafkav1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesKafka(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesKafka Suite")
}

func int32Ptr(i int32) *int32    { return &i }
func boolPtr(b bool) *bool       { return &b }
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

// dualRolePool returns the minimal valid pool — one node carrying both
// KRaft roles with persistent storage (the dev shape every test mutates
// from).
func dualRolePool() *KubernetesKafkaNodePool {
	return &KubernetesKafkaNodePool{
		Name:     "dual-role",
		Roles:    []string{"controller", "broker"},
		Replicas: 1,
		Storage:  &KubernetesKafkaStorage{Size: "5Gi"},
	}
}

// plainListener returns the minimal valid listener.
func plainListener() *KubernetesKafkaListener {
	return &KubernetesKafkaListener{
		Name: "plain",
		Port: 9092,
	}
}

var _ = ginkgo.Describe("KubernetesKafka Validation Tests", func() {
	var input *KubernetesKafka

	ginkgo.BeforeEach(func() {
		input = &KubernetesKafka{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesKafka",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-kafka",
			},
			Spec: &KubernetesKafkaSpec{
				Namespace: literal("kafka"),
				NodePools: []*KubernetesKafkaNodePool{dualRolePool()},
				Listeners: []*KubernetesKafkaListener{plainListener()},
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

		ginkgo.It("separate controller and broker pools should be valid (production shape)", func() {
			input.Spec.NodePools = []*KubernetesKafkaNodePool{
				{Name: "controller", Roles: []string{"controller"}, Replicas: 3, Storage: &KubernetesKafkaStorage{Size: "10Gi"}},
				{Name: "broker", Roles: []string{"broker"}, Replicas: 3, Storage: &KubernetesKafkaStorage{Size: "100Gi"}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("ephemeral storage without a size should be valid (dev shape)", func() {
			input.Spec.NodePools[0].Storage = &KubernetesKafkaStorage{Type: stringPtr("ephemeral")}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("jbod storage with volumes should be valid", func() {
			input.Spec.NodePools[0].Storage = &KubernetesKafkaStorage{
				Type: stringPtr("jbod"),
				Volumes: []*KubernetesKafkaStorageVolume{
					{Id: 0, Size: "100Gi", KraftMetadata: true},
					{Id: 1, Size: "100Gi"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("every listener exposure type should be valid on a TLS listener with hosts where required", func() {
			for _, listenerType := range []string{"internal", "cluster-ip", "nodeport", "loadbalancer", "route"} {
				input.Spec.Listeners = []*KubernetesKafkaListener{{
					Name: "lst",
					Port: 9095,
					Type: stringPtr(listenerType),
					Tls:  true,
				}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("an ingress listener with bootstrap and broker hosts should be valid (the operator's own requirement)", func() {
			input.Spec.Listeners = []*KubernetesKafkaListener{{
				Name: "ext",
				Port: 9095,
				Type: stringPtr("ingress"),
				Tls:  true,
				Configuration: &KubernetesKafkaListenerConfiguration{
					Class:     "nginx",
					Bootstrap: &KubernetesKafkaListenerBootstrap{Host: "kafka.example.com"},
					Brokers: []*KubernetesKafkaListenerBroker{
						{Broker: 0, Host: "b0.kafka.example.com"},
					},
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("scram authentication on a plaintext listener should be valid (SASL_PLAINTEXT)", func() {
			input.Spec.Listeners[0].Authentication = &KubernetesKafkaListenerAuthentication{Type: "scram-sha-512"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("simple authorization with super users should be valid", func() {
			input.Spec.Authorization = &KubernetesKafkaAuthorization{
				Type:       "simple",
				SuperUsers: []string{"User:CN=admin"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("custom authorization with an authorizer class should be valid", func() {
			input.Spec.Authorization = &KubernetesKafkaAuthorization{
				Type:            "custom",
				AuthorizerClass: "com.example.MyAuthorizer",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full surface (cruise control, exporter, metrics, cas, rack, jvm, windows) should be valid", func() {
			input.Spec.CruiseControl = &KubernetesKafkaCruiseControl{
				Enabled:            true,
				Config:             map[string]string{"hard.goals": "RackAwareGoal"},
				AutoRebalanceModes: []string{"add-brokers", "remove-brokers"},
			}
			input.Spec.KafkaExporter = &KubernetesKafkaExporter{Enabled: true, TopicRegex: ".*"}
			input.Spec.Metrics = &KubernetesKafkaMetrics{Enabled: true}
			input.Spec.ClusterCa = &KubernetesKafkaCa{ValidityDays: int32Ptr(730), RenewalDays: int32Ptr(60)}
			input.Spec.ClientsCa = &KubernetesKafkaCa{ValidityDays: int32Ptr(365), RenewalDays: int32Ptr(30)}
			input.Spec.Rack = &KubernetesKafkaRack{TopologyKey: "topology.kubernetes.io/zone"}
			input.Spec.Jvm = &KubernetesKafkaJvm{Xms: "2g", Xmx: "2g"}
			input.Spec.MaintenanceTimeWindows = []string{"* * 0-3 ? * SUN"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When node pools are invalid", func() {
		ginkgo.It("an empty node pool list should fail (min_items)", func() {
			input.Spec.NodePools = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("no controller role anywhere should fail (spec.node_pools.controller_role_present)", func() {
			input.Spec.NodePools = []*KubernetesKafkaNodePool{
				{Name: "broker", Roles: []string{"broker"}, Replicas: 3, Storage: &KubernetesKafkaStorage{Size: "10Gi"}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("no broker role anywhere should fail (spec.node_pools.broker_role_present)", func() {
			input.Spec.NodePools = []*KubernetesKafkaNodePool{
				{Name: "controller", Roles: []string{"controller"}, Replicas: 3, Storage: &KubernetesKafkaStorage{Size: "10Gi"}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("duplicate pool names should fail (spec.node_pools.unique_names)", func() {
			input.Spec.NodePools = []*KubernetesKafkaNodePool{
				{Name: "pool", Roles: []string{"controller"}, Replicas: 1, Storage: &KubernetesKafkaStorage{Size: "1Gi"}},
				{Name: "pool", Roles: []string{"broker"}, Replicas: 1, Storage: &KubernetesKafkaStorage{Size: "1Gi"}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an invalid role should fail (spec.node_pools.roles.enum)", func() {
			input.Spec.NodePools[0].Roles = []string{"controller", "zookeeper"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a repeated role within one pool should fail (spec.node_pools.roles.unique)", func() {
			input.Spec.NodePools[0].Roles = []string{"broker", "broker"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an uppercase pool name should fail (DNS-1123 label)", func() {
			input.Spec.NodePools[0].Name = "Brokers"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero replicas should fail (gte 1)", func() {
			input.Spec.NodePools[0].Replicas = 0
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("persistent-claim storage without a size should fail (spec.storage.persistent_requires_size)", func() {
			input.Spec.NodePools[0].Storage = &KubernetesKafkaStorage{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("jbod storage without volumes should fail (spec.storage.jbod_requires_volumes)", func() {
			input.Spec.NodePools[0].Storage = &KubernetesKafkaStorage{Type: stringPtr("jbod")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("volumes on non-jbod storage should fail (spec.storage.volumes_only_for_jbod)", func() {
			input.Spec.NodePools[0].Storage = &KubernetesKafkaStorage{
				Size:    "10Gi",
				Volumes: []*KubernetesKafkaStorageVolume{{Id: 0, Size: "10Gi"}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown storage type should fail (spec.storage.type_enum)", func() {
			input.Spec.NodePools[0].Storage = &KubernetesKafkaStorage{Type: stringPtr("nfs"), Size: "10Gi"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When listeners are invalid", func() {
		ginkgo.It("an empty listener list should fail (min_items)", func() {
			input.Spec.Listeners = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("duplicate listener names should fail (spec.listeners.unique_names)", func() {
			input.Spec.Listeners = []*KubernetesKafkaListener{
				{Name: "plain", Port: 9092},
				{Name: "plain", Port: 9093},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("duplicate listener ports should fail (spec.listeners.unique_ports)", func() {
			input.Spec.Listeners = []*KubernetesKafkaListener{
				{Name: "one", Port: 9092},
				{Name: "two", Port: 9092},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a listener name over 11 characters should fail (spec.listeners.name.format)", func() {
			input.Spec.Listeners[0].Name = "waytoolongname"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a listener name with a dash should fail (lowercase alphanumerics only)", func() {
			input.Spec.Listeners[0].Name = "my-listener"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a port below 9092 should fail (reserved for internal listeners)", func() {
			input.Spec.Listeners[0].Port = 9091
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown listener type should fail (spec.listeners.type_enum)", func() {
			input.Spec.Listeners[0].Type = stringPtr("hostport")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an ingress listener without TLS should fail (spec.listeners.tls_required_types)", func() {
			input.Spec.Listeners = []*KubernetesKafkaListener{{
				Name: "ext",
				Port: 9095,
				Type: stringPtr("ingress"),
				Tls:  false,
				Configuration: &KubernetesKafkaListenerConfiguration{
					Bootstrap: &KubernetesKafkaListenerBootstrap{Host: "kafka.example.com"},
					Brokers:   []*KubernetesKafkaListenerBroker{{Broker: 0, Host: "b0.example.com"}},
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("mutual-TLS authentication on a plaintext listener should fail (spec.listeners.tls_auth_requires_tls)", func() {
			input.Spec.Listeners[0].Authentication = &KubernetesKafkaListenerAuthentication{Type: "tls"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an ingress listener without hosts should fail (spec.listeners.ingress_requires_hosts)", func() {
			input.Spec.Listeners = []*KubernetesKafkaListener{{
				Name: "ext",
				Port: 9095,
				Type: stringPtr("ingress"),
				Tls:  true,
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown authentication type should fail (spec.listeners.authentication.type_enum)", func() {
			input.Spec.Listeners[0].Tls = true
			input.Spec.Listeners[0].Authentication = &KubernetesKafkaListenerAuthentication{Type: "oauth"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an invalid externalTrafficPolicy should fail (spec.listeners.configuration.etp_enum)", func() {
			input.Spec.Listeners[0].Configuration = &KubernetesKafkaListenerConfiguration{
				ExternalTrafficPolicy: stringPtr("Regional"),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an invalid preferred node-port address type should fail (spec.listeners.configuration.npat_enum)", func() {
			input.Spec.Listeners[0].Configuration = &KubernetesKafkaListenerConfiguration{
				PreferredNodePortAddressType: stringPtr("PublicIP"),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("duplicate broker IDs in a listener's brokers list should fail (brokers.unique_ids)", func() {
			input.Spec.Listeners[0].Configuration = &KubernetesKafkaListenerConfiguration{
				Brokers: []*KubernetesKafkaListenerBroker{
					{Broker: 0, AdvertisedHost: "a"},
					{Broker: 0, AdvertisedHost: "b"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When cluster-level blocks are invalid", func() {
		ginkgo.It("an unknown authorization type should fail (spec.authorization.type_enum)", func() {
			input.Spec.Authorization = &KubernetesKafkaAuthorization{Type: "keycloak"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("custom authorization without an authorizer class should fail (custom_requires_class)", func() {
			input.Spec.Authorization = &KubernetesKafkaAuthorization{Type: "custom"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an authorizer class on simple authorization should fail (class_only_for_custom)", func() {
			input.Spec.Authorization = &KubernetesKafkaAuthorization{
				Type:            "simple",
				AuthorizerClass: "com.example.MyAuthorizer",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown auto-rebalance mode should fail (auto_rebalance_modes_enum)", func() {
			input.Spec.CruiseControl = &KubernetesKafkaCruiseControl{
				Enabled:            true,
				AutoRebalanceModes: []string{"rebalance-everything"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero CA validity days should fail (gte 1)", func() {
			input.Spec.ClusterCa = &KubernetesKafkaCa{ValidityDays: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a rack block without a topology key should fail (required)", func() {
			input.Spec.Rack = &KubernetesKafkaRack{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
