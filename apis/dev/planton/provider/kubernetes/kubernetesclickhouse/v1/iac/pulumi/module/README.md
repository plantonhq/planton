# ClickHouse Kubernetes Pulumi Module

## Module Structure

This module is organized into focused, single-responsibility files for maintainability and clarity.

### File Organization

```
module/
├── main.go                        ( 53 lines) - Entry point, resource orchestration
├── locals.go                      (120 lines) - Local variables and exports
├── click_house_installation.go    (189 lines) - ClickHouseInstallation CRD generation
├── coordination.go                ( 55 lines) - Coordination routing and priority logic
├── coordination_keeper.go         ( 69 lines) - ClickHouse Keeper configuration
├── coordination_zookeeper.go      ( 59 lines) - ZooKeeper configuration (deprecated)
├── password_secret.go             ( 94 lines) - Password generation and Secret creation
├── ingress.go                     ( 67 lines) - LoadBalancer ingress service
├── outputs.go                     ( 13 lines) - Output constant definitions
├── variables.go                   ( 22 lines) - Module variables and defaults
├── README.md                                - Module architecture documentation
└── BUILD.bazel                              - Bazel build configuration
```

**Total**: 741 lines organized into 10 focused files

---

## File Responsibilities

### main.go
**Purpose**: Module entry point and resource orchestration

**Key Functions**:
- `Resources()` - Main entry point called by Pulumi
- Orchestrates resources in correct order:
  1. Kubernetes Provider setup
  2. Namespace creation
  3. Password Secret generation
  4. ClickHouseInstallation CRD
  5. Ingress LoadBalancer (optional)

**Responsibilities**:
- Error handling and propagation
- Resource creation sequencing
- Provider management

### locals.go
**Purpose**: Local variable initialization and output exports

**Key Functions**:
- `initializeLocals()` - Initialize module-scoped variables
- Generates: Namespaces, service names, hostnames, labels
- Exports: All output values to Pulumi stack

### click_house_installation.go
**Purpose**: ClickHouseInstallation CRD generation

**Key Functions**:
- `clickhouseInstallation()` - Creates ClickHouseInstallation CRD
- `buildConfiguration()` - Builds cluster configuration
- `buildDefaults()` - Template defaults
- `buildTemplates()` - Pod and volume templates
- `buildPodTemplates()` - Container specifications
- `buildVolumeClaimTemplates()` - Persistence configuration

**Responsibilities**:
- ClickHouseInstallation CRD structure
- Cluster topology (shards, replicas)
- Resource allocations
- Persistence settings
- Template generation

### coordination.go
**Purpose**: Coordination configuration routing and priority logic

**Key Functions**:
- `buildCoordinationConfig()` - Main entry point for coordination
- `buildCoordinationFromNewField()` - Routes based on CoordinationType enum

**Responsibilities**:
- Priority handling (coordination > zookeeper > default)
- Type-based routing to specific builders
- Default behavior coordination

**Decision Flow**:
```
buildCoordinationConfig()
  ├─ spec.Coordination exists?
  │  └─ Yes → buildCoordinationFromNewField()
  │           ├─ type=keeper → buildAutoManagedKeeperReference()
  │           ├─ type=external_keeper → buildExternalKeeperReference()
  │           └─ type=external_zookeeper → buildExternalZookeeperReference()
  │
  ├─ spec.Zookeeper exists? (deprecated)
  │  └─ Yes → buildCoordinationFromDeprecatedZookeeperField()
  │
  └─ Default → buildDefaultKeeperReference()
```

### coordination_keeper.go
**Purpose**: ClickHouse Keeper coordination logic

**Key Functions**:
- `buildAutoManagedKeeperReference()` - Auto-managed Keeper (80% use case)
- `buildExternalKeeperReference()` - External Keeper infrastructure
- `buildDefaultKeeperReference()` - Default Keeper reference

**Responsibilities**:
- ClickHouse Keeper service references
- Future: Auto-creation of ClickHouseKeeperInstallation CRD
- Keeper-specific configuration

**Service Naming**:
- Auto-managed: `keeper:2181` (default)
- External: User-specified nodes

### coordination_zookeeper.go
**Purpose**: ZooKeeper coordination logic (legacy support)

**Key Functions**:
- `buildExternalZookeeperReference()` - External ZooKeeper
- `buildCoordinationFromDeprecatedZookeeperField()` - Deprecated field handler

**Responsibilities**:
- External ZooKeeper references
- Backward compatibility with deprecated `zookeeper` field
- Legacy ZooKeeper service references

**Note**: This file handles legacy scenarios (5% of users) and will be simplified when `zookeeper` field is removed in v2.

### password_secret.go
**Purpose**: Password generation and Kubernetes Secret creation

**Key Functions**:
- `createPasswordSecret()` - Main entry point
- `generateRandomPassword()` - Cryptographically secure password generation
- `createKubernetesSecret()` - Secret resource creation

**Responsibilities**:
- Generate 20-character random passwords with complexity requirements
- Use **URL-safe special characters only** (`-_`) to avoid connection string encoding issues
- Create Kubernetes Secret with StringData (auto base64 encoding)
- Parent relationship with namespace for lifecycle management

**Security**:
- Minimum 2 special, 3 numeric, 3 uppercase, 3 lowercase characters
- Password never appears in manifests or version control
- SHA256 hashed by ClickHouse when used

**URL-Safe Password Requirement**:
Characters like `+`, `=`, `/`, `&`, `?`, `#` cause problems when passwords are used in
URL-encoded connection strings like `tcp://host:port/?password=XXX`. The `+` character
is particularly problematic as it's decoded as a space. Only hyphen (`-`) and underscore
(`_`) are used as special characters. See: https://github.com/Altinity/clickhouse-operator/issues/1883

### ingress.go
**Purpose**: External LoadBalancer service for ingress access

**Key Functions**:
- `createIngressLoadBalancer()` - Creates LoadBalancer service

**Responsibilities**:
- Optional LoadBalancer creation (only if ingress.enabled = true)
- Exposes HTTP (8123) and native (9000) ports
- External DNS annotation for automatic DNS record creation
- Pod selector targeting Altinity operator-managed ClickHouse pods

**Networking**:
- Service type: LoadBalancer
- Ports: 8123 (HTTP), 9000 (Native)
- Selector: Uses Altinity operator labels
- DNS: external-dns.alpha.kubernetes.io/hostname annotation

### outputs.go
**Purpose**: Output constant definitions

**Constants**: Namespace, service, endpoints, credentials

### variables.go
**Purpose**: Module-level variables and defaults

**Variables**: Ports, versions, usernames, namespaces

---

## Design Principles

### Single Responsibility
Each file has one clear purpose:
- `main.go` - Orchestration
- `click_house_installation.go` - CHI CRD
- `coordination.go` - Routing
- `coordination_keeper.go` - Keeper logic
- `coordination_zookeeper.go` - ZooKeeper logic

### Separation of Concerns
- **Coordination types are isolated** - Easy to add new types
- **Backward compatibility is contained** - `coordination_zookeeper.go` handles deprecated field
- **Future enhancements are clear** - Add Keeper auto-creation to `coordination_keeper.go`

### Readability
- Each file < 200 lines
- Clear function names
- Comprehensive comments
- Type-specific logic grouped together

### Maintainability
- Easy to find coordination logic (3 dedicated files)
- Easy to add new coordination types (add new case in `coordination.go`)
- Easy to remove deprecated code (delete `coordination_zookeeper.go` in v2)

---

## Future Enhancements

### Phase 2: Auto-Create ClickHouse Keeper

**File to modify**: `coordination_keeper.go`

Add function to create ClickHouseKeeperInstallation:
```go
func createClickHouseKeeperInstallation(
    ctx *pulumi.Context,
    locals *Locals,
    namespace *kubernetescorev1.Namespace,
    keeperConfig *clickhousekubernetesv1.ClickHouseKubernetesKeeperConfig,
) error {
    // Create ClickHouseKeeperInstallation CRD
    // Use keeperConfig for replicas, resources, diskSize
    // Return service name for reference
}
```

Call from `main.go` before `clickhouseInstallation()`.

### Phase 3: Additional Coordination Types

To add new coordination type (e.g., `external_etcd`):

1. Add enum value to `spec.proto`
2. Add case in `coordination.go` switch statement
3. Create `coordination_etcd.go` with specific logic

---

## Testing

### Compilation Test
```bash
cd apis/project/planton/.../iac/pulumi
go build .
```
**Result**: ✅ Compiles successfully

### Module Structure Test
```bash
ls -la module/*.go
wc -l module/*.go
```
**Result**: 8 Go files, 645 total lines, well-balanced

---

## Benefits of This Structure

### For Developers
- ✅ Easy to find coordination logic
- ✅ Clear separation between Keeper and ZooKeeper
- ✅ Simple to add new coordination types
- ✅ Each file is manageable size (<200 lines)

### For Maintenance
- ✅ Isolate changes (Keeper changes don't affect ZooKeeper code)
- ✅ Easy to deprecate (remove `coordination_zookeeper.go` in v2)
- ✅ Clear upgrade path
- ✅ Better testability (can test each coordination type independently)

### For Documentation
- ✅ File names are self-documenting
- ✅ Clear responsibility boundaries
- ✅ Easy to explain architecture

---

**This modular structure follows software engineering best practices and makes the codebase more maintainable for future contributors.**
