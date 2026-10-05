<p align="center">
  <img src="logo.svg" alt="AWS REST API VPC Link" width="80"/>
</p>

# AWS REST API VPC Link

Create and manage an [API Gateway v1 VPC link](https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-nlb-for-vpclink-using-console.html)
— the NLB attachment that lets REST API integrations reach private
services inside a VPC.

One link is shared by many APIs and owns its own network attachment,
which is why it is its own kind rather than a field on
[AwsRestApiGateway](../awsrestapigateway).

HTTP APIs use a different link that attaches to subnets directly —
that is [AwsHttpApiVpcLink](../awshttpapivpclink). The two are not
interchangeable.

## What Gets Created

- **A VPC link** fronting exactly one Network Load Balancer. AWS
  cannot change the balancer after create — a different NLB means a
  new link.

Provisioning takes several minutes while AWS builds the network
attachment. Creating the link is free; standard NLB charges apply to
the balancer it fronts.

See [v1alpha1/reference.md](v1alpha1/reference.md) for the full field
reference generated from the spec proto.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
