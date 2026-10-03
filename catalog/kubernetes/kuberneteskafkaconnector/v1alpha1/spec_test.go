package kuberneteskafkaconnectorv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesKafkaConnector(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesKafkaConnector Suite")
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

var _ = ginkgo.Describe("KubernetesKafkaConnector Validation Tests", func() {
	var input *KubernetesKafkaConnector

	ginkgo.BeforeEach(func() {
		input = &KubernetesKafkaConnector{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesKafkaConnector",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "orders-source",
			},
			Spec: &KubernetesKafkaConnectorSpec{
				Namespace:      literal("kafka"),
				ConnectCluster: literal("my-connect"),
				ConnectorClass: "org.apache.kafka.connect.file.FileStreamSourceConnector",
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("connect_cluster as a reference should be valid", func() {
			input.Spec.ConnectCluster = valueFrom(catalogkind.CatalogKind_KubernetesKafkaConnect, "my-connect", "status.outputs.connect_name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "kafka", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("every connector state should be valid", func() {
			for _, state := range []string{"running", "paused", "stopped"} {
				input.Spec.State = stringPtr(state)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("connector config without the retired plugin-version key should be valid", func() {
			input.Spec.Config = map[string]string{
				"file":  "/tmp/source.txt",
				"topic": "orders",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("auto restart with a max-restarts bound should be valid", func() {
			input.Spec.AutoRestart = &KubernetesKafkaConnectorAutoRestart{Enabled: true, MaxRestarts: int32Ptr(5)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full surface (tasks_max, version, offset targets) should be valid", func() {
			input.Spec.TasksMax = int32Ptr(4)
			input.Spec.Version = "3.1.0"
			input.Spec.ListOffsets = &KubernetesKafkaConnectorListOffsets{ToConfigMap: literal("orders-offsets")}
			input.Spec.AlterOffsets = &KubernetesKafkaConnectorAlterOffsets{FromConfigMap: literal("orders-offsets-override")}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail (required)", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("missing connect_cluster should fail (required)", func() {
			input.Spec.ConnectCluster = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("missing connector_class should fail (required)", func() {
			input.Spec.ConnectorClass = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero tasks_max should fail (gte 1)", func() {
			input.Spec.TasksMax = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("connector.plugin.version in config should fail (spec.config.no_connector_plugin_version)", func() {
			input.Spec.Config = map[string]string{"connector.plugin.version": "3.1.0"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown state should fail (spec.state_enum)", func() {
			input.Spec.State = stringPtr("suspended")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero max_restarts should fail (gte 1)", func() {
			input.Spec.AutoRestart = &KubernetesKafkaConnectorAutoRestart{Enabled: true, MaxRestarts: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("list_offsets without a target ConfigMap should fail (required)", func() {
			input.Spec.ListOffsets = &KubernetesKafkaConnectorListOffsets{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("alter_offsets without a source ConfigMap should fail (required)", func() {
			input.Spec.AlterOffsets = &KubernetesKafkaConnectorAlterOffsets{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
