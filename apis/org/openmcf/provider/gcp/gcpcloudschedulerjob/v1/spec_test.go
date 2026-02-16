package gcpcloudschedulerjobv1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpCloudSchedulerJobSpec Suite")
}

var _ = ginkgo.Describe("GcpCloudSchedulerJobSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpCloudSchedulerJob with HTTP target.
	minimalHttp := func() *GcpCloudSchedulerJob {
		return &GcpCloudSchedulerJob{
			ApiVersion: "gcp.openmcf.org/v1",
			Kind:       "GcpCloudSchedulerJob",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-job",
			},
			Spec: &GcpCloudSchedulerJobSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-gcp-project",
					},
				},
				Location: "us-central1",
				Schedule: "0 9 * * 1",
				HttpTarget: &GcpCloudSchedulerJobHttpTarget{
					Uri: "https://example.com/api/trigger",
				},
			},
		}
	}

	// Helper to build a minimal valid GcpCloudSchedulerJob with Pub/Sub target.
	minimalPubsub := func() *GcpCloudSchedulerJob {
		return &GcpCloudSchedulerJob{
			ApiVersion: "gcp.openmcf.org/v1",
			Kind:       "GcpCloudSchedulerJob",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-job",
			},
			Spec: &GcpCloudSchedulerJobSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-gcp-project",
					},
				},
				Location: "us-central1",
				Schedule: "*/5 * * * *",
				PubsubTarget: &GcpCloudSchedulerJobPubsubTarget{
					TopicName: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
							Value: "projects/my-gcp-project/topics/my-topic",
						},
					},
				},
			},
		}
	}

	// Helper to build a minimal valid GcpCloudSchedulerJob with App Engine target.
	minimalAppEngine := func() *GcpCloudSchedulerJob {
		return &GcpCloudSchedulerJob{
			ApiVersion: "gcp.openmcf.org/v1",
			Kind:       "GcpCloudSchedulerJob",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-job",
			},
			Spec: &GcpCloudSchedulerJobSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-gcp-project",
					},
				},
				Location: "us-central1",
				Schedule: "0 0 * * *",
				AppEngineHttpTarget: &GcpCloudSchedulerJobAppEngineHttpTarget{
					RelativeUri: "/tasks/daily-cleanup",
				},
			},
		}
	}

	// Helper to build a StringValueOrRef with a literal value.
	strRef := func(val string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: val,
			},
		}
	}

	// Suppress unused variable warnings for helpers.
	_ = minimalPubsub
	_ = minimalAppEngine
	_ = strRef

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid HTTP target spec", func() {
		msg := minimalHttp()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a minimal valid Pub/Sub target spec", func() {
		msg := minimalPubsub()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a minimal valid App Engine target spec", func() {
		msg := minimalAppEngine()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept HTTP target with all HTTP methods", func() {
		methods := []string{"POST", "GET", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"}
		for _, method := range methods {
			msg := minimalHttp()
			msg.Spec.HttpTarget.HttpMethod = method
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "method %s should be valid", method)
		}
	})

	ginkgo.It("should accept App Engine target with all HTTP methods", func() {
		methods := []string{"POST", "GET", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"}
		for _, method := range methods {
			msg := minimalAppEngine()
			msg.Spec.AppEngineHttpTarget.HttpMethod = method
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "method %s should be valid", method)
		}
	})

	ginkgo.It("should accept HTTP target with OIDC token", func() {
		msg := minimalHttp()
		msg.Spec.HttpTarget.OidcToken = &GcpCloudSchedulerJobOidcToken{
			ServiceAccountEmail: strRef("sa@project.iam.gserviceaccount.com"),
			Audience:            "https://example.com",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept HTTP target with OAuth token", func() {
		msg := minimalHttp()
		msg.Spec.HttpTarget.OauthToken = &GcpCloudSchedulerJobOAuthToken{
			ServiceAccountEmail: strRef("sa@project.iam.gserviceaccount.com"),
			Scope:               "https://www.googleapis.com/auth/cloud-platform",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept HTTP target with headers", func() {
		msg := minimalHttp()
		msg.Spec.HttpTarget.Headers = map[string]string{
			"Content-Type": "application/json",
			"X-Custom":     "value",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept HTTP target with body", func() {
		msg := minimalHttp()
		msg.Spec.HttpTarget.HttpMethod = "POST"
		msg.Spec.HttpTarget.Body = "eyJhY3Rpb24iOiAicmVwb3J0In0=" // {"action": "report"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept Pub/Sub target with data", func() {
		msg := minimalPubsub()
		msg.Spec.PubsubTarget.Data = "eyJldmVudCI6ICJkYWlseV9yZXBvcnQifQ==" // {"event": "daily_report"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept Pub/Sub target with attributes", func() {
		msg := minimalPubsub()
		msg.Spec.PubsubTarget.Attributes = map[string]string{
			"source": "scheduler",
			"type":   "daily",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept App Engine target with routing", func() {
		msg := minimalAppEngine()
		msg.Spec.AppEngineHttpTarget.AppEngineRouting = &GcpCloudSchedulerJobAppEngineRouting{
			Service: "my-service",
			Version: "v1",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept valid job_name with letters, numbers, hyphens", func() {
		msg := minimalHttp()
		msg.Spec.JobName = "daily-report-job-123"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept valid job_name with underscores", func() {
		msg := minimalHttp()
		msg.Spec.JobName = "daily_report_job"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept empty job_name (falls back to metadata.name)", func() {
		msg := minimalHttp()
		msg.Spec.JobName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept time_zone", func() {
		msg := minimalHttp()
		msg.Spec.TimeZone = "America/New_York"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept description", func() {
		msg := minimalHttp()
		msg.Spec.Description = "Triggers daily report generation"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept attempt_deadline", func() {
		msg := minimalHttp()
		msg.Spec.AttemptDeadline = "300s"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept paused=true", func() {
		msg := minimalHttp()
		msg.Spec.Paused = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept retry config with all fields", func() {
		msg := minimalHttp()
		msg.Spec.RetryConfig = &GcpCloudSchedulerJobRetryConfig{
			RetryCount:         3,
			MaxRetryDuration:   "3600s",
			MinBackoffDuration: "5s",
			MaxBackoffDuration: "3600s",
			MaxDoublings:       5,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a fully featured HTTP spec", func() {
		msg := minimalHttp()
		msg.Spec.JobName = "full-featured-job"
		msg.Spec.TimeZone = "America/Los_Angeles"
		msg.Spec.Description = "Full-featured scheduled job"
		msg.Spec.AttemptDeadline = "600s"
		msg.Spec.HttpTarget.HttpMethod = "POST"
		msg.Spec.HttpTarget.Body = "eyJhY3Rpb24iOiAicnVuIn0="
		msg.Spec.HttpTarget.Headers = map[string]string{
			"Content-Type": "application/json",
		}
		msg.Spec.HttpTarget.OidcToken = &GcpCloudSchedulerJobOidcToken{
			ServiceAccountEmail: strRef("invoker@project.iam.gserviceaccount.com"),
			Audience:            "https://my-service-abc123.run.app",
		}
		msg.Spec.RetryConfig = &GcpCloudSchedulerJobRetryConfig{
			RetryCount:         3,
			MaxRetryDuration:   "1800s",
			MinBackoffDuration: "5s",
			MaxBackoffDuration: "600s",
			MaxDoublings:       3,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a fully featured Pub/Sub spec", func() {
		msg := minimalPubsub()
		msg.Spec.JobName = "pubsub-publisher-job"
		msg.Spec.TimeZone = "Europe/London"
		msg.Spec.Description = "Publishes daily report trigger to Pub/Sub"
		msg.Spec.PubsubTarget.Data = "eyJldmVudCI6ICJkYWlseV9yZXBvcnQifQ=="
		msg.Spec.PubsubTarget.Attributes = map[string]string{
			"source": "cloud-scheduler",
			"type":   "daily-report",
		}
		msg.Spec.RetryConfig = &GcpCloudSchedulerJobRetryConfig{
			RetryCount:   5,
			MaxDoublings: 3,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a fully featured App Engine spec", func() {
		msg := minimalAppEngine()
		msg.Spec.JobName = "app-engine-cron"
		msg.Spec.TimeZone = "Asia/Tokyo"
		msg.Spec.Description = "Triggers nightly cleanup on App Engine"
		msg.Spec.AttemptDeadline = "900s"
		msg.Spec.AppEngineHttpTarget.HttpMethod = "POST"
		msg.Spec.AppEngineHttpTarget.Body = "eyJjbGVhbnVwIjogdHJ1ZX0="
		msg.Spec.AppEngineHttpTarget.Headers = map[string]string{
			"Content-Type": "application/json",
		}
		msg.Spec.AppEngineHttpTarget.AppEngineRouting = &GcpCloudSchedulerJobAppEngineRouting{
			Service:  "cleanup-service",
			Version:  "v2",
			Instance: "instance-001",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept description at max length (500 chars)", func() {
		msg := minimalHttp()
		msg.Spec.Description = strings.Repeat("a", 500)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept OIDC token without audience", func() {
		msg := minimalHttp()
		msg.Spec.HttpTarget.OidcToken = &GcpCloudSchedulerJobOidcToken{
			ServiceAccountEmail: strRef("sa@project.iam.gserviceaccount.com"),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept OAuth token without scope", func() {
		msg := minimalHttp()
		msg.Spec.HttpTarget.OauthToken = &GcpCloudSchedulerJobOAuthToken{
			ServiceAccountEmail: strRef("sa@project.iam.gserviceaccount.com"),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject missing api_version", func() {
		msg := minimalHttp()
		msg.ApiVersion = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject wrong api_version", func() {
		msg := minimalHttp()
		msg.ApiVersion = "wrong.version/v1"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject wrong kind", func() {
		msg := minimalHttp()
		msg.Kind = "WrongKind"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing metadata", func() {
		msg := minimalHttp()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing spec", func() {
		msg := minimalHttp()
		msg.Spec = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing project_id", func() {
		msg := minimalHttp()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing location", func() {
		msg := minimalHttp()
		msg.Spec.Location = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing schedule", func() {
		msg := minimalHttp()
		msg.Spec.Schedule = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject no target specified", func() {
		msg := &GcpCloudSchedulerJob{
			ApiVersion: "gcp.openmcf.org/v1",
			Kind:       "GcpCloudSchedulerJob",
			Metadata:   &shared.CloudResourceMetadata{Name: "test"},
			Spec: &GcpCloudSchedulerJobSpec{
				ProjectId: strRef("my-project"),
				Location:  "us-central1",
				Schedule:  "0 9 * * 1",
				// No target set.
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one"))
	})

	ginkgo.It("should reject multiple targets (HTTP + Pub/Sub)", func() {
		msg := minimalHttp()
		msg.Spec.PubsubTarget = &GcpCloudSchedulerJobPubsubTarget{
			TopicName: strRef("projects/p/topics/t"),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one"))
	})

	ginkgo.It("should reject multiple targets (HTTP + App Engine)", func() {
		msg := minimalHttp()
		msg.Spec.AppEngineHttpTarget = &GcpCloudSchedulerJobAppEngineHttpTarget{
			RelativeUri: "/tasks/run",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one"))
	})

	ginkgo.It("should reject multiple targets (Pub/Sub + App Engine)", func() {
		msg := minimalPubsub()
		msg.Spec.AppEngineHttpTarget = &GcpCloudSchedulerJobAppEngineHttpTarget{
			RelativeUri: "/tasks/run",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one"))
	})

	ginkgo.It("should reject all three targets", func() {
		msg := minimalHttp()
		msg.Spec.PubsubTarget = &GcpCloudSchedulerJobPubsubTarget{
			TopicName: strRef("projects/p/topics/t"),
		}
		msg.Spec.AppEngineHttpTarget = &GcpCloudSchedulerJobAppEngineHttpTarget{
			RelativeUri: "/tasks/run",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one"))
	})

	ginkgo.It("should reject both OAuth and OIDC tokens on HTTP target", func() {
		msg := minimalHttp()
		msg.Spec.HttpTarget.OauthToken = &GcpCloudSchedulerJobOAuthToken{
			ServiceAccountEmail: strRef("sa@project.iam.gserviceaccount.com"),
		}
		msg.Spec.HttpTarget.OidcToken = &GcpCloudSchedulerJobOidcToken{
			ServiceAccountEmail: strRef("sa@project.iam.gserviceaccount.com"),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("oauth_token"))
	})

	ginkgo.It("should reject invalid HTTP method on HTTP target", func() {
		msg := minimalHttp()
		msg.Spec.HttpTarget.HttpMethod = "INVALID"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("http_method"))
	})

	ginkgo.It("should reject invalid HTTP method on App Engine target", func() {
		msg := minimalAppEngine()
		msg.Spec.AppEngineHttpTarget.HttpMethod = "CONNECT"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("http_method"))
	})

	ginkgo.It("should reject HTTP target missing uri", func() {
		msg := minimalHttp()
		msg.Spec.HttpTarget.Uri = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject Pub/Sub target missing topic_name", func() {
		msg := minimalPubsub()
		msg.Spec.PubsubTarget.TopicName = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject App Engine target missing relative_uri", func() {
		msg := minimalAppEngine()
		msg.Spec.AppEngineHttpTarget.RelativeUri = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject App Engine relative_uri not starting with /", func() {
		msg := minimalAppEngine()
		msg.Spec.AppEngineHttpTarget.RelativeUri = "tasks/run"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("relative_uri"))
	})

	ginkgo.It("should reject job_name starting with a number", func() {
		msg := minimalHttp()
		msg.Spec.JobName = "123-invalid"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("job_name"))
	})

	ginkgo.It("should reject job_name with spaces", func() {
		msg := minimalHttp()
		msg.Spec.JobName = "my job"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject description exceeding 500 characters", func() {
		msg := minimalHttp()
		msg.Spec.Description = strings.Repeat("a", 501)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject OAuth token missing service_account_email", func() {
		msg := minimalHttp()
		msg.Spec.HttpTarget.OauthToken = &GcpCloudSchedulerJobOAuthToken{
			ServiceAccountEmail: nil,
			Scope:               "https://www.googleapis.com/auth/cloud-platform",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject OIDC token missing service_account_email", func() {
		msg := minimalHttp()
		msg.Spec.HttpTarget.OidcToken = &GcpCloudSchedulerJobOidcToken{
			ServiceAccountEmail: nil,
			Audience:            "https://example.com",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	// ──────────────── Proto Wire Format Cases ────────────────

	ginkgo.It("should round-trip through proto marshal/unmarshal", func() {
		msg := minimalHttp()
		msg.Spec.JobName = "roundtrip-test"
		msg.Spec.TimeZone = "Etc/UTC"
		msg.Spec.Description = "Test job"
		msg.Spec.HttpTarget.HttpMethod = "POST"
		msg.Spec.HttpTarget.Headers = map[string]string{"X-Test": "value"}

		data, err := proto.Marshal(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		decoded := &GcpCloudSchedulerJob{}
		err = proto.Unmarshal(data, decoded)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		gomega.Expect(decoded.Spec.JobName).To(gomega.Equal("roundtrip-test"))
		gomega.Expect(decoded.Spec.Schedule).To(gomega.Equal("0 9 * * 1"))
		gomega.Expect(decoded.Spec.HttpTarget.Uri).To(gomega.Equal("https://example.com/api/trigger"))
		gomega.Expect(decoded.Spec.HttpTarget.HttpMethod).To(gomega.Equal("POST"))
		gomega.Expect(decoded.Spec.HttpTarget.Headers["X-Test"]).To(gomega.Equal("value"))
	})
})
