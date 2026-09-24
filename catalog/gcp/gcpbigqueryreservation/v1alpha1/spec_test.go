package gcpbigqueryreservationv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpBigQueryReservationSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func valueFrom(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: name}},
	}
}

var _ = ginkgo.Describe("GcpBigQueryReservationSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpBigQueryReservation {
		return &GcpBigQueryReservation{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpBigQueryReservation",
			Metadata:   &shared.CloudResourceMetadata{Name: "analytics"},
			Spec:       &GcpBigQueryReservationSpec{},
		}
	}
	projectAssignment := func(project, jobType string) *GcpBigQueryReservationAssignment {
		return &GcpBigQueryReservationAssignment{
			Assignee: &GcpBigQueryReservationAssignee{ProjectId: litRef(project)},
			JobType:  jobType,
		}
	}

	ginkgo.It("should accept a zero-baseline reservation with nothing else set", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set, with assignments at all three levels", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("bq-admin")
		msg.Spec.Location = "EU"
		msg.Spec.ReservationName = "analytics-prod"
		msg.Spec.SlotCapacity = 100
		msg.Spec.Edition = "ENTERPRISE_PLUS"
		msg.Spec.AutoscaleMaxSlots = 400
		msg.Spec.IgnoreIdleSlots = true
		msg.Spec.Concurrency = 50
		msg.Spec.ReservationGroup = valueFrom("tier-1")
		msg.Spec.SecondaryLocation = "europe-west4"
		msg.Spec.Labels = map[string]string{"team": "data"}
		msg.Spec.Assignments = []*GcpBigQueryReservationAssignment{
			projectAssignment("analytics", "QUERY"),
			projectAssignment("analytics", "PIPELINE"),
			{Assignee: &GcpBigQueryReservationAssignee{FolderId: valueFrom("data-folder")}, JobType: "QUERY"},
			{Assignee: &GcpBigQueryReservationAssignee{OrganizationId: "123456789"}, JobType: "CONTINUOUS"},
			{
				Assignee:  &GcpBigQueryReservationAssignee{ProjectId: litRef("analytics")},
				JobType:   "QUERY",
				Principal: "principal://iam.googleapis.com/projects/-/serviceAccounts/etl@analytics.iam.gserviceaccount.com",
			},
		}
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a duplicate assignment", func() {
		msg := minimal()
		msg.Spec.Assignments = []*GcpBigQueryReservationAssignment{
			projectAssignment("analytics", "QUERY"),
			projectAssignment("analytics", "QUERY"),
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Assignments = []*GcpBigQueryReservationAssignment{
			{Assignee: &GcpBigQueryReservationAssignee{FolderId: valueFrom("f")}, JobType: "QUERY"},
			{Assignee: &GcpBigQueryReservationAssignee{FolderId: valueFrom("f")}, JobType: "QUERY"},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require exactly one assignee and a job type from Google's list", func() {
		msg := minimal()
		msg.Spec.Assignments = []*GcpBigQueryReservationAssignment{{Assignee: &GcpBigQueryReservationAssignee{}, JobType: "QUERY"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Assignments = []*GcpBigQueryReservationAssignment{{
			Assignee: &GcpBigQueryReservationAssignee{ProjectId: litRef("a"), OrganizationId: "1"}, JobType: "QUERY",
		}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Assignments = []*GcpBigQueryReservationAssignment{projectAssignment("a", "JOB_TYPE_UNSPECIFIED")}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Assignments = []*GcpBigQueryReservationAssignment{{JobType: "QUERY"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed organization ID and principal", func() {
		msg := minimal()
		msg.Spec.Assignments = []*GcpBigQueryReservationAssignment{{
			Assignee: &GcpBigQueryReservationAssignee{OrganizationId: "organizations/1"}, JobType: "QUERY",
		}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		assignment := projectAssignment("a", "QUERY")
		assignment.Principal = "user:ana@example.com"
		msg.Spec.Assignments = []*GcpBigQueryReservationAssignment{assignment}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject negative capacity, an unknown edition, a malformed name, and an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.SlotCapacity = -1
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.AutoscaleMaxSlots = -50
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Edition = "FLAT_RATE"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.ReservationName = "analytics_prod"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
