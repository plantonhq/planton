package ocinosqltablev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOciNosqlTableSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciNosqlTableSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidNosqlTable() *OciNosqlTable {
	return &OciNosqlTable{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciNosqlTable",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-table",
		},
		Spec: &OciNosqlTableSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
			Name:          "test_table",
			DdlStatement:  "CREATE TABLE test_table (id INTEGER, name STRING, PRIMARY KEY(id))",
			TableLimits: &OciNosqlTableSpec_TableLimits{
				MaxReadUnits:    50,
				MaxWriteUnits:   50,
				MaxStorageInGbs: 25,
			},
		},
	}
}

var _ = ginkgo.Describe("OciNosqlTableSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_nosql_table", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidNosqlTable()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with on_demand capacity mode", func() {
				input := minimalValidNosqlTable()
				input.Spec.TableLimits.CapacityMode = OciNosqlTableSpec_TableLimits_on_demand
				input.Spec.TableLimits.MaxReadUnits = 0
				input.Spec.TableLimits.MaxWriteUnits = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with explicit provisioned capacity mode", func() {
				input := minimalValidNosqlTable()
				input.Spec.TableLimits.CapacityMode = OciNosqlTableSpec_TableLimits_provisioned
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with is_auto_reclaimable set", func() {
				input := minimalValidNosqlTable()
				input.Spec.IsAutoReclaimable = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with a simple index", func() {
				input := minimalValidNosqlTable()
				input.Spec.Indexes = []*OciNosqlTableSpec_Index{
					{
						Name: "idx_name",
						Keys: []*OciNosqlTableSpec_Index_IndexKey{
							{ColumnName: "name"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple indexes", func() {
				input := minimalValidNosqlTable()
				input.Spec.Indexes = []*OciNosqlTableSpec_Index{
					{
						Name: "idx_name",
						Keys: []*OciNosqlTableSpec_Index_IndexKey{
							{ColumnName: "name"},
						},
					},
					{
						Name: "idx_created",
						Keys: []*OciNosqlTableSpec_Index_IndexKey{
							{ColumnName: "created_at"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with a JSON index", func() {
				input := minimalValidNosqlTable()
				input.Spec.Indexes = []*OciNosqlTableSpec_Index{
					{
						Name: "idx_profile_email",
						Keys: []*OciNosqlTableSpec_Index_IndexKey{
							{
								ColumnName:    "profile",
								JsonFieldType: "STRING",
								JsonPath:      "email",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with a composite index", func() {
				input := minimalValidNosqlTable()
				input.Spec.Indexes = []*OciNosqlTableSpec_Index{
					{
						Name: "idx_composite",
						Keys: []*OciNosqlTableSpec_Index_IndexKey{
							{ColumnName: "tenant_id"},
							{ColumnName: "created_at"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via valueFrom ref", func() {
				input := minimalValidNosqlTable()
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

			ginkgo.It("should not return a validation error with full configuration", func() {
				input := minimalValidNosqlTable()
				input.Spec.TableLimits.CapacityMode = OciNosqlTableSpec_TableLimits_provisioned
				input.Spec.TableLimits.MaxReadUnits = 200
				input.Spec.TableLimits.MaxWriteUnits = 100
				input.Spec.TableLimits.MaxStorageInGbs = 100
				input.Spec.IsAutoReclaimable = true
				input.Spec.Indexes = []*OciNosqlTableSpec_Index{
					{
						Name: "idx_name",
						Keys: []*OciNosqlTableSpec_Index_IndexKey{
							{ColumnName: "name"},
						},
					},
					{
						Name: "idx_profile_email",
						Keys: []*OciNosqlTableSpec_Index_IndexKey{
							{
								ColumnName:    "profile",
								JsonFieldType: "STRING",
								JsonPath:      "email",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_nosql_table", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidNosqlTable()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidNosqlTable()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidNosqlTable()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciNosqlTable{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciNosqlTable",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-table"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidNosqlTable()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is empty", func() {
				input := minimalValidNosqlTable()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ddl_statement is empty", func() {
				input := minimalValidNosqlTable()
				input.Spec.DdlStatement = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when table_limits is missing", func() {
				input := minimalValidNosqlTable()
				input.Spec.TableLimits = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when max_storage_in_gbs is zero", func() {
				input := minimalValidNosqlTable()
				input.Spec.TableLimits.MaxStorageInGbs = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when provisioned mode with zero max_read_units", func() {
				input := minimalValidNosqlTable()
				input.Spec.TableLimits.CapacityMode = OciNosqlTableSpec_TableLimits_provisioned
				input.Spec.TableLimits.MaxReadUnits = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when provisioned mode with zero max_write_units", func() {
				input := minimalValidNosqlTable()
				input.Spec.TableLimits.CapacityMode = OciNosqlTableSpec_TableLimits_provisioned
				input.Spec.TableLimits.MaxWriteUnits = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when unspecified mode with zero throughput", func() {
				input := minimalValidNosqlTable()
				input.Spec.TableLimits.MaxReadUnits = 0
				input.Spec.TableLimits.MaxWriteUnits = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when index name is empty", func() {
				input := minimalValidNosqlTable()
				input.Spec.Indexes = []*OciNosqlTableSpec_Index{
					{
						Name: "",
						Keys: []*OciNosqlTableSpec_Index_IndexKey{
							{ColumnName: "name"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when index has no keys", func() {
				input := minimalValidNosqlTable()
				input.Spec.Indexes = []*OciNosqlTableSpec_Index{
					{
						Name: "idx_empty",
						Keys: []*OciNosqlTableSpec_Index_IndexKey{},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when index key column_name is empty", func() {
				input := minimalValidNosqlTable()
				input.Spec.Indexes = []*OciNosqlTableSpec_Index{
					{
						Name: "idx_bad",
						Keys: []*OciNosqlTableSpec_Index_IndexKey{
							{ColumnName: ""},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
