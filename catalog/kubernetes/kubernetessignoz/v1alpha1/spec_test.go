package kubernetessignozv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesSignoz(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesSignoz Suite")
}

func int32Ptr(i int32) *int32 { return &i }

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

// testClickHouse returns the required telemetry-store connection shaped
// like a KubernetesClickHouse composition with literal values.
func testClickHouse() *KubernetesSignozClickHouse {
	return &KubernetesSignozClickHouse{
		Host:     literal("clickhouse-analytics.data.svc.cluster.local"),
		Username: "signoz",
		PasswordSecret: &KubernetesSignozClickHousePassword{
			SecretName: literal("analytics-clickhouse-auth"),
			SecretKey:  "signoz",
		},
	}
}

var _ = ginkgo.Describe("KubernetesSignoz Validation Tests", func() {
	var input *KubernetesSignoz

	ginkgo.BeforeEach(func() {
		input = &KubernetesSignoz{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesSignoz",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "observe",
			},
			Spec: &KubernetesSignozSpec{
				Namespace:  literal("observability"),
				Clickhouse: testClickHouse(),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec (namespace + clickhouse connection) should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "observability", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a clickhouse connection composed entirely from KubernetesClickHouse references should be valid", func() {
			input.Spec.Clickhouse = &KubernetesSignozClickHouse{
				Host:        valueFrom(catalogkind.CatalogKind_KubernetesClickHouse, "analytics", "status.outputs.service_name"),
				ClusterName: valueFrom(catalogkind.CatalogKind_KubernetesClickHouse, "analytics", "status.outputs.cluster_name"),
				Username:    "signoz",
				PasswordSecret: &KubernetesSignozClickHousePassword{
					SecretName: valueFrom(catalogkind.CatalogKind_KubernetesClickHouse, "analytics", "status.outputs.auth_secret_name"),
					SecretKey:  "signoz",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a clickhouse connection over verified TLS should be valid", func() {
			input.Spec.Clickhouse.Secure = true
			input.Spec.Clickhouse.Verify = true
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("explicit non-default clickhouse ports should be valid", func() {
			input.Spec.Clickhouse.TcpPort = int32Ptr(19000)
			input.Spec.Clickhouse.HttpPort = int32Ptr(18123)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a fully declared server block should be valid", func() {
			input.Spec.Server = &KubernetesSignozServer{
				DiskSize:    stringPtr("2Gi"),
				ExternalUrl: "https://signoz.example.com",
				Smtp: &KubernetesSignozSmtp{
					Address:  "smtp.example.com:587",
					From:     "signoz@example.com",
					Username: "smtp-user",
					PasswordSecret: &KubernetesSignozSecretKeyRef{
						Name: "smtp-auth",
						Key:  "password",
					},
				},
				Env: map[string]string{
					"signoz_prometheus_active__query__tracker_enabled": "true",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("unauthenticated SMTP (an internal relay) should be valid", func() {
			input.Spec.Server = &KubernetesSignozServer{
				Smtp: &KubernetesSignozSmtp{
					Address: "mail-relay.internal:25",
					From:    "signoz@example.com",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("collector receiver toggles should be valid", func() {
			jaegerOff := false
			input.Spec.OtelCollector = &KubernetesSignozOtelCollector{
				JaegerReceiverEnabled: &jaegerOff,
				ZipkinReceiverEnabled: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("collector autoscaling with sane bounds should be valid", func() {
			input.Spec.OtelCollector = &KubernetesSignozOtelCollector{
				Autoscaling: &KubernetesSignozOtelCollectorAutoscaling{
					Enabled:                        true,
					MinReplicas:                    int32Ptr(2),
					MaxReplicas:                    int32Ptr(10),
					TargetCpuUtilizationPercent:    int32Ptr(60),
					TargetMemoryUtilizationPercent: int32Ptr(70),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a spec without the clickhouse connection should fail (composed, never bundled)", func() {
			input.Spec.Clickhouse = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a clickhouse connection without a host should fail", func() {
			input.Spec.Clickhouse.Host = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a clickhouse connection without a username should fail", func() {
			input.Spec.Clickhouse.Username = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a clickhouse connection without a password secret should fail", func() {
			input.Spec.Clickhouse.PasswordSecret = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a password secret without a key should fail", func() {
			input.Spec.Clickhouse.PasswordSecret.SecretKey = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a password secret without a name should fail", func() {
			input.Spec.Clickhouse.PasswordSecret.SecretName = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("certificate verification without TLS should fail", func() {
			input.Spec.Clickhouse.Verify = true
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an out-of-range clickhouse tcp port should fail", func() {
			input.Spec.Clickhouse.TcpPort = int32Ptr(70000)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a zero clickhouse http port should fail", func() {
			input.Spec.Clickhouse.HttpPort = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an SMTP address without a port should fail", func() {
			input.Spec.Server = &KubernetesSignozServer{
				Smtp: &KubernetesSignozSmtp{
					Address: "smtp.example.com",
					From:    "signoz@example.com",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("SMTP without a From address should fail", func() {
			input.Spec.Server = &KubernetesSignozServer{
				Smtp: &KubernetesSignozSmtp{Address: "smtp.example.com:587"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an SMTP password without a username should fail", func() {
			input.Spec.Server = &KubernetesSignozServer{
				Smtp: &KubernetesSignozSmtp{
					Address: "smtp.example.com:587",
					From:    "signoz@example.com",
					PasswordSecret: &KubernetesSignozSecretKeyRef{
						Name: "smtp-auth",
						Key:  "password",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a malformed server disk size should fail", func() {
			input.Spec.Server = &KubernetesSignozServer{DiskSize: stringPtr("big")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero collector replicas should fail", func() {
			input.Spec.OtelCollector = &KubernetesSignozOtelCollector{Replicas: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("collector autoscaling with max below min should fail", func() {
			input.Spec.OtelCollector = &KubernetesSignozOtelCollector{
				Autoscaling: &KubernetesSignozOtelCollectorAutoscaling{
					Enabled:     true,
					MinReplicas: int32Ptr(5),
					MaxReplicas: int32Ptr(2),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
