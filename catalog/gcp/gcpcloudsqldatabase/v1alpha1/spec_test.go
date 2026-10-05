package gcpcloudsqldatabasev1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpCloudSqlDatabaseSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func fromRef(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

var _ = ginkgo.Describe("GcpCloudSqlDatabaseSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpCloudSqlDatabase {
		return &GcpCloudSqlDatabase{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpCloudSqlDatabase",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-database",
			},
			Spec: &GcpCloudSqlDatabaseSpec{
				Instance:     litRef("orders-db"),
				DatabaseName: "orders",
			},
		}
	}

	expectValid := func(r *GcpCloudSqlDatabase) {
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
	}

	expectInvalid := func(r *GcpCloudSqlDatabase, substr string) {
		err := validator.Validate(r)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), substr)).To(
			gomega.BeTrue(), "expected error to contain %q, got: %s", substr, err)
	}

	ginkgo.It("accepts a minimal database", func() {
		expectValid(minimal())
	})

	ginkgo.It("rejects a missing instance", func() {
		r := minimal()
		r.Spec.Instance = nil
		expectInvalid(r, "instance")
	})

	ginkgo.It("accepts an instance reference", func() {
		r := minimal()
		r.Spec.Instance = fromRef("orders-db")
		expectValid(r)
	})

	ginkgo.It("rejects a missing database_name", func() {
		r := minimal()
		r.Spec.DatabaseName = ""
		expectInvalid(r, "database_name")
	})

	ginkgo.It("rejects an overlong database_name", func() {
		r := minimal()
		r.Spec.DatabaseName = strings.Repeat("a", 129)
		expectInvalid(r, "database_name")
	})

	ginkgo.It("accepts a project_id reference", func() {
		r := minimal()
		r.Spec.ProjectId = fromRef("my-project")
		expectValid(r)
	})

	ginkgo.It("accepts MySQL charset and collation", func() {
		r := minimal()
		r.Spec.Charset = "utf8mb4"
		r.Spec.Collation = "utf8mb4_0900_ai_ci"
		expectValid(r)
	})

	ginkgo.It("rejects a wrong kind constant", func() {
		r := minimal()
		r.Kind = "GcpCloudSql"
		expectInvalid(r, "kind")
	})

	ginkgo.It("accepts every valid deletion_policy", func() {
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			r := minimal()
			r.Spec.DeletionPolicy = policy
			expectValid(r)
		}
	})

	ginkgo.It("rejects an unknown deletion_policy", func() {
		r := minimal()
		r.Spec.DeletionPolicy = "RETAIN"
		expectInvalid(r, "deletion_policy")
	})
})
