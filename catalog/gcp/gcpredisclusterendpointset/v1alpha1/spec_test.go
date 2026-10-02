package gcpredisclusterendpointsetv1alpha1

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
	ginkgo.RunSpecs(t, "GcpRedisClusterEndpointSetSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func nameRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: v}},
	}
}

func pathRef(name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: name, FieldPath: fieldPath}},
	}
}

var _ = ginkgo.Describe("GcpRedisClusterEndpointSetSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	connection := func(rule string, attachmentPath string) *GcpRedisClusterEndpointSetConnection {
		return &GcpRedisClusterEndpointSetConnection{
			ForwardingRule:    nameRef(rule),
			PscConnectionId:   pathRef(rule, "status.outputs.psc_connection_id"),
			Address:           nameRef(rule + "-ip"),
			Network:           nameRef("consumer-vpc"),
			ServiceAttachment: pathRef("orders-cache", attachmentPath),
		}
	}

	// One consumer VPC with a connection per cluster service attachment.
	minimal := func() *GcpRedisClusterEndpointSet {
		return &GcpRedisClusterEndpointSet{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpRedisClusterEndpointSet",
			Metadata:   &shared.CloudResourceMetadata{Name: "orders-cache-endpoints"},
			Spec: &GcpRedisClusterEndpointSetSpec{
				Cluster: nameRef("orders-cache"),
				Region:  "us-central1",
				Endpoints: []*GcpRedisClusterEndpointSetEndpoint{{
					Connections: []*GcpRedisClusterEndpointSetConnection{
						connection("orders-cache-discovery", "status.outputs.discovery_service_attachment"),
						connection("orders-cache-primary", "status.outputs.primary_service_attachment"),
					},
				}},
			},
		}
	}

	ginkgo.It("should accept one consumer network with two connections", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept literals and a second consumer network with an explicit project", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("cache-project")
		msg.Spec.Cluster = litRef("projects/cache-project/locations/us-central1/clusters/orders-cache")
		msg.Spec.Endpoints = append(msg.Spec.Endpoints, &GcpRedisClusterEndpointSetEndpoint{
			Connections: []*GcpRedisClusterEndpointSetConnection{{
				ForwardingRule:    litRef("https://www.googleapis.com/compute/v1/projects/consumer/regions/us-central1/forwardingRules/redis-disc"),
				PscConnectionId:   litRef("1234567890"),
				Address:           litRef("10.20.0.5"),
				Network:           litRef("projects/consumer/global/networks/partner-vpc"),
				ServiceAttachment: litRef("projects/cache-project/regions/us-central1/serviceAttachments/sa-1"),
				ProjectId:         litRef("consumer"),
			}},
		})
		msg.Spec.DeletionPolicy = "ABANDON"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require cluster, region, and at least one endpoint", func() {
		msg := minimal()
		msg.Spec.Cluster = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Region = "us"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Endpoints = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require at least one connection per endpoint", func() {
		msg := minimal()
		msg.Spec.Endpoints[0].Connections = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require every identifying field on a connection", func() {
		for _, mutate := range []func(*GcpRedisClusterEndpointSetConnection){
			func(c *GcpRedisClusterEndpointSetConnection) { c.ForwardingRule = nil },
			func(c *GcpRedisClusterEndpointSetConnection) { c.PscConnectionId = nil },
			func(c *GcpRedisClusterEndpointSetConnection) { c.Address = nil },
			func(c *GcpRedisClusterEndpointSetConnection) { c.Network = nil },
			func(c *GcpRedisClusterEndpointSetConnection) { c.ServiceAttachment = nil },
		} {
			msg := minimal()
			mutate(msg.Spec.Endpoints[0].Connections[0])
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		}
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
