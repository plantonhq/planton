package kubernetesnatsv1alpha1

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

func TestKubernetesNats(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesNats Suite")
}

func int32Ptr(i int32) *int32 { return &i }
func boolPtr(b bool) *bool    { return &b }
func strPtr(s string) *string { return &s }

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

func user(name string) *KubernetesNatsUser {
	return &KubernetesNatsUser{Username: name}
}

var _ = ginkgo.Describe("KubernetesNats Validation Tests", func() {
	var input *KubernetesNats

	ginkgo.BeforeEach(func() {
		input = &KubernetesNats{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesNats",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "messaging",
			},
			Spec: &KubernetesNatsSpec{
				Namespace: literal("nats"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec (every optional block omitted) should be valid", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "nats", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a 3-server cluster should be valid", func() {
			input.Spec.Cluster = &KubernetesNatsCluster{Enabled: true, Replicas: int32Ptr(3)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("jetstream sizing with a storage-class reference should be valid", func() {
			input.Spec.JetStream = &KubernetesNatsJetStream{
				DiskSize:           strPtr("20Gi"),
				StorageClass:       valueFrom(catalogkind.CatalogKind_KubernetesStorageClass, "fast-ssd", "metadata.name"),
				MaxFileStore:       strPtr("18Gi"),
				MemoryStoreMaxSize: strPtr("1Gi"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("flat users with permissions should be valid", func() {
			input.Spec.Auth = &KubernetesNatsAuth{
				Users: []*KubernetesNatsUser{
					{Username: "orders-service", Permissions: &KubernetesNatsPermissions{
						PublishAllow:   []string{"orders.>"},
						SubscribeAllow: []string{"orders.>", "_INBOX.>"},
					}},
					{Username: "auditor", Permissions: &KubernetesNatsPermissions{
						SubscribeAllow: []string{">"},
						PublishDeny:    []string{">"},
					}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("multi-tenant accounts should be valid", func() {
			input.Spec.Auth = &KubernetesNatsAuth{
				Accounts: []*KubernetesNatsAccount{
					{Name: "team-a", Users: []*KubernetesNatsUser{user("svc-a")}, JetStreamEnabled: true},
					{Name: "team-b", Users: []*KubernetesNatsUser{user("svc-b")}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a no_auth_user naming a declared flat user should be valid", func() {
			input.Spec.Auth = &KubernetesNatsAuth{
				Users:      []*KubernetesNatsUser{user("app"), user("guest")},
				NoAuthUser: "guest",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a no_auth_user naming an account user should be valid", func() {
			input.Spec.Auth = &KubernetesNatsAuth{
				Accounts: []*KubernetesNatsAccount{
					{Name: "public", Users: []*KubernetesNatsUser{user("guest")}},
				},
				NoAuthUser: "guest",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("tls with a certificate reference should be valid", func() {
			input.Spec.Tls = &KubernetesNatsTls{
				SecretName:    valueFrom(catalogkind.CatalogKind_KubernetesCertificate, "nats-cert", "status.outputs.secret_name"),
				VerifyClients: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("websocket on a custom port should be valid", func() {
			input.Spec.Websocket = &KubernetesNatsWebsocket{Enabled: true, Port: int32Ptr(9222)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("mqtt with default jetstream should be valid", func() {
			input.Spec.Mqtt = &KubernetesNatsMqtt{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("mqtt with jetstream explicitly enabled should be valid", func() {
			input.Spec.JetStream = &KubernetesNatsJetStream{Enabled: boolPtr(true)}
			input.Spec.Mqtt = &KubernetesNatsMqtt{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("leafnodes should be valid", func() {
			input.Spec.Leafnodes = &KubernetesNatsLeafnodes{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("metrics with exporter and pod monitor should be valid", func() {
			input.Spec.Metrics = &KubernetesNatsMetrics{ExporterEnabled: true, PodMonitorEnabled: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a LoadBalancer service with annotations should be valid", func() {
			input.Spec.Service = &KubernetesNatsService{
				Type: KubernetesNatsService_load_balancer,
				Annotations: map[string]string{
					"service.beta.kubernetes.io/aws-load-balancer-type": "nlb",
					"external-dns.alpha.kubernetes.io/hostname":         "nats.example.com",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("image overrides should be valid", func() {
			input.Spec.Images = &KubernetesNatsImages{
				Nats: &kubernetes.ContainerImage{Repo: "mirror.example.com/nats", Tag: "2.14.2-alpine"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a cluster of 1 should fail (floor is 2)", func() {
			input.Spec.Cluster = &KubernetesNatsCluster{Enabled: true, Replicas: int32Ptr(1)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a cluster of 10 should fail (cap is 9)", func() {
			input.Spec.Cluster = &KubernetesNatsCluster{Enabled: true, Replicas: int32Ptr(10)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a bad jetstream disk size should fail", func() {
			input.Spec.JetStream = &KubernetesNatsJetStream{DiskSize: strPtr("ten gigs")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("declaring users AND accounts together should fail", func() {
			input.Spec.Auth = &KubernetesNatsAuth{
				Users:    []*KubernetesNatsUser{user("app")},
				Accounts: []*KubernetesNatsAccount{{Name: "team-a", Users: []*KubernetesNatsUser{user("svc-a")}}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an empty auth block should fail", func() {
			input.Spec.Auth = &KubernetesNatsAuth{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a no_auth_user that is not declared should fail", func() {
			input.Spec.Auth = &KubernetesNatsAuth{
				Users:      []*KubernetesNatsUser{user("app")},
				NoAuthUser: "guest",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("duplicate flat usernames should fail", func() {
			input.Spec.Auth = &KubernetesNatsAuth{
				Users: []*KubernetesNatsUser{user("app"), user("app")},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("the same username in two accounts should fail (auth Secret keys collide)", func() {
			input.Spec.Auth = &KubernetesNatsAuth{
				Accounts: []*KubernetesNatsAccount{
					{Name: "team-a", Users: []*KubernetesNatsUser{user("svc")}},
					{Name: "team-b", Users: []*KubernetesNatsUser{user("svc")}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an account without users should fail", func() {
			input.Spec.Auth = &KubernetesNatsAuth{
				Accounts: []*KubernetesNatsAccount{{Name: "team-a"}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("mqtt with jetstream explicitly disabled should fail", func() {
			input.Spec.JetStream = &KubernetesNatsJetStream{Enabled: boolPtr(false)}
			input.Spec.Mqtt = &KubernetesNatsMqtt{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a pod monitor without the exporter should fail", func() {
			input.Spec.Metrics = &KubernetesNatsMetrics{PodMonitorEnabled: true}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("tls without a secret name should fail", func() {
			input.Spec.Tls = &KubernetesNatsTls{VerifyClients: true}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a bad username should fail", func() {
			input.Spec.Auth = &KubernetesNatsAuth{
				Users: []*KubernetesNatsUser{user("bad user!")},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a websocket port out of range should fail", func() {
			input.Spec.Websocket = &KubernetesNatsWebsocket{Enabled: true, Port: int32Ptr(70000)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a bad account name should fail", func() {
			input.Spec.Auth = &KubernetesNatsAuth{
				Accounts: []*KubernetesNatsAccount{{Name: "team a", Users: []*KubernetesNatsUser{user("svc")}}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
