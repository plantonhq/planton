# AWS ECS Environment

The **AWS ECS Environment** InfraChart provisions all cloud resources required to run a containerised service on Amazon
ECS—optionally terminating TLS directly on an Application Load Balancer (ALB).
Like the GKE chart, it now leverages **Jinja‑based conditionals**, so you can turn HTTPS support on or off with a single
flag.

Chart manifests live in the [`templates`](templates) directory; every tunable value is documented in [
`values.yaml`](values.yaml).

---

## Included Cloud Resources (conditional)

| Resource                                         | Always created | Controlled by boolean flag |
|--------------------------------------------------|----------------|----------------------------|
| **Custom VPC** (+ public & private subnets, NAT) | Yes            | —                          |
| **Security Group**                               | Yes            | —                          |
| **Route 53 Hosted Zone**                         | Yes            | —                          |
| **Elastic Container Registry (ECR) Repo**        | Yes            | —                          |
| **ECS Cluster (Fargate + Spot)**                 | Yes            | —                          |
| **Application Load Balancer (ALB)**              | Yes            | —                          |
| **ACM Certificate (DNS‑validated)**              | *No*           | `httpsEnabled`             |
| **ECS Service (+ Task Def)**                     | Yes            | —                          |
| **IAM Task‑Execution Role**                      | Yes            | —                          |

### How the `httpsEnabled` flag works

* `httpsEnabled: true` →

    * Renders an **`AwsCertManagerCert`** resource.
    * Adds an `ssl:` block (certificate ARN) to the ALB spec.
    * Configures the ECS service listener to **443**.
* `httpsEnabled: false` →

    * Omits the certificate and `ssl:` configuration.
    * Sets the listener to plain **80**.

---

## Chart Input Values

Booleans are shown as **unquoted YAML booleans** (`true` /`false`) to avoid string/boolean casting issues.

| Parameter                        | Description                        | Example / Options       | Required / Default   |
|----------------------------------|------------------------------------|-------------------------|----------------------|
| **availability\_zone\_1**        | First AZ for the subnet pair       | `us‑east‑1a`            | Default `us‑east‑1a` |
| **availability\_zone\_2**        | Second AZ for the subnet pair      | `us‑east‑1b`            | Default `us‑east‑1b` |
| **domain\_name**                 | Route 53 zone domain               | `example.com`           | Required             |
| **load\_balancer\_domain\_name** | DNS name served by the ALB         | `app.example.com`       | Required             |
| **service\_name**                | ECS service / task family name     | `nginx`                 | Default `nginx`      |
| **service\_image\_repo\_name**   | ECR repository for images          | `shopping-cart-service` | Required             |
| **service\_port**                | Container port exposed by the task | `8080`                  | Default `8080`       |
| **httpsEnabled**                 | Create ACM cert & terminate TLS    | `true` / `false`        | **Default:** `true`  |
| **alb\_idle\_timeout\_seconds**  | ALB idle timeout                   | `60`                    | Default `60`         |

> **Tip:**Flip `httpsEnabled` to `false` for quick internal environments where plain HTTP is sufficient.

---

## Customisation & Management

* Toggle `httpsEnabled` per environment (dev vs prod) in a higher‑priority`values.yaml`.
* Change `service_port` to expose a different container port; the security group rule automatically follows.
* All cross‑resource references are wired with `valueFrom`; you rarely need to touch the templates.

---

## Important Notes

* Ensure your **Route 53** zone (`domain_name`) already exists, or delegate the domain before applying the chart.
* When `httpsEnabled: true`, ACM issues a *DNS‑validated* certificate—Route 53 records must be publicly resolvable
  during validation.
* The chart creates a **public** ALB. If you need an internal ALB, you can extend the template with a `publicAlbEnabled`
  flag following the same pattern.

---

© Planton. Licensed under [Apache-2.0](../../../LICENSE).
