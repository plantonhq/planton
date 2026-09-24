# An App Runner service and an EKS node group reach what they name and never live in it

## What changed

- **An App Runner VPC ingress connection's `vpc_id` is containment-exempt.** `AwsAppRunnerServiceVpcIngressConnection.vpc_id` names the VPC the service is *published into* over PrivateLink: clients inside that VPC reach the service, but the service runs on App Runner's managed infrastructure and is never deployed into the network (a VPC connector is what places a service's egress in a VPC, and that remains placement). Until now the reference was placement by omission, so a private service with an ingress connection was drawn inside the VPC that can merely reach it.
- **An EKS node group's and a Fargate profile's `subnet_ids` are containment-exempt.** `AwsEksNodeGroupSpec.subnet_ids` and `AwsEksFargateProfileSpec.subnet_ids` sat beside each kind's `cluster_name` as a second placement claim. A node group is a member of its cluster; its subnets are where its nodes attach. With both references placement, a group on subnets its cluster does not name fell to a mixed-kind tie -- drawn by luck on a project diagram, floating outside its cluster on the estate. The same correction the AKS node pool's node and pod subnets already carry.
- The containment-decision registry moves exactly those three lines from `contained` to `exempt`; nothing else moved.

## Why

`container_kind` says a kind is a box other resources nest inside; `containment_exempt` says a reference into such a box is access, not placement. A VPC is the box every AWS network resource lives in, and a subnet is the box every AWS compute resource lives in -- which is exactly why references into them must be read carefully. An ingress connection *opens a door* into a VPC. A node group *attaches nodes* to subnets while belonging to a cluster. Neither is a tenant of what it names, and a diagram that put them there would be a true sentence about the wrong subject.

## How to check

```bash
go test ./shared/cloudresourcekind/ -run TestContainmentDecisions   # green; the golden carries the three exempt lines
grep -n containment_exempt catalog/aws/awsapprunnerservice/v1alpha1/spec.proto catalog/aws/awseksnodegroup/v1alpha1/spec.proto catalog/aws/awseksfargateprofile/v1alpha1/spec.proto
```
