# Overview

The AwsIamInstanceProfile API resource provisions an IAM instance profile: the
container that delivers an IAM role to EC2 instances. EC2 cannot assume a role
directly -- it can only be launched with an instance profile carrying one.

## Why We Created This API Resource

The instance profile is a real AWS object with its own lifecycle, sitting
between a role and the EC2-shaped resources that use it. Modeling it as a
first-class component keeps the identity graph honest. It lets you:

- **Keep roles universal**: a role serves Lambda, ECS, EKS, and EC2 alike; only
  EC2 needs the profile wrapper, so only EC2 topologies create one.
- **Compose by reference**: the profile references an `AwsIamRole`'s
  `role_name` output, and EC2 instances, launch templates, and Auto Scaling
  groups reference the profile's `instance_profile_arn` output.
- **Swap roles without churn**: the carried role can be changed in place -- the
  profile (and everything referencing it) stays put while running instances
  pick up the new role's credentials on their next metadata refresh.

## Key Features

### Role Delivery

- **One role per profile**: mirrors the AWS limit exactly; the role is attached
  by name, as the AWS API requires.
- **Reference or literal**: point at an `AwsIamRole` component or pass the name
  of a role that exists outside Planton.

### Organization

- **IAM path**: group profiles into path hierarchies (e.g. `/compute/`) that
  IAM policies can match with wildcards.

## Benefits

- **Honest composition**: the EC2 identity chain (role → profile → instance) is
  three referenceable nodes, not a hidden side effect of any one of them.
- **Contained blast radius**: profile changes never risk the role or its
  attachments; role swaps never replace the profile.
- **Consistency**: identical behavior across Terraform and Pulumi.

## Stack outputs

- `instance_profile_arn`: ARN of the profile (what an EC2 instance references)
- `instance_profile_name`: friendly name (launch templates take the profile by name)
- `instance_profile_id`: stable unique ID AWS assigns to the profile
- `role_name`: name of the IAM role the profile carries
