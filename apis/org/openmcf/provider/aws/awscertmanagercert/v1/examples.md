# AWS Cert Manager Cert Examples

## Create Using CLI

Create a YAML manifest using one of the examples below. After the YAML is created, apply it with OpenMCF:

```shell
openmcf pulumi up --manifest <yaml-path> --stack <stack-name>
```

Or, if using Terraform:

```shell
openmcf terraform apply --manifest <yaml-path> --stack <stack-name>
```

---

## Basic Example (Single Domain)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCertManagerCert
metadata:
  name: single-cert-example
spec:
  region: us-east-1
  primaryDomainName: example.com
  route53HostedZoneId: Z123456ABCXYZ
```

This manifest requests an SSL certificate for `example.com` and uses Route53 hosted zone `Z123456ABCXYZ` for DNS validation. The validation method defaults to DNS.

---

## Multiple Domains Example

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCertManagerCert
metadata:
  name: multi-cert-example
spec:
  region: us-east-1
  primaryDomainName: myapp.io
  alternateDomainNames:
    - api.myapp.io
    - blog.myapp.io
    - www.myapp.io
  route53HostedZoneId: Z987654XYZABC
  validationMethod: DNS
```

In this example, we issue a certificate for `myapp.io` as the primary domain and add three alternate domains as Subject Alternative Names (SANs). DNS validation records are automatically created in the Route53 hosted zone.

---

## Wildcard Domain Example

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCertManagerCert
metadata:
  name: wildcard-example
spec:
  region: us-east-1
  primaryDomainName: '*.example.com'
  route53HostedZoneId: Z999999EXAMPLE
```

This manifest requests a wildcard certificate for `*.example.com`, allowing you to secure any subdomain under `example.com` without creating separate certificates. Wildcard certificates require DNS validation.

---

## Wildcard with Root Domain

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCertManagerCert
metadata:
  name: wildcard-plus-root
spec:
  region: us-east-1
  primaryDomainName: '*.production.com'
  alternateDomainNames:
    - production.com
  route53HostedZoneId: Z1111111111111
```

This configuration creates a certificate that covers both the wildcard (`*.production.com`) and the root domain (`production.com`). This is useful when you need to secure both `production.com` and `api.production.com`, `www.production.com`, etc.

---

## Minimal Example

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCertManagerCert
metadata:
  name: minimal-cert
spec:
  region: us-east-1
  primaryDomainName: simple.io
  route53HostedZoneId: Z222222222222
```

Here, we only specify the required fields. The certificate will be created for `simple.io` with DNS validation (default), and validation records will be automatically managed in the hosted zone.

---

## Using Foreign Key References

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCertManagerCert
metadata:
  name: cert-with-ref
spec:
  region: us-east-1
  primaryDomainName: app.mydomain.com
  alternateDomainNames:
    - api.mydomain.com
  route53HostedZoneId:
    valueFrom:
      kind: AwsRoute53Zone
      name: my-hosted-zone
      field: status.outputs.zone_id
```

This example demonstrates using a foreign key reference to dynamically retrieve the Route53 hosted zone ID from another OpenMCF resource. This pattern promotes composable infrastructure definitions.

---

## Email Validation Example (Alternative)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCertManagerCert
metadata:
  name: email-validation-cert
spec:
  region: us-east-1
  primaryDomainName: legacy.com
  route53HostedZoneId: Z333333333333
  validationMethod: EMAIL
```

This example uses email validation instead of DNS. AWS will send validation emails to the domain's administrative contacts. **Note**: Email validation is not recommended as it requires manual intervention and doesn't support wildcard certificates. DNS validation is the preferred method.

---

## After Deploying

Once you've chosen one of these examples, apply it with OpenMCF using Pulumi or Terraform:

```shell
openmcf pulumi up --manifest aws-cert-manager-cert.yaml --stack myorg/dev
```

or

```shell
openmcf terraform apply --manifest aws-cert-manager-cert.yaml --stack myorg/dev
```

When the command completes successfully, your SSL certificate will be provisioned in ACM. You can confirm by checking the AWS console or by using the AWS CLI:

```shell
aws acm list-certificates
```

You should see your new certificate and any domain validation records created within Route53. The certificate status should show as "ISSUED" once validation is complete.

---

Enjoy secure connections with OpenMCF!
