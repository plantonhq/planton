package alicloudrocketmqinstancev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	fkv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAliCloudRocketmqInstanceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AliCloudRocketmqInstanceSpec Validation Tests")
}

var _ = ginkgo.Describe("AliCloudRocketmqInstanceSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "dev-mq",
				},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region:        "cn-hangzhou",
					SeriesCode:    "standard",
					SubSeriesCode: "single_node",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-abc123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with professional cluster_ha and topics", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "prod-mq",
					Org:  "acme-corp",
					Env:  "production",
				},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region:        "cn-shanghai",
					SeriesCode:    "professional",
					SubSeriesCode: "cluster_ha",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-prod"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-prod-a"},
					},
					MsgProcessSpec: "rmq.p2.4xlarge",
					PaymentType:    proto.String("PayAsYouGo"),
					Topics: []*AliCloudRocketmqTopic{
						{TopicName: "order-events", MessageType: proto.String("NORMAL")},
						{TopicName: "payment-events", MessageType: proto.String("FIFO")},
						{TopicName: "delay-notifications", MessageType: proto.String("DELAY")},
					},
					ConsumerGroups: []*AliCloudRocketmqConsumerGroup{
						{ConsumerGroupId: "GID_order_processor"},
						{
							ConsumerGroupId:   "GID_payment_processor",
							DeliveryOrderType: proto.String("Orderly"),
							ConsumeRetryPolicy: &AliCloudRocketmqConsumeRetryPolicy{
								RetryPolicy:   proto.String("FixedRetryPolicy"),
								MaxRetryTimes: proto.Int32(5),
							},
						},
					},
					Tags: map[string]string{"team": "platform"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with subscription billing and auto-renew", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "prepaid-mq",
				},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region:        "cn-hangzhou",
					SeriesCode:    "ultimate",
					SubSeriesCode: "cluster_ha",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					PaymentType:     proto.String("Subscription"),
					Period:          proto.Int32(12),
					PeriodUnit:      proto.String("Month"),
					AutoRenew:       proto.Bool(true),
					AutoRenewPeriod: proto.Int32(3),
					MsgProcessSpec:  "rmq.u2.4xlarge",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with internet access enabled", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "internet-mq",
				},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region:        "cn-hangzhou",
					SeriesCode:    "professional",
					SubSeriesCode: "cluster_ha",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					InternetInfo: &AliCloudRocketmqInternetInfo{
						Enabled:     proto.Bool(true),
						FlowOutType: proto.String("payByTraffic"),
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with product info and encryption", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "encrypted-mq",
				},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region:        "cn-hangzhou",
					SeriesCode:    "professional",
					SubSeriesCode: "cluster_ha",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					MsgProcessSpec: "rmq.p2.4xlarge",
					ProductInfo: &AliCloudRocketmqProductInfo{
						MessageRetentionTime: proto.Int32(168),
						SendReceiveRatio:     proto.Float64(0.3),
						AutoScaling:          proto.Bool(true),
						TraceOn:              proto.Bool(true),
						StorageEncryption:    proto.Bool(true),
						StorageSecretKey:     "kms-key-abc123",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with consumer group using default retry policy", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "cg-test",
				},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region:        "cn-hangzhou",
					SeriesCode:    "standard",
					SubSeriesCode: "single_node",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					ConsumerGroups: []*AliCloudRocketmqConsumerGroup{
						{
							ConsumerGroupId: "GID_simple",
						},
						{
							ConsumerGroupId: "GID_with_dlq",
							ConsumeRetryPolicy: &AliCloudRocketmqConsumeRetryPolicy{
								RetryPolicy:           proto.String("FixedRetryPolicy"),
								MaxRetryTimes:         proto.Int32(3),
								DeadLetterTargetTopic: "dead-letter-topic",
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with TRANSACTION message type", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "tx-mq",
				},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region:        "cn-hangzhou",
					SeriesCode:    "professional",
					SubSeriesCode: "cluster_ha",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					Topics: []*AliCloudRocketmqTopic{
						{TopicName: "tx-topic", MessageType: proto.String("TRANSACTION")},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "wrong/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region: "cn-hangzhou", SeriesCode: "standard", SubSeriesCode: "single_node",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region: "cn-hangzhou", SeriesCode: "standard", SubSeriesCode: "single_node",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Spec: &AliCloudRocketmqInstanceSpec{
					Region: "cn-hangzhou", SeriesCode: "standard", SubSeriesCode: "single_node",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when spec is missing", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when region is empty", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region: "", SeriesCode: "standard", SubSeriesCode: "single_node",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when series_code is invalid", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region: "cn-hangzhou", SeriesCode: "enterprise", SubSeriesCode: "single_node",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when sub_series_code is invalid", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region: "cn-hangzhou", SeriesCode: "standard", SubSeriesCode: "multi_node",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vpc_id is missing", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region: "cn-hangzhou", SeriesCode: "standard", SubSeriesCode: "single_node",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when payment_type is invalid", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region: "cn-hangzhou", SeriesCode: "standard", SubSeriesCode: "single_node",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					PaymentType: proto.String("Free"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when message_type is invalid", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region: "cn-hangzhou", SeriesCode: "standard", SubSeriesCode: "single_node",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					Topics: []*AliCloudRocketmqTopic{
						{TopicName: "bad-topic", MessageType: proto.String("BROADCAST")},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when topic_name is empty", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region: "cn-hangzhou", SeriesCode: "standard", SubSeriesCode: "single_node",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					Topics: []*AliCloudRocketmqTopic{
						{TopicName: ""},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when consumer_group_id is empty", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region: "cn-hangzhou", SeriesCode: "standard", SubSeriesCode: "single_node",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					ConsumerGroups: []*AliCloudRocketmqConsumerGroup{
						{ConsumerGroupId: ""},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when delivery_order_type is invalid", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region: "cn-hangzhou", SeriesCode: "standard", SubSeriesCode: "single_node",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					ConsumerGroups: []*AliCloudRocketmqConsumerGroup{
						{ConsumerGroupId: "GID_test", DeliveryOrderType: proto.String("Random")},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when auto_renew_period is invalid", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region: "cn-hangzhou", SeriesCode: "standard", SubSeriesCode: "single_node",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					AutoRenewPeriod: proto.Int32(5),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when max_retry_times exceeds 1000", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region: "cn-hangzhou", SeriesCode: "standard", SubSeriesCode: "single_node",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					ConsumerGroups: []*AliCloudRocketmqConsumerGroup{
						{
							ConsumerGroupId: "GID_test",
							ConsumeRetryPolicy: &AliCloudRocketmqConsumeRetryPolicy{
								MaxRetryTimes: proto.Int32(1001),
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when flow_out_type is invalid", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region: "cn-hangzhou", SeriesCode: "standard", SubSeriesCode: "single_node",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					InternetInfo: &AliCloudRocketmqInternetInfo{
						Enabled:     proto.Bool(true),
						FlowOutType: proto.String("payByRequest"),
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when retry_policy is invalid", func() {
			input := &AliCloudRocketmqInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudRocketmqInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRocketmqInstanceSpec{
					Region: "cn-hangzhou", SeriesCode: "standard", SubSeriesCode: "single_node",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					ConsumerGroups: []*AliCloudRocketmqConsumerGroup{
						{
							ConsumerGroupId: "GID_test",
							ConsumeRetryPolicy: &AliCloudRocketmqConsumeRetryPolicy{
								RetryPolicy: proto.String("ExponentialBackoff"),
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
