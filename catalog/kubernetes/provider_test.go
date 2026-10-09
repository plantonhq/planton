package kubernetes

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestKubernetesProviderConfig(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesProviderConfig Suite")
}

func gkeConfig() *KubernetesProviderConfigGcpGke {
	return &KubernetesProviderConfigGcpGke{
		ClusterEndpoint: "https://34.100.155.147",
		ClusterCaData:   "dGVzdC1jYS1kYXRh",
	}
}

func aksConfig() *KubernetesProviderConfigAzureAks {
	return &KubernetesProviderConfigAzureAks{
		ClusterEndpoint: "https://demo-aks.hcp.eastus.azmk8s.io",
		ClusterCaData:   "dGVzdC1jYS1kYXRh",
		TenantId:        "11111111-2222-3333-4444-555555555555",
	}
}

var _ = ginkgo.Describe("KubernetesProviderConfigGcpGke credentials", func() {

	ginkgo.Context("each credential alone, or none (the ambient chain)", func() {
		ginkgo.It("should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(gkeConfig())).To(gomega.BeNil())

			withKey := gkeConfig()
			withKey.ServiceAccountKey = `{"type":"service_account"}`
			gomega.Expect(protovalidate.Validate(withKey)).To(gomega.BeNil())

			withToken := gkeConfig()
			withToken.AccessToken = "ya29.test-token"
			gomega.Expect(protovalidate.Validate(withToken)).To(gomega.BeNil())
		})
	})

	ginkgo.Context("a service-account key and an access token together", func() {
		ginkgo.It("should be refused by kubernetes.gcp_gke.one_credential", func() {
			input := gkeConfig()
			input.ServiceAccountKey = `{"type":"service_account"}`
			input.AccessToken = "ya29.test-token"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("set at most one of service_account_key and access_token"))
		})
	})
})

var _ = ginkgo.Describe("KubernetesProviderConfigAzureAks credentials", func() {

	ginkgo.Context("each credential alone, or none (the ambient chain)", func() {
		ginkgo.It("should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(aksConfig())).To(gomega.BeNil())

			withSecret := aksConfig()
			withSecret.ClientId = "test-client-id"
			withSecret.ClientSecret = "test-client-secret"
			gomega.Expect(protovalidate.Validate(withSecret)).To(gomega.BeNil())

			withToken := aksConfig()
			withToken.AccessToken = "eyJ0eXAi.test-token"
			gomega.Expect(protovalidate.Validate(withToken)).To(gomega.BeNil())
		})
	})

	ginkgo.Context("a client secret and an access token together", func() {
		ginkgo.It("should be refused by kubernetes.azure_aks.one_credential", func() {
			input := aksConfig()
			input.ClientId = "test-client-id"
			input.ClientSecret = "test-client-secret"
			input.AccessToken = "eyJ0eXAi.test-token"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("set at most one of client_secret and access_token"))
		})
	})

	ginkgo.Context("a client secret without its identity", func() {
		ginkgo.It("should be refused by kubernetes.azure_aks.secret_requires_identity", func() {
			input := aksConfig()
			input.ClientSecret = "test-client-secret"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("tenant_id and client_id are required when client_secret is set"))
		})
	})
})
