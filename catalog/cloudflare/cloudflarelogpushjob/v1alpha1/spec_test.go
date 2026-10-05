package cloudflarelogpushjobv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestCloudflareLogpushJobSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareLogpushJobSpec Custom Validation Tests")
}

const testAccountID = "0da42c8d2132a9ddaf714f9e7c920711"

func ref(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func validJob(spec *CloudflareLogpushJobSpec) *CloudflareLogpushJob {
	return &CloudflareLogpushJob{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareLogpushJob",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-logpush-job",
		},
		Spec: spec,
	}
}

func zoneSpec() *CloudflareLogpushJobSpec {
	return &CloudflareLogpushJobSpec{
		ZoneId:          ref("023e105f4ecef8ad9ca31a8372d0c353"),
		Dataset:         "http_requests",
		DestinationConf: ref("r2://logs-bucket/{DATE}?account-id=abc&access-key-id=k&secret-access-key=s"),
	}
}

var _ = ginkgo.Describe("CloudflareLogpushJobSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal zone-scoped job", func() {
			gomega.Expect(protovalidate.Validate(validJob(zoneSpec()))).To(gomega.BeNil())
		})

		ginkgo.It("should accept an account-scoped job with an account dataset", func() {
			spec := &CloudflareLogpushJobSpec{
				AccountId:       testAccountID,
				Dataset:         "audit_logs",
				DestinationConf: ref("https://splunk.example.com/services/collector/raw?channel=abc"),
			}
			gomega.Expect(protovalidate.Validate(validJob(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept full output options and upload tuning", func() {
			spec := zoneSpec()
			spec.Name = "http-to-r2"
			spec.Enabled = proto.Bool(true)
			spec.Filter = `{"where":{"and":[{"key":"ClientRequestHost","operator":"eq","value":"example.com"}]}}`
			spec.MaxUploadBytes = proto.Int64(5000000)
			spec.MaxUploadIntervalSeconds = proto.Int64(30)
			spec.MaxUploadRecords = proto.Int64(1000)
			spec.OutputOptions = &CloudflareLogpushJobOutputOptions{
				OutputType:      "ndjson",
				TimestampFormat: "rfc3339",
				SampleRate:      proto.Float64(0.5),
				FieldNames:      []string{"ClientIP", "EdgeStartTimestamp", "RayID"},
			}
			gomega.Expect(protovalidate.Validate(validJob(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept the instant-logs kind and unset tuning as zero", func() {
			spec := zoneSpec()
			spec.Kind = "edge"
			spec.MaxUploadBytes = proto.Int64(0)
			gomega.Expect(protovalidate.Validate(validJob(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept the challenge-issuing arm", func() {
			spec := zoneSpec()
			spec.GenerateOwnershipChallenge = true
			gomega.Expect(protovalidate.Validate(validJob(spec))).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject both scopes set", func() {
			spec := zoneSpec()
			spec.AccountId = testAccountID
			gomega.Expect(protovalidate.Validate(validJob(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject neither scope set", func() {
			spec := zoneSpec()
			spec.ZoneId = nil
			gomega.Expect(protovalidate.Validate(validJob(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown dataset", func() {
			spec := zoneSpec()
			spec.Dataset = "http_request"
			gomega.Expect(protovalidate.Validate(validJob(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing destination", func() {
			spec := zoneSpec()
			spec.DestinationConf = nil
			gomega.Expect(protovalidate.Validate(validJob(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown job kind", func() {
			spec := zoneSpec()
			spec.Kind = "instant-logs"
			gomega.Expect(protovalidate.Validate(validJob(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject max_upload_bytes below the 5 MB floor", func() {
			spec := zoneSpec()
			spec.MaxUploadBytes = proto.Int64(100)
			gomega.Expect(protovalidate.Validate(validJob(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject max_upload_interval_seconds above 300", func() {
			spec := zoneSpec()
			spec.MaxUploadIntervalSeconds = proto.Int64(301)
			gomega.Expect(protovalidate.Validate(validJob(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown output type", func() {
			spec := zoneSpec()
			spec.OutputOptions = &CloudflareLogpushJobOutputOptions{OutputType: "json"}
			gomega.Expect(protovalidate.Validate(validJob(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a sample rate above 1", func() {
			spec := zoneSpec()
			spec.OutputOptions = &CloudflareLogpushJobOutputOptions{SampleRate: proto.Float64(1.5)}
			gomega.Expect(protovalidate.Validate(validJob(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown timestamp format", func() {
			spec := zoneSpec()
			spec.OutputOptions = &CloudflareLogpushJobOutputOptions{TimestampFormat: "iso8601"}
			gomega.Expect(protovalidate.Validate(validJob(spec))).NotTo(gomega.BeNil())
		})
	})
})
