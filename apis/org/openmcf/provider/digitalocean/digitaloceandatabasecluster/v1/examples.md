# DigitalOcean Database Cluster Examples

Complete, copy-paste ready YAML manifests for common database deployment scenarios.

---

## Example 1: Development PostgreSQL (Single Node)

**Use Case**: Non-production PostgreSQL for local development or testing.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDatabaseCluster
metadata:
  name: dev-postgres
spec:
  clusterName: dev-postgres
  engine: postgres
  engineVersion: "16"
  region: nyc3
  sizeSlug: db-s-1vcpu-1gb
  nodeCount: 1
  enablePublicConnectivity: true  # For dev access from local machine
```

**Notes:**
- Single node = no HA (acceptable for dev/test)
- Public connectivity enabled for ease of access
- Smallest/cheapest size slug ($15/month)

---

## Example 2: Production PostgreSQL (High Availability with VPC)

**Use Case**: Mission-critical PostgreSQL with automatic failover.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDatabaseCluster
metadata:
  name: prod-postgres
spec:
  clusterName: prod-postgres
  engine: postgres
  engineVersion: "16"
  region: nyc3
  sizeSlug: db-s-4vcpu-8gb
  nodeCount: 3  # Primary + 2 standbys for maximum HA
  vpc:
    value: "12345678-1234-1234-1234-123456789012"
  storageGib: 200
  enablePublicConnectivity: false
```

**Notes:**
- 3 nodes for maximum resilience
- VPC-private (no public access)
- Custom storage (200GB instead of default)
- Production-sized instance (4 vCPU, 8GB RAM)

**Required Follow-Up:**
1. Configure PgBouncer connection pool (separate resource)
2. Set up firewall rules (tag-based or VPC CIDR)
3. Configure application with private_uri output

---

## Example 3: MySQL for Web Applications

**Use Case**: MySQL cluster for content management systems or e-commerce.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDatabaseCluster
metadata:
  name: prod-mysql
spec:
  clusterName: prod-mysql
  engine: mysql
  engineVersion: "8"
  region: sfo3
  sizeSlug: db-s-2vcpu-4gb
  nodeCount: 2
  vpc:
    value: "vpc-87654321"
  storageGib: 100
  enablePublicConnectivity: false
```

**Notes:**
- MySQL 8 (latest stable)
- 2-node HA configuration
- San Francisco region for US West users
- Moderate storage for typical web app

**MySQL Gotchas:**
- No SUPER privileges (cannot SET GLOBAL variables)
- Restricted mysqldump (use DigitalOcean backups instead)
- See research doc for migration considerations

---

## Example 4: Redis Cache Cluster

**Use Case**: In-memory caching layer for application performance.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDatabaseCluster
metadata:
  name: prod-redis-cache
spec:
  clusterName: prod-redis-cache
  engine: redis
  engineVersion: "7"
  region: nyc3
  sizeSlug: db-s-2vcpu-4gb
  nodeCount: 2
  vpc:
    value: "vpc-cache-12345"
  enablePublicConnectivity: false
```

**Notes:**
- Redis 7 (actually Valkey due to SSPL licensing)
- HA configuration (primary + standby)
- VPC-private for security
- No clustering mode (single-instance limit)

**Redis Limitations:**
- No Redis Cluster mode support
- Persistence is available but Redis should be treated as ephemeral
- Use for caching/sessions, not primary data store

---

## Example 5: MongoDB Document Store (High Availability)

**Use Case**: Document-oriented database for flexible schema applications.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDatabaseCluster
metadata:
  name: prod-mongodb
spec:
  clusterName: prod-mongodb
  engine: mongodb
  engineVersion: "7.0"
  region: fra1
  sizeSlug: db-s-4vcpu-8gb
  nodeCount: 3
  vpc:
    value: "vpc-eu-67890"
  storageGib: 500
  enablePublicConnectivity: false
```

**Notes:**
- MongoDB 7.0 (latest)
- 3-node replica set for HA
- Frankfurt region for EU data residency
- Large storage (500GB) for document collections

**MongoDB Limitations:**
- No sharding support (replica sets only)
- Max 3 nodes (cannot scale beyond this)

---

## Example 6: Production PostgreSQL with VPC Reference

**Use Case**: Reference VPC from another OpenMCF resource.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDatabaseCluster
metadata:
  name: app-database
spec:
  clusterName: app-database
  engine: postgres
  engineVersion: "15"
  region: nyc3
  sizeSlug: db-s-4vcpu-8gb
  nodeCount: 3
  vpc:
    ref:
      kind: DigitalOceanVpc
      name: production-vpc
      field_path: status.outputs.vpc_id
  storageGib: 250
  enablePublicConnectivity: false
```

**Notes:**
- Uses foreign key reference to DigitalOceanVpc resource
- field_path extracts vpc_id from VPC resource status
- Ensures database and VPC are deployed in dependency order

---

## Example 7: Multi-Environment Setup

**Use Case**: Consistent configuration across dev/staging/prod with parameter differences.

**Development:**
```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDatabaseCluster
metadata:
  name: dev-app-db
spec:
  clusterName: dev-app-db
  engine: postgres
  engineVersion: "16"
  region: nyc3
  sizeSlug: db-s-1vcpu-1gb
  nodeCount: 1
  enablePublicConnectivity: true
```

**Staging:**
```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDatabaseCluster
metadata:
  name: staging-app-db
spec:
  clusterName: staging-app-db
  engine: postgres
  engineVersion: "16"
  region: nyc3
  sizeSlug: db-s-2vcpu-4gb
  nodeCount: 2
  vpc:
    value: "vpc-staging-123"
  enablePublicConnectivity: false
```

**Production:**
```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDatabaseCluster
metadata:
  name: prod-app-db
spec:
  clusterName: prod-app-db
  engine: postgres
  engineVersion: "16"
  region: nyc3
  sizeSlug: db-s-4vcpu-8gb
  nodeCount: 3
  vpc:
    value: "vpc-prod-456"
  storageGib: 300
  enablePublicConnectivity: false
```

**Pattern Highlights:**
- Dev: Single node, public access, smallest size
- Staging: 2 nodes (HA), VPC-private, moderate size
- Production: 3 nodes (max HA), VPC-private, large storage

---

## Configuration Patterns Summary

| Use Case | Engine | Nodes | Size | Public | VPC | Monthly Cost |
|----------|--------|-------|------|--------|-----|--------------|
| Dev/Test | Any | 1 | db-s-1vcpu-1gb | Yes | Optional | ~$15 |
| Small Prod | postgres/mysql | 2 | db-s-2vcpu-4gb | No | Required | ~$120 |
| Large Prod | Any | 3 | db-s-4vcpu-8gb | No | Required | ~$360 |
| Cache Layer | redis | 2 | db-s-2vcpu-4gb | No | Required | ~$120 |
| Document Store | mongodb | 3 | db-s-4vcpu-8gb | No | Required | ~$360 |

---

## Deployment Checklist

Before deploying, ensure:

- ✅ `cluster_name` is unique in your DigitalOcean account
- ✅ `engine_version` is supported (check DigitalOcean docs for current versions)
- ✅ `size_slug` is valid for selected engine
- ✅ `node_count` ≥ 2 for production (HA requirement)
- ✅ `vpc` is specified for production deployments
- ✅ `enable_public_connectivity: false` for production
- ✅ `storage_gib` accounts for expected data growth (cannot shrink)

---

## Next Steps After Deployment

1. **Retrieve Connection Details**:
   ```bash
   openmcf get outputs <cluster-name>
   # Outputs: connection_uri, host, port, username, password
   ```

2. **Configure Firewall** (separate resource):
   ```yaml
   apiVersion: digital-ocean.openmcf.org/v1
   kind: DigitalOceanDatabaseFirewall
   metadata:
     name: prod-db-firewall
   spec:
     cluster_id: ${cluster_id_from_outputs}
     rules:
       - type: vpc
         value: "10.10.0.0/16"
   ```

3. **Set Up PgBouncer** (PostgreSQL only):
   ```yaml
   apiVersion: digital-ocean.openmcf.org/v1
   kind: DigitalOceanDatabaseConnectionPool
   metadata:
     name: app-pool
   spec:
     cluster_id: ${cluster_id}
     name: app-pool
     mode: transaction
     size: 25
   ```

4. **Test Connectivity**:
   ```bash
   # PostgreSQL
   psql "${connection_uri}"
   
   # MySQL
   mysql -h ${host} -P ${port} -u ${username} -p
   
   # Redis
   redis-cli -h ${host} -p ${port} -a ${password}
   
   # MongoDB
   mongosh "${connection_uri}"
   ```

---

For more details, see:
- [README.md](./README.md) - Component overview and best practices
- [docs/README.md](./docs/README.md) - Comprehensive production guide
- [iac/pulumi/README.md](./iac/pulumi/README.md) - Pulumi module usage
- [iac/tf/README.md](./iac/tf/README.md) - Terraform module usage

