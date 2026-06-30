# OpenBAO Ingress Verification

This document describes how to verify that the Istio Gateway API ingress for a KubernetesOpenBao deployment is working correctly.

## Prerequisites

- [OpenBAO CLI](https://openbao.org/docs/install/) (`bao`) installed on your machine
- Access to the deployment's external hostname
- The OpenBAO root token or a valid service token

### Installing the OpenBAO CLI

```bash
# macOS (Homebrew)
brew install openbao

# Verify installation
bao version
```

## Step 1: Check Server Status

Set the environment variables for your deployment and check server status:

```bash
export BAO_ADDR="https://<your-openbao-hostname>"
export BAO_TOKEN="<your-token>"

bao status
```

**Expected output** (healthy, unsealed server):

```
Key             Value
---             -----
Seal Type       shamir
Initialized     true
Sealed          false
Total Shares    1
Threshold       1
Version         2.x.x
Storage Type    file
HA Enabled      false
```

If this succeeds, the ingress is routing traffic correctly:
- DNS resolves the hostname to the Istio ingress gateway
- The TLS certificate is valid (cert-manager issued it)
- The Gateway is terminating TLS and forwarding to the OpenBao service on port 8200

**Common failures at this step:**

| Symptom | Likely Cause |
|---------|-------------|
| `connection refused` | Gateway or HTTPRoute not created, or DNS not resolving |
| `certificate signed by unknown authority` | cert-manager Certificate not ready, or ClusterIssuer misconfigured |
| `503 Service Unavailable` | HTTPRoute `backendRef` pointing to wrong service name or port |
| `Sealed: true` | Server needs unsealing (not an ingress issue) |

## Step 2: Verify Secrets Engines

List the mounted secrets engines to confirm full API access:

```bash
bao secrets list
```

**Expected output** (for a standard deployment with KV v2 and Transit):

```
Path          Type         Description
----          ----         -----------
cubbyhole/    cubbyhole    per-token private secret storage
identity/     identity     identity store
secret/       kv           n/a
sys/          system       system endpoints
transit/      transit      n/a
```

## Step 3: Test KV v2 Read/Write

Write a test secret, read it back, then clean up:

```bash
# Write
bao kv put secret/ingress-test message="ingress is working"

# Read
bao kv get secret/ingress-test

# Clean up
bao kv delete secret/ingress-test
```

## Step 4: Test Transit Engine

Create a test encryption key, verify it, then clean up:

```bash
# Create a key
bao write -f transit/keys/ingress-test type=aes256-gcm96

# Encrypt some data
bao write transit/encrypt/ingress-test plaintext=$(echo -n "hello" | base64)

# Clean up (must enable deletion first)
bao write transit/keys/ingress-test/config deletion_allowed=true
bao delete transit/keys/ingress-test
```

## Step 5: Verify HTTP-to-HTTPS Redirect

Confirm that HTTP requests are redirected to HTTPS with a 301:

```bash
curl -v -o /dev/null http://<your-openbao-hostname>/v1/sys/health 2>&1 | grep "< HTTP"
```

**Expected output:**

```
< HTTP/1.1 301 Moved Permanently
```

## Troubleshooting

### Inspecting Gateway API Resources

If ingress is not working, inspect the Kubernetes resources directly:

```bash
# Check the Certificate status
kubectl get certificate -n istio-ingress | grep <metadata-name>
kubectl describe certificate <metadata-name>-certificate -n istio-ingress

# Check the Gateway
kubectl get gateway -n istio-ingress | grep <metadata-name>
kubectl describe gateway <metadata-name>-external -n istio-ingress

# Check the HTTPRoutes
kubectl get httproute -n <namespace> | grep <metadata-name>
kubectl describe httproute <metadata-name>-https -n <namespace>
kubectl describe httproute <metadata-name>-http-redirect -n <namespace>
```

### Resource Naming Convention

All ingress resources are named using the deployment's `metadata.name`:

| Resource | Name Pattern | Namespace |
|----------|-------------|-----------|
| Certificate | `{name}-certificate` | `istio-ingress` |
| Gateway | `{name}-external` | `istio-ingress` |
| HTTPRoute (HTTPS) | `{name}-https` | deployment namespace |
| HTTPRoute (redirect) | `{name}-http-redirect` | deployment namespace |
| TLS Secret | `{name}-tls` | `istio-ingress` |

### Port-Forwarding as a Bypass

If ingress is not working but you need immediate access, use port-forwarding:

```bash
kubectl port-forward -n <namespace> service/<metadata-name> 8200:8200
export BAO_ADDR="http://localhost:8200"
bao status
```

This bypasses the ingress entirely and connects directly to the OpenBao service.
