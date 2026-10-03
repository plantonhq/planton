package gcplogbucketv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpLogBucketSpec Suite")
}

// folderRef is a literal folder ID in the reference field's literal arm.
func folderRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

var _ = ginkgo.Describe("GcpLogBucketSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpLogBucket {
		return &GcpLogBucket{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpLogBucket",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-log-bucket",
			},
			Spec: &GcpLogBucketSpec{
				BucketId: "audit-logs",
			},
		}
	}

	folderScoped := func() *GcpLogBucket {
		b := minimal()
		b.Spec.Scope = &GcpLogBucketScope{FolderId: folderRef("123456789")}
		return b
	}

	ginkgo.Context("minimal manifest", func() {
		ginkgo.It("passes validation", func() {
			gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
		})
	})

	ginkgo.Context("bucket_id", func() {
		ginkgo.It("is required", func() {
			b := minimal()
			b.Spec.BucketId = ""
			gomega.Expect(validator.Validate(b)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("scope", func() {
		ginkgo.It("accepts each single arm", func() {
			for _, scope := range []*GcpLogBucketScope{
				{FolderId: folderRef("123456789")},
				{OrganizationId: "987654321"},
				{BillingAccount: "012345-6789AB-CDEF01"},
			} {
				b := minimal()
				b.Spec.Scope = scope
				gomega.Expect(validator.Validate(b)).To(gomega.Succeed())
			}
		})

		ginkgo.It("rejects two arms at once", func() {
			b := minimal()
			b.Spec.Scope = &GcpLogBucketScope{
				FolderId:       folderRef("123456789"),
				OrganizationId: "987654321",
			}
			gomega.Expect(validator.Validate(b)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("retention_days", func() {
		ginkgo.It("accepts the documented range and rejects values beyond it", func() {
			b := minimal()
			b.Spec.RetentionDays = 3650
			gomega.Expect(validator.Validate(b)).To(gomega.Succeed())

			b.Spec.RetentionDays = 3651
			gomega.Expect(validator.Validate(b)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("locked", func() {
		ginkgo.It("is project-scope only", func() {
			b := minimal()
			b.Spec.Locked = true
			gomega.Expect(validator.Validate(b)).To(gomega.Succeed(), "default project scope")

			b = folderScoped()
			b.Spec.Locked = true
			gomega.Expect(validator.Validate(b)).ToNot(gomega.Succeed(), "folder scope")
		})
	})

	ginkgo.Context("enable_analytics", func() {
		ginkgo.It("is project-scope only", func() {
			b := minimal()
			b.Spec.EnableAnalytics = proto.Bool(true)
			gomega.Expect(validator.Validate(b)).To(gomega.Succeed(), "project scope")

			b = folderScoped()
			b.Spec.EnableAnalytics = proto.Bool(true)
			gomega.Expect(validator.Validate(b)).ToNot(gomega.Succeed(), "folder scope")

			b = folderScoped()
			b.Spec.EnableAnalytics = proto.Bool(false)
			gomega.Expect(validator.Validate(b)).ToNot(gomega.Succeed(), "explicit false is still a project-scope-only argument")
		})
	})

	ginkgo.Context("linked_bigquery_dataset", func() {
		ginkgo.It("requires enable_analytics true", func() {
			b := minimal()
			b.Spec.LinkedBigqueryDataset = &GcpLogBucketLinkedDataset{LinkId: "audit_logs"}
			gomega.Expect(validator.Validate(b)).ToNot(gomega.Succeed(), "analytics unset")

			b.Spec.EnableAnalytics = proto.Bool(true)
			gomega.Expect(validator.Validate(b)).To(gomega.Succeed(), "analytics enabled")
		})

		ginkgo.It("requires link_id", func() {
			b := minimal()
			b.Spec.EnableAnalytics = proto.Bool(true)
			b.Spec.LinkedBigqueryDataset = &GcpLogBucketLinkedDataset{}
			gomega.Expect(validator.Validate(b)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("scope_settings", func() {
		ginkgo.It("requires folder or organization scope", func() {
			b := minimal()
			b.Spec.ScopeSettings = &GcpLogBucketScopeSettings{DisableDefaultSink: true}
			gomega.Expect(validator.Validate(b)).ToNot(gomega.Succeed(), "project scope")

			b = folderScoped()
			b.Spec.ScopeSettings = &GcpLogBucketScopeSettings{DisableDefaultSink: true}
			gomega.Expect(validator.Validate(b)).To(gomega.Succeed(), "folder scope")
		})
	})

	ginkgo.Context("index_configs", func() {
		ginkgo.It("requires field_path and type, capped at 20 entries", func() {
			b := minimal()
			b.Spec.IndexConfigs = []*GcpLogBucketIndexConfig{{FieldPath: "jsonPayload.status"}}
			gomega.Expect(validator.Validate(b)).ToNot(gomega.Succeed(), "missing type")

			b.Spec.IndexConfigs[0].Type = "INDEX_TYPE_STRING"
			gomega.Expect(validator.Validate(b)).To(gomega.Succeed())

			b.Spec.IndexConfigs = nil
			for i := 0; i < 21; i++ {
				b.Spec.IndexConfigs = append(b.Spec.IndexConfigs, &GcpLogBucketIndexConfig{
					FieldPath: "jsonPayload.f",
					Type:      "INDEX_TYPE_STRING",
				})
			}
			gomega.Expect(validator.Validate(b)).ToNot(gomega.Succeed(), "21 entries")
		})
	})

	ginkgo.Context("log_views", func() {
		ginkgo.It("requires view_id", func() {
			b := minimal()
			b.Spec.LogViews = []*GcpLogBucketLogView{{Filter: `severity>=ERROR`}}
			gomega.Expect(validator.Validate(b)).ToNot(gomega.Succeed())

			b.Spec.LogViews[0].ViewId = "errors-only"
			gomega.Expect(validator.Validate(b)).To(gomega.Succeed())
		})
	})

	ginkgo.Context("deletion_policy", func() {
		ginkgo.It("accepts the documented values and rejects others", func() {
			for _, v := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
				b := minimal()
				b.Spec.DeletionPolicy = v
				gomega.Expect(validator.Validate(b)).To(gomega.Succeed(), "value %q", v)
			}
			b := minimal()
			b.Spec.DeletionPolicy = "KEEP"
			gomega.Expect(validator.Validate(b)).ToNot(gomega.Succeed())
		})
	})
})
