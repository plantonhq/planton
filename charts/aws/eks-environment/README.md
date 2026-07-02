# EKS Environment InfraChart

This chart provisions a **complete, production‑ready Kubernetes environment on AWS**:

* Custom VPC across two AZs, composed from standalone networking primitives — Internet Gateway, one public and one private subnet per AZ (each with its own route table), Elastic IP(s) and NAT gateway(s)
* Selectable private-subnet egress via `nat_mode`: `single` (one shared NAT, cost‑conscious), `per_az` (one NAT per AZ, highly available), or `none`
* IAM roles for control‑plane and nodes
* Optional customer‑managed KMS key for secrets encryption
* Private or restricted API endpoint with CloudWatch control‑plane logs
* Managed node group with autoscaling, Spot or On‑Demand instances
* Optional Route 53 public zone
* Toggleable Kubernetes add‑ons (Cert‑Manager, External‑DNS, Istio, etc.)

Edit **values.yaml** to tailor the deployment; each `*Enabled` boolean cleanly removes its add‑on.

© Planton. Licensed under [Apache-2.0](../../../LICENSE).
