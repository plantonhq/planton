package cloudflareworkerv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

const validAccountID = "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"

func ref(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func validWorker() *CloudflareWorker {
	return &CloudflareWorker{
		ApiVersion: "cloudflare.openmcf.org/v1",
		Kind:       "CloudflareWorker",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-worker"},
		Spec: &CloudflareWorkerSpec{
			AccountId:  validAccountID,
			WorkerName: "test-worker",
			Source:     &CloudflareWorkerSpec_Content{Content: "export default { fetch() { return new Response('ok'); } }"},
		},
	}
}

func TestCloudflareWorkerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareWorkerSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareWorkerSpec Custom Validation Tests", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a minimal inline-content worker", func() {
			gomega.Expect(protovalidate.Validate(validWorker())).To(gomega.BeNil())
		})

		ginkgo.It("accepts an r2_bundle source", func() {
			in := validWorker()
			in.Spec.Source = &CloudflareWorkerSpec_R2Bundle{R2Bundle: &CloudflareWorkerScriptBundle{Bucket: "builds", Path: "worker.js"}}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts grouped bindings, routing, and schedules", func() {
			in := validWorker()
			in.Spec.Vars = map[string]string{"LOG_LEVEL": "info"}
			in.Spec.Secrets = []*CloudflareWorkerSecretBinding{{Name: "API_KEY", Value: ref("secret")}}
			in.Spec.KvNamespaces = []*CloudflareWorkerKvBinding{{Name: "CONFIG", NamespaceId: ref("0f1e2d3c4b5a69788796a5b4c3d2e1f0")}}
			in.Spec.D1Databases = []*CloudflareWorkerD1Binding{{Name: "DB", DatabaseId: ref("9a1b2c3d-4e5f-6a7b-8c9d-0e1f2a3b4c5d")}}
			in.Spec.R2Buckets = []*CloudflareWorkerR2Binding{{Name: "ASSETS", BucketName: ref("media"), Jurisdiction: "eu"}}
			in.Spec.Services = []*CloudflareWorkerServiceBinding{{Name: "AUTH", Service: ref("auth-worker"), Environment: "production", Entrypoint: "AuthEntrypoint"}}
			in.Spec.WorkersDev = &CloudflareWorkerWorkersDev{Enabled: true}
			in.Spec.CustomDomains = []*CloudflareWorkerCustomDomain{{Hostname: "api.example.com", ZoneId: ref("023e105f4ecef8ad9ca31a8372d0c353")}}
			in.Spec.Routes = []*CloudflareWorkerRoute{{ZoneId: ref("023e105f4ecef8ad9ca31a8372d0c353"), Pattern: "api.example.com/*"}}
			in.Spec.Schedules = []string{"0 * * * *"}
			in.Spec.Observability = &CloudflareWorkerObservability{Enabled: true, HeadSamplingRate: 1}
			in.Spec.Placement = &CloudflareWorkerPlacement{Mode: "smart"}
			in.Spec.Limits = &CloudflareWorkerLimits{CpuMs: 50, Subrequests: 100}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a worker with no script source", func() {
			in := validWorker()
			in.Spec.Source = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a non-hex account_id", func() {
			in := validWorker()
			in.Spec.AccountId = "nope"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing worker_name", func() {
			in := validWorker()
			in.Spec.WorkerName = ""
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a secret binding with no value", func() {
			in := validWorker()
			in.Spec.Secrets = []*CloudflareWorkerSecretBinding{{Name: "API_KEY"}}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a kv binding with an empty namespace_id", func() {
			in := validWorker()
			in.Spec.KvNamespaces = []*CloudflareWorkerKvBinding{{Name: "CONFIG", NamespaceId: ref("")}}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid r2 binding jurisdiction", func() {
			in := validWorker()
			in.Spec.R2Buckets = []*CloudflareWorkerR2Binding{{Name: "ASSETS", BucketName: ref("media"), Jurisdiction: "mars"}}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a head_sampling_rate above 1", func() {
			in := validWorker()
			in.Spec.Observability = &CloudflareWorkerObservability{Enabled: true, HeadSamplingRate: 2}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid placement mode", func() {
			in := validWorker()
			in.Spec.Placement = &CloudflareWorkerPlacement{Mode: "random"}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a bad compatibility_date format", func() {
			in := validWorker()
			in.Spec.CompatibilityDate = "2024/01/01"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})
