package azurecontainerappenvironmentdaprcomponentv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureContainerAppEnvironmentDaprComponentSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureContainerAppEnvironmentDaprComponentSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func ref(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

func strPtr(s string) *string { return &s }

// minimalSpec returns a valid state-store component.
func minimalSpec() *AzureContainerAppEnvironmentDaprComponent {
	return &AzureContainerAppEnvironmentDaprComponent{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureContainerAppEnvironmentDaprComponent",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-dapr-component",
		},
		Spec: &AzureContainerAppEnvironmentDaprComponentSpec{
			ContainerAppEnvironmentId: literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.App/managedEnvironments/env"),
			ComponentName:             "statestore",
			ComponentType:             "state.azure.blobstorage",
			Version:                   "v1",
		},
	}
}

var _ = ginkgo.Describe("AzureContainerAppEnvironmentDaprComponentSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("accepts a minimal component", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("accepts metadata with literal values and secret references", func() {
			input := minimalSpec()
			input.Spec.Secrets = []*AzureContainerAppEnvironmentDaprComponentSecret{{
				Name:  "account-key",
				Value: "base64key==",
			}}
			input.Spec.Metadata = []*AzureContainerAppEnvironmentDaprComponentMetadata{
				{Name: "accountName", Value: literal("mystorageaccount")},
				{Name: "containerName", Value: literal("state")},
				{Name: "accountKey", SecretName: "account-key"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a metadata value referenced from another resource's output", func() {
			// The keyless-auth shape: azureClientId tracks a managed
			// identity's client_id output instead of shipping a
			// connection-string secret.
			input := minimalSpec()
			input.Spec.Metadata = []*AzureContainerAppEnvironmentDaprComponentMetadata{
				{Name: "namespaceName", Value: literal("my-bus.servicebus.windows.net")},
				{Name: "azureClientId", Value: ref("workload-identity")},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts scopes, init timeout, and ignore_errors", func() {
			input := minimalSpec()
			input.Spec.Scopes = []string{"orders", "billing"}
			input.Spec.InitTimeout = strPtr("10m")
			ignoreErrors := true
			input.Spec.IgnoreErrors = &ignoreErrors
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a single-letter component name", func() {
			input := minimalSpec()
			input.Spec.ComponentName = "s"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("rejects a missing environment reference", func() {
			input := minimalSpec()
			input.Spec.ContainerAppEnvironmentId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a missing component name", func() {
			input := minimalSpec()
			input.Spec.ComponentName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a component name starting with a digit", func() {
			input := minimalSpec()
			input.Spec.ComponentName = "1store"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a component name with consecutive hyphens", func() {
			input := minimalSpec()
			input.Spec.ComponentName = "state--store"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a missing component type", func() {
			input := minimalSpec()
			input.Spec.ComponentType = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a missing version", func() {
			input := minimalSpec()
			input.Spec.Version = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an init timeout without a unit", func() {
			input := minimalSpec()
			input.Spec.InitTimeout = strPtr("30")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an init timeout with a fractional value", func() {
			input := minimalSpec()
			input.Spec.InitTimeout = strPtr("1.5m")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a metadata entry carrying both a value and a secret reference", func() {
			input := minimalSpec()
			input.Spec.Metadata = []*AzureContainerAppEnvironmentDaprComponentMetadata{{
				Name:       "accountKey",
				Value:      literal("literal"),
				SecretName: "account-key",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a secret without a value", func() {
			input := minimalSpec()
			input.Spec.Secrets = []*AzureContainerAppEnvironmentDaprComponentSecret{{
				Name: "account-key",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a secret with an invalid name", func() {
			input := minimalSpec()
			input.Spec.Secrets = []*AzureContainerAppEnvironmentDaprComponentSecret{{
				Name:  "Account_Key",
				Value: "base64key==",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
