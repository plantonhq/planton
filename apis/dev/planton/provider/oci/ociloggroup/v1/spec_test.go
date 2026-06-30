package ociloggroupv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestOciLogGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciLogGroupSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidLogGroup() *OciLogGroup {
	return &OciLogGroup{
		ApiVersion: "oci.planton.dev/v1",
		Kind:       "OciLogGroup",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-log-group",
		},
		Spec: &OciLogGroupSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
		},
	}
}

func newServiceLog(displayName string) *OciLogGroupSpec_Log {
	return &OciLogGroupSpec_Log{
		DisplayName: displayName,
		LogType:     OciLogGroupSpec_Log_service,
		Configuration: &OciLogGroupSpec_Log_ServiceLogConfiguration{
			Service:  "flowlogs",
			Resource: newStringValueOrRef("ocid1.vcn.oc1..example"),
			Category: "all",
		},
	}
}

func newCustomLog(displayName string) *OciLogGroupSpec_Log {
	return &OciLogGroupSpec_Log{
		DisplayName: displayName,
		LogType:     OciLogGroupSpec_Log_custom,
	}
}

var _ = ginkgo.Describe("OciLogGroupSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_log_group", func() {

			ginkgo.It("should not return a validation error for minimal valid fields (no logs)", func() {
				input := minimalValidLogGroup()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with description", func() {
				input := minimalValidLogGroup()
				input.Spec.Description = "Application audit logs"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with a custom log", func() {
				input := minimalValidLogGroup()
				input.Spec.Logs = []*OciLogGroupSpec_Log{
					newCustomLog("app-custom-log"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with a service log", func() {
				input := minimalValidLogGroup()
				input.Spec.Logs = []*OciLogGroupSpec_Log{
					newServiceLog("vcn-flow-log"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with service log including parameters", func() {
				input := minimalValidLogGroup()
				svcLog := newServiceLog("bucket-write-log")
				svcLog.Configuration.Service = "objectstorage"
				svcLog.Configuration.Category = "write"
				svcLog.Configuration.Resource = newStringValueOrRef("ocid1.bucket.oc1..example")
				svcLog.Configuration.Parameters = map[string]string{
					"prefix": "/data",
				}
				input.Spec.Logs = []*OciLogGroupSpec_Log{svcLog}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with configuration compartment_id override", func() {
				input := minimalValidLogGroup()
				svcLog := newServiceLog("cross-compartment-log")
				svcLog.Configuration.CompartmentId = newStringValueOrRef("ocid1.compartment.oc1..other")
				input.Spec.Logs = []*OciLogGroupSpec_Log{svcLog}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with is_enabled explicitly set", func() {
				input := minimalValidLogGroup()
				log := newCustomLog("disabled-log")
				log.IsEnabled = proto.Bool(false)
				input.Spec.Logs = []*OciLogGroupSpec_Log{log}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with valid retention durations", func() {
				for _, days := range []int32{30, 60, 90, 120, 150, 180} {
					input := minimalValidLogGroup()
					log := newCustomLog("retained-log")
					log.RetentionDuration = proto.Int32(days)
					input.Spec.Logs = []*OciLogGroupSpec_Log{log}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should not return a validation error with multiple logs (mixed types)", func() {
				input := minimalValidLogGroup()
				input.Spec.Logs = []*OciLogGroupSpec_Log{
					newCustomLog("app-log"),
					newServiceLog("vcn-flow-log"),
					newServiceLog("lb-access-log"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via valueFrom ref", func() {
				input := minimalValidLogGroup()
				input.Spec.CompartmentId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-compartment",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with resource via valueFrom ref", func() {
				input := minimalValidLogGroup()
				svcLog := newServiceLog("flow-log")
				svcLog.Configuration.Resource = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-vcn",
						},
					},
				}
				input.Spec.Logs = []*OciLogGroupSpec_Log{svcLog}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full configuration", func() {
				input := minimalValidLogGroup()
				input.Spec.Description = "Production audit and flow logs"
				svcLog := newServiceLog("vcn-flow-log")
				svcLog.Configuration.CompartmentId = newStringValueOrRef("ocid1.compartment.oc1..other")
				svcLog.Configuration.Parameters = map[string]string{"filter": "all"}
				svcLog.IsEnabled = proto.Bool(true)
				svcLog.RetentionDuration = proto.Int32(90)
				customLog := newCustomLog("app-audit-log")
				customLog.IsEnabled = proto.Bool(true)
				customLog.RetentionDuration = proto.Int32(180)
				input.Spec.Logs = []*OciLogGroupSpec_Log{svcLog, customLog}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_log_group", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidLogGroup()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidLogGroup()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidLogGroup()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciLogGroup{
					ApiVersion: "oci.planton.dev/v1",
					Kind:       "OciLogGroup",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-log-group"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidLogGroup()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when log display_name is empty", func() {
				input := minimalValidLogGroup()
				input.Spec.Logs = []*OciLogGroupSpec_Log{
					{
						DisplayName: "",
						LogType:     OciLogGroupSpec_Log_custom,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when log_type is unspecified", func() {
				input := minimalValidLogGroup()
				input.Spec.Logs = []*OciLogGroupSpec_Log{
					{
						DisplayName: "bad-log",
						LogType:     OciLogGroupSpec_Log_unspecified,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when service log has no configuration", func() {
				input := minimalValidLogGroup()
				input.Spec.Logs = []*OciLogGroupSpec_Log{
					{
						DisplayName: "svc-log-no-config",
						LogType:     OciLogGroupSpec_Log_service,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when configuration service is empty", func() {
				input := minimalValidLogGroup()
				input.Spec.Logs = []*OciLogGroupSpec_Log{
					{
						DisplayName: "bad-svc-log",
						LogType:     OciLogGroupSpec_Log_service,
						Configuration: &OciLogGroupSpec_Log_ServiceLogConfiguration{
							Service:  "",
							Resource: newStringValueOrRef("ocid1.vcn.oc1..example"),
							Category: "all",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when configuration resource is missing", func() {
				input := minimalValidLogGroup()
				input.Spec.Logs = []*OciLogGroupSpec_Log{
					{
						DisplayName: "bad-svc-log",
						LogType:     OciLogGroupSpec_Log_service,
						Configuration: &OciLogGroupSpec_Log_ServiceLogConfiguration{
							Service:  "flowlogs",
							Resource: nil,
							Category: "all",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when configuration category is empty", func() {
				input := minimalValidLogGroup()
				input.Spec.Logs = []*OciLogGroupSpec_Log{
					{
						DisplayName: "bad-svc-log",
						LogType:     OciLogGroupSpec_Log_service,
						Configuration: &OciLogGroupSpec_Log_ServiceLogConfiguration{
							Service:  "flowlogs",
							Resource: newStringValueOrRef("ocid1.vcn.oc1..example"),
							Category: "",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when retention_duration is not a 30-day increment", func() {
				input := minimalValidLogGroup()
				log := newCustomLog("bad-retention")
				log.RetentionDuration = proto.Int32(45)
				input.Spec.Logs = []*OciLogGroupSpec_Log{log}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when retention_duration exceeds 180", func() {
				input := minimalValidLogGroup()
				log := newCustomLog("bad-retention")
				log.RetentionDuration = proto.Int32(210)
				input.Spec.Logs = []*OciLogGroupSpec_Log{log}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when retention_duration is below 30", func() {
				input := minimalValidLogGroup()
				log := newCustomLog("bad-retention")
				log.RetentionDuration = proto.Int32(15)
				input.Spec.Logs = []*OciLogGroupSpec_Log{log}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

		})
	})
})
